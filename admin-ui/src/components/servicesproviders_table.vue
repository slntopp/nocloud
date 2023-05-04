<template>
  <nocloud-table
    :filters-items="filterItems"
    :filters-values="selectedFiltres"
    @input:filter="selectedFiltres[$event.key] = $event.value"
    table-name="serviceProvidersTable"
    :loading="loading"
    :items="filteredSp"
    :value="selected"
    @input="handleSelect"
    :single-select="singleSelect"
    :headers="Headers"
    item-key="uuid"
    :footer-error="fetchError"
  >
    <template v-slot:[`item.state`]="{ value }">
      <v-chip small :color="chipsColor(value)">
        {{ value }}
      </v-chip>
    </template>
    <template v-slot:[`item.meta`]="{ item }">
      <icon-title-preview
        :is-mdi="false"
        :title="item.meta.service.title"
        :icon="item.meta.service.icon"
      />
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import { filterArrayByTitleAndUuid } from "@/functions";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";

const Headers = [
  { text: "Title", value: "titleLink" },
  { text: "Type", value: "type", customFilter: true },
  { text: "State", value: "state", customFilter: true },
  { text: "Preview", value: "meta" },
  {
    text: "UUID",
    align: "start",
    sortable: true,
    value: "uuid",
  },
];

export default {
  name: "servicesProviders-table",
  components: {
    IconTitlePreview,
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
      allTypes: [],
      selectedFiltres: { type: [], state: [] },
      stateColorMap: {
        running: "success",
        operation: "success",
        unknown: "error",
        deleted: "error",
        failure: "error",
      },
    };
  },
  methods: {
    handleSelect(item) {
      this.$emit("input", item);
    },
    chipsColor(state) {
      if (!state) {
        return "gray";
      }
      return this.stateColorMap[state.toLowerCase()] || "";
    },
  },
  computed: {
    tableData() {
      return this.$store.getters["servicesProviders/all"].map((el) => ({
        titleLink: el.title,
        title: el.title,
        type: el.type,
        uuid: el.uuid,
        route: {
          name: "ServicesProvider",
          params: { uuid: el.uuid },
        },
        meta: el.meta,
        state: el?.state?.state ?? "UNKNOWN",
        region: el.secrets?.endpoint?.split("-")[1] ?? "-",
      }));
    },
    searchParam() {
      return this.$store.getters["appSearch/param"];
    },
    filteredSp() {
      const sp = this.tableData.filter((sp) => {
        return Object.keys(this.selectedFiltres).every(
          (key) =>
            this.selectedFiltres[key].length === 0 ||
            this.selectedFiltres[key].includes(sp[key])
        );
      });
      if (this.searchParam) {
        return filterArrayByTitleAndUuid(sp, this.searchParam);
      }
      return sp;
    },
    filterItems() {
      return {
        type: this.allTypes,
        state: Object.keys(this.stateColorMap).map((k) => k.toUpperCase()),
      };
    },
  },
  created() {
    this.loading = true;
    this.$store
      .dispatch("servicesProviders/fetch", false)
      .then(({ pool }) => {
        pool.forEach((el) => {
          if (el.type === "ovh") {
            const isRegionIncludes = Headers.find(
              (el) => el.value === "region"
            );

            if (!isRegionIncludes) {
              Headers.push({ text: "Region", value: "region" });
            }
          }
        });
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

    const types = require.context(
      "@/components/modules/",
      true,
      /serviceCreate\.vue$/
    );

    types.keys().forEach((key) => {
      const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/serviceCreate\.vue/i);
      if (matched && matched.length > 1) {
        this.allTypes.push(matched[1]);
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
