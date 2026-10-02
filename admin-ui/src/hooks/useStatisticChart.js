import { ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce, formatWarsawDate } from "@/functions";

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
    name: `${formatWarsawDate(periods[key][0])}/${formatWarsawDate(
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

export function buildSeriesFromTimeseries(timeseries, field, getValue, label) {
  const getter = getValue || ((row, key) => row?.[key] || 0);
  const data = (timeseries || []).map((row) => getter(row, field));
  const name = label || field;
  return {
    series: [{ name, data }],
    categories: (timeseries || []).map((row) => row.ts),
    summary: { [name]: data.reduce((acc, value) => acc + (Number(value) || 0), 0) },
  };
}

export function buildGroupedSeries(timeseries, nameField, valueField, getValue) {
  const getter = getValue || ((row, key) => row?.[key] || 0);
  const names = [];
  const categories = [];
  (timeseries || []).forEach((row) => {
    const name = row[nameField];
    if (name != null && name !== "" && !names.includes(name)) names.push(name);
    if (!categories.includes(row.ts)) categories.push(row.ts);
  });
  const series = names.map((name) => ({
    name,
    data: categories.map((ts) => {
      const row = (timeseries || []).find(
        (item) => item.ts === ts && item[nameField] === name
      );
      return row ? getter(row, valueField) : 0;
    }),
  }));
  const summary = {};
  series.forEach((serie) => {
    summary[serie.name] = serie.data.reduce((acc, value) => acc + (Number(value) || 0), 0);
  });
  return { series, categories, summary };
}

export function limitTop(series, summary, limit) {
  const ranked = [...series].sort(
    (a, b) => (summary[b.name] || 0) - (summary[a.name] || 0)
  );
  if (ranked.length <= limit) {
    return { series: ranked, summary };
  }
  const top = ranked.slice(0, limit);
  const rest = ranked.slice(limit);
  const other = {
    name: "Other",
    data: (top[0]?.data || []).map((_, index) =>
      rest.reduce((acc, serie) => acc + (Number(serie.data[index]) || 0), 0)
    ),
  };
  const next = {};
  top.forEach((serie) => {
    next[serie.name] = summary[serie.name] || 0;
  });
  next.Other = rest.reduce((acc, serie) => acc + (summary[serie.name] || 0), 0);
  return { series: [...top, other], summary: next };
}

export function bucketLabels(start, end, periodType) {
  if (!start || !end) return [];
  const startText = formatWarsawDate(start);
  const endText = formatWarsawDate(end);
  const today = formatWarsawDate(new Date());
  if (periodType === "year-1_month") {
    const year = Number(startText.slice(0, 4));
    const lastMonth = year === Number(today.slice(0, 4)) ? Number(today.slice(5, 7)) : 12;
    return Array.from({ length: lastMonth }, (_, index) =>
      `${year}-${String(index + 1).padStart(2, "0")}-01`
    );
  }
  if (periodType === "years-1_year") {
    const from = Number(startText.slice(0, 4));
    const to = Math.min(Number(endText.slice(0, 4)), Number(today.slice(0, 4)));
    const labels = [];
    for (let year = from; year <= to; year++) labels.push(`${year}-01-01`);
    return labels;
  }
  const capped = endText > today ? today : endText;
  const labels = [];
  let cursor = Date.parse(`${startText}T12:00:00Z`);
  const last = Date.parse(`${capped}T12:00:00Z`);
  while (cursor <= last) {
    labels.push(new Date(cursor).toISOString().slice(0, 10));
    cursor += 86400000;
  }
  return labels;
}
