<template>
  <div class="pa-4">
    <div class="d-flex" v-if="!item">
      <h1 class="page__title">Create price model</h1>
      <plan-wiki-icon />
    </div>

    <v-form v-model="isValid" ref="form">
      <v-row>
        <v-col cols="1" class="align-center d-flex">
          <v-subheader>Price model type</v-subheader>
        </v-col>
        <v-col cols="3">
          <v-autocomplete
            label="Type"
            v-model="plan.type"
            :items="types"
            :rules="[rules.required]"
          />
          <v-text-field
            label="Type name"
            v-if="plan.type === 'custom'"
            v-model="customTitle"
            :rules="[rules.required]"
          />
        </v-col>
        <v-col cols="1" class="align-center d-flex">
          <v-subheader>Price model title</v-subheader>
        </v-col>
        <v-col cols="3" class="align-center d-flex">
          <v-text-field
            label="Title"
            v-model="plan.title"
            :rules="[rules.required]"
          />
        </v-col>
        <v-col cols="1" class="align-center d-flex">
          <v-subheader>Price model kind</v-subheader>
        </v-col>
        <v-col cols="3" class="align-center d-flex">
          <v-radio-group
            :disabled="allowedKinds.length === 1"
            row
            mandatory
            v-model="selectedKind"
          >
            <confirm-dialog @cancel="changePlan(true)" @confirm="changePlan">
              <div class="d-flex">
                <v-radio
                  v-for="item in kinds"
                  :key="item"
                  :value="item"
                  :label="item.toLowerCase()"
                />
              </div>
            </confirm-dialog>
          </v-radio-group>
        </v-col>
        <template v-if="plan.kind === 'STATIC'">
          <v-col cols="1" class="align-center d-flex">
            <v-subheader>Default tariff</v-subheader>
          </v-col>
          <v-col cols="3" class="align-center d-flex">
            <v-autocomplete
              label="Tariff"
              v-model="plan.meta.product"
              :items="Object.keys(plan.products)"
            />
          </v-col>
        </template>

        <template v-if="plan.kind === 'DYNAMIC'">
          <v-col cols="1" class="align-center d-flex">
            <v-subheader>Linked price model</v-subheader>
          </v-col>
          <v-col cols="3" class="align-center d-flex">
            <v-autocomplete
              clearable
              @change="plan.meta.linkedPlan = $event ?? undefined"
              label="Price model"
              :value="plan.meta.linkedPlan"
              :items="filteredPlans"
            />
          </v-col>
        </template>

        <v-col cols="1" class="align-center d-flex">
          <v-subheader>Public</v-subheader>
        </v-col>
        <v-col cols="1" class="align-center d-flex">
          <v-switch style="width: fit-content" v-model="plan.public" />
        </v-col>
        <template>
          <v-col cols="1" class="align-center d-flex">
            <v-subheader>Auto start</v-subheader>
          </v-col>
          <v-col cols="1" class="align-center d-flex">
            <v-switch
              style="width: fit-content"
              v-model="plan.meta.auto_start"
            />
          </v-col>
        </template>
      </v-row>

      <v-col :cols="viewport > 2560 ? 6 : 12">
        <v-divider />
      </v-col>

      <v-col :cols="viewport > 2560 ? 6 : 12">
        <component
          v-if="!productsHide"
          :is="template"
          :rules="rules"
          :type="plan.type"
          :resources="plan.resources"
          :products="filteredProducts"
          @change:resource="(data) => changeConfig(data, 'resource')"
          @change:product="(data) => changeConfig(data, 'product')"
          @change:meta="(data) => changeMetaConfig(data, 'meta')"
        />
      </v-col>

      <v-row>
        <v-col>
          <v-btn class="mr-2" v-if="isEdit" @click="isDialogVisible = true">
            Save
          </v-btn>

          <v-dialog persistent v-else :max-width="600" v-model="isSetSpDialog">
            <template v-slot:activator="{ on, attrs }">
              <v-btn v-bind="attrs" v-on="on" class="mr-2" :loading="isLoading">
                Create
              </v-btn>
            </template>

            <v-card color="background-light">
              <v-card-title>Connect plan to sp</v-card-title>
              <v-card-subtitle
                >You can also connect plan later.</v-card-subtitle
              >

              <nocloud-table
                :items="typedSp"
                :headers="spHeaders"
                single-select
                v-model="selectedSp"
              />

              <v-card-actions class="d-flex justify-end">
                <v-btn @click="tryToSend('create')"> No </v-btn>
                <v-btn
                  :disabled="!selectedSp"
                  @click="tryToSend('create', true)"
                  class="mr-2"
                >
                  Connect
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>

          <download-template-button
            :template="plan"
            :type="selectedFileType"
            @click:xlsx="downloadPlanXlsx([plan])"
            :name="downloadedFileName"
          />
          <v-select
            style="max-width: 80px"
            :items="fileTypes"
            label="File type"
            v-model="selectedFileType"
            class="d-inline-block mx-1"
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
          <v-btn
            class="mr-2"
            :loading="isLoading && savePlanAction === 'create'"
            :disabled="isLoading && savePlanAction !== 'create'"
            @click="tryToSend('create')"
          >
            Create
          </v-btn>
          <v-btn
            v-if="item"
            :loading="isLoading && savePlanAction === 'edit'"
            :disabled="isLoading && savePlanAction !== 'edit'"
            @click="tryToSend('edit')"
          >
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
import { downloadPlanXlsx, getTimestamp } from "@/functions.js";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";
import PlanWikiIcon from "@/components/ui/planWikiIcon.vue";
import NocloudTable from "@/components/table.vue";

export default {
  name: "plansCreate-view",
  mixins: [snackbar],
  components: {
    NocloudTable,
    PlanWikiIcon,
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
      meta: {
        auto_start: true,
      },
      fee: null,
    },
    rules: {
      required: (v) => !!v || "This field is required!",
      price: (v) => (v !== "" && +v >= 0) || "Wrong price",
    },

    isDialogVisible: false,
    isVisible: true,
    isValid: false,
    isFeeValid: true,
    isLoading: false,
    savePlanAction: "",
    selectedFileType: "JSON",
    fileTypes: ["JSON", "YAML", "XLSX"],

    isSetSpDialog: false,
    selectedSp: null,
    spHeaders: [{ text: "Title", value: "title" }],
  }),
  methods: {
    downloadPlanXlsx,
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
      else if (type === "resource") this.$set(id, key, value);
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
    checkName({ title, uuid }, obj, num = 2) {
      const value = obj.find((el) => el.title === title && el.uuid !== uuid);
      const oldTitle = title.split(" ");

      if (oldTitle.length > 1 && num !== 2) {
        oldTitle[oldTitle.length - 1] = num;
      } else oldTitle.push(num);

      const plan = { title: oldTitle.join(" "), uuid };

      if (value) return this.checkName(plan, obj, num + 1);
      else return title;
    },
    async tryToSend(action, bindPlan = false) {
      if (!this.isValid || !this.isFeeValid) {
        this.$refs.form.validate();
        return this.showSnackbarError({ message: "Validation failed!" });
      }

      if (action === "create") delete this.plan.uuid;
      if (this.plan.type === "custom") {
        this.plan.type = this.customTitle;
      }

      this.isLoading = true;
      this.isSetSpDialog = false;
      this.savePlanAction = action;
      this.plan.title = this.checkName(this.plan, this.plans);

      const id = this.$route.params?.planId;

      //quick periodKind fix
      const periodMap = {
        2592000: "CALENDAR_MONTH",
        31536000: "CALENDAR_YEAR",
      };

      Object.keys(this.plan.products || {}).forEach((key) => {
        this.plan.products[key].periodKind = periodMap[
          this.plan.products[key].period
        ]
          ? periodMap[this.plan.products[key].period]
          : "DEFAULT";
      });

      this.plan.resources = this.plan.resources.map((r) => {
        r.periodKind = periodMap[r.period] ? periodMap[r.period] : "DEFAULT";
        return r;
      });

      try {
        //update or create descriptions
        const descriptionPromises = [
          ...this.plan.resources.map((resource, index) =>
            this.updateOrCreateDescription(resource, "resources", index)
          ),
          ...Object.keys(this.plan.products).map((key) =>
            this.updateOrCreateDescription(
              this.plan.products[key],
              "products",
              key
            )
          ),
        ];

        const descriptions = await Promise.all(descriptionPromises);
        descriptions.forEach(({ descriptionId, type, id }) => {
          this.plan[type][id].descriptionId = descriptionId;
        });

        const request =
          action === "edit"
            ? api.plans.update(id, this.plan)
            : api.plans.create(this.plan);

        const data = await request;
        if (bindPlan) {
          await api.servicesProviders.bindPlan(this.selectedSp[0].uuid, [
            data.uuid,
          ]);
        }

        const message =
          action === "edit"
            ? "Price model edited successfully"
            : "Price model created successfully";
        this.showSnackbarSuccess({ message });

        if (action !== "edit") {
          this.$router.push({ name: "Plans" });
        }
        this.isDialogVisible = false;
      } catch (err) {
        this.showSnackbarError({ message: err });
      } finally {
        this.isLoading = false;
        this.savePlanAction = "";
      }
    },
    async updateOrCreateDescription(item, type, id) {
      const { descriptionId, description } = item;
      if (descriptionId) {
        await api.patch("/descs/" + descriptionId, {
          text: description,
        });
        return {
          descriptionId,
          type,
          id,
        };
      }

      const data = await api.put("/descs", { text: description });
      return {
        descriptionId: data.uuid,
        type,
        id,
      };
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
  },
  created() {
    this.$store.dispatch("plans/fetch", { silent: true }).catch((err) => {
      const message = err.response?.data?.message ?? err.message ?? err;

      this.showSnackbarError({ message });
      console.error(err);
    });

    this.$store.dispatch("servicesProviders/fetch");

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
      const hidden = ["ovh", "goget", "acronis", "cpanel", "keyweb", "openai"];
      return hidden.some((h) => this.plan.type?.includes(h));
    },
    filteredProducts() {
      return this.plan.products;
    },
    downloadedFileName() {
      return this.plan.title
        ? this.plan.title.replaceAll(" ", "_")
        : "unknown_price_model";
    },
    allowedKinds() {
      const allowed = [];

      switch (this.plan.type) {
        case "openai": {
          allowed.push("DYNAMIC");
          break;
        }
        case "keyweb":
        case "ovh vps":
        case "ovh dedicated":
        case "ovh cloud":
        case "opensrs":
        case "empty":
        case "cpanel": {
          allowed.push("STATIC");
          break;
        }
        default: {
          allowed.push("DYNAMIC", "STATIC");
        }
      }

      return allowed;
    },
    typedSp() {
      return this.$store.getters["servicesProviders/all"].filter(
        (sp) => sp.type == this.plan.type.split(" ")[0]
      );
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
    allowedKinds(newVal) {
      if (!newVal.includes(this.plan.type)) {
        if (!this.isEdit) {
          this.plan.kind = newVal[0];
        }
        this.selectedKind = this.plan.kind;
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
