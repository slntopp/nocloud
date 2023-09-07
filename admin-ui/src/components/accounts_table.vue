<template>
  <nocloud-table
    table-name="accounts"
    :headers="headers"
    :items="filtredAccounts"
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
    <template v-slot:[`item.access.level`]="{ value }">
      <v-chip :color="colorChip(value)">
        {{ value }}
      </v-chip>
    </template>
    <template v-slot:[`item.namespace`]="{ item }">
      {{ getName(item.uuid) }}
    </template>
    <template v-slot:[`item.currency`]="{ item }">
      {{ item.currency || defaultCurrency }}
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import Balance from "./balance.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import { mapGetters } from "vuex";
import { filterByKeysAndParam } from "@/functions";

export default {
  name: "accounts-table",
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
      selected: this.value,
      loading: false,
      fetchError: "",
      headers: [
        { text: "Title", value: "title" },
        { text: "UUID", value: "uuid" },
        { text: "Balance", value: "balance" },
        { text: "Email", value: "data.email" },
        { text: "Client currency", value: "currency" },
        { text: "Access level", value: "access.level" },
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
    getName(uuid) {
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
  },
  computed: {
    ...mapGetters("appSearch", {
      searchParam: "customSearchParam",
      searchParams: "customParams",
    }),
    tableData() {
      return this.$store.getters["accounts/all"];
    },
    filtredAccounts() {
      const searchParams = { ...this.searchParams };

      if (this.namespace) {
        searchParams["access.namespace"] = [{ value: this.namespace }];
      }

      const accounts = this.tableData.filter((a) =>
        Object.keys(searchParams).every((k) => {
          return (
            !searchParams?.[k] ||
            !searchParams[k].length ||
            searchParams[k]?.find((t) => {
              let key = k;
              let data = { ...a };
              k.split(".").forEach((subKey, index) => {
                if (index === k.split(".").length - 1) {
                  key = subKey;
                  return;
                }
                data = a[subKey];
              });
              return t.value === data[key];
            })
          );
        })
      );

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
  mounted() {
    setTimeout(() => {
      this.$store.commit("appSearch/setVariants", {
        "access.level": {
          items: Object.keys(this.levelColorMap),
          isArray: true,
          title: "Access",
        },
      });
    }, 0);
  },
  watch: {
    tableData() {
      this.fetchError = "";
    },
  },
};
</script>

<style></style>
