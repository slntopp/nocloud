<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <instance-actions :template="template" />
      </v-col>
    </v-row>
    <v-row align="center">
      <v-col>
        <v-text-field
          readonly
          label="instance uuid"
          style="display: inline-block; width: 330px"
          :value="template.uuid"
          :append-icon="
            copyed === 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'
          "
          @click:append="addToClipboard(template.uuid, 'rootUUID')"
        />
      </v-col>
      <v-col v-if="template.state">
        <v-text-field
          readonly
          label="state"
          style="display: inline-block; width: 150px"
          :value="template.state.meta?.state_str || template.state.state"
        />
      </v-col>
      <v-col v-if="template.state?.meta.lcm_state_str">
        <v-text-field
          readonly
          label="lcm state"
          style="display: inline-block; width: 150px"
          :value="template.state?.meta.lcm_state_str"
        />
      </v-col>
      <v-col v-if="template.state?.meta.networking?.public">
        <div>
          <span class="mr-4">ips</span>
          <instance-ip-menu :item="template" />
        </div>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="price model"
          :append-icon="editPriceModelComponent ? 'mdi-pencil' : null"
          @click:append="priceModelDialog = true"
          style="display: inline-block; width: 150px"
          :value="template.billingPlan.title"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field
          @click:append="goTo('NamespacePage', { namespaceId: namespace.uuid })"
          readonly
          append-icon="mdi-login"
          label="namespace"
          :value="!namespace ? '' : 'NS_' + namespace.title"
        />
      </v-col>
      <v-col>
        <v-text-field
          @click:append="goTo('Account', { accountId: account.uuid })"
          readonly
          append-icon="mdi-login"
          label="account"
          :value="account?.title"
        />
      </v-col>
      <v-col>
        <v-text-field
          @click:append="goTo('Service', { serviceId: service.uuid })"
          readonly
          append-icon="mdi-login"
          label="service"
          :value="!service ? '' : 'SRV_' + service.title"
        />
      </v-col>
      <v-col>
        <v-text-field
          @click:append="goTo('ServicesProvider', { uuid: sp.uuid })"
          readonly
          append-icon="mdi-login"
          label="service provider"
          :value="sp?.title"
        />
      </v-col>
    </v-row>

    <component
      :is="templates[template.type] ?? templates.custom"
      :template="template"
      @refresh="refreshInstance"
    />

    <v-btn
      :to="{
        name: 'Instance edit',
        params: {
          instanceId: template.uuid,
        },
      }"
    >
      Edit
    </v-btn>

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
    <template v-if="editPriceModelComponent">
      <component
        :is="editPriceModelComponent"
        v-model="priceModelDialog"
        :template="template"
        :plans="filtredPlans"
        @refresh="refreshInstance"
        :service="service"
      />
    </template>
  </v-card>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import instanceActions from "@/components/instance/controls.vue";
import JsonTextarea from "@/components/JsonTextarea.vue";
import instanceIpMenu from "../ui/instanceIpMenu.vue";
import { mapGetters } from "vuex";
import EditPriceModel from "@/components/modules/ione/editPriceModel.vue";

export default {
  name: "instance-info",
  components: {
    EditPriceModel,
    nocloudTable,
    instanceActions,
    JsonTextarea,
    instanceIpMenu,
  },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    copyed: null,
    templates: {},
    priceModelDialog: false,
  }),
  methods: {
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((res) => {
            console.error(res);
          });
      } else {
        this.showSnackbarError({
          message: "Clipboard is not supported!",
        });
      }
    },
    goTo(name, params) {
      this.$router.push({ name, params });
    },
    refreshInstance() {
      this.$store.dispatch("services/fetch", this.template.uuid);
      this.$store.dispatch("servicesProviders/fetch");
    },
  },
  computed: {
    ...mapGetters("namespaces", { namespaces: "all" }),
    ...mapGetters("accounts", { accounts: "all" }),
    ...mapGetters("services", { services: "all" }),
    ...mapGetters("plans", { plans: "all" }),
    ...mapGetters("servicesProviders", { servicesProviders: "all" }),
    namespace() {
      return this.namespaces?.find(
        (n) => n.uuid == this.template.access.namespace
      );
    },
    account() {
      if (!this.namespace) {
        return;
      }
      return this.accounts?.find(
        (a) => a.uuid == this.namespace.access.namespace
      );
    },
    service() {
      return this.services?.find((s) => s.uuid == this.template.service);
    },
    sp() {
      return this.servicesProviders?.find((sp) => sp.uuid == this.template.sp);
    },
    filtredPlans() {
      return this.plans.filter(
        (p) =>
          p.type === this.template.type || p.type.includes(this.template.type)
      );
    },
    editPriceModelComponent() {
      const types = require.context(
        "@/components/modules/",
        true,
        /editPriceModel\.vue$/
      );

      if (types.keys().includes(`./${this.template.type}/editPriceModel.vue`)) {
        return () =>
          import(
            `@/components/modules/${this.template.type}/editPriceModel.vue`
          );
      }
      return null;
    },
  },
  created() {
    const types = require.context(
      "@/components/modules/",
      true,
      /instanceCard\.vue$/
    );

    types.keys().forEach((key) => {
      const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/instanceCard\.vue/i);

      if (matched && matched.length > 1) {
        this.templates[matched[1]] = () =>
          import(`@/components/modules/${matched[1]}/instanceCard.vue`);
      }
    });
  },
};
</script>
