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

    <div class="mt-4" v-if="!isPricesLoading">
      <v-btn class="mx-1" @click="setSellToAllTariffs(true)">Enable all</v-btn>
      <v-btn class="mx-1" @click="setSellToAllTariffs(false)"
        >Disable all</v-btn
      >
    </div>

    <div>
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
import { getMarginedValue, getTimestamp } from "@/functions";
import PlanOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";

export default {
  name: "plan-prices",
  components: {
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
      { text: "Period", value: "period", width: 220 },
      { text: "Price", value: "price", width: 150 },
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

        return price;
      });
      this.isPricesLoading = false;
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
          products[item.key] = {
            title: item.name,
            kind: "PREPAID",
            price: item.price,
            period: getTimestamp(item.period),
            sorter: item.sorter,
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
    this.products = this.template.products;
  },
  computed: {
    sps() {
      return this.$store.getters["servicesProviders/all"];
    },
  },
};
</script>
