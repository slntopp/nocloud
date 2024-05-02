<template>
  <v-container>
    <v-row class="align-center">
      <v-col cols="4">
        <v-text-field label="domain" v-model="searchDomainString" />
      </v-col>
      <v-col cols="2">
        <v-btn
          @click="searchDomains"
          :disabled="!searchDomainString"
          small
          dark
          :loading="isDomainsLoading"
        >
          <v-icon dark> mdi-magnify </v-icon>
          Search
        </v-btn>
      </v-col>
      <v-col cols="2" class="d-flex justify-center align-center">
        <p v-if="!isPriceLoading">
          Price:
          {{ domainPrice ? `${domainPrice} USD` : "select domain and period" }}
        </p>
        <v-progress-circular v-else indeterminate color="primary" />
      </v-col>
      <v-col cols="2">
        <v-select
          @change="(newVal) => $emit('input:period', +newVal)"
          v-model="selectedPeriodIndex"
          :items="domainPeriods"
        >
          <template v-slot:selection="{ item }">
            {{ item + " years" }}
          </template>
          <template v-slot:item="{ item }">
            {{ item + " years" }}
          </template>
        </v-select>
      </v-col>
    </v-row>
    <v-row class="flex-column pa-md-5">
      <nocloud-table
        table-name="open-srs-domains"
        @input="(item) => changeDomain(item)"
        :footer-error="tableError"
        item-key="domain"
        single-select
        no-hide-uuid
        v-model="selectedDomain"
        :items="domains"
        :headers="headers"
        :loading="isDomainsLoading"
      />
    </v-row>
  </v-container>
</template>

<script>
import api from "@/api";
import { levenshtein } from "@/functions";
import NocloudTable from "@/components/table.vue";

export default {
  name: "domains-table",
  components: { NocloudTable },
  emits: ["input:domain", "input:period", "input:price"],
  props: ["sp-uuid"],
  data: () => ({
    headers: [
      { text: "Domain", value: "domain" },
      { text: "Status", value: "status" },
    ],
    tableError: "",
    prices: {},
    selectedPeriodIndex: 0,
    domains: [],
    searchDomainString: "",
    selectedDomain: [],
    isDomainsLoading: false,
    isPriceLoading: false,
  }),
  methods: {
    searchDomains() {
      this.isDomainsLoading = true;
      api.servicesProviders
        .action({
          uuid: this.spUuid,
          action: "get_domains",
          params: {
            searchString: this.searchDomainString,
            gTLD: true,
            p_ccTLD: false,
            m_ccTLD: false,
          },
        })
        .then((data) => {
          this.domains = this.sortDomainsLSM(
            data.meta.domains.filter((d) => d.status !== "undetermined"),
            this.searchDomainString.toLowerCase()
          );
        })
        .finally(() => {
          this.isDomainsLoading = false;
        });
    },
    changeDomain(newVal, index) {
      const currentDomain = newVal[0];

      if (!currentDomain) {
        this.tableError = "";
        return this.resetDomain(index);
      } else if (currentDomain.status === "taken") {
        this.tableError = "This domain already taken!";
        return this.resetDomain(index);
      }

      this.isPriceLoading = true;

      this.$emit("input:domain", currentDomain.domain);
      this.tableError = "";

      api.servicesProviders
        .action({
          uuid: this.spUuid,
          action: "get_domain_price",
          params: {
            domain: currentDomain.domain,
          },
        })
        .then((data) => {
          this.prices = data.meta.prices;
        })
        .finally(() => {
          this.isPriceLoading = false;
        });
    },
    resetDomain() {
      this.$emit("input:domain", "");
      this.prices = [];
    },
    sortDomainsLSM(domains, searchkey) {
      return domains.sort(function (a, b) {
        return (
          levenshtein(a.domain, searchkey) - levenshtein(b.domain, searchkey)
        );
      });
    },
  },
  computed: {
    domainPeriods() {
      if (Object.keys(this.prices).length === 0) {
        return null;
      }

      return Object.keys(this.prices);
    },
    domainPrice() {
      return this.prices[this.selectedPeriodIndex];
    },
  },
  watch: {
    domainPrice(value) {
      this.$emit("input:price", value);
    },
  },
};
</script>

<style scoped></style>
