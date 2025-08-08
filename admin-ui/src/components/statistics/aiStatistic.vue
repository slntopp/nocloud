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
      :not-comparable="fields !== 'revenue'"
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
          :stacked="fields !== 'revenue'"
        />
      </template>

      <template v-slot:options>
        <div style="display: flex; flex-wrap: wrap; width: 610px">
          <div style="display: flex">
            <v-autocomplete
              style="width: 480px; margin-left: 5px"
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
              style="width: 120px; margin-left: 5px"
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
              style="width: 300px; margin-left: 5px"
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
              style="width: 300px; margin-left: 5px"
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
        <v-card-title>{{
          `${point?.series}: ${point?.category} - ${point?.value} ${defaultCurrency.code}`
        }}</v-card-title>

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
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";
import { formatToYYMMDD } from "@/functions";
import api from "@/api";
import router from "../../router";
import useCurrency from "@/hooks/useCurrency";

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

const { defaultCurrency } = useCurrency();

const allFields = ref([
  { label: "Revenue", value: "revenue" },
  { label: "By Accounts", value: "accounts" },
  { label: "By Models", value: "models" },
  { label: "By Providers", value: "providers" },
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

const selectedProviders = ref([]);

const series = ref([]);
const categories = ref([]);
const summary = ref({});

const isDataLoading = ref(false);
const chartData = ref();

const aiConfig = ref({ models: [] });

const isSelectedPointOpen = ref(false);
const selectedDataPoints = ref([]);

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
    accepted.includes(model.value)
  );
});

const chartOptions = computed(() => {
  const options = {
    tooltip: {
      enabled: true,
      shared: true,
      intersect: false,
      theme: store.getters["app/theme"],
      custom: function ({ dataPointIndex, w }) {
        const defaultCurrencyCode = defaultCurrency.value.code;
        const fullAccountsMap = fullAccounts.value;
        const categories = w.globals.labels;
        const category = categories[dataPointIndex];

        let html = `<div class="apexcharts-tooltip-title">Date: ${category}</div>`;
        for (let i = 0; i < series.value.length; i++) {
          const val = series.value[i].data?.[dataPointIndex];
          if (val == null) continue;

          let seriesName = w.globals.seriesNames[i];

          const color = w.globals.colors[i];

          const meta = series.value?.[i]?.meta?.[dataPointIndex] || {};

          const { accounts = [], agents = [], models = [] } = meta;

          let accountLinks = accounts
            .map((id) => {
              const label = fullAccountsMap.get(id)?.label || id;
              return `<a href="${window.location.origin}/admin/accounts/${id}">${label}</a>`;
            })
            .join(", ");

          const agentsList = agents.join(", ");
          let modelsList = models.join(", ");

          if (fields.value === "accounts") {
            seriesName =
              fullAccounts.value.get(seriesName)?.label || seriesName;
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
      },
      style: {
        fontSize: "14px",
        fontFamily: "Inter, sans-serif",
      },
    },

    yaxis: {
      labels: {
        formatter: function (val) {
          return val?.toFixed(2);
        },
      },
    },
    chart: {
      events: {
        dataPointSelection: onColumnClick,
      },
    },
  };

  if (fields.value !== "revenue") {
    options.legend = {
      formatter: (val) => {
        const total = summary.value[val];

        if (fields.value === "accounts") {
          return `${fullAccounts.value.get(val)?.label || val} ${total} ${
            defaultCurrency.value.code
          }`;
        }

        return `${val} ${total} ${defaultCurrency.value.code}`;
      },
    };
  }

  return options;
});

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
        raw: fields.value !== "revenue",
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
    });
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

const onColumnClick = (...args) => {
  const { dataPointIndex } = args[2];
  selectedDataPoints.value = series.value
    .map((serie) => ({
      series: serie.name,
      meta: serie.meta[dataPointIndex],
      value: serie.data[dataPointIndex],
      category: categories.value[dataPointIndex],
    }))
    .filter((v) => !!v.value);

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
  }
);

watch(fields, () => {
  if (fields.value !== "revenue") {
    comparable.value = false;
  }
});

watch([chartData, fields], () => {
  if (!chartData.value || !fields.value) {
    return;
  }

  const newSeries = [];
  const newCategories = [];
  summary.value = {};

  const tempData = JSON.parse(JSON.stringify(chartData.value));

  if (fields.value === "revenue") {
    Object.keys(
      comparable.value ? periods.value : { first: period.value }
    ).forEach((key) => {
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

      if (!tempData[1]) {
        newCategories.push(first.ts);
      } else if (!newCategories.includes(index + 1)) {
        newCategories.push(index + 1);
      }

      [first, second].forEach((value, index) => {
        if (!value) {
          return;
        }
        newSeries[index].data.push(getValue(value?.[fields.value] || 0));
        newSeries[index].meta.push({
          accounts: value?.accounts || [],
          models: value?.models || [],
          agents: value?.agents || ["api"],
        });
      });
    }

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getValue(
        serie.data.reduce((acc, a) => acc + a, 0) || 0
      );
    });
  } else {
    const dataMap = {};
    tempData[0].timeseries.forEach((timeserie) => {
      const ts = timeserie.ts.split("T")[0];

      let dataKey = "";

      if (fields.value === "accounts") {
        dataKey = timeserie.account;
      } else if (fields.value === "models") {
        dataKey = timeserie.model;
      } else if (fields.value === "providers") {
        dataKey = Object.keys(modelsByProviders.value).find((provider) =>
          modelsByProviders.value[provider].includes(timeserie.model)
        );
      }

      if (!dataMap[dataKey]) {
        dataMap[dataKey] = [];
      }

      dataMap[dataKey][ts] = {
        revenue: (dataMap[dataKey]?.[ts]?.revenue || 0) + timeserie.revenue,
        accounts:
          dataMap[dataKey]?.[ts]?.accounts &&
          !dataMap[dataKey]?.[ts]?.accounts.includes(timeserie.account)
            ? [...dataMap[dataKey]?.[ts]?.accounts, timeserie.account]
            : dataMap[dataKey]?.[ts]?.accounts || [],
        models:
          dataMap[dataKey]?.[ts]?.models &&
          !dataMap[dataKey]?.[ts]?.models.includes(timeserie.model)
            ? [...dataMap[dataKey]?.[ts]?.models, timeserie.model]
            : dataMap[dataKey]?.[ts]?.models || [],
        agents:
          dataMap[dataKey]?.[ts]?.agents &&
          !dataMap[dataKey]?.[ts]?.agents.includes(timeserie.agent || "api")
            ? [...dataMap[dataKey]?.[ts]?.agents, timeserie.agent || "api"]
            : dataMap[dataKey]?.[ts]?.agents || [],
      };
    });

    Object.keys(dataMap).forEach((account) => {
      newSeries.push({ name: account, data: [], meta: [] });

      const data = Array(period.value[1].getDate() - 1).fill(null);
      const meta = Array(period.value[1].getDate() - 1).fill(null);

      Object.keys(dataMap[account]).forEach((key) => {
        const date = new Date(key);
        data[date.getDate() - 1] = getValue(dataMap[account][key]?.revenue);

        meta[date.getDate() - 1] = {
          accounts: dataMap[account][key]?.accounts,
          models: dataMap[account][key]?.models,
          agents: dataMap[account][key]?.agents,
        };
      });

      newSeries[newSeries.length - 1].data = data;
      newSeries[newSeries.length - 1].meta = meta;
    });

    const current = new Date(period.value[0]);
    const end = new Date(period.value[1]);

    while (current <= end) {
      const formatted = current.toISOString().slice(0, 10);
      newCategories.push(formatted);
      current.setDate(current.getDate() + 1);
    }

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getValue(
        Object.keys(dataMap[serie.name]).reduce(
          (acc, key) => acc + dataMap[serie.name][key]?.revenue,
          0
        ) || 0
      );
    });

    console.log(summary.value);
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
