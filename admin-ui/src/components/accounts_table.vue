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
      <v-chip :color="item.access.color">
        {{ item.access.level }}
      </v-chip>
    </template>
    <template v-slot:[`item.data.regular_payment`]="{ value, item }">
      <v-switch
        :disabled="
          !!changeRegularPaymentUuid && changeRegularPaymentUuid !== item.uuid
        "
        :loading="
          !!changeRegularPaymentUuid && changeRegularPaymentUuid === item.uuid
        "
        @change="changeRegularPayment(item, $event)"
        :input-value="value"
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
  mixins: [search({name:"accounts-table"})],
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
      return this.$store.getters["accounts/all"].map((a) => ({
        ...a,
        access: {
          ...a.access,
          color: this.colorChip(a.access.level),
        },
        balance: a.balance || 0,
        currency: a.currency || this.defaultCurrency,
        namespace: this.getNamespaceName(a.uuid),
        data: {
          ...a.data,
          date_create: formatSecondsToDate(a.data?.date_create),
          regular_payment:
            a.data?.regular_payment === undefined ||
            a.data?.regular_payment === true,
        },
      }));
    },
    filteredAccounts() {
      const filter = { ...this.filter };

      const accounts = this.accounts.filter((a) => {
        return Object.keys(filter).every((key) => {
          let data;
          if (key === "namespace") {
            data = this.getNamespaceName(a.uuid);
          } else {
            data = getDeepObjectValue(a, key);
          }

          return compareSearchValue(
            data,
            filter[key],
            this.searchFields.find((f) => f.key === key)
          );
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
      return [
        {
          title: "Title",
          key: "title",
          type: "input",
        },
        {
          title: "Status",
          key: "status",
          type: "select",
          items: [...new Set(this.accounts.map((a) => a.status))],
        },
        { title: "Balance", key: "balance", type: "number-range" },
        { title: "Email", key: "data.email", type: "input" },
        { title: "Created date", key: "data.date_create", type: "date" },
        { title: "Country", key: "data.country", type: "input" },
        { title: "Address", key: "data.address", type: "input" },
        {
          title: "Client currency",
          key: "currency",
          type: "select",
          items: this.$store.getters["currencies/all"].filter(
            (c) => c !== "NCU"
          ),
        },
        {
          title: "Access level",
          key: "access.level",
          type: "select",
          items: Object.keys(this.levelColorMap),
        },
        {
          title: "Invoice based",
          key: "data.regular_payment",
          type: "logic-select",
        },
        {
          title: "Group(NameSpace)",
          key: "namespace",
          type: "select",
          items: [
            ...new Set(this.accounts.map((a) => this.getNamespaceName(a.uuid))),
          ],
        },
      ];
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
      this.$store.commit("appSearch/setFields", this.searchFields);
    },
    value() {
      this.selected = this.value;
    },
  },
};
</script>

<style></style>
