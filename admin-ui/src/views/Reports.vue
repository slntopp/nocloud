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
      table-name="reports"
      :headers="reportsHeaders"
      :items="reports"
      :loading="isLoading"
      :footer-error="fetchError"
      :server-items-length="count"
      :server-side-page="page"
      sort-by="exec"
      sort-desc
      @update:options="onUpdateOptions"
      no-hide-uuid
    >
      <template v-slot:[`item.uuid`]="{ value }">
        <router-link
          v-if="!isInstancesLoading"
          :to="{ name: 'Instance', params: { instanceId: value } }"
        >
          {{ getShortName(getInstance(value)?.title || value, 100) }}
        </router-link>

        <v-skeleton-loader type="text" v-else />
      </template>
      <template v-slot:[`item.totalDefaultPreview`]="{ item }">
        {{
          `${formatPrice(
            convertTo(item.total, item.currency),
            defaultCurrency
          )} ${defaultCurrency?.code}`
        }}
      </template>

      <template v-slot:[`item.totalPreview`]="{ item }">
        {{
          `${formatPrice(item.total, item.currency)} ${item.currency?.code}`
        }}
      </template>
    </nocloud-table>
  </v-card>
</template>

<script setup>
import { ref, computed, watch, onMounted } from "vue";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import DatePicker from "@/components/ui/datePicker.vue";
import useCurrency from "@/hooks/useCurrency";
import { getShortName } from "@/functions";
import { formatPrice } from "../functions";

const store = useStore();
const { convertTo, defaultCurrency } = useCurrency();

const reports = ref([]);
const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const isInstancesLoading = ref(false);
const fetchError = ref("");
const options = ref({});
const durationFilter = ref({ to: "", from: "" });

const reportsHeaders = [
  { text: "Instance", value: "uuid" },
  { text: "Total", value: "totalPreview" },
  { text: "Total in default currency", value: "totalDefaultPreview" },
];

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => {
      onUpdateOptions(options.value);
    },
  });
});

const isLoading = computed(() => {
  return isFetchLoading.value || isCountLoading.value;
});

const requestOptions = computed(() => ({
  page: page.value,
  limit: options.value.itemsPerPage,
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
    const { reports: result } = await api.reports.base_list(
      requestOptions.value
    );

    reports.value = result.map((r) => {
      return {
        uuid: r.uuid,
        totalPreview: `${r.total.toFixed(2)} ${r.currency?.code}`,
        total: r.total,
        currency: r.currency,
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
    count.value = +(await api.reports.base_count(requestOptions.value)).total;
  } finally {
    isCountLoading.value = false;
  }
};

const getInstance = (uuid) => {
  return store.getters["instances/cached"].get(uuid) || uuid;
};

watch(
  durationFilter,
  () => {
    onUpdateOptions(options.value);
  },
  { deep: true }
);

watch(reports, async () => {
  isInstancesLoading.value = true;
  try {
    await Promise.all(
      reports.value.map(({ uuid }) =>
        store.dispatch("instances/fetchToCached", uuid)
      )
    );
  } finally {
    isInstancesLoading.value = false;
  }
});
</script>

<script>
export default {
  name: "reports-page",
};
</script>

<style scoped lang="scss"></style>
