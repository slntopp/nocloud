<template>
  <nocloud-table :loading="loading" :items="filtredNamespaces" :headers="headers" :value="selected" @input="handleSelect"
    :single-select="singleSelect" :footer-error="fetchError">
    
    <template v-slot:[`item.access`]="{ item }">
      <v-chip color="info">
        {{ getName(item.access.namespace) }} ({{ item.access.level }})
      </v-chip>
    </template>
    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'NamespacePage', params: { namespaceId: item.uuid } }">
        {{ 'NS_' + item.title  }}
      </router-link>
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import { filterArrayByTitleAndUuid } from "@/functions";

export default {
  name: "namespaces-table",
  components: {
    "nocloud-table": noCloudTable,
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
      headers: [
        { text: "Title", value: "title" },
        { text: "UUID", value: "uuid" },
        { text: "Access", value: "access" }
      ],
      selected: this.value,
      loading: false,
      fetchError: "",
    };
  },
  methods: {
    handleSelect(item) {
      this.$emit("input", item);
    },
    getName(account) {
      return this.accounts.find(({ uuid }) => account === uuid)?.title ?? '';
    },
  },
  computed: {
    tableData() {
      return this.$store.getters["namespaces/all"];
    },
    filtredNamespaces() {
      if (this.searchParam) {
        return filterArrayByTitleAndUuid(this.tableData, this.searchParam);
      }
      return this.tableData;
    },
    accounts() {
      return this.$store.getters["accounts/all"];
    },
  },
  created() {
    this.loading = true;
    this.$store.dispatch("accounts/fetch");
    this.$store.dispatch("namespaces/fetch")
      .then(() => {
        this.fetchError = "";
      })
      .finally(() => {
        this.loading = false;
      })
      .catch((err) => {
        console.log(`err`, err);
        this.fetchError = "Can't reach the server";
        if (err.response) {
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
