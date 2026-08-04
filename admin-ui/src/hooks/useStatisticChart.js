import { ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce, formatToYYMMDD } from "@/functions";

export const statisticPeriodProps = {
  period: { type: Array, default: () => [] },
  periodType: { type: String, default: "month" },
  type: { type: String, default: "bar" },
  periods: { type: Object, default: () => ({ first: [], second: [] }) },
  periodOffset: { type: Number, default: 0 },
  periodsFirstOffset: { type: Number, default: 0 },
  periodsSecondOffset: { type: Number, default: -1 },
};

export const statisticPeriodEmits = [
  "update:period",
  "update:periods",
  "update:period-type",
  "update:type",
  "update:period-offset",
  "update:periods-first-offset",
  "update:periods-second-offset",
];

// Fetches a `statistic/getForChart` series, debouncing repeated calls while
// keeping the very first load immediate. `entity`/`params` may be plain
// values or refs; `extraDeps` are additional refs to re-fetch on.
export function useStatisticData({
  entity,
  periodType,
  period,
  periods,
  comparable,
  params,
  extraDeps = [],
}) {
  const store = useStore();
  const chartData = ref(null);
  const isDataLoading = ref(false);

  async function fetchData() {
    isDataLoading.value = true;

    try {
      chartData.value = await store.dispatch("statistic/getForChart", {
        entity: entity?.value ?? entity,
        periodType: periodType.value,
        periods: !comparable.value
          ? [period.value]
          : [periods.value.first, periods.value.second],
        params: params?.value,
      });
    } finally {
      isDataLoading.value = false;
    }
  }

  const fetchDataDebounced = debounce(fetchData, 1000);

  debounce(fetchData, 100)();

  watch([period, periods, comparable, ...extraDeps], () => {
    if (!chartData.value) {
      fetchData();
    } else {
      fetchDataDebounced();
    }
  });

  return { chartData, isDataLoading, fetchData };
}

// Builds the two-series "compare period A vs period B" chart shape shared by
// every statistic widget: one series per period, aligned by index (day 1 of
// period A next to day 1 of period B, etc), plus a value-summed summary.
export function buildComparablePeriodSeries(
  tempData,
  periods,
  fieldKey,
  getValue = (row, key) => row?.[key] || 0
) {
  const series = Object.keys(periods).map((key) => ({
    name: `${formatToYYMMDD(periods[key][0])}/${formatToYYMMDD(
      periods[key][1]
    )}`,
    data: [],
  }));
  const categories = [];

  const length = Math.max(
    tempData[0]?.timeseries?.length || 0,
    tempData[1]?.timeseries?.length || 0
  );

  for (let index = 0; index < length; index++) {
    const first = tempData[0]?.timeseries?.[index];
    const second = tempData[1]?.timeseries?.[index];

    if (!categories.includes(index + 1)) categories.push(index + 1);

    series[0].data.push(getValue(first, fieldKey));
    series[1].data.push(getValue(second, fieldKey));
  }

  const summary = {};
  series.forEach((serie) => {
    summary[serie.name] = serie.data.reduce((acc, v) => acc + (v || 0), 0);
  });

  return { series, categories, summary };
}
