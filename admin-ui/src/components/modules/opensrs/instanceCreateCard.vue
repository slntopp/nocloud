<template>
  <v-card
    :id="instance.uuid"
    class="mb-4 pa-2"
    elevation="0"
    color="background"
  >
    <v-row>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) => $emit('set-value', { key: 'title', value: newVal })
          "
          label="title"
          :value="instance.title"
        />
      </v-col>
      <v-col v-if="showRemove" class="d-flex justify-end">
        <v-btn @click="() => $emit('remove')"> remove </v-btn>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="6">
        <v-text-field
          label="first name"
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.first_name',
                value: newVal,
              })
          "
          :value="instance.resources.user.first_name"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.last_name',
                value: newVal,
              })
          "
          label="last name"
          :value="instance.resources.user.last_name"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.org_name',
                value: newVal,
              })
          "
          label="organization name"
          :value="instance.resources.user.org_name"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.address1',
                value: newVal,
              })
          "
          label="address1"
          :value="instance.resources.user.address1"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.address2',
                value: newVal,
              })
          "
          label="address2"
          :value="instance.resources.user.address2"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.city',
                value: newVal,
              })
          "
          label="city"
          :value="instance.resources.user.city"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.country',
                value: newVal,
              })
          "
          label="country"
          :value="instance.resources.user.country"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', { key: 'resources.user.state', value: newVal })
          "
          label="state"
          :value="instance.resources.user.state"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.user.postal_code',
                value: newVal,
              })
          "
          label="postal_code"
          :value="instance.resources.user.postal_code"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', { key: 'resources.user.phone', value: newVal })
          "
          label="phone"
          :value="instance.resources.user.phone"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', { key: 'resources.user.email', value: newVal })
          "
          label="email"
          :value="instance.resources.user.email"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.reg_username',
                value: newVal,
              })
          "
          label="reg_username"
          :value="instance.resources.reg_username"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.reg_password',
                value: newVal,
              })
          "
          label="reg_password"
          :value="instance.resources.reg_password"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="4">
        <v-switch
          @change="
            (newVal) =>
              $emit('set-value', { key: 'resources.auto_renew', value: newVal })
          "
          :value="instance.resources.auto_renew"
          label="auto_renew"
        />
      </v-col>
      <v-col cols="4">
        <v-switch
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.who_is_privacy',
                value: newVal,
              })
          "
          :value="instance.resources.who_is_privacy"
          label="who_is_privacy"
        />
      </v-col>
      <v-col cols="4">
        <v-switch
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.lock_domain',
                value: newVal,
              })
          "
          :value="instance.resources.lock_domain"
          label="lock_domain"
        />
      </v-col>
    </v-row>
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
          Price: {{ domainPrice || "select domain and period" }}
        </p>
        <v-progress-circular v-else indeterminate color="primary" />
      </v-col>
      <v-col cols="2">
        <v-select
          @change="
            (newVal) =>
              $emit('set-value', { key: 'resources.period', value: +newVal })
          "
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
        table-name="openSrsServiceCreate"
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
  </v-card>
</template>

<script>
import api from "@/api";
import { levenshtein } from "@/functions";
import NocloudTable from "@/components/table.vue";

export default {
  name: "instanceCreateCard",
  components: { NocloudTable },
  emits: ["set-value", "remove"],
  props: {
    instance: {},
    "show-remove": { type: Boolean, default: true },
    "sp-uuid":{}
  },
  data: () => ({
    headers: [
      { text: "domain", value: "domain" },
      { text: "status", value: "status" },
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
    changeDomain(newVal) {
      const currentDomain = newVal[0];

      if (!currentDomain) {
        this.tableError = "";
        return this.resetDomain();
      } else if (currentDomain.status === "taken") {
        this.tableError = "This domain already taken!";
        return this.resetDomain();
      }

      this.isPriceLoading = true;

      this.$emit("set-value", {
        key: "resources.domain",
        value: currentDomain.domain,
      });
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
      this.$emit("set-value", {
        key: "resources.domain",
        value: "",
      });
      this.prices = [];
    },
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
};
</script>

<style scoped></style>
