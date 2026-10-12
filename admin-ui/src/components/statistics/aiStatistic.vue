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
      :not-comparable="false"
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
          :stacked="fields !== 'revenue'"
          :value-formatter="valueFormatter"
          :tooltip-formatter="tooltipFormatter"
          :legend-formatter="fields !== 'revenue' ? legendFormatter : null"
          :on-point-click="onColumnClick"
        />
      </template>

      <template v-slot:options>
        <div style="display: flex; flex-wrap: wrap; width: 610px">
          <div style="display: flex">
            <v-autocomplete
              style="width: 480px; margin-right: 15px"
              item-text="label"
              item-value="value"
              :loading="isAccountsLoading"
              label="Accounts"
              clearable
              multiple
              hide-details
              :items="!isAccountsLoading ? [...fullAccounts.values()] : []"
              v-model="selectedAccounts"
            />

            <v-autocomplete
              style="width: 120px; margin-right: 15px"
              item-text="label"
              item-value="value"
              placeholder="All"
              label="Agents"
              clearable
              hide-details
              :items="agentTypes"
              v-model="agentType"
            />
          </div>
          <div style="display: flex">
            <v-autocomplete
              style="width: 300px; margin-right: 15px"
              item-text="label"
              item-value="value"
              label="Models"
              clearable
              multiple
              hide-details
              chips
              small-chips
              deletable-chips
              :items="
                selectedProviders.length
                  ? currentModels
                  : [...allModels.values()]
              "
              v-model="selectedModels"
            />

            <v-autocomplete
              style="width: 300px; margin-right: 15px"
              item-text="label"
              item-value="value"
              label="Providers"
              clearable
              multiple
              chips
              deletable-chips
              small-chips
              hide-details
              :items="
                Object.keys(modelsByProviders).map((a) => ({
                  label: a,
                  modelsByProviders: a,
                }))
              "
              v-model="selectedProviders"
            />
          </div>
        </div>
      </template>
    </statistic-item>

    <v-dialog v-model="isSelectedPointOpen" max-width="50vw">
      <v-card
        v-for="point in selectedDataPoints"
        :key="point.series"
        class="data_point_menu"
      >
        <v-card-title>
          <router-link
            v-if="fields === 'accounts'"
            :to="{ name: 'Account', params: { accountId: point?.series } }"
          >
            {{ getSeriesName(point?.series) }}
          </router-link>

          <span v-else>
            {{ getSeriesName(point?.series) }}
          </span>

          {{ `: ${point?.category} - ${point?.value} ${defaultCurrency.code}` }}
        </v-card-title>

        <template v-if="fields !== 'accounts'">
          <v-card-subtitle class="subtitle">Accounts:</v-card-subtitle>
          <v-list class="list">
            <v-list-item
              class="list_item"
              v-for="account of point?.meta?.accounts || []"
              :key="account"
            >
              <v-list-item-title>
                <a @click="openAccount(account)">{{
                  fullAccounts.get(account)?.label || account
                }}</a>
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </template>

        <template v-if="fields !== 'models'">
          <v-card-subtitle class="subtitle">Models:</v-card-subtitle>
          <v-list class="list">
            <v-list-item
              class="list_item"
              v-for="model of point?.meta?.models || []"
              :key="model"
            >
              <v-list-item-title>
                {{ model }}
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </template>

        <v-card-subtitle class="subtitle">Agents:</v-card-subtitle>
        <v-list class="list">
          <v-list-item
            class="list_item"
            v-for="agent of point?.meta?.agents || []"
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
import { computed, onMounted, ref, shallowRef, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";
import { formatWarsawDate } from "@/functions";
import api from "@/api";
import router from "../../router";
import useCurrency from "@/hooks/useCurrency";
import {
  statisticPeriodProps,
  statisticPeriodEmits,
  bucketLabels,
  limitTop,
} from "@/hooks/useStatisticChart";

const store = useStore();

const props = defineProps(statisticPeriodProps);
const { period, periodType, periods, type } = toRefs(props);

const emit = defineEmits(statisticPeriodEmits);

const { defaultCurrency } = useCurrency();

const allFields = ref([
  { label: "Revenue", value: "revenue" },
  { label: "By Accounts", value: "accounts" },
  { label: "By Models", value: "models" },
  { label: "By Providers", value: "providers" },
]);
const fields = ref("revenue");

const comparable = ref(false);

const agentTypes = ref([
  { label: "Bots", value: "nocloud_bots" },
  { label: "Chats", value: "chats" },
  { label: "API", value: "" },
  { label: "All", value: "all" },
]);
const agentType = ref("all");

const allModels = shallowRef(new Set());
const selectedModels = ref([]);

const isAccountsLoading = ref(false);
const allAccounts = ref([]);
const fullAccounts = shallowRef(new Map());
const selectedAccounts = ref([]);

const selectedProviders = ref([]);

const series = shallowRef([]);
const categories = shallowRef([]);
const summary = shallowRef({});

const isDataLoading = ref(false);
const chartData = shallowRef();

const aiConfig = ref({ models: [] });

const isSelectedPointOpen = ref(false);
const selectedDataPoints = ref([]);

const processedDataCache = new Map();
const CACHE_MAX_SIZE = 3;

function getCacheKey(params) {
  return JSON.stringify({
    periodType: params.periodType,
    periods: params.periods,
    fields: params.fields,
    models: params.models?.sort(),
    accounts: params.accounts?.sort(),
    agents: params.agents?.sort(),
  });
}

function getCachedProcessedData(key) {
  return processedDataCache.get(key) || null;
}

function setCachedProcessedData(key, data) {
  if (processedDataCache.size >= CACHE_MAX_SIZE) {
    const firstKey = processedDataCache.keys().next().value;
    processedDataCache.delete(firstKey);
  }

  processedDataCache.set(key, data);
}

function getSeriesName(series) {
  if (fields.value === "accounts") {
    return fullAccounts.value.get(series)?.label || series;
  }

  return series;
}

onMounted(async () => {
  try {
    const response = await api.get("/api/openai/get_config");
    aiConfig.value = response.cfg;
  } catch (e) {
    console.log(e);
  }
});

const modelsByProviders = computed(() => {
  const result = {};

  Object.keys(aiConfig.value.models).map((key) => {
    const model = aiConfig.value.models[key];

    if (!result[model.provider]) {
      result[model.provider] = [];
    }

    result[model.provider].push(key);
  });
  return result;
});

const providersModels = computed(() => {
  return selectedProviders.value.reduce((acc, key) => {
    acc.push(...modelsByProviders.value[key]);
    return acc;
  }, []);
});

const currentModels = computed(() => {
  const accepted = [...providersModels.value];

  return [...allModels.value.values()].filter((model) =>
    accepted.includes(model.value),
  );
});

const valueFormatter = (val) => val?.toFixed?.(2) ?? val;

const legendFormatter = (name) => {
  const total = summary.value[name];
  const label =
    fields.value === "accounts" ? fullAccounts.value.get(name)?.label || name : name;
  return `${label} ${total} ${defaultCurrency.value.code}`;
};

const tooltipFormatter = (params) => {
  const list = Array.isArray(params) ? params : [params];
  const dataPointIndex = list[0]?.dataIndex;
  const category = list[0]?.axisValueLabel ?? categories.value[dataPointIndex];

  const defaultCurrencyCode = defaultCurrency.value.code;
  const fullAccountsMap = fullAccounts.value;

  const rows = series.value
    .map((serie, index) => ({ ...serie, index }))
    .sort((a, b) => (b.data?.[dataPointIndex] || 0) - (a.data?.[dataPointIndex] || 0))
    .slice(0, 5);

  let html = `<div style="font-weight:600;margin-bottom:4px;">Date: ${category}</div>`;

  for (const row of rows) {
    const val = row.data?.[dataPointIndex];
    if (val == null) continue;

    const color = list.find((p) => p.seriesIndex === row.index)?.color;
    const meta = row.meta?.[dataPointIndex] || {};
    const { accounts = [], agents = [], models = [] } = meta;

    let seriesName = row.name;
    let accountLinks = accounts
      .map((id) => {
        const label = fullAccountsMap.get(id)?.label || id;
        return `<a href="${window.location.origin}/admin/accounts/${id}">${label}</a>`;
      })
      .join(", ");

    const agentsList = agents.join(", ");
    let modelsList = models.join(", ");

    if (fields.value === "accounts") {
      seriesName = fullAccounts.value.get(seriesName)?.label || seriesName;
      accountLinks = null;
    } else if (fields.value === "models") {
      modelsList = null;
    }

    html += `
        <div style="margin: 6px 0;">
          <div style="display: flex; align-items: center; gap: 6px;">
            <span style="background:${color};width:10px;height:10px;border-radius:50%;display:inline-block;"></span>
            <span><strong>${seriesName}:</strong> ${val} ${defaultCurrencyCode}</span>
          </div>
          <div style="padding-left: 16px; font-size: 12px; white-space: normal; word-wrap: break-word;">
            ${accountLinks ? `Accounts: ${accountLinks}<br/>` : ""}
            ${modelsList ? `Models: ${modelsList}<br/>` : ""}
            ${agentsList ? `Agents: ${agentsList}<br/>` : ""}
          </div>
        </div>
      `;
  }

  return html;
};

async function fetchData() {
  isDataLoading.value = true;

  try {
    chartData.value = await store.dispatch(
      "statistic/getForChart",
      {
        entity: "ai/transactions",
        periodType: periodType.value,
        periods: !comparable.value
          ? [period.value]
          : [periods.value.first, periods.value.second],
        params: {
          raw: false,
          group_by:
            fields.value === "accounts"
              ? "account"
              : fields.value === "models" || fields.value === "providers"
              ? "model"
              : undefined,
          models: selectedModels.value.length
            ? selectedModels.value
            : providersModels.value.length
            ? providersModels.value
            : undefined,
          accounts: selectedAccounts.value.length
            ? selectedAccounts.value
            : undefined,
          agents:
            agentType.value !== "all"
              ? [agentType.value]
              : agentTypes.value.map((a) => a.value).filter((a) => a !== "all"),
        },
        withCache: false,
      },
    );
  } finally {
    isDataLoading.value = false;
  }
}

function getValue(value) {
  const num = Number(value);
  if (isNaN(num)) return NaN;

  const absNum = Math.abs(num);

  const precision = absNum < 1 ? 4 : 2;
  return Number(num.toFixed(precision));
}

const onColumnClick = ({ dataIndex: dataPointIndex }) => {
  const points = series.value
    .map((serie) => ({
      series: serie.name,
      meta: serie.meta[dataPointIndex],
      value: serie.data[dataPointIndex],
      category: categories.value[dataPointIndex],
    }))
    .filter((v) => !!v.value);

  points.sort((a, b) => b.value - a.value);

  selectedDataPoints.value = points;

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

  if (accountsToFetch.length === 0) {
    return;
  }

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
    }),
  );

  isAccountsLoading.value = false;
}

const fetchAccountsDebounced = debounce(fetchAccounts, 200);

watch(
  [
    period,
    periods,
    comparable,
    agentType,
    selectedModels,
    selectedAccounts,
    selectedProviders,
    fields,
  ],
  () => {
    if (!chartData.value) {
      fetchData();
    } else {
      fetchDataDebounced();
    }
  },
);

watch([chartData, fields], () => {
  if (!chartData.value || !fields.value) {
    return;
  }

  const cacheKey = getCacheKey({
    periodType: periodType.value,
    periods: !comparable.value
      ? [period.value]
      : [periods.value.first, periods.value.second],
    fields: fields.value,
    models: selectedModels.value.length
      ? selectedModels.value
      : providersModels.value.length
      ? providersModels.value
      : undefined,
    accounts: selectedAccounts.value.length
      ? selectedAccounts.value
      : undefined,
    agents:
      agentType.value !== "all"
        ? [agentType.value]
        : agentTypes.value.map((a) => a.value).filter((a) => a !== "all"),
  });

  const cached = getCachedProcessedData(cacheKey);
  if (cached) {
    series.value = cached.series;
    categories.value = cached.categories;
    summary.value = cached.summary;
    return;
  }

  const newSeries = [];
  const newCategories = [];
  const newSummary = {};

  const tempData = structuredClone
    ? structuredClone(chartData.value)
    : JSON.parse(JSON.stringify(chartData.value));

  if (fields.value === "revenue") {
    const ranges = comparable.value ? periods.value : { first: period.value };
    const labels = bucketLabels(ranges.first[0], ranges.first[1], periodType.value);
    Object.keys(ranges).forEach((key, seriesIndex) => {
      const rangeLabels = bucketLabels(ranges[key][0], ranges[key][1], periodType.value);
      const byDay = {};
      (tempData[seriesIndex]?.timeseries || []).forEach((row) => {
        byDay[String(row.ts).split("T")[0]] = row;
      });
      newSeries.push({
        name: `${formatWarsawDate(ranges[key][0])}/${formatWarsawDate(ranges[key][1])}`,
        data: labels.map((_, index) => getValue(byDay[rangeLabels[index]]?.revenue || 0)),
        meta: labels.map((_, index) => byDay[rangeLabels[index]] || {}),
      });
    });
    newCategories.push(...labels);
    newSeries.forEach((serie) => {
      newSummary[serie.name] = getValue(
        serie.data.reduce((acc, value) => acc + (Number(value) || 0), 0) || 0,
      );
    });
  } else {
    const labels = bucketLabels(period.value[0], period.value[1], periodType.value);
    const blocks = comparable.value ? tempData : [tempData[0]];
    const maps = blocks.map((block) => {
      const map = {};
      (block?.timeseries || []).forEach((row) => {
        const day = String(row.ts).split("T")[0];
        let dataKey = row.key || row.model || row.account || "unknown";
        if (fields.value === "providers") {
          dataKey =
            Object.keys(modelsByProviders.value).find((provider) =>
              (modelsByProviders.value[provider] || []).includes(row.key || row.model),
            ) || "unknown";
        }
        map[dataKey] = map[dataKey] || {};
        map[dataKey][day] = (map[dataKey][day] || 0) + Number(row.revenue || 0);
      });
      return map;
    });
    const names = new Set();
    Object.keys(maps[0] || {}).forEach((name) => names.add(name));
    let built = [...names].map((name) => ({
      name,
      data: labels.map((label) => getValue(maps[0][name]?.[label] || 0)),
      meta: labels.map(() => ({})),
    }));
    const builtSummary = {};
    built.forEach((serie) => {
      builtSummary[serie.name] = getValue(
        serie.data.reduce((acc, value) => acc + (Number(value) || 0), 0) || 0,
      );
    });
    const limited = limitTop(built, builtSummary, 10);
    limited.series.forEach((serie) => {
      if (!serie.meta) serie.meta = (serie.data || []).map(() => ({}));
    });
    newSeries.push(...limited.series);
    Object.assign(newSummary, limited.summary);
    if (comparable.value && maps[1]) {
      const prevLabels = bucketLabels(
        periods.value.second[0],
        periods.value.second[1],
        periodType.value,
      );
      limited.series.forEach((serie) => {
        if (serie.name === "Other") return;
        const data = labels.map((_, index) =>
          getValue(maps[1][serie.name]?.[prevLabels[index]] || 0),
        );
        newSeries.push({
          name: `${serie.name} prev`,
          data,
          meta: labels.map(() => ({})),
        });
        newSummary[`${serie.name} prev`] = getValue(
          data.reduce((acc, value) => acc + (Number(value) || 0), 0) || 0,
        );
      });
    }
    newCategories.push(...labels);
  }

  const accounts = [];
  tempData.forEach((data) => {
    data.summary?.models?.forEach((model) =>
      allModels.value.add({ label: model, value: model }),
    );

    data.summary?.accounts?.forEach((account) => accounts.push(account));
  });

  allAccounts.value = [
    ...new Set([...allAccounts.value, ...accounts]).values(),
  ];

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
  summary.value = newSummary;

  setCachedProcessedData(cacheKey, {
    series: newSeries,
    categories: newCategories.map((c) => c.toString().split("T")[0]),
    summary: newSummary,
  });
});

watch(
  allAccounts,
  (accounts) => {
    fetchAccountsDebounced(accounts);
  },
  { deep: true },
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
