<template>
  <nocloud-table
    :table-name="tableName"
    class="mt-4"
    :value="value"
    @input="emit('input', $event)"
    sort-by="created"
    sort-desc
    :items="promocodes"
    :headers="headers"
    :loading="isLoading"
    :server-items-length="count"
    :server-side-page="page"
    @update:options="setOptions"
    :show-select="!hideSelect"
  >
    <template v-slot:[`item.dueDate`]="{ item }">
      {{ formatSecondsToDate(item.dueDate, true) || "Unlimited" }}
    </template>

    <template v-slot:[`item.code`]="{ item }">
      <div>
        <v-chip>{{ item.code }}</v-chip>
        <v-btn icon @click="copyCode(item)" class="ml-2">
          <v-icon>mdi-content-copy</v-icon>
        </v-btn>
      </div>
    </template>

    <template v-slot:[`item.uses`]="{ item }">
      {{ (item.uses || []).length }} {{ item.limit ? `/ ${item.limit}` : "" }}
    </template>

    <template v-slot:[`item.created`]="{ item }">
      {{ formatSecondsToDate(item.created, true) }}
    </template>

    <template v-slot:[`item.status`]="{ item }">
      <promocode-status-chip :item="item" />
    </template>

    <template v-slot:[`item.condition`]="{ item }">
      <promocode-condition-chip :item="item" />
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <router-link
        :to="{ name: 'Promocode page', params: { uuid: item.uuid } }"
      >
        {{ item.title }}
      </router-link>
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import promocodeConditionChip from "@/components/promocode/ui/promocodeConditionChip.vue";
import promocodeStatusChip from "@/components/promocode/ui/promocodeStatusChip.vue";
import { addToClipboard, debounce, formatSecondsToDate } from "../functions";
import { ref, computed, watch, toRefs, onMounted } from "vue";
import { useStore } from "@/store";
import useSearch from "@/hooks/useSearch";
import {
  PromocodeCondition,
  PromocodeStatus,
} from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";

const props = defineProps({
  tableName: { type: String, default: "promocode-table" },
  value: {},
  customFilter: {},
  noSearch: { type: Boolean, default: false },
  hideSelect: { type: Boolean, default: false },
  refetch: { type: Boolean, default: false },
});

const { tableName, value, refetch, noSearch, customFilter } = toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();
useSearch({
  name: props.tableName,
  noSearch: props.noSearch,
});

const count = ref(0);
const page = ref(1);
const isFetchLoading = ref(true);
const isCountLoading = ref(true);
const options = ref({});
const fetchError = ref("");

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Code", value: "code" },
  { text: "Created date ", value: "created" },
  { text: "Due date", value: "dueDate" },
  { text: "Uses", value: "uses" },
  { text: "Status", value: "status" },
  { text: "Condition", value: "condition" },
]);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: fetchPromocodes,
  });

  if (!noSearch.value) {
    store.commit("appSearch/setFields", searchFields.value);
  }
});

const promocodes = computed(() => store.getters["promocodes/all"]);
const isLoading = computed(() => isFetchLoading.value || isCountLoading.value);

const filter = computed(() => store.getters["appSearch/filter"]);
const searchParam = computed(() => store.getters["appSearch/param"]);
const searchFields = computed(() => [
  {
    title: "Status",
    key: "status",
    type: "select",
    items: Object.keys(PromocodeStatus)
      .filter((value) => !Number.isInteger(+value))
      .map((key) => ({
        text: key,
        value: PromocodeStatus[key],
      })),
  },
  {
    title: "Condition",
    key: "condition",
    type: "select",
    items: Object.keys(PromocodeCondition)
      .filter((value) => !Number.isInteger(+value))
      .map((key) => ({
        text: key,
        value: PromocodeCondition[key],
      })),
  },
  { title: "Due date", key: "dueDate", type: "date" },
  {
    title: "Uses",
    key: "uses",
    type: "number-range",
  },
  {
    title: "Limit",
    key: "limit",
    type: "number-range",
  },
  { title: "Created date", key: "created", type: "date" },
]);

const promocodesFilters = computed(() => {
  const filters = {};

  if (noSearch.value) {
    for (const key of Object.keys(customFilter.value)) {
      filters[key] = customFilter.value[key];
    }
  } else {
    const datekeys = ["created", "dueDate"];

    for (const key of Object.keys(filter.value)) {
      const value = filter.value[key];

      if (
        !value ||
        (Array.isArray(value) && !value.length) ||
        (typeof value === "object" && !Object.keys(value).length)
      ) {
        continue;
      }

      if (value?.to || value?.from) {
        const total = {};
        if (value?.to) {
          total.to = +value?.to;
        }
        if (value?.from) {
          total.from = +value?.from;
        }

        filters[key] = total;
        continue;
      }

      if (datekeys.includes(key)) {
        let dates = [];

        if (value[0]) {
          dates.push(new Date(value[0]).getTime() / 1000);
        }
        if (value[1]) {
          dates.push(new Date(value[1]).getTime() / 1000);
        }

        dates = dates.sort();

        const result = { from: dates[0] };
        if (dates[1]) {
          result.to = dates[1];
        }
        filters[key] = result;
        continue;
      }

      filters[key] = filter.value[key];
    }
  }

  if (searchParam.value) {
    filters.search_param = searchParam.value;
  }

  return filters;
});

const listOptions = computed(() => ({
  filters: promocodesFilters.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));

const setOptions = (newOptions) => {
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const fetchPromocodes = async () => {
  init();

  isFetchLoading.value = true;
  fetchError.value = "";
  try {
    await store.dispatch("promocodes/fetch", listOptions.value);
  } catch (e) {
    fetchError.value = e.message;
  } finally {
    isFetchLoading.value = false;
  }
};

const init = async () => {
  isCountLoading.value = true;
  try {
    const { total } = await store.dispatch("promocodes/count", {
      filters: listOptions.value.filters,
    });
    count.value = Number(total);
  } finally {
    isCountLoading.value = false;
  }
};

const fetchPromocodesDebounce = debounce(fetchPromocodes, 300);

const copyCode = (item) => {
  addToClipboard(item.code);
};

watch(
  promocodesFilters,
  () => {
    page.value = 1;
    fetchPromocodesDebounce();
  },
  { deep: true }
);
watch(options, fetchPromocodesDebounce);
watch(refetch, fetchPromocodesDebounce);

watch([searchFields, noSearch], () => {
  if (!noSearch.value) {
    store.commit("appSearch/setFields", searchFields.value);
  }
});
</script>

<script>
export default {
  name: "promocodes-table",
};
</script>
