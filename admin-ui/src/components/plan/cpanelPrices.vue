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
        <v-progress-linear v-if="isPricesLoading" indeterminate class="pt-1" />

        <div v-if="tab === 'Tariffs'">
          <div class="mt-4" v-if="!isPricesLoading">
            <v-btn class="mx-1" @click="setSellToAllTariffs(true)"
              >Enable all</v-btn
            >
            <v-btn class="mx-1" @click="setSellToAllTariffs(false)"
              >Disable all</v-btn
            >
          </div>

          <nocloud-table
            table-name="cpanel-prices"
            class="pa-4"
            item-key="key"
            :show-select="false"
            :items="prices"
            :headers="headers"
          >
            <template v-slot:[`item.addons`]="{ item }">
              <v-dialog width="90vw">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn icon v-bind="attrs" v-on="on">
                    <v-icon> mdi-menu-open </v-icon>
                  </v-btn>
                </template>

                <nocloud-table
                  table-name="cpanel-addons-prices"
                  class="pa-4"
                  item-key="id"
                  :show-select="false"
                  :items="item.addons.filter((a) => a.public)"
                  :headers="addonsHeaders"
                >
                  <template v-slot:[`item.sell`]="{ item: addon }">
                    <v-switch v-model="addon.sell" />
                  </template>
                </nocloud-table>
              </v-dialog>
            </template>
            <template v-slot:[`item.enabled`]="{ item }">
              <v-switch v-model="item.enabled" />
            </template>
            <template v-slot:[`item.name`]="{ item }">
              <v-text-field v-model="item.name" />
            </template>
            <template v-slot:[`item.price`]="{ item }">
              <v-text-field type="number" v-model.number="item.price" />
            </template>
            <template v-slot:[`item.sorter`]="{ item }">
              <v-text-field type="number" v-model.number="item.sorter" />
            </template>
            <template v-slot:[`item.period`]="{ item }">
              <date-field
                :period="item.period"
                @changeDate="item.period = $event"
              />
            </template>
          </nocloud-table>
        </div>
        <plans-resources-table
          v-show="tab === 'Addons'"
          :default-virtual="true"
          @change:resource="changeCustomAddons($event)"
          :resources="customAddonsResources"
          type="ovh dedicated"
        />
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
import { getBillingPeriod, getMarginedValue, getTimestamp } from "@/functions";
import PlanOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import plansResourcesTable from "@/components/plans_resources_table.vue";

export default {
  name: "plan-prices",
  components: {
    plansResourcesTable,
    ConfirmDialog,
    PlanOpensrs,
    DateField,
    nocloudTable,
  },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    prices: [],
    fee: {},
    products: [],
    isPricesLoading: false,
    isValid: false,
    isSaveLoading: false,
    headers: [
      { text: "name", value: "name", width: "220px" },
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
      { text: "Enabled", value: "enabled" },
    ],
    addonsHeaders: [
      { text: "Title", value: "title" },
      {
        text: "Period",
        value: "period",
      },
      { text: "Price", value: "price" },
      {
        text: "Sell",
        value: "sell",
        width: 100,
      },
    ],
    tabs: ["Tariffs", "Addons"],
    tabsIndex: 0,
    customAddonsResources: [],
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
        price.enabled = !!product;
        const date = new Date(price.period * 1000);
        const time = date.toUTCString().split(" ");

        price.period = {
          day: `${date.getUTCDate() - 1}`,
          month: `${date.getUTCMonth()}`,
          year: `${date.getUTCFullYear() - 1970}`,
          quarter: "0",
          week: "0",
          time: time.at(-2),
        };

        price.addons = JSON.parse(
          JSON.stringify(this.customAddonsResources)
        ).map((a) => ({ ...a, sell: product?.meta?.addons?.includes(a.key) }));

        return price;
      });
      this.isPricesLoading = false;
    },
    changeCustomAddons({ key, value, id }) {
      if (key === "resources") {
        this.customAddonsResources = value;
        return;
      }

      this.customAddonsResources = this.customAddonsResources.map((c) => {
        if (c.id === id) {
          if (key === "date") {
            c.period = getTimestamp(value);
          } else {
            c[key] = value;
          }
        }
        return c;
      });
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

      resources.push(...this.customAddonsResources);

      this.prices
        .filter((p) => p.enabled)
        .forEach((item) => {
          products[item.key] = {
            title: item.name,
            kind: "PREPAID",
            price: item.price,
            period: getTimestamp(item.period),
            sorter: item.sorter,
            meta: {
              addons: item.addons
                .filter((a) => a.sell && a.public)
                .map((a) => a.key),
            },
            resources: {
              model: item.key,
              bandwidth: item.BWLIMIT || undefined,
              ssd: item.QUOTA || undefined,
              email: item.MAXPOP || undefined,
              mysql: item.MAXSQL || undefined,
              websites: 1 + +item.MAXADDON || undefined,
            },
          };
        });

      this.isSaveLoading = true;
      try {
        await api.plans.update(this.template.uuid, {
          ...this.template,
          products,
          resources,
        });
        this.showSnackbarSuccess({ message: "Plan save successfully" });
      } catch (e) {
        this.showSnackbarError({ message: "Error on save plan" });
      } finally {
        this.isSaveLoading = false;
      }
    },
  },
  mounted() {
    this.customAddonsResources = this.template.resources
      .filter((r) => r.virtual)
      .map((a) => ({ ...a, period: getBillingPeriod(a.period) }));
    this.fetchPrices();
    this.products = this.template.products;
  },
  computed: {
    sps() {
      return this.$store.getters["servicesProviders/all"];
    },
  },
};
</script>
