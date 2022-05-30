<template>
  <nocloud-table
    :loading="loading"
    :items="tableData"
    :value="selected"
    @input="handleSelect"
    :single-select="singleSelect"
    :headers="Headers"
    item-key="original"
    :footer-error="fetchError"
  >
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";

const Headers = [{ text: "title", value: "titleLink" }];

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
      this.fetchError = '';
    }
  }
};
</script>

<style></style>
