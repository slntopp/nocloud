<template>
  <div>
    <statistic-item
      :period="period"
      :periodType="periodType"
      :periods="periods"
      :type="type"
      :period-offset="periodOffset"
      :periods-first-offset="periodsFirstOffset"
      :periods-second-offset="periodsSecondOffset"
      @input:period="emit('update:period', $event)"
      @input:period-type="emit('update:period-type', $event)"
      @input:periods="emit('update:periods', $event)"
      @input:type="emit('update:type', $event)"
      @input:period-offset="emit('update:period-offset', $event)"
      @input:periods-first-offset="emit('update:periods-first-offset', $event)"
      @input:periods-second-offset="
        emit('update:periods-second-offset', $event)
      "
      :loading="isDataLoading"
      :all-fields="allFields"
      :fields="fields"
      @input:fields="fields = $event"
      :comparable="comparable"
      @input:comparable="comparable = $event"
    >
      <template v-slot:content>
        <default-chart
          description="AI statistics"
          :type="type"
          :series="series"
          :categories="categories"
          :summary="summary"
          :options="chartOptions"
        />
      </template>

      <template v-slot:options>
        <v-select
          style="width: 250px; margin-left: 5px"
          item-text="label"
          item-value="value"
          :loading="isAccountsLoading"
          label="Accounts"
          clearable
          multiple
          :items="!isAccountsLoading ? [...fullAccounts.values()] : []"
          v-model="selectedAccounts"
        />

        <v-select
          style="width: 250px; margin-left: 5px"
          item-text="label"
          item-value="value"
          label="Models"
          clearable
          multiple
          :items="[...allModels.values()]"
          v-model="selectedModels"
        />

        <v-select
          style="width: 120px; margin-left: 5px"
          item-text="label"
          item-value="value"
          placeholder="All"
          label="Agents"
          clearable
          :items="agentTypes"
          v-model="agentType"
        />
      </template>
    </statistic-item>

    <v-dialog v-model="isSelectedPointOpen" max-width="50vw">
      <v-card class="data_point_menu">
        <v-card-title>{{
          `${selectedDataPoint?.series}: ${selectedDataPoint?.category} - ${selectedDataPoint?.value}`
        }}</v-card-title>
        <v-card-subtitle class="subtitle">Accounts:</v-card-subtitle>
        <v-list class="list">
          <v-list-item
            class="list_item"
            v-for="account of selectedDataPoint?.meta.accounts || []"
            :key="account"
          >
            <v-list-item-title>
              <a @click="openAccount(account)">{{
                fullAccounts.get(account)?.label || account
              }}</a>
            </v-list-item-title>
          </v-list-item>
        </v-list>
        <v-card-subtitle class="subtitle">Models:</v-card-subtitle>
        <v-list class="list">
          <v-list-item
            class="list_item"
            v-for="model of selectedDataPoint?.meta.models || []"
            :key="model"
          >
            <v-list-item-title>
              {{ model }}
            </v-list-item-title>
          </v-list-item>
        </v-list>

        <v-card-subtitle class="subtitle">Agents:</v-card-subtitle>
        <v-list class="list">
          <v-list-item
            class="list_item"
            v-for="agent of selectedDataPoint?.meta.agents || []"
            :key="agent"
          >
            <v-list-item-title>
              {{ agent }}
            </v-list-item-title>
          </v-list-item>
        </v-list>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import StatisticItem from "@/components/statistics/statisticItem.vue";
import { computed, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";
import { formatToYYMMDD } from "@/functions";
import api from "@/api";
import router from "../../router";

const store = useStore();

const props = defineProps({
  period: { type: Array, default: () => [] },
  periodType: { type: String, default: "month" },
  type: { type: String, default: "bar" },
  periods: { type: Object, default: () => ({ first: [], second: [] }) },
  periodOffset: { type: Number, default: 0 },
  periodsFirstOffset: { type: Number, default: 0 },
  periodsSecondOffset: { type: Number, default: -1 },
});
const { period, periodType, periods, type } = toRefs(props);

const emit = defineEmits([
  "update:period",
  "update:periods",
  "update:period-type",
  "update:type",
  "update:period-offset",
  "update:periods-first-offset",
  "update:periods-second-offset",
]);

const allFields = ref([
  { label: "Revenue", value: "revenue" },
  { label: "Count", value: "count" },
]);
const fields = ref("revenue");
const comparable = ref(true);

const agentTypes = ref([
  { label: "Bots", value: "nocloud_bots" },
  { label: "Chats", value: "chats" },
  { label: "API", value: "" },
  { label: "All", value: "all" },
]);
const agentType = ref("all");

const allModels = ref(new Set());
const selectedModels = ref([]);

const isAccountsLoading = ref(false);
const allAccounts = ref([]);
const fullAccounts = ref(new Map());
const selectedAccounts = ref([]);

const series = ref([]);
const categories = ref([]);
const summary = ref({});

const isDataLoading = ref(false);
const chartData = ref();

const isSelectedPointOpen = ref(false);
const selectedDataPoint = ref();

const chartOptions = computed(() => ({
  tooltip: {
    enabled: true,
    shared: true,
    intersect: false,
    theme: store.getters["app/theme"],
    y: {
      formatter: () => "",
      title: {
        formatter: (date) => {
          return `${date} `;
        },
      },
    },
    x: {
      formatter: (val, opts) => {
        const result = [`Date: ${val}`];
        const dataIndex = opts.dataPointIndex;
        const seriesIndex = opts.seriesIndex;
        const { accounts, agents, models } =
          series.value[seriesIndex].meta[dataIndex];

        result.push(
          `Accounts: ${accounts
            .map((id) => ({
              label: fullAccounts.value.get(id)?.label || id,
              id,
            }))
            .map(
              ({ id, label }) =>
                `<a href="${window.location.origin}/admin/accounts/${id}">${label}</a>`
            )}`
        );
        result.push(`Models: ${models.join(", ")}`);

        result.push(`Agents: ${agents.join(", ")}`);

        return result.join("<br/>");
      },
    },
    style: {
      fontSize: "14px",
      fontFamily: "Inter, sans-serif",
    },
  },
  chart: {
    events: {
      dataPointSelection: onColumnClick,
    },
  },
}));

async function fetchData() {
  isDataLoading.value = true;

  try {
    chartData.value = await store.dispatch("statistic/getForChart", {
      entity: "ai/transactions",
      periodType: periodType.value,
      periods: !comparable.value
        ? [period.value]
        : [periods.value.first, periods.value.second],
      params: {
        models: selectedModels.value.length ? selectedModels.value : undefined,
        accounts: selectedAccounts.value.length
          ? selectedAccounts.value
          : undefined,
        agents:
          agentType.value !== "all"
            ? [agentType.value]
            : agentTypes.value.map((a) => a.value).filter((a) => a !== "all"),
      },
    });
  } finally {
    isDataLoading.value = false;
  }
}

function getValue(value) {
  const num = Number(value);
  if (isNaN(num)) return NaN;

  const decimalPart = num.toString().split(".")[1];

  if (decimalPart && decimalPart.length > 2) {
    return Number(num.toFixed(2));
  }

  return num;
}

const onColumnClick = (...args) => {
  const { seriesIndex, dataPointIndex } = args[2];
  selectedDataPoint.value = {
    series: series.value[seriesIndex]?.name,
    meta: series.value[seriesIndex].meta[dataPointIndex],
    value: series.value[seriesIndex].data[dataPointIndex],
    category: categories.value[dataPointIndex],
  };
  isSelectedPointOpen.value = true;
};

const openAccount = (id) => {
  const routeData = router.resolve({
    name: "Account",
    params: { accountId: id },
  });
  window.open(routeData.href, "_blank");
};

const fetchDataDebounced = debounce(fetchData, 1000);

debounce(fetchData, 100)();

async function fetchAccounts(accounts) {
  const accountsToFetch = [];
  accounts.forEach((account) => {
    if (!fullAccounts.value.has(account)) {
      accountsToFetch.push(account);
    }
  });
  isAccountsLoading.value = true;
  await Promise.allSettled(
    accountsToFetch.map(async (uuid) => {
      try {
        fullAccounts.value.set(uuid, api.accounts.get(uuid));
        const account = await fullAccounts.value.get(uuid);

        fullAccounts.value.set(account.uuid, {
          label: account.title,
          value: account.uuid,
        });
      } catch (e) {
        fullAccounts.value.delete(uuid);
      }
    })
  );

  isAccountsLoading.value = false;
}

const fetchAccountsDebounced = debounce(fetchAccounts, 200);

watch(
  [period, periods, comparable, agentType, selectedModels, selectedAccounts],
  () => {
    if (!chartData.value) {
      fetchData();
    } else {
      fetchDataDebounced();
    }
  }
);

watch([chartData, fields], () => {
  if (!chartData.value || !fields.value) {
    return;
  }

  const newSeries = [];
  const newCategories = [];
  summary.value = {};

  const tempData = JSON.parse(JSON.stringify(chartData.value));

  if (!comparable.value) {
    newSeries.push({
      name: allFields.value.find((field) => field.value === fields.value).label,
      data: [],
      meta: [],
      id: fields.value,
    });

    tempData[0].timeseries?.forEach((timeseries) => {
      newCategories.push(timeseries.ts);
      newSeries.forEach((serie) => {
        serie.data.push(getValue(timeseries[serie.id] || 0));
        serie.meta.push({
          accounts: timeseries.accounts || [],
          models: timeseries.models || [],
          agents: timeseries.agents || ["api"],
        });
      });
    });

    newSeries.forEach((serie) => {
      tempData[0].summary?.models?.forEach((model) =>
        allModels.value.add({ label: model, value: model })
      );
      const accounts = [];

      tempData[0].summary?.accounts?.forEach((account) =>
        accounts.push(account)
      );
      allAccounts.value = [
        ...new Set([...allAccounts.value, ...accounts]).values(),
      ];

      summary.value[serie.name] = getValue(
        tempData[0].summary?.[serie.id] || 0
      );
    });
  } else {
    Object.keys(periods.value).forEach((key) => {
      newSeries.push({
        name: `${formatToYYMMDD(periods.value[key][0])}/${formatToYYMMDD(
          periods.value[key][1]
        )}`,
        data: [],
        meta: [],
      });
    });

    for (
      let index = 0;
      index <
      Math.max(
        tempData[0]?.timeseries?.length || 0,
        tempData[1]?.timeseries?.length || 0
      );
      index++
    ) {
      const first = tempData[0]?.timeseries?.[index];
      const second = tempData[1]?.timeseries?.[index];

      if (!newCategories.includes(index + 1)) {
        newCategories.push(index + 1);
      }

      [first, second].forEach((value, index) => {
        newSeries[index].data.push(getValue(value?.[fields.value] || 0));
        newSeries[index].meta.push({
          accounts: value?.accounts || [],
          models: value?.models || [],
          agents: value?.agents || ["api"],
        });
      });
    }

    const accounts = [];
    tempData.forEach((data) => {
      data.summary?.models?.forEach((model) =>
        allModels.value.add({ label: model, value: model })
      );

      data.summary?.accounts?.forEach((account) => accounts.push(account));
    });

    allAccounts.value = [
      ...new Set([...allAccounts.value, ...accounts]).values(),
    ];

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getValue(
        serie.data.reduce((acc, a) => acc + a, 0) || 0
      );
    });
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
});

watch(
  allAccounts,
  (accounts) => {
    fetchAccountsDebounced(accounts);
  },
  { deep: true }
);
</script>

<style lang="scss" scoped>
.data_point_menu {
  padding: 25px;
  .subtitle {
    margin: 0px;
    padding: 0px;
  }
  .list {
    margin: 0px;
    padding: 0px;
  }
  .list_item {
    margin: 0px;
    padding: 0px;
  }
}
</style>
