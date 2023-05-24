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
    :filters-items="filterItems"
    :filters-values="selectedFilter"
    @input:filter="selectedFilter[$event.key] = $event.value"
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
        :currency="item.currency"
        @click="goToBalance(item.uuid)"
        v-if="item.balance"
        :value="item.balance"
      />
      <template v-else>-</template>
    </template>
    <template v-slot:[`item.access.level`]="{ value }">
      <v-chip :color="colorChip(value)">
        {{ value }}
      </v-chip>
    </template>
    <template v-slot:[`item.namespace`]="{ item }">
      {{ "NS_" + getName(item.uuid) }}
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import Balance from "./balance.vue";
import { filterArrayByTitleAndUuid } from "@/functions";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";

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
    searchParam: {
      type: String,
      default: "",
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
        { text: "Access level", value: "access.level", customFilter: true },
        { text: "Group(NameSpace)", value: "namespace" },
      ],
      levelColorMap: {
        ROOT: "info",
        ADMIN: "success",
        MGMT: "warning",
        READ: "gray",
        NONE: "error",
      },
      selectedFilter: { "access.level": [] },
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
    tableData() {
      return this.$store.getters["accounts/all"];
    },
    filtredAccounts() {
      const accounts = this.tableData.filter(
        (a) =>
          this.selectedFilter["access.level"].length === 0 ||
          this.selectedFilter["access.level"].includes(a.access.level)
      );

      if (this.searchParam) {
        return filterArrayByTitleAndUuid(accounts, this.searchParam);
      }
      return accounts;
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    filterItems() {
      return { "access.level": Object.keys(this.levelColorMap) };
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
    tableData() {
      this.fetchError = "";
    },
  },
};
</script>

<style></style>
