<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <instance-actions
          @refresh="refreshInstance"
          :sp="sp"
          :account="account"
          :copy-template="copyInstance"
          :template="template"
          :addons="addons"
          :copy-addons="copyAddons"
        />
      </v-col>
    </v-row>
    <v-card-title class="primary--text">Client info</v-card-title>
    <v-row>
      <v-col>
        <route-text-field
          :to="{
            name: 'NamespacePage',
            params: { namespaceId: namespace?.uuid },
          }"
          label="Group(Namespace)"
          :value="!namespace ? '' : namespace.title"
        />
      </v-col>
      <v-col>
        <div class="d-flex justify-center align-center">
          <route-text-field
            :to="{ name: 'Account', params: { accountId: account?.uuid } }"
            label="Account"
            :value="account?.title"
          />
          <v-btn icon @click="moveDialog = true">
            <v-icon size="30">mdi-arrow-up-bold</v-icon>
          </v-btn>
        </div>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Email"
          :value="account?.data?.email"
          append-icon="mdi-content-copy"
          @click:append="addToClipboard(account?.data?.email)"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Balance"
          :value="account?.balance?.toFixed(2) || 0"
        />
      </v-col>
    </v-row>
    <!--    <component-->
    <!--      v-if="!type.includes('ovh') && !type.includes('ione')"-->
    <!--      :is="templates[type] ?? templates.custom"-->
    <!--      :template="template"-->
    <!--      @refresh="refreshInstance"-->
    <!--    />-->
    <template>
      <v-card-title class="primary--text">Instance info</v-card-title>
      <v-row>
        <v-col>
          <v-text-field
            v-model="copyInstance.title"
            ref="instance-title"
            label="Instance name"
          >
            <template v-slot:append>
              <v-icon class="mr-2" @click="$refs['instance-title'].focus()"
                >mdi-pencil</v-icon
              >
              <login-in-account-icon
                :uuid="account?.uuid"
                :instanceId="template.uuid"
                :type="template.type"
              />
            </template>
          </v-text-field>
        </v-col>
        <v-col>
          <v-text-field
            :value="template.uuid"
            readonly
            label="UUID"
            append-icon="mdi-content-copy"
            @click:append="addToClipboard(template.uuid)"
          />
        </v-col>
        <v-col>
          <route-text-field
            :to="{ name: 'Service', params: { serviceId: service?.uuid } }"
            :value="!service ? '' : service.title"
            label="Service"
          />
        </v-col>
        <v-col>
          <route-text-field
            :to="{ name: 'ServicesProvider', params: { uuid: sp?.uuid } }"
            :value="sp?.title"
            label="Service provider"
          />
        </v-col>
        <v-col>
          <v-text-field
            label="Location"
            :value="template.config?.location || 'Unknown'"
            @input="update({ key: 'config.location', value: $event })"
          />
        </v-col>
        <v-col>
          <v-text-field readonly :value="type" label="Type" />
        </v-col>
      </v-row>

      <nocloud-expansion-panels
        title="Description"
        class="mb-5"
        v-if="template.billingPlan.products[template.product]"
      >
        <rich-editor
          class="pa-5"
          disabled
          :value="
            template.billingPlan.products[template.product].meta?.description
          "
        />
        <div class="d-flex justify-end align-center">
          <v-btn
            class="mx-2"
            @click="
              addToClipboard(
                template.billingPlan.products[template.product].meta
                  ?.description || ''
              )
            "
            >copy</v-btn
          >
          <v-btn class="mx-2" @click="goToPlan">edit</v-btn>
        </div>
      </nocloud-expansion-panels>

      <component
        @update="update"
        :is="additionalInstanceInfoComponent"
        :sp="sp"
        :account="account"
        :template="template"
      />
      <v-card-title class="primary--text">Billing info</v-card-title>
      <component
        @update="update"
        :is="billingInfoComponent"
        :template="copyInstance"
        :service="service"
        :plans="plans"
        :sp="sp"
        :addons="copyAddons"
        :account="account"
        @refresh="refreshInstance"
      />
    </template>
    <move-instance
      @refresh="refreshInstance"
      :account="account"
      :template="template"
      v-model="moveDialog"
    />

    <div v-if="billingLabelComponent" class="billing-label">
      <component
        @update="update"
        :is="billingLabelComponent"
        v-if="Object.keys(copyInstance).length"
        :account="account"
        :addons="copyAddons"
        :template="copyInstance"
      />
    </div>
  </v-card>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import instanceActions from "@/components/instance/controls.vue";
import JsonTextarea from "@/components/JsonTextarea.vue";
import { mapGetters } from "vuex";
import RouteTextField from "@/components/ui/routeTextField.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import MoveInstance from "@/components/dialogs/moveInstance.vue";
import { addToClipboard } from "@/functions";
import RichEditor from "@/components/ui/richEditor.vue";
import NocloudExpansionPanels from "@/components/ui/nocloudExpansionPanels.vue";

export default {
  name: "instance-info",
  components: {
    NocloudExpansionPanels,
    RichEditor,
    MoveInstance,
    LoginInAccountIcon,
    RouteTextField,
    nocloudTable,
    instanceActions,
    JsonTextarea,
  },
  mixins: [snackbar],
  props: {
    template: { type: Object, required: true },
    account: { type: Object, required: true },
    addons: { type: Array, required: true },
  },
  data: () => ({
    templates: {},
    moveDialog: false,
    copyInstance: {},
    copyAddons: [],
  }),
  methods: {
    addToClipboard,
    refreshInstance() {
      this.$store.dispatch("reloadBtn/onclick");
    },
    update({ type = "template", ...params }) {
      if (type === "addons") {
        return this.updateCopyAddons(params);
      }

      return this.updateCopyInstance(params);
    },
    updateCopyAddons({ key, value }) {
      const keys = key.split(".");
      if (keys.length) {
        const lastKey = keys.pop();
        let temp = this.copyAddons;
        keys.forEach((key, index, array) => {
          if (["com", "net", "org"].includes(key)) {
            temp = temp[array[index - 1] + "." + key];
          } else if (temp[key]) {
            temp = temp[key];
          }
        });
        temp[lastKey] = value;
      } else {
        this.copyAddons[key] = value;
      }
      this.copyAddons = [...this.copyAddons];
    },
    updateCopyInstance({ key, value }) {
      const keys = key.split(".");
      if (keys.length) {
        const lastKey = keys.pop();
        let temp = this.copyInstance;
        keys.forEach((key, index, array) => {
          if (["com", "net", "org"].includes(key)) {
            temp = temp[array[index - 1] + "." + key];
          } else if (temp[key]) {
            temp = temp[key];
          }
        });
        temp[lastKey] = value;
      } else {
        this.copyInstance[key] = value;
      }
      this.copyInstance = { ...this.copyInstance };
      console.log(this.copyInstance);
      
    },
    goToPlan() {
      this.$router.push({
        name: "Plan",
        params: { planId: this.template.billingPlan.uuid },
      });
    },
  },
  computed: {
    ...mapGetters("namespaces", { namespaces: "all" }),
    ...mapGetters("services", { services: "all" }),
    ...mapGetters("plans", { plans: "all" }),
    ...mapGetters("servicesProviders", { servicesProviders: "all" }),
    namespace() {
      return this.namespaces?.find((n) => n.uuid == this.template.namespace);
    },
    service() {
      return this.services?.find((s) => s?.uuid == this.template.service);
    },
    sp() {
      return this.servicesProviders?.find((sp) => sp?.uuid == this.template.sp);
    },
    plan() {
      return this.template.billingPlan;
    },
    additionalInstanceInfoComponent() {
      return () =>
        import(`@/components/modules/${this.type}/additionalInstanceInfo.vue`);
    },
    billingInfoComponent() {
      return () => import(`@/components/modules/${this.type}/billingInfo.vue`);
    },
    billingLabelComponent() {
      return () => import(`@/components/modules/${this.type}/billingLabel.vue`);
    },
    type() {
      return this.template.type;
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
  mounted() {
    //break reactivity
    if (this.template.config?.regular_payment === undefined) {
      this.$set(this.template.config, "regular_payment", true);
    }

    this.copyInstance = JSON.parse(JSON.stringify(this.template));
    this.copyAddons = JSON.parse(JSON.stringify(this.addons));
  },
  watch: {
    template: {
      handler(newVal) {
        this.copyInstance = JSON.parse(JSON.stringify(newVal));
      },
      deep: true,
    },
  },
};
</script>

<style scoped lang="scss">
.billing-label {
  position: absolute;
  top: 0;
  right: 5px;
  z-index: 2;
  max-width: 450px;
  width: 40%;
}
</style>
