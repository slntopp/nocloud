<template>
  <v-form v-model="formValid" ref="form">
    <v-container>
      <v-row justify="start">
        <v-col cols="4" md="3" lg="2">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="account"
            item-text="title"
            item-value="uuid"
            :items="accounts"
            v-model="accountId"
            :rules="rules.req"
          />
        </v-col>
        <v-col v-if="servicesByAccount.length > 2" cols="4" md="3" lg="2">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="service"
            item-text="title"
            :items="servicesByAccount"
            return-object
            v-model="service"
            :rules="rules.req"
          />
        </v-col>
        <v-col cols="4" md="3" lg="2">
          <v-autocomplete
            label="type"
            :items="typeItems"
            v-model="type"
            :rules="rules.req"
          />
        </v-col>
        <v-col cols="4" md="3" lg="2">
          <v-autocomplete
            :filter="defaultFilterObject"
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
        <v-col
          v-if="serviceInstanceGroupTitles.length > 2"
          cols="4"
          md="3"
          lg="2"
        >
          <v-autocomplete
            :items="serviceInstanceGroupTitles"
            v-model="instanceGroupTitle"
            label="instanceGroup"
            :rules="rules.req"
          />
        </v-col>
      </v-row>
      <component
        :is-edit="isEdit || this.$route.params.instanceId"
        v-if="isDataLoading"
        @set-value="setValue"
        :instance-group="instanceGroup"
        @set-instance-group="instanceGroup = $event"
        @set-instance="instance = $event"
        @set-meta="meta = $event"
        :instance="instance"
        :plans="plans"
        :account-id="accountId"
        :plan-rules="planRules"
        :sp-uuid="serviceProviderId"
        :is="templates[type] ?? templates.custom"
      />
      <v-row class="mx-5" justify="end">
        <v-btn :loading="isCreateLoading" @click="save">Create</v-btn>
      </v-row>
    </v-container>
  </v-form>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import api from "@/api";
import { defaultFilterObject } from "@/functions";

export default {
  name: "instance-create",
  data: () => ({
    isEdit: false,
    isCreateLoading: false,

    typeItems: [],
    templates: [],
    type: "ione",
    customTypeName: null,
    instance: {},
    service: {},
    accountId: null,
    instanceGroup: {},
    instanceGroupTitle: "",
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
    defaultFilterObject,
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
    async save() {
      if (!this.$refs.form.validate()) {
        return;
      }

      this.instance.config.auto_start =
        this.instance.billing_plan.meta.auto_start;
      if (typeof this.instance.billing_plan === "string") {
        this.instance.billing_plan = { uuid: this.instance.billing_plan };
      }

      this.isCreateLoading = true;
      try {
        const namespaceUuid = this.namespaces.find(
          (n) => n.access.namespace == this.accountId
        )?.uuid;

        if (!this.service) {
          this.service = await api.services.create({
            namespace: namespaceUuid,
            service: {
              version: "1",
              title: this.accounts.find((a) => a.uuid === this.accountId).title,
              context: {},
              instancesGroups: [],
            },
          });
        }

        let igIndex = this.service.instancesGroups.findIndex(
          (i) => i.title === this.instanceGroupTitle
        );

        if (igIndex === -1) {
          this.service.instancesGroups.push({
            title:
              this.instanceGroupTitle ||
              this.accounts.find((ac) => ac.uuid === this.accountId).title +
                new Date().getTime(),
            type: this.customTypeName || this.type,
            instances: [],
            sp: this.serviceProviderId,
          });
          igIndex = this.service.instancesGroups.length - 1;
        }

        this.service.instancesGroups[igIndex] = {
          ...this.service.instancesGroups[igIndex],
          ...this.instanceGroup,
        };

        if (this.isEdit) {
          const instanceIndex = this.service.instancesGroups[
            igIndex
          ].instances.findIndex((i) => i.uuid === this.instance.uuid);
          this.service.instancesGroups[igIndex].instances[instanceIndex] =
            this.instance;
        } else {
          this.service.instancesGroups[igIndex].instances.push(this.instance);
        }

        if (this.service.instancesGroups[igIndex].type === "ione") {
          this.service.instancesGroups[igIndex].resources = {
            ...(this.service.instancesGroups[igIndex].resources || {}),
            ips_public: this.service.instancesGroups[igIndex].instances
              .filter((i) => i.state?.state !== "DELETED")
              .reduce((acc, i) => acc + +i?.resources.ips_public, 0),
          };
        }

        const data = {
          namespace: namespaceUuid,
          service: this.service,
        };

        const res = await api.services.testConfig(data);

        if (!res.result) throw res;
        await api.services._update(data.service);
        this.showSnackbarSuccess({
          message: this.isEdit
            ? "instance updated successfully"
            : "instance created successfully",
        });
        api.services.up(data.service.uuid);
        this.$router.push({ name: "Instances" });
      } catch (err) {
        const opts = {
          message: err.errors.map((error) => error),
        };
        this.showSnackbarError(opts);
      } finally {
        this.isCreateLoading = false;
      }
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
    servicesByAccount() {
      return this.services.filter(
        (s) =>
          s.access.namespace ===
          this.namespaces.find((n) => n.access.namespace === this.accountId)
            ?.uuid
      );
    },
    accounts() {
      return this.$store.getters["accounts/all"];
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    serviceInstanceGroups() {
      if (
        !this.service ||
        !this.service.instancesGroups ||
        !this.type ||
        !this.serviceProviderId
      ) {
        return [];
      }

      return this.service.instancesGroups.filter(
        (ig) => ig.type === this.type && ig.sp === this.serviceProviderId
      );
    },
    serviceInstanceGroupTitles() {
      const igs = this.serviceInstanceGroups.map((i) => i.title);

      return igs;
    },
    planRules() {
      return this.rules.req;
    },
    isDataLoading() {
      return (
        this.type &&
        this.serviceProviderId &&
        (!(this.isEdit || this.$route.params.instanceId) ||
          ((this.isEdit || this.$route.params.instanceId) &&
            this.plans.list.length))
      );
    },
  },
  async created() {
    this.$store.dispatch("servicesProviders/fetch", { anonymously: false });

    const types = require.context(
      "@/components/modules/",
      true,
      /instanceCreate\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(
        /\.\/([A-Za-z0-9-_,\s]*)\/instanceCreate\.vue/i
      );
      if (matched && matched.length > 1) {
        const type = matched[1];
        this.typeItems.push(type);
        this.templates[type] = () =>
          import(`@/components/modules/${type}/instanceCreate.vue`);
      }
    });

    await Promise.all([
      this.$store.dispatch("accounts/fetch"),
      this.$store.dispatch("namespaces/fetch"),
      this.$store.dispatch("services/fetch"),
    ]);
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
            this.instanceGroupTitle = ig.title;
            this.instance = instance;
          }
        });
      });
      return;
    }

    let { type, serviceId, accountId, serviceProviderId } = this.$route.params;

    if (type) {
      this.service = this.services.find((s) => s.uuid === serviceId);
      if (!this.typeItems.includes(type)) {
        this.customTypeName = type;
        type = "custom";
      }
      this.type = type;
      this.serviceProviderId = serviceProviderId;
    }
    this.accountId = accountId;
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
      this.instanceGroupTitle = this.service.instancesGroups.find(
        (ig) => ig.sp === sp_uuid
      )?.title;
    },
    instanceGroupTitle(newVal) {
      this.instanceGroup = this.serviceInstanceGroups.find(
        (ig) => ig.title === newVal
      );
    },
    accountId() {
      this.service = this.servicesByAccount?.[0];
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
