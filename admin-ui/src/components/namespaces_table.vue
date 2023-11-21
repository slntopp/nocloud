<template>
  <nocloud-table
    table-name="namespaces"
    :loading="isLoading"
    :items="filtredNamespaces"
    :headers="headers"
    :value="selected"
    @input="handleSelect"
    :single-select="singleSelect"
    :footer-error="fetchError"
  >
    <template v-slot:[`item.access`]="{ item }">
      <v-chip color="info">
        {{ getName(item.access.namespace) }} ({{ item.access.level }})
      </v-chip>
    </template>
    <template v-slot:[`item.title`]="{ item }">
      <router-link
        :to="{ name: 'NamespacePage', params: { namespaceId: item.uuid } }"
      >
        {{ item.title }}
      </router-link>
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import {
  compareSearchValue,
  filterArrayByTitleAndUuid,
  getDeepObjectValue,
} from "@/functions";
import search from "@/mixins/search";
import { mapGetters } from "vuex";

export default {
  name: "namespaces-table",
  mixins: [search("namespaces-table")],
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
  },
  data() {
    return {
      headers: [
        { text: "Title", value: "title" },
        { text: "UUID", value: "uuid" },
        { text: "Access", value: "access" },
      ],
      selected: this.value,
      fetchError: "",
    };
  },
  methods: {
    handleSelect(item) {
      this.$emit("input", item);
    },
    getName(account) {
      return this.accounts.find(({ uuid }) => account === uuid)?.title ?? "";
    },
  },
  computed: {
    ...mapGetters("appSearch", ["filter", "param"]),
    ...mapGetters("namespaces", { tableData: "all", isLoading: "isLoading" }),
    filtredNamespaces() {
      const namespaces = this.tableData.filter((n) =>
        Object.keys(this.filter).every((key) => {
          const data = getDeepObjectValue(n, key);

          return compareSearchValue(
            data,
            this.filter[key],
            this.searchFields.find((f) => f.key === key)
          );
        })
      );

      if (this.param) {
        return filterArrayByTitleAndUuid(namespaces, this.param);
      }
      return namespaces;
    },
    accessLevels() {
      return [...new Set(this.tableData.map((n) => n.access.level)), "NONE"];
    },
    searchFields() {
      return [
        {
          title: "Title",
          key: "title",
          type: "input",
        },
        {
          title: "Access",
          key: "access.level",
          type: "select",
          items: this.accessLevels,
        },
      ];
    },
    accounts() {
      return this.$store.getters["accounts/all"];
    },
  },
  created() {
    this.$store.dispatch("accounts/fetch");
    this.$store
      .dispatch("namespaces/fetch")
      .then(() => {
        this.fetchError = "";
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
      this.$store.commit("appSearch/setFields", this.searchFields);
    },
  },
};
</script>

<style></style>
