<template>
  <div class="pa-4">
    <div class="d-flex">
      <h1 class="page__title" v-if="!item">Create price model</h1>
      <v-icon class="mx-3" large color="light" @click="openPlanWiki">
        mdi-information-outline
      </v-icon>
    </div>

    <v-form v-model="isValid" ref="form">
      <v-row>
        <v-col :cols="viewport > 1600 ? 6 : 12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Price model type</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-autocomplete
                label="Type"
                v-model="plan.type"
                :items="types"
                :rules="generalRule"
              />
              <v-text-field
                label="Type name"
                v-if="plan.type === 'custom'"
                v-model="customTitle"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Price model title</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                label="Title"
                v-model="plan.title"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Price model kind</v-subheader>
            </v-col>
            <v-col cols="9">
              <confirm-dialog @cancel="changePlan(true)" @confirm="changePlan">
                <v-radio-group row mandatory v-model="selectedKind">
                  <v-radio
                    v-for="item in kinds"
                    :key="item"
                    :value="item"
                    :label="item.toLowerCase()"
                  />
                </v-radio-group>
              </confirm-dialog>
            </v-col>
          </v-row>

          <v-row align="center" v-if="plan.kind === 'STATIC'">
            <v-col cols="3">
              <v-subheader>Default tariff</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-autocomplete
                label="Tariff"
                v-model="plan.meta.product"
                :items="Object.keys(plan.products)"
              />
            </v-col>
          </v-row>

          <v-row v-if="plan.kind === 'DYNAMIC'">
            <v-col cols="3">
              <v-subheader>Linked price model</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-autocomplete
                clearable
                @change="plan.meta.linkedPlan = $event ?? undefined"
                label="Price model"
                :value="plan.meta.linkedPlan"
                :items="filteredPlans"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Public</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-switch style="width: fit-content" v-model="plan.public" />
            </v-col>
          </v-row>
          <v-row align="center" v-if="plan.type === 'empty'">
            <v-col cols="3">
              <v-subheader>Auto start</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-switch
                style="width: fit-content"
                v-model="plan.meta.auto_start"
              />
            </v-col>
          </v-row>
        </v-col>

        <v-col :cols="viewport > 2560 ? 6 : 12">
          <v-divider />
        </v-col>

        <v-col :cols="viewport > 2560 ? 6 : 12">
          <component
            v-if="!productsHide"
            :is="template"
            :type="plan.type"
            :resources="plan.resources"
            :products="filteredProducts"
            @change:resource="(data) => changeConfig(data, 'resource')"
            @change:product="(data) => changeConfig(data, 'product')"
            @change:meta="(data) => changeMetaConfig(data, 'meta')"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-btn class="mr-2" v-if="isEdit" @click="isDialogVisible = true">
            Save
          </v-btn>
          <v-btn
            v-else
            class="mr-2"
            :loading="isLoading"
            @click="tryToSend('create')"
          >
            Create
          </v-btn>
          <download-template-button
            :template="plan"
            :type="isJson ? 'JSON' : 'YAML'"
            :name="downloadedFileName"
          />
          <v-switch
            class="d-inline-block mr-2"
            style="margin-top: 5px; padding-top: 0"
            v-model="isJson"
            :label="!isJson ? 'YAML' : 'JSON'"
          />
          <v-file-input
            class="file-input"
            v-if="!isEdit"
            :label="`upload ${isJson ? 'json' : 'yaml'} price model...`"
            :accept="isJson ? '.json' : '.yaml'"
            @change="onJsonInputChange"
          />
        </v-col>
      </v-row>
    </v-form>

    <v-dialog :max-width="600" v-model="isDialogVisible">
      <v-card color="background-light">
        <v-card-title
          >Do you really want to change your current price model?</v-card-title
        >
        <v-card-subtitle
          >You can also create a new price model based on the current
          one.</v-card-subtitle
        >
        <v-card-actions>
          <v-btn class="mr-2" :loading="isLoading" @click="tryToSend('create')">
            Create
          </v-btn>
          <v-btn v-if="item" :loading="isLoading" @click="tryToSend('edit')">
            Edit
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import confirmDialog from "@/components/confirmDialog.vue";
import planOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import JsonEditor from "@/components/JsonEditor.vue";
import { readJSONFile, readYAMLFile, getTimestamp } from "@/functions.js";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";

export default {
  name: "plansCreate-view",
  mixins: [snackbar],
  components: {
    DownloadTemplateButton,
    confirmDialog,
    planOpensrs,
    JsonEditor,
  },
  props: { item: { type: Object }, isEdit: { type: Boolean, default: false } },
  data: () => ({
    types: [],
    kinds: ["DYNAMIC", "STATIC"],
    selectedKind: "",
    customTitle: "",
    plan: {
      title: "",
      type: "custom",
      kind: "DYNAMIC",
      public: true,
      resources: [],
      products: {},
      meta: {},
      fee: null,
    },
    generalRule: [(v) => !!v || "This field is required!"],

    isDialogVisible: false,
    isVisible: true,
    isValid: false,
    isFeeValid: true,
    isLoading: false,
    isJson: true,
  }),
  methods: {
    changeConfig({ key, value, id }, type) {
      try {
        value = JSON.parse(value);
      } catch {
        value;
      }

      const configs =
        type === "resource"
          ? this.plan.resources
          : Object.values(this.plan.products);
      const product = configs.find((el) => el.id === id);

      switch (key) {
        case "key":
          if (type === "product") {
            const [oldKey = ""] =
              Object.entries(this.plan.products).find(
                ([, el]) => el.id === id
              ) ?? [];

            delete this.plan.products[oldKey];
            this.plan.products[value] = product;
            return;
          }
          break;
        case "date":
          this.setPeriod(value, id);
          return;
        case "resources":
          this.plan.resources = value;
          return;
        case "products":
          this.plan.products = value;
          return;
        case "amount":
          key = "resources";
      }

      if (product) product[key] = value;
    },
    changeMetaConfig({ key, value, id }) {
      try {
        value = JSON.parse(value);
      } catch {
        value;
      }

      const product = Object.values(this.plan.products).find(
        (product) => product.id === id
      );

      this.$set(product.meta, key, value);
      this.plan.meta = Object.assign({}, this.plan.meta);
    },
    tryToSend(action) {
      let message = "";

      if (!this.isValid || !this.isFeeValid) {
        this.$refs.form.validate();
        message = "Validation failed!";
      }

      if (!message) {
        message = this.checkPlanPeriods(this.plan);
      }

      if (message) {
        this.showSnackbarError({ message });
        return;
      }
      if (action === "create") delete this.plan.uuid;
      if (this.plan.type === "custom") {
        this.plan.type = this.customTitle;
      }

      function checkName({ title, uuid }, obj, num = 2) {
        const value = obj.find((el) => el.title === title && el.uuid !== uuid);
        const oldTitle = title.split(" ");

        if (oldTitle.length > 1 && num !== 2) {
          oldTitle[oldTitle.length - 1] = num;
        } else oldTitle.push(num);

        const plan = { title: oldTitle.join(" "), uuid };

        if (value) return checkName(plan, obj, num + 1);
        else return title;
      }

      this.isLoading = true;
      this.plan.title = checkName(this.plan, this.plans);

      const id = this.$route.params?.planId;
      const request =
        action === "edit"
          ? api.plans.update(id, this.plan)
          : api.plans.create(this.plan);

      request
        .then(() => {
          this.showSnackbarSuccess({
            message:
              action === "edit"
                ? "Price model edited successfully"
                : "Price model created successfully",
          });
          if (action !== "edit") {
            this.$router.push({ name: "Plans" });
          }
          this.isDialogVisible = false;
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    checkPeriods(periods) {
      const wrongPeriod = periods.find((p) => p.period === 0);

      return (
        wrongPeriod &&
        `Period cannot be zero in an ${
          wrongPeriod.key || wrongPeriod.title
        } config`
      );
    },
    checkPlanPeriods(plan) {
      if (
        !Object.keys(plan.products).length &&
        !Array.isArray(plan.resources)
      ) {
        return;
      } else if (plan.products) {
        return this.checkPeriods(Object.values(plan.products));
      } else {
        return this.checkPeriods(plan.resources);
      }
    },
    setPeriod(date, id) {
      const period = getTimestamp(date);
      const resource = this.plan.resources.find((el) => el.id === id);
      const product = Object.values(this.plan.products).find(
        (el) => el.id === id
      );

      if (this.plan.kind === "DYNAMIC") this.plan.products = {};
      if (resource) resource.period = period;
      else if (product) product.period = period;
    },
    getItem(item = this.item) {
      if (Object.keys(item).length > 0) {
        if (!this.types.includes(item.type)) {
          this.customTitle = item.type;
          item.type = "custom";
        }

        this.plan = item;
        this.isVisible = false;
        this.selectedKind = item.kind;

        item.resources.forEach((_, i) => {
          this.plan.resources[i].id = Math.random().toString(16).slice(2);
        });
        Object.entries(item.products).forEach(([key]) => {
          this.plan.products[key].id = Math.random().toString(16).slice(2);
        });
      }
    },
    changePlan(isReset) {
      if (isReset) {
        return (this.selectedKind =
          this.item?.kind || this.selectedKind === "STATIC"
            ? "DYNAMIC"
            : "STATIC");
      }
      this.plan.kind = this.selectedKind;
    },
    openPlanWiki() {
      window.open(
        "https://github.com/slntopp/nocloud/wiki/Billing-Plans",
        "_blank"
      );
    },
    setPlan(res) {
      const requiredKeys = [
        "resources",
        "products",
        "title",
        "public",
        "type",
        "kind",
      ];

      for (const key of requiredKeys) {
        if (res[key] === undefined) {
          throw new Error("JSON need keys:" + requiredKeys.join(", "));
        }
      }

      if (!this.types.includes(res.type)) {
        throw new Error(`Type ${res.type} not exists!`);
      }

      if (!this.kinds.includes(res.kind)) {
        throw new Error(`Kind ${res.kind} not exists!`);
      }

      this.getItem(res);
    },
    onJsonInputChange(file) {
      if (this.isJson) {
        readJSONFile(file)
          .then((res) => this.setPlan(res))
          .catch(({ message }) => {
            this.showSnackbarError({ message });
          });
      } else {
        readYAMLFile(file)
          .then((res) => this.setPlan(res))
          .catch(({ message }) => {
            this.showSnackbarError({ message });
          });
      }
    },
  },
  created() {
    this.$store.dispatch("plans/fetch", { silent: true }).catch((err) => {
      const message = err.response?.data?.message ?? err.message ?? err;

      this.showSnackbarError({ message });
      console.error(err);
    });

    if (this.isEdit) {
      this.plan.resources = this.item.resources;
    }
    const types = require.context(
      "@/components/modules/",
      true,
      /serviceProviders\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(
        /\.\/([A-Za-z0-9-_,\s]*)\/serviceProviders\.vue/i
      );
      if (matched && matched.length > 1) {
        if (matched[1] === "ovh") {
          this.types.push("ovh vps", "ovh dedicated", "ovh cloud");
        } else {
          this.types.push(matched[1]);
        }
      }
    });

    if (this.item) this.getItem();
  },
  computed: {
    template() {
      const type = this.plan.kind === "DYNAMIC" ? "resources" : "products";

      return () => import(`@/components/plans_${type}_table.vue`);
    },
    plans() {
      return this.$store.getters["plans/all"];
    },
    filteredPlans() {
      const items = this.plans.filter(
        (plan) => plan.type === this.plan.type && plan.uuid !== this.plan.uuid
      );

      return items.map((item) => ({ text: item.title, value: item.uuid }));
    },
    viewport() {
      return document.documentElement.clientWidth;
    },
    productsHide() {
      const hidden = ["ovh", "goget", "acronis", "cpanel"];
      return hidden.some((h) => this.plan.type.includes(h));
    },
    filteredProducts() {
      if (!this.searchParam) {
        return this.plan.products;
      }

      const filtered = {};
      Object.keys(this.plan.products).forEach((key) => {
        if (key === this.searchParam) {
          filtered[key] = this.plan.products[key];
          return;
        }

        if (
          this.plan.products[key]?.title
            .toLowerCase()
            .startsWith(this.searchParam)
        ) {
          filtered[key] = this.plan.products[key];
        }
      });

      return filtered;
    },
    downloadedFileName() {
      return this.plan.title
        ? this.plan.title.replaceAll(" ", "_")
        : "unknown_price_model";
    },
  },
  watch: {
    "plan.kind"() {
      if (!this.isEdit) {
        if (this.plan.kind === "STATIC") {
          this.plan.products = {};
        } else {
          this.plan.resources = [];
        }
      }
    },
    "plan.type"(newVal) {
      if (!this.isEdit) {
        this.plan.meta.auto_start = newVal === "empty" ? false : undefined;
      }
    },
  },
};
</script>

<style scoped>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}

.theme--dark.v-tabs-items {
  background: var(--v-background-base);
}

.mw-20 {
  max-width: 150px;
}

.file-input {
  max-width: 200px;
  min-width: 200px;
  margin-top: 0;
  padding-top: 0;
}
</style>
