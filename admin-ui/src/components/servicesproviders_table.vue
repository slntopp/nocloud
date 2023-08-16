<template>
  <nocloud-table
    table-name="service-providers-table"
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
      <icon-title-preview :is-mdi="false" :instance="getShowcase(item)" />
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import { filterArrayByTitleAndUuid } from "@/functions";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";

const Headers = [
  { text: "Title", value: "titleLink" },
  { text: "Type", value: "type" },
  { text: "State", value: "state" },
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
    getShowcase(item) {
      return Object.values(item.meta?.showcase ?? {})[0];
    },
    getInstanceTypes() {
      const types = require.context(
        "@/components/modules/",
        true,
        /serviceCreate\.vue$/
      );

      types.keys().forEach((key) => {
        const matched = key.match(
          /\.\/([A-Za-z0-9-_,\s]*)\/serviceCreate\.vue/i
        );
        if (matched && matched.length > 1) {
          this.allTypes.push(matched[1]);
        }
      });
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
      return this.$store.getters["appSearch/customSearchParam"];
    },
    filteredSp() {
      const sp = this.tableData.filter((sp) => {
        return Object.keys(this.searchParams || {}).every(
          (key) =>
            this.searchParams[key].length === 0 ||
            this.searchParams[key].find((p) => sp[key]?.includes(p.value))
        );
      });
      if (this.searchParam) {
        return filterArrayByTitleAndUuid(sp, this.searchParam);
      }
      return sp;
    },
    searchParams() {
      return this.$store.getters["appSearch/customParams"];
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

    this.getInstanceTypes();
  },
  watch: {
    tableData() {
      this.fetchError = "";
    },
    allTypes(val) {
      if (val) {
        this.$store.commit("appSearch/setSearchName", "services-providers");
        this.$store.commit("appSearch/setVariants", {
          type: { items: this.allTypes, title: "Type", isArray: true },
          state: {
            items: Object.keys(this.stateColorMap).map((k) => k.toUpperCase()),
            title: "State",
            isArray: true,
          },
        });
      }
    },
  },
};
</script>

<style></style>
