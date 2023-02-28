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
        <v-col :cols="(viewport > 1600) ? 6 : 12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Price model type</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-select
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

          <v-row align="center" v-if="selectedKind === 'STATIC'">
            <v-col cols="3">
              <v-subheader>Default tariff</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-select
                label="Tariff"
                v-model="plan.meta.product"
                :items="Object.keys(plan.products)"
              />
            </v-col>
          </v-row>

          <v-row>
            <v-col cols="3">
              <v-subheader>Meta</v-subheader>
            </v-col>
            <v-col cols="9">
              <json-editor
                :json="plan.meta"
                @changeValue="(data) => plan.meta = data"
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
        </v-col>

        <v-col />
        <v-divider />

        <v-col :cols="(viewport > 2200) ? 6 : 12">
          <component
            v-if="!['ovh vps', 'ovh dedicated', 'goget'].includes(plan.type)"
            :is="template"
            :resources="plan.resources"
            :products="plan.products"
            @change:resource="(data) => changeConfig(data, 'resource')"
            @change:product="(data) => changeConfig(data, 'product')"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-btn
            class="mr-2"
            v-if="isEdit"
            @click="isDialogVisible = true"
          >
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
            <v-btn class="mr-2" @click="downloadFile">
              Download {{ isJson ? "JSON" : "YAML" }}
            </v-btn>
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
        <v-card-title>Do you really want to change your current price model?</v-card-title>
        <v-card-subtitle>You can also create a new price model based on the current one.</v-card-subtitle>
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
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import confirmDialog from "@/components/confirmDialog.vue";
import planOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import JsonEditor from "@/components/JsonEditor.vue";

import {
  downloadJSONFile,
  readJSONFile,
  readYAMLFile,
  downloadYAMLFile,
} from "@/functions.js";

export default {
  name: "plansCreate-view",
  mixins: [snackbar],
  components: { confirmDialog, planOpensrs, JsonEditor },
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
      try { value = JSON.parse(value) }
      catch { value }

      const configs = (type === "resource")
        ? this.plan.resources
        : Object.values(this.plan.products);
      const product = configs.find((el) => el.id === id);

      switch (key) {
        case "key":
          if (type === "product") {
            const [oldKey = ''] = Object.entries(this.plan.products).find((el) => el.id === id) ?? [];

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
      if (action === 'create') delete this.plan.uuid;
      if (this.plan.type === 'custom') {
        this.plan.type = this.customTitle;
      }

      function checkName({ title, uuid }, obj, num = 2) {
        const value = obj.find((el) => el.title === title && el.uuid !== uuid);
        const oldTitle = title.split(' ');

        if (oldTitle.length > 1 && num !== 2) {
          oldTitle[oldTitle.length - 1] = num;
        }
        else oldTitle.push(num);

        const plan = { title: oldTitle.join(' '), uuid }

        if (value) return checkName(plan, obj, num + 1);
        else return title;
      }

      this.isLoading = true;
      this.plan.title = checkName(this.plan, this.plans);

      const id = this.$route.params?.planId;
      const request = (action === 'edit')
        ? api.plans.update(id, this.plan)
        : api.plans.create(this.plan);

      request.then(() => {
        this.showSnackbarSuccess({
          message: (action === 'edit')
            ? "Price model edited successfully"
            : "Price model created successfully",
        });
        setTimeout(() => {
          this.$router.push({ name: "Plans" });
        }, 100);
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
      const period = this.getTimestamp(date);
      const resource = this.plan.resources.find((el) => el.id === id);
      const product = Object.values(this.plan.products).find((el) => el.id === id);

      if (this.plan.kind === "DYNAMIC") {
        resource.period = period;
        this.plan.products = {};
      } else if (product) {
        product.period = period;
      }
    },
    getTimestamp({ day, month, year, quarter, week, time }) {
      year = +year + 1970;
      month = +month + quarter * 3 + 1;
      day = +day + week * 7 + 1;

      if (`${day}`.length < 2) day = "0" + day;
      if (`${month}`.length < 2) month = "0" + month;
      let seconds = Date.parse(`${year}-${month}-${day}T${time}Z`) / 1000;

      if (month > 1) {
        seconds -= 60 * 60 * 24 * (month - 1);
      }

      return seconds;
    },
    getItem(item = this.item) {
      if (Object.keys(item).length > 0) {
        if (!this.types.includes(item.type)) {
          this.customTitle = item.type;
          item.type = 'custom';
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
        this.selectedKind = this.item.kind;
        return;
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
      const requiredKeys = ["resources", "products", "title", "public", "type", "kind"];

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
    downloadFile() {
      const name = this.plan.title
        ? this.plan.title.replaceAll(" ", "_")
        : "unknown_price_model";
      if (this.isJson) {
        downloadJSONFile(this.plan, name);
      } else {
        downloadYAMLFile(this.plan, name);
      }
    },
  },
  created() {
    this.$store.dispatch('plans/fetch', { silent: true })
      .catch((err) => {
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
        if (matched[1] === 'ovh') {
          this.types.push('ovh vps', 'ovh dedicated');
        } else {
          this.types.push(matched[1]);
        }
      }
    });

    if (this.item) this.getItem();
  },
  computed: {
    template() {
      const type = (this.plan.kind === "DYNAMIC") ? "resources" : "products";

      return () => import(`@/components/plans_${type}_table.vue`);
    },
    plans() {
      return this.$store.getters['plans/all'];
    },
    viewport() {
      return document.documentElement.clientWidth;
    }
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
    }
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
