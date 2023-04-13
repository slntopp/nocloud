<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <nocloud-table
      table-text="cpanel-prices"
      class="pa-4"
      item-key="text"
      :show-select="false"
      :items="prices"
      :headers="headers"
      :loading="isPricesLoading"
    >
      <template v-slot:[`item.isSell`]="{ item }">
        <v-switch @change="changeSell(item, $event)" :input-value="item.isSell" />
      </template>
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field type="number" v-model.number="item.price" />
      </template>
      <template v-slot:[`item.period`]="{ item }">
        <date-field :period="item.period" @changeDate="item.period = $event" />
      </template>
    </nocloud-table>
    <v-card-actions class="d-flex justify-end">
      <v-btn :loading="isSaveLoading || isPricesLoading" @click="savePrices">save</v-btn>
    </v-card-actions>
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
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import DateField from "@/components/date.vue";
import { getTimestamp } from "@/functions";

export default {
  name: "plan-prices",
  components: { DateField, nocloudTable },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    prices: [],
    products: [],
    isPricesLoading: false,
    isSaveLoading: false,
    headers: [
      { text: "name", value: "name" },
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
      { text: "Period", value: "period" },
      { text: "Price", value: "price" },
      { text: "Sell", value: "isSell" },
    ],
  }),
  methods: {
    async fetchPrices() {
      this.isPricesLoading = true;
      await this.$store.dispatch("servicesProviders/fetch");
      const sp = this.sps.find((sp) => sp.type === "cpanel");

      const res = await api.servicesProviders.action({
        action: "plans",
        uuid: sp.uuid,
      });
      this.prices = res.meta.pkg.map((el) => {
        const price = { ...el };
        const product = this.template.products[el.name];
        price.price = product?.price || 0;
        price.period = product?.period || 0;
        price.isSell = !!product;
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
    changeSell(item, val) {
      if (val) {
        if (!getTimestamp(item.period) || !item.price) {
          this.$set(item, "isSell", false);
          return this.showSnackbarError({
            message: "Price and period required",
          });
        }

        return (this.products[item.name] = {
          title: item.name,
          kind: "PREPAID",
          price: item.price,
          period: getTimestamp(item.period),
          resources: {
            model: item.name,
          },
        });
      }

      this.products[item.name] = undefined;
    },
    savePrices() {
      this.isSaveLoading = true;
      try {
        api.plans.update(this.template.uuid, {
          ...this.template,
          products: this.products,
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
