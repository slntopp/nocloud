<template>
  <v-form v-model="formValid" ref="form">
    <v-container>
      <v-row justify="start">
        <v-col cols="4" md="3" lg="2">
          <v-select
            label="service"
            item-text="title"
            :items="services"
            return-object
            v-model="service"
            :rules="rules.req"
          />
        </v-col>
        <v-col cols="4" md="3" lg="2">
          <v-select
            label="type"
            :items="typeItems"
            v-model="type"
            :rules="rules.req"
          />
        </v-col>
        <v-col cols="4" md="3" lg="2">
          <v-select
            label="service provider"
            item-value="uuid"
            item-text="title"
            v-model="serviceProviderId"
            :items="servicesProviders"
            :rules="rules.req"
          />
        </v-col>
        <v-col v-if="type === 'custom'" cols="4" md="3" lg="2">
          <v-text-field
            v-model="customTypeName"
            label="Type name"
            :rules="rules.req"
          />
        </v-col>
        <v-col cols="4" md="3" lg="2">
          <v-autocomplete
            @input.native="customInstanceGroup = $event.target.value"
            :items="serviceInstanceGroups"
            v-model="instanceGroup"
            label="instanceGroup"
            :rules="rules.req"
          />
        </v-col>
      </v-row>
      <v-row>
        <component
          :is-edit="isEdit || this.$route.params.instanceId"
          v-if="isDataLoading"
          @set-value="setValue"
          @set-instance="instance = $event"
          @set-meta="meta = $event"
          :instance="instance"
          :plans="plans"
          :plan-rules="planRules"
          :sp-uuid="serviceProviderId"
          :is="templates[type] ?? templates.custom"
        />
      </v-row>
      <v-row class="mx-5" justify="end">
        <v-btn @click="save">Save</v-btn>
      </v-row>
    </v-container>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-form>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import api from "@/api";

export default {
  name: "service-create",
  data: () => ({
    isEdit: false,

    typeItems: [],
    templates: [],
    type: "ione",
    customTypeName: null,
    instance: {},
    service: {},
    instanceGroup: "",
    customInstanceGroup: "",
    serviceProviderId: null,

    meta: {},

    plans: {
      list: [],
      products: [],
    },

    formValid: false,
    rules: {
      req: [(v) => !!v || "required field"],
    },
  }),
  mixins: [snackbar],
  methods: {
    setValue({ key, value }) {
      const keys = key.split(".");
      switch (keys.length) {
        case 1: {
          return (this.instance[keys[0]] = value);
        }
        case 2: {
          return (this.instance[keys[0]][keys[1]] = value);
        }
        case 3: {
          return (this.instance[keys[0]][keys[1]][keys[2]] = value);
        }
      }
    },
    save() {
      if (!this.$refs.form.validate()) {
        return;
      }
      let igIndex = this.service.instancesGroups.findIndex(
        (i) => i.title === this.instanceGroup
      );

      if (igIndex === -1) {
        this.service.instancesGroups.push({
          title: this.instanceGroup,
          type: this.customTypeName || this.type,
          instances: [],
          sp: this.serviceProviderId,
        });
        igIndex = this.service.instancesGroups.length - 1;
      }

      if (this.type === "ione") {
        this.instance.billing_plan = this.plans.list.find(
          (p) => p.uuid === this.instance.billing_plan
        );
      }

      if (this.isEdit) {
        const instanceIndex = this.service.instancesGroups[
          igIndex
        ].instances.findIndex((i) => i.uuid === this.instance.uuid);
        this.service.instancesGroups[igIndex].instances[instanceIndex] =
          this.instance;
      } else {
        this.service.instancesGroups[igIndex].instances.push(this.instance);
      }

      const data = {
        namespace: this.service.access.namespace,
        service: this.service,
      };

      this.isLoading = true;
      api.services
        .testConfig(data)
        .then((res) => {
          if (res.result) api.services._update(data.service);
          else throw res;
        })
        .then(() => {
          this.showSnackbarSuccess({
            message: this.isEdit
              ? "instance updated successfully"
              : "instance created successfully",
          });
          this.$router.push({ name: "Instances" });
        })
        .catch((err) => {
          const opts = {
            message: err.errors.map((error) => error),
          };
          this.showSnackbarError(opts);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
  },
  computed: {
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"].filter(
        (el) => el.type === this.type
      );
    },
    services() {
      return this.$store.getters["services/all"];
    },
    serviceInstanceGroups() {
      if (
        !this.service.instancesGroups ||
        !this.type ||
        !this.serviceProviderId
      ) {
        return [];
      }

      const igs = this.service.instancesGroups
        .filter(
          (ig) => ig.type === this.type && ig.sp === this.serviceProviderId
        )
        .map((ig) => ig.title);

      if (this.customInstanceGroup) {
        igs.unshift(this.customInstanceGroup);
      }

      return igs;
    },
    planRules() {
      return this.plansVisible ? this.rules.req : [];
    },
    isDataLoading() {
      return (
        this.type &&
        this.serviceProviderId &&
        (!(this.isEdit ||  this.$route.params.instanceId) || ((this.isEdit || this.$route.params.instanceId) && this.plans.list.length))
      );
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetch");
    this.$store.dispatch("services/fetch").then(() => {
      const instanceId = this.$route.params.instanceId;
      if (instanceId) {
        this.services.forEach((s) => {
          s.instancesGroups.forEach((ig) => {
            this.isEdit = true;
            const instance = ig.instances.find((i) => i.uuid === instanceId);
            if (instance) {
              this.type = ig.type;
              this.service = s;
              this.serviceProviderId = ig.sp;
              this.instanceGroup = ig.title;
              this.instance = instance;
            }
          });
        });
        return;
      }

      let type = this.$route.params.type;

      if (type && this.$route.params.serviceId) {
        this.service = this.services.find(
          (s) => s.uuid === this.$route.params.serviceId
        );
        if (!this.typeItems.includes(type)) {
          this.customTypeName = type;
          type = "custom";
        }
        this.type = type;
        this.serviceProviderId = this.$route.params.serviceProviderId;
      }
    });

    const types = require.context(
      "@/components/modules/",
      true,
      /serviceCreate\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/serviceCreate\.vue/i);
      if (matched && matched.length > 1) {
        const type = matched[1];
        this.typeItems.push(type);
        this.templates[type] = () =>
          import(`@/components/modules/${type}/instanceCreate.vue`);
      }
    });
  },
  watch: {
    serviceProviderId(sp_uuid) {
      if (!sp_uuid) return;
      this.plans.list = [];
      api.plans.list({ sp_uuid, anonymously: false }).then((res) => {
        res.pool.forEach((plan) => {
          const end = plan.uuid.length > 8 ? "..." : "";
          const title = `${plan.title} (${plan.uuid.slice(0, 8)}${end})`;
          this.plans.list.push({ ...plan, title });
        });
      });
      this.instanceGroup = this.service.instancesGroups.find(
        (ig) => ig.sp === sp_uuid
      )?.title;
    },
    type() {
      if (this.isEdit) {
        this.isEdit = false;
        return;
      }
      if (this.type !== "custom") {
        this.customTypeName = null;
      }
      this.isEdit = false;
    },
    ["plans.list"](newVal) {
      if (newVal && this.instance?.billingPlan?.uuid) {
        this.instance.billing_plan = this.instance?.billingPlan?.uuid;
        delete this.instance.billingPlan;
      }
    },
  },
};
</script>