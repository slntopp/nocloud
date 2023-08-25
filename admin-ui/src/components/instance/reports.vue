<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <div style="max-width: 300px" class="mx-3">
        <date-picker label="from" v-model="durationFilter.from" />
      </div>
      <div style="max-width: 300px" class="mx-3">
        <date-picker label="to" v-model="durationFilter.to" />
      </div>
    </v-row>
    <nocloud-table
      table-name="instance-reports"
      :headers="instanceReportsHeaders"
      :items="instanceReports"
      :loading="isLoading"
      :footer-error="fetchError"
      :server-items-length="count"
      :server-side-page="page"
      sort-by="exec"
      sort-desc
      @update:options="onUpdateOptions"
      no-hide-uuid
      :itemsPerPageOptions="itemsPerPageOptions"
    >
    </nocloud-table>
  </v-card>
</template>

<script setup>
import { toRefs, ref, computed, watch } from "vue";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import useRate from "@/hooks/useRate";
import DatePicker from "@/components/ui/datePicker.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const instanceReports = ref([]);
const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const fetchError = ref("");
const options = ref({});
const itemsPerPageOptions = ref([5, 10, 15, 25]);
const durationFilter = ref({ to: "", from: "" });

const userCurrency = ref("");
const { rate, convertTo } = useRate(userCurrency);

const instanceReportsHeaders = [
  { text: "Duration", value: "duration" },
  { text: "Executed date", value: "exec" },
  { text: "Total", value: "totalPreview" },
  { text: "Total in default currency", value: "totalDefaultPreview" },
  { text: "Product or resource", value: "item" },
];

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const isLoading = computed(() => {
  return isFetchLoading.value || isCountLoading.value;
});

const requestOptions = computed(() => ({
  page: page.value,
  limit: options.value.itemsPerPage,
  instanceUuid: template.value.uuid,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
  from: durationFilter.value.from
    ? new Date(durationFilter.value.from).getTime() / 1000
    : undefined,
  to: durationFilter.value.to
    ? new Date(durationFilter.value.to).getTime() / 1000
    : undefined,
}));

const onUpdateOptions = async (newOptions) => {
  setOptions(newOptions);
  page.value = newOptions.page;
  init();
  isFetchLoading.value = true;
  try {
    const { records: result } = await api.reports.list(requestOptions.value);
    if (result.length) {
      userCurrency.value = result[0].currency;
    }

    instanceReports.value = result.map((r) => {
      return {
        totalPreview: `${r.total} ${r.currency}`,
        total: r.total,
        duration: `${new Date(r.start * 1000).toLocaleString()} - ${new Date(
          r.end * 1000
        ).toLocaleString()}`,
        exec: new Date(r.exec * 1000).toLocaleString(),
        item: r.product || r.resource,
        totalDefaultPreview: rate.value
          ? `${convertTo(r.total)} ${defaultCurrency.value}`
          : null,
      };
    });
  } finally {
    isFetchLoading.value = false;
  }
};

const setOptions = (newOptions) => {
  const sortByReplaceKeys = {
    totalPreview: "total",
    totalDefaultPreview: "total",
    duration: "start",
  };
  options.value = {
    ...newOptions,
    sortBy: newOptions.sortBy.map((k) => sortByReplaceKeys[k] || k),
  };
};

const init = async () => {
  isCountLoading.value = true;
  try {
    count.value = +(await api.reports.count(requestOptions.value)).total;
  } finally {
    isCountLoading.value = false;
  }
};

watch(rate, () => {
  instanceReports.value = instanceReports.value.map((r) => ({
    ...r,
    totalDefaultPreview: `${convertTo(r.total)} ${defaultCurrency.value}`,
  }));
});

watch(
  durationFilter,
  () => {
    onUpdateOptions(options.value);
  },
  { deep: true }
);
</script>

<script>
export default {
  name: "instance-report",
};
</script>

<style scoped lang="scss"></style>
