<template>
  <nocloud-table
    :loading="loading"
    :items="filtredZones"
    :value="selected"
    @input="handleSelect"
    :single-select="singleSelect"
    :headers="Headers"
    item-key="original"
    :footer-error="fetchError"
    table-name="zonesTable"
  >
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import { filterArrayIncludes } from "@/functions";

const Headers = [{ text: "Title", value: "titleLink" }];

export default {
  name: "zones-table",
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
      Headers,
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
      return this.$store.getters["dns/all"].map((el) => ({
        titleLink: el.replace(/\.$/, ""),
        original: el,
        route: { name: "Zone manager", params: { dnsname: el } },
      }));
    },
    filtredZones() {
      if (this.searchParam) {
        return filterArrayIncludes(this.tableData, {
          keys: ["titleLink"],
          value: this.searchParam,
        });
      }
      return this.tableData;
    },
  },
  created() {
    this.loading = true;
    this.$store
      .dispatch("dns/fetch")
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
