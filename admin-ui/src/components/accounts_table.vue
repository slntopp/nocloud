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
      </div>
    </template>
    <template v-slot:[`item.balance`]="{ item }">
      <balance v-if="item.balance" :value="item.balance" />
      <template v-else>-</template>
    </template>
    <template v-slot:[`item.access`]="{ item }">
      <v-chip :color="colorChip(item.access.level)">
        {{ item.access.level }}
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

export default {
  name: "accounts-table",
  components: {
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
        { text: "Access level", value: "access" },
        { text: "Group(NameSpace)", value: "namespace" },
      ],
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
      switch (level) {
        case "ROOT":
          return "info";
        case "ADMIN":
          return "success";
        case "MGMT":
          return "warning";
        case "READ":
          return "gray";
        case "NONE":
          return "error";
      }
    },
  },
  computed: {
    tableData() {
      return this.$store.getters["accounts/all"];
    },
    filtredAccounts() {
      if (this.searchParam) {
        return filterArrayByTitleAndUuid(this.tableData, this.searchParam);
      }
      return this.tableData;
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
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
