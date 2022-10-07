<template>
  <nocloud-table
    :loading="loading"
    :items="filtredNamespaces"
    :value="selected"
    @input="handleSelect"
    :single-select="singleSelect"
    :footer-error="fetchError"
  >
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
      selected: this.value,
      loading: false,
      fetchError: "",
    };
  },
  methods: {
    handleSelect(item) {
      this.$emit("input", item);
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
  },
  created() {
    this.loading = true;
    this.$store
      .dispatch("namespaces/fetch")
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
