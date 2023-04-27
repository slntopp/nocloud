<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <instance-actions :template="template" />
      </v-col>
    </v-row>
    <v-card-title class="primary--text">Client info</v-card-title>
    <v-row>
      <v-col>
        <route-text-field
          :to="{
            name: 'NamespacePage',
            params: { namespaceId: namespace.uuid },
          }"
          label="Group(Namespace)"
          :value="!namespace ? '' : 'NS_' + namespace.title"
        />
      </v-col>
      <v-col>
        <div class="d-flex justify-center align-center">
          <route-text-field
            :to="{ name: 'Account', params: { accountId: account.uuid } }"
            label="Account"
            :value="account?.title"
          />
          <v-btn icon @click="moveDialog = true">
            <v-icon size="30">mdi-arrow-up-bold</v-icon>
          </v-btn>
        </div>
      </v-col>
      <v-col>
        <v-text-field readonly label="email" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="balance"
          :value="account?.balance.toFixed(2)"
        />
      </v-col>
    </v-row>
    <component
      v-if="!template.type.includes('ovh')"
      :is="templates[template.type] ?? templates.custom"
      :template="template"
      @refresh="refreshInstance"
    />
    <template v-else>
      <v-card-title class="primary--text">Instance info</v-card-title>
      <v-row>
        <v-col>
          <v-text-field v-model="instance.title" label="Instance title">
            <template v-slot:append>
              <v-icon class="mr-2">mdi-pencil</v-icon>
              <login-in-account-icon :uuid="account.uuid" />
            </template>
          </v-text-field>
        </v-col>
        <v-col>
          <v-text-field :value="template.uuid" readonly label="UUID" />
        </v-col>
        <v-col>
          <route-text-field
            :to="{ name: 'Service', params: { serviceId: service.uuid } }"
            :value="!service ? '' : 'SRV_' + service.title"
            label="Service"
          />
        </v-col>
        <v-col>
          <route-text-field
            :to="{ name: 'ServicesProvider', params: { uuid: sp.uuid } }"
            :value="sp?.title"
            label="Service provider"
          />
        </v-col>
        <v-col>
          <v-text-field readonly :value="template.type" label="Type" />
        </v-col>
      </v-row>
      <component :is="additionalInstanceInfoComponent" :template="template" />
      <v-card-title class="primary--text">Billing info</v-card-title>
      <component
        :is="billingInfoComponent"
        :template="template"
        :plans="plans"
        @refresh="refreshInstance"
      />
    </template>

    <v-btn @click="save" :loading="isSaveLoading"> Save </v-btn>

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
    <move-instance
      @refresh="refreshInstance"
      :account="account"
      :services="services"
      :namespaces="namespaces"
      :accounts="accounts"
      :template="template"
      v-model="moveDialog"
    />

    <div
      v-if="billingLabelComponent"
      style="position: absolute; top: 0; right: 75px"
    >
      <component :is="billingLabelComponent" :template="instance" />
    </div>
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
import RouteTextField from "@/components/ui/routeTextField.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import MoveInstance from "@/components/dialogs/moveInstance.vue";
import api from "@/api";

export default {
  name: "instance-info",
  components: {
    MoveInstance,
    LoginInAccountIcon,
    RouteTextField,
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
    moveDialog: false,
    instance: {},
    isSaveLoading: false,
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
    refreshInstance() {
      this.$store.dispatch("services/fetch", this.template.uuid);
      this.$store.dispatch("servicesProviders/fetch");
    },
    save() {
      const instance = this.instance;
      const service = JSON.parse(JSON.stringify(this.service));

      const igIndex = service.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === this.template.uuid)
      );
      const instanceIndex = service.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === this.template.uuid);

      service.instancesGroups[igIndex].instances[instanceIndex] = instance;

      this.isSaveLoading = true;
      api.services
        ._update(service)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Instance saved successfully",
          });

          this.refreshInstance();
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isSaveLoading = false;
        });
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
    additionalInstanceInfoComponent() {
      return () =>
        import(
          `@/components/modules/${this.template.type}/additionalInstanceInfo.vue`
        );
    },
    billingInfoComponent() {
      return () =>
        import(`@/components/modules/${this.template.type}/billingInfo.vue`);
    },
    billingLabelComponent() {
      return () =>
        import(`@/components/modules/${this.instance.type}/billingLabel.vue`);
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

    this.instance = JSON.parse(JSON.stringify(this.template));
  },
  watch: {
    template: {
      handler(newVal) {
        this.instance = JSON.parse(JSON.stringify(newVal));
      },
      deep: true,
    },
  },
};
</script>
