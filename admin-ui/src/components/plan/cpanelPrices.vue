<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-expansion-panels v-if="!isPricesLoading">
      <v-expansion-panel>
        <v-expansion-panel-header color="background">
          Margin rules:
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background">
          <plan-opensrs
            :fee="fee"
            :isEdit="true"
            @changeFee="changeFee"
            @onValid="(data) => (isValid = data)"
          />
          <confirm-dialog
            text="This will apply the rules markup parameters to all prices"
            @confirm="setFee"
          >
            <v-btn class="mt-4" color="secondary">Set rules</v-btn>
          </confirm-dialog>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <v-tabs
      class="rounded-t-lg"
      v-model="tabsIndex"
      background-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab">{{ tab }}</v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabsIndex"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item v-for="tab in tabs" :key="tab">
        <div v-if="tab === 'Prices'">
          <div class="mt-4" v-if="!isPricesLoading">
            <v-btn class="mx-1" @click="setSellToAllTariffs(true)"
              >Enable all</v-btn
            >
            <v-btn class="mx-1" @click="setSellToAllTariffs(false)"
              >Disable all</v-btn
            >
          </div>
          <nocloud-table
            :loading="isPricesLoading"
            table-name="cpanel-prices"
            class="pa-4"
            item-key="key"
            :show-select="false"
            :items="prices"
            :headers="headers"
          >
            <template v-slot:[`item.enabled`]="{ item }">
              <v-switch v-model="item.enabled" />
            </template>
            <template v-slot:[`item.name`]="{ item }">
              <v-text-field v-model="item.name" />
            </template>
            <template v-slot:[`item.price`]="{ item }">
              <v-text-field type="number" v-model.number="item.price" />
            </template>
            <template v-slot:[`item.emptyBind`]="{ item }">
              <v-autocomplete
                class="empty-bind-field"
                v-model="item.emptyBind"
                :items="emptyPackageItems"
                :loading="isEmptyPlansLoading"
                :menu-props="{ maxHeight: 360, offsetY: true }"
                item-text="text"
                item-value="value"
                clearable
                dense
                hide-details
                placeholder="Empty package"
              />
            </template>
            <template v-slot:[`item.addons`]="{ item }">
              <product-addons-dialog
                @change:addons="item.addons = $event"
                :addons="item.addons"
              />
            </template>
            <template v-slot:[`item.sorter`]="{ item }">
              <v-text-field type="number" v-model.number="item.sorter" />
            </template>
            <template v-slot:[`item.period`]="{ item }">
              <date-field
                :period="item.period"
                :periodKind="item.periodKind"
                @changeDate="item.period = $event"
                @changePeriodKind="item.periodKind = $event"
              />
            </template>
          </nocloud-table>
        </div>

        <div class="os-tab__card" v-else>
          <plan-addons-table
            @change:addons="planAddons = $event"
            :addons="template.addons"
          />
        </div>
      </v-tab-item>
    </v-tabs-items>
    <v-card-actions class="d-flex justify-end">
      <v-btn
        :loading="isSaveLoading"
        :disabled="isPricesLoading"
        @click="savePrices"
        >save</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import DateField from "@/components/date.vue";
import { getMarginedValue } from "@/functions";
import PlanOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import planAddonsTable from "@/components/planAddonsTable.vue";
import productAddonsDialog from "@/components/product_addons_dialog.vue";

import { ListRequest } from "nocloud-proto/proto/es/billing/billing_pb";

function unwrapMetaString(value) {
  if (typeof value === "string") {
    return value.trim();
  }
  if (value && typeof value === "object" && typeof value.stringValue === "string") {
    return value.stringValue.trim();
  }
  return "";
}

function emptyBindValue(product) {
  const plan = unwrapMetaString(product?.resources?.empty_plan) || unwrapMetaString(product?.meta?.empty_plan);
  const pkg = unwrapMetaString(product?.resources?.empty_product) || unwrapMetaString(product?.meta?.empty_product);
  if (!plan || !pkg) {
    return "";
  }
  return `${plan}/${pkg}`;
}

function parseEmptyBind(value) {
  const text = String(value || "").trim();
  const slash = text.indexOf("/");
  if (slash < 1 || slash === text.length - 1) {
    return { plan: "", product: "" };
  }
  return { plan: text.slice(0, slash), product: text.slice(slash + 1) };
}

export default {
  name: "plan-prices",
  components: {
    ConfirmDialog,
    PlanOpensrs,
    DateField,
    nocloudTable,
    planAddonsTable,
    productAddonsDialog,
  },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    tabs: ["Prices", "Addons"],
    tabsIndex: 0,
    planAddons: [],
    prices: [],
    fee: {},
    products: [],
    emptyPlans: [],
    isPricesLoading: false,
    isEmptyPlansLoading: false,
    isValid: false,
    isSaveLoading: false,
    headers: [
      { text: "Title", value: "name", width: "220px" },
      { text: "BWLIMIT", value: "BWLIMIT" },
      { text: "CGI", value: "CGI" },
      { text: "CPMOD", value: "CPMOD" },
      { text: "DIGESTAUTH", value: "DIGESTAUTH" },
      { text: "FEATURELIST", value: "FEATURELIST" },
      { text: "HASSHELL", value: "HASSHELL" },
      { text: "IP", value: "IP" },
      { text: "LANG", value: "LANG" },
      { text: "MAXADDON", value: "MAXADDON" },
      { text: "MAXFTP", value: "MAXFTP" },
      { text: "MAXLST", value: "MAXLST" },
      { text: "MAXPARK", value: "MAXPARK" },
      { text: "MAXPOP", value: "MAXPOP" },
      { text: "MAXSQL", value: "MAXSQL" },
      { text: "MAXSUB", value: "MAXSUB" },
      { text: "MAX_DEFER_FAIL_PERCENTAGE", value: "MAX_DEFER_FAIL_PERCENTAGE" },
      { text: "MAX_EMAIL_PER_HOUR", value: "MAX_EMAIL_PER_HOUR" },
      { text: "MAX_EMAILACCT_QUOTA", value: "MAX_EMAILACCT_QUOTA" },
      { text: "QUOTA", value: "QUOTA" },
      { text: "lve_cpu", value: "lve_cpu" },
      { text: "lve_ep", value: "lve_ep" },
      { text: "lve_io", value: "lve_io" },
      { text: "lve_iops", value: "lve_iops" },
      { text: "lve_mem", value: "lve_mem" },
      { text: "lve_ncpu", value: "lve_ncpu" },
      { text: "lve_nproc", value: "lve_nproc" },
      { text: "lve_cpu", value: "lve_cpu" },
      { text: "lve_pmem", value: "lve_pmem" },
      { text: "Sorter", value: "sorter" },
      { text: "Addons", value: "addons" },
      { text: "Period", value: "period", width: 220 },
      { text: "Price", value: "price", width: 150 },
      { text: "Package", value: "emptyBind", width: 280 },
      { text: "Enabled", value: "enabled" },
    ],
  }),
  methods: {
    async fetchPrices() {
      this.isPricesLoading = true;
      await this.$store.dispatch("servicesProviders/fetch", {
        anonymously: true,
      });
      const sp = this.sps.find(
        (sp) =>
          sp.type === "cpanel" && sp.meta.plans?.includes(this.template.uuid)
      );
      if (!sp) {
        this.isPricesLoading = false;
        return this.showSnackbarError({
          message: "Bind plan to cpanel service provider",
        });
      }
      const res = await api.servicesProviders.action({
        action: "plans",
        uuid: sp.uuid,
      });
      this.prices = res.meta.pkg.map((el) => {
        const price = { ...el };
        const product = this.template.products[el.name];
        price.key = el.name;
        price.price = product?.price || 0;
        price.period = product?.period || 3600 * 24 * 30;
        price.sorter = product?.sorter || 0;
        price.addons = product?.addons || [];
        price.enabled = !!product;
        price.periodKind = product?.periodKind || "CALENDAR_MONTH";
        price.emptyBind = emptyBindValue(product);
        return price;
      });
      this.isPricesLoading = false;
    },
    async fetchEmptyPlans() {
      this.isEmptyPlansLoading = true;
      try {
        const response = await this.$store.getters["plans/plansClient"].listPlans(
          ListRequest.fromJson({
            anonymously: false,
            showDeleted: false,
            limit: "500",
            filters: { type: ["empty"] },
          })
        );
        this.emptyPlans = response.toJson().pool || [];
      } catch (error) {
        this.emptyPlans = [];
        this.showSnackbarError({
          message: error.response?.data?.message || "Error during fetch empty plans",
        });
      } finally {
        this.isEmptyPlansLoading = false;
      }
    },
    changeFee(value) {
      this.fee = JSON.parse(JSON.stringify(value));
    },
    setFee() {
      this.prices.forEach((t) => {
        t.price = getMarginedValue(this.fee, t.price);
      });
    },
    setSellToAllTariffs(value) {
      this.prices.forEach((t) => {
        t.enabled = value;
      });
    },
    async savePrices() {
      const products = {};
      const resources = [];

      this.prices
        .filter((p) => p.enabled)
        .forEach((item) => {
          const previous = this.template.products?.[item.key] || {};
          const meta = { ...(previous.meta || {}) };
          delete meta.site_credit;
          delete meta.site_models;
          delete meta.empty_plan;
          delete meta.empty_product;
          const bind = parseEmptyBind(item.emptyBind);
          const resources = {
            model: item.key,
            bandwidth: item.BWLIMIT || undefined,
            ssd: item.QUOTA || undefined,
            email: item.MAXPOP || undefined,
            mysql: item.MAXSQL || undefined,
            websites: 1 + +item.MAXADDON || undefined,
          };
          if (bind.plan && bind.product) {
            resources.empty_plan = bind.plan;
            resources.empty_product = bind.product;
          }
          products[item.key] = {
            title: item.name,
            kind: "PREPAID",
            price: item.price,
            period: item.period,
            sorter: item.sorter,
            addons: item.addons,
            periodKind: item.periodKind,
            meta,
            resources,
          };
        });

      this.isSaveLoading = true;
      try {
        await api.plans.update(this.template.uuid, {
          ...this.template,
          products,
          resources,
          addons: this.planAddons,
        });
        this.showSnackbarSuccess({
          message: "Price model edited successfully",
        });
      } catch (e) {
        this.showSnackbarError({ message: "Error on save plan" });
      } finally {
        this.isSaveLoading = false;
      }
    },
  },
  mounted() {
    this.fetchPrices();
    this.fetchEmptyPlans();
    this.products = this.template.products;
    this.planAddons = this.template.addons;
  },
  computed: {
    sps() {
      return this.$store.getters["servicesProviders/all"];
    },
    emptyPackageItems() {
      const items = [];
      const plans = (this.emptyPlans || [])
        .slice()
        .sort((a, b) => String(a.title || a.uuid).localeCompare(String(b.title || b.uuid)));
      for (const plan of plans) {
        const products = plan.products || {};
        const keys = Object.keys(products).sort((a, b) => {
          const left = products[a] || {};
          const right = products[b] || {};
          const sorter = (left.sorter || 0) - (right.sorter || 0);
          if (sorter) {
            return sorter;
          }
          return String(left.title || a).localeCompare(String(right.title || b));
        });
        if (keys.length === 0) {
          continue;
        }
        items.push({ header: plan.title || plan.uuid });
        for (const key of keys) {
          const product = products[key] || {};
          items.push({
            text: product.title || key,
            value: `${plan.uuid}/${key}`,
          });
        }
      }
      const known = new Set(items.map((item) => item.value).filter(Boolean));
      for (const price of this.prices) {
        if (price.emptyBind && !known.has(price.emptyBind)) {
          items.push({ header: "Saved" });
          items.push({ text: price.emptyBind, value: price.emptyBind });
          known.add(price.emptyBind);
        }
      }
      return items;
    },
  },
};
</script>

<style scoped>
.empty-bind-field {
  min-width: 220px;
}
</style>
