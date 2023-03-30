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
    </nocloud-table>

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

export default {
  name: "plan-prices",
  components: { nocloudTable },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    prices: [],
    isPricesLoading: false,
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
      { text: "_PACKAGE_EXTENSIONS", value: "_PACKAGE_EXTENSIONS" },
    ],
  }),
  methods: {},
  async mounted() {
    this.isPricesLoading = true;
    await this.$store.dispatch("servicesProviders/fetch");
    const sp = this.sps.find((sp) => sp.type === "cpanel");

    const res = await api.servicesProviders.action({
      action: "plans",
      uuid: sp.uuid,
    });
    this.prices = res.meta.pkg;
    this.isPricesLoading = false;
  },
  computed: {
    sps() {
      return this.$store.getters["servicesProviders/all"];
    },
  },
};
</script>
