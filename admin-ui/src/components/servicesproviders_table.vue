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
      <v-chip small :color="getStateChipColor(value)">
        {{ value }}
      </v-chip>
    </template>
    <template v-slot:[`item.status`]="{ value }">
      <v-chip small :color="getStatusChipColor(value)">
        {{ value }}
      </v-chip>
    </template>
  </nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import {
  compareSearchValue,
  filterArrayByTitleAndUuid,
  getDeepObjectValue,
  getShortName
} from "@/functions";
import search from "@/mixins/search";
import { mapGetters } from "vuex";

const Headers = [
  { text: "Title", value: "titleLink" },
  { text: "Type", value: "type" },
  { text: "State", value: "state" },
  { text: "Status", value: "status" },
  {
    text: "UUID",
    align: "start",
    sortable: true,
    value: "uuid",
  },
];

const statusMap = {
  DEL: { title: "DELETED", color: "blue-grey darken-2" },
  UNSPECIFIED: { title: "ACTIVE", color: "success" },
  UNKNOWN: {
    title: "UNKNOWN",
    color: "red darken-2",
  },
};

export default {
  name: "servicesProviders-table",
  mixins: [
    search({
      name: "service-providers-table",
      defaultLayout: { title: "Default", filter: { status: ["ACTIVE"] } },
    }),
  ],
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
      allTypes: [],
      stateColorMap: {
        running: "success",
        operation: "success",
        unknown: "error",
        deleted: "error",
        failure: "error",
      },
      statusMap,
    };
  },
  methods: {
    handleSelect(item) {
      this.$emit("input", item);
    },
    getStateChipColor(state) {
      if (!state) {
        return "gray";
      }
      return this.stateColorMap[state.toLowerCase()] || "";
    },
    getStatusChipColor(status) {
      const statusKey = Object.keys(this.statusMap).find(
        (key) => this.statusMap[key].title === status
      );
      return this.statusMap[statusKey]?.color;
    },
    getInstanceTypes() {
      const types = require.context(
        "@/components/modules/",
        true,
        /instanceCreate\.vue$/
      );

      types.keys().forEach((key) => {
        const matched = key.match(
          /\.\/([A-Za-z0-9-_,\s]*)\/instanceCreate\.vue/i
        );
        if (matched && matched.length > 1) {
          this.allTypes.push(matched[1]);
        }
      });
    },
  },
  computed: {
    ...mapGetters("appSearch", { filter: "filter", searchParam: "param" }),
    tableData() {
      return this.$store.getters["servicesProviders/all"].map((el) => ({
        titleLink: getShortName(el.title,45),
        title: el.title,
        type: el.type,
        uuid: el.uuid,
        status: statusMap[el.status]?.title || statusMap.UNKNOWN.title,
        route: {
          name: "ServicesProvider",
          params: { uuid: el.uuid },
        },
        meta: el.meta,
        state: el?.state?.state ?? "UNKNOWN",
        region: el.secrets?.endpoint?.split("-")[1] ?? "-",
      }));
    },
    filteredSp() {
      const sp = this.tableData.filter((sp) =>
        Object.keys(this.filter).every((key) => {
          const data = getDeepObjectValue(sp, key);

          return compareSearchValue(
            data,
            this.filter[key],
            this.searchFields.find((f) => f.key === key)
          );
        })
      );
      if (this.searchParam) {
        return filterArrayByTitleAndUuid(sp, this.searchParam);
      }
      return sp;
    },
    searchFields() {
      return [
        { title: "Title", type: "input", key: "title" },
        { items: this.allTypes, title: "Type", type: "select", key: "type" },
        {
          items: Object.keys(this.statusMap).map(
            (key) => this.statusMap[key].title
          ),
          title: "Status",
          type: "select",
          key: "status",
        },
        { title: "Region", type: "input", key: "region" },
        {
          items: Object.keys(this.stateColorMap).map((k) => k.toUpperCase()),
          title: "State",
          type: "select",
          key: "state",
        },
      ];
    },
  },
  created() {
    this.loading = true;
    this.$store
      .dispatch("servicesProviders/fetch", {
        anonymously: false,
        showDeleted: true,
      })
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
        this.$store.commit("appSearch/setFields", this.searchFields);
      }
    },
  },
};
</script>

<style></style>
