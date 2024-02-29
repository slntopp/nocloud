<template>
  <nocloud-table
    table-name="accounts"
    :headers="headers"
    :items="filteredAccounts"
    :value="selected"
    :loading="loading"
    :single-select="singleSelect"
    :footer-error="fetchError"
    @input="handleSelect"
  >
    <template v-slot:[`item.title`]="{ item }">
      <div class="d-flex justify-space-between">
        <router-link
          :to="{ name: 'Account', params: { accountId: item.uuid } }"
        >
          {{ item.title }}
        </router-link>
        <div>
          <v-icon
            @click="
              $router.push({
                name: 'Account',
                params: { accountId: item.uuid },
                query: { tab: 2 },
              })
            "
            class="ml-5"
            >mdi-calendar-multiple</v-icon
          >
          <login-in-account-icon
            class="ml-5"
            v-if="['ROOT', 'ADMIN'].includes(item.access.level)"
            :uuid="item.uuid"
          />
        </div>
      </div>
    </template>
    <template v-slot:[`item.balance`]="{ item }">
      <balance
        :hide-currency="true"
        :currency="item.currency"
        @click="goToBalance(item.uuid)"
        :value="item.balance"
      />
    </template>

    <template v-slot:[`item.data.date_create`]="{ value }">
      {{ formatSecondsToDate(value) }}
    </template>

    <template v-slot:[`item.address`]="{ item }">
      <v-tooltip bottom>
        <template v-slot:activator="{ on, attrs }">
          <span v-bind="attrs" v-on="on">
            {{ item.data?.city || item.data?.address }}
          </span>
        </template>
        <span>{{ item.data?.address }}</span>
      </v-tooltip>
    </template>
    <template v-slot:[`item.access.level`]="{ item }">
      <v-chip :color="colorChip(item.access.level)">
        {{ item.access.level }}
      </v-chip>
    </template>
    <template v-slot:[`item.data.regular_payment`]="{ item }">
      <v-switch
        :disabled="
          !!changeRegularPaymentUuid && changeRegularPaymentUuid !== item.uuid
        "
        :loading="
          !!changeRegularPaymentUuid && changeRegularPaymentUuid === item.uuid
        "
        @change="changeRegularPayment(item, $event)"
        :input-value="
          item.data?.regular_payment === undefined ||
          item.data?.regular_payment === true
        "
      >
      </v-switch>
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import Balance from "./balance.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import { mapGetters } from "vuex";
import {
  compareSearchValue,
  filterByKeysAndParam,
  formatSecondsToDate,
  getDeepObjectValue,
} from "@/functions";
import api from "@/api";
import search from "@/mixins/search";

export default {
  name: "accounts-table",
  mixins: [search({ name: "accounts-table" })],
  components: {
    LoginInAccountIcon,
    "nocloud-table": noCloudTable,
    Balance,
  },
  props: {
    value: {
      type: Array,
      default: () => [],
    },
    "single-select": {
      type: Boolean,
      default: false,
    },
    notFiltered: { type: Boolean, default: false },
    namespace: {
      type: String,
    },
  },
  data() {
    return {
      selected: [],
      loading: false,
      fetchError: "",
      changeRegularPaymentUuid: "",
      headers: [
        { text: "Title", value: "title" },
        { text: "UUID", value: "uuid" },
        { text: "Status", value: "status" },
        { text: "Balance", value: "balance" },
        { text: "Email", value: "data.email" },
        { text: "Created date", value: "data.date_create" },
        { text: "Country", value: "data.country" },
        { text: "Address", value: "address" },
        { text: "Client currency", value: "currency" },
        { text: "Access level", value: "access.level" },
        { text: "Invoice based", value: "data.regular_payment" },
        { text: "Group(NameSpace)", value: "namespace" },
      ],
      levelColorMap: {
        ROOT: "info",
        ADMIN: "success",
        MGMT: "warning",
        READ: "gray",
        NONE: "error",
      },
    };
  },
  methods: {
    formatSecondsToDate,
    handleSelect(item) {
      this.$emit("input", item);
    },
    getNamespaceName(uuid) {
      return (
        this.namespaces.find(({ access }) => access.namespace === uuid)
          ?.title ?? ""
      );
    },
    colorChip(level) {
      return this.levelColorMap[level];
    },
    goToBalance(uuid) {
      this.$router.push({ name: "Transactions", query: { account: uuid } });
    },
    async changeRegularPayment(item, value) {
      this.changeRegularPaymentUuid = item.uuid;
      try {
        if (!item.data) {
          item.data = {};
        }
        item.data.regular_payment = value;
        await api.accounts.update(item.uuid, item).catch((err) => {
          this.showSnackbarError({ message: err });
        });
        this.$store.commit("accounts/pushAccount", item);
      } catch {
        this.showSnackbarError({
          message: "Error while change invoice based",
        });
      } finally {
        this.changeRegularPaymentUuid = "";
      }
    },
  },
  computed: {
    ...mapGetters("appSearch", {
      searchParam: "param",
      filter: "filter",
    }),
    accounts() {
      return this.$store.getters["accounts/all"];
    },
    filteredAccounts() {
      if (this.notFiltered) {
        return this.accounts;
      }

      const filter = { ...this.filter };
      const filterKeys = Object.keys(filter).filter((key) => !!filter[key]);

      const accounts = this.accounts.filter((a) => {
        return filterKeys.every((key) => {
          let data;
          if (key === "namespace") {
            data = this.getNamespaceName(a.uuid);
          } else {
            data = getDeepObjectValue(a, key);
          }

          return compareSearchValue(data, filter[key], this.searchFields[key]);
        });
      });

      if (this.searchParam) {
        return filterByKeysAndParam(
          accounts,
          ["title", "uuid", "data.email"],
          this.searchParam
        );
      }
      return accounts;
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    searchFields() {
      return {
        title: {
          title: "Title",
          type: "input",
        },
        status: {
          title: "Status",
          type: "select",
          items: [...new Set(this.accounts.map((a) => a.status))],
        },
        balance: { title: "Balance", type: "number-range" },
        "data.email": { title: "Email", type: "input" },
        "data.date_create": { title: "Created date", type: "date" },
        "data.country": { title: "Country", type: "input" },
        "data.address": { title: "Address", type: "input" },
        currency: {
          title: "Client currency",
          type: "select",
          items: this.$store.getters["currencies/all"].filter(
            (c) => c !== "NCU"
          ),
        },
        "access.level": {
          title: "Access level",
          type: "select",
          items: Object.keys(this.levelColorMap),
        },
        "data.regular_payment": {
          title: "Invoice based",
          type: "logic-select",
        },
        namespace: {
          title: "Group(NameSpace)",
          type: "select",
          items: [...new Set(this.namespaces.map((n) => n.title))],
        },
      };
    },
  },
  created() {
    this.loading = true;
    this.$store.dispatch("namespaces/fetch");
    this.$store
      .dispatch("accounts/fetch")
      .then(() => {
        this.fetchError = "";
      })
      .finally(() => {
        this.loading = false;
      })
      .catch((err) => {
        console.error(err.toJSON());
        this.fetchError = "Can't reach the server";
        if (err.response && err.response.data.message) {
          this.fetchError += `: [ERROR]: ${err.response.data.message}`;
        } else {
          this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
        }
      });
  },
  watch: {
    accounts() {
      this.fetchError = "";
    },
    searchFields() {
      this.$store.commit(
        "appSearch/setFields",
        Object.keys(this.searchFields).map((key) => ({
          ...this.searchFields[key],
          key,
        }))
      );
    },
    value() {
      this.selected = this.value;
    },
  },
};
</script>

<style></style>
