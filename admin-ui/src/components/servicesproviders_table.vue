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
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import { filterArrayByTitleAndUuid } from "@/functions";

const Headers = [
  { text: "title", value: "titleLink" },
  { text: "type", value: "type", customFilter: true },
  { text: "state", value: "state" },
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
    searchParams: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      selected: this.value,
      loading: false,
      Headers,
      fetchError: "",
      allTypes: [],
      selectedFiltres: { type: [] },
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
      switch (state.toLowerCase()) {
        case "running":
        case "operation":
          return "success";
        case "unknown":
        case "deleted":
        case "failure":
          return "error";

        default:
          return "gray";
      }
    },
    filterSpByTypes(spArray, types) {
      return spArray.filter((sp) => types.includes(sp.type));
    },
  },
  computed: {
    tableData() {
      return this.$store.getters["servicesProviders/all"].map((el) => ({
        titleLink: el.title,
        type: el.type,
        uuid: el.uuid,
        route: {
          name: "ServicesProvider",
          params: { uuid: el.uuid },
        },
        state: el?.state?.state ?? "UNKNOWN",
        region: el.secrets?.endpoint?.split("-")[1] ?? "-",
      }));
    },
    filteredSp() {
      const isAdvanced = this.selectedFiltres.type?.length > 0;
      if (this.searchParams.param || isAdvanced) {
        const filtred =
          !this.selectedFiltres?.type?.includes("all") && isAdvanced > 0
            ? this.filterSpByTypes(
                this.tableData,
                this.selectedFiltres.type
              )
            : this.tableData;

        return this.searchParams.param
          ? filterArrayByTitleAndUuid(
              filtred,
              this.searchParams.param,
              true,
              "titleLink"
            )
          : filtred;
      }
      return this.tableData;
    },
    filterItems() {
      return {
        type: this.allTypes,
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
              Headers.push({ text: "region", value: "region" });
            }
            this.$store.dispatch("servicesProviders/fetchById", el.uuid);
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

    this.allTypes.push('all')
  },
  watch: {
    tableData() {
      this.fetchError = "";
    },
  },
};
</script>

<style></style>
