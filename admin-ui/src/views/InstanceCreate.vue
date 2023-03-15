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
          v-if="type"
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
      password: [
        (v) => !!v || "password required",
        (v) => v.length > 6 || "password must be at least 6 characters length",
      ],
      nubmer: [(v) => Number(v) == v || "must be a correct number"],
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

      this.service.instancesGroups[igIndex].instances.push(this.instance);

      const data = {
        namespace: this.service.access.namespace,
        service: this.service,
      };

      this.isLoading = true;
      api.services
        .testConfig(data)
        .then((res) => {
          console.log("good 12");
          if (res.result) api.services._update(data.service);
          else throw res;
        })
        .then(() => {
          console.log("good");
          this.showSnackbarSuccess({
            message: "instance created successfully",
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
  },
  created() {
    this.$store.dispatch("services/fetch").then(() => {
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
      }
    });
    this.$store.dispatch("servicesProviders/fetch", false);

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
      api.plans.list({ sp_uuid, anonymously: false }).then((res) => {
        res.pool.forEach((plan) => {
          const end = plan.uuid.length > 8 ? "..." : "";
          const title = `${plan.title} (${plan.uuid.slice(0, 8)}${end})`;

          this.plans.list = [];
          this.plans.list.push({ ...plan, title });
        });
      });
    },
    type() {
      this.instanceGroup = null;
      this.serviceProviderId = null;
      this.customInstanceGroup = null;
      if (this.type !== "custom") {
        this.customTypeName = null;
      }
    },
  },
};
</script>
