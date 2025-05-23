<template>
  <div class="chart_container">
    <div class="chart_options">
      <div class="d-flex align-center">
        <template v-if="!comparable">
          <span
            class="period_title mr-2"
            :style="`background-color: ${colors[0]};`"
          >
          </span>

          <div class="current_duration">
            <v-btn
              x-small
              @click="emit('input:period-offset', periodOffset - 1)"
            >
              <v-icon>mdi-minus</v-icon>
            </v-btn>
            <span class="current_duration_info"
              >{{ formatDate(period[0]) }} - {{ formatDate(period[1]) }}</span
            >
            <v-btn
              x-small
              @click="emit('input:period-offset', periodOffset + 1)"
            >
              <v-icon>mdi-plus</v-icon>
            </v-btn>
          </div>
        </template>

        <div class="current_periods" v-else>
          <div class="d-flex align-center">
            <span
              class="period_title mr-2"
              :style="`background-color: ${colors[0]};`"
            >
            </span>
            <div class="current_duration">
              <v-btn
                x-small
                @click="
                  emit('input:periods-first-offset', periodsFirstOffset - 1)
                "
              >
                <v-icon>mdi-minus</v-icon>
              </v-btn>
              <span class="current_duration_info"
                >{{ formatDate(periods.first[0]) }} -
                {{ formatDate(periods.first[1]) }}</span
              >
              <v-btn
                x-small
                @click="
                  emit('input:periods-first-offset', periodsFirstOffset + 1)
                "
              >
                <v-icon>mdi-plus</v-icon>
              </v-btn>
            </div>
          </div>
          <div class="d-flex my-2 align-center">
            <span
              class="period_title mr-2"
              :style="`background-color: ${colors[1]};`"
            >
            </span>

            <div class="current_duration">
              <v-btn
                x-small
                @click="
                  emit('input:periods-second-offset', periodsSecondOffset - 1)
                "
              >
                <v-icon>mdi-minus</v-icon>
              </v-btn>
              <span class="current_duration_info"
                >{{ formatDate(periods.second[0]) }} -
                {{ formatDate(periods.second[1]) }}</span
              >
              <v-btn
                x-small
                @click="
                  emit('input:periods-second-offset', periodsSecondOffset + 1)
                "
              >
                <v-icon>mdi-plus</v-icon>
              </v-btn>
            </div>
          </div>
        </div>
      </div>

      <div class="d-flex alingn-center">
        <v-switch
          v-if="periods && !notComparable"
          class="ml-2 mr-2"
          label="Comparison"
          :input-value="comparable"
          @change="emit('input:comparable', $event)"
        />
        <slot name="options" />
        <v-select
          v-if="allFields.length"
          class="ml-2"
          style="max-width: 240px"
          label="Fields"
          :multiple="fieldsMultiple"
          :value="fields"
          @input="emit('input:fields', $event)"
          :items="allFields"
          item-text="label"
          item-value="value"
        >
          <template v-slot:selection="{ item, index }">
            <span v-if="index === 0">{{ item.label }}</span>
            <span v-if="index === 1" class="grey--text text-caption">
              (+{{ fields?.length - 1 }} others)
            </span>
          </template>
        </v-select>
        <v-select
          class="ml-2"
          style="width: 150px"
          :value="periodType"
          @input="emit('input:period-type', $event)"
          :items="durationOptions"
          item-text="label"
          item-value="value"
        />
        <v-select
          class="ml-2"
          style="width: 75px"
          :value="type"
          @input="emit('input:type', $event)"
          :items="typeOptions"
          item-text="label"
          item-value="value"
        />
      </div>
    </div>
    <div class="chart">
      <slot v-if="!loading" name="content" />
      <v-skeleton-loader
        v-else
        class="mx-auto pa-5"
        width="100%"
        height="600"
        type="image"
      ></v-skeleton-loader>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { toRefs, watch } from "vue";
import { getApexChartsColors } from "@/functions";

const props = defineProps({
  loading: { type: Boolean, default: false },
  comparable: { type: Boolean, default: false },
  notComparable: { type: Boolean, default: false },
  fieldsMultiple: { type: Boolean, default: false },
  period: { type: Array, default: () => [] },
  periodType: { type: String, default: "month" },
  description: { type: String, default: "" },
  type: { type: String, default: "chart" },
  allFields: { type: Array, default: () => [] },
  fields: {},
  periods: { type: Object, default: () => ({ second: [], first: [] }) },
  periodOffset: { type: Number, default: 0 },
  periodsFirstOffset: { type: Number, default: 0 },
  periodsSecondOffset: { type: Number, default: -1 },
});

const {
  period,
  periodType,
  loading,
  allFields,
  fields,
  fieldsMultiple,
  comparable,
  periods,
  type,
  periodOffset,
  periodsFirstOffset,
  periodsSecondOffset,
} = toRefs(props);

const emit = defineEmits([
  "input:period",
  "input:period-type",
  "input:periods",
  "input:type",
  "input:comparable",
  "input:fields",
  "input:period-offset",
  "input:periods-first-offset",
  "input:periods-second-offset",
]);

const durationOptions = [
  { label: "Week", value: "week" },
  { label: "Month", value: "month" },
  { label: "Month by week", value: "month-7_days" },
  { label: "Year", value: "year-7_days" },
  { label: "Year by month", value: "year-1_month" },
];

const typeOptions = [
  { label: "Bar", value: "bar" },
  { label: "Line", value: "line" },
  { label: "Area", value: "area" },
];

const colors = getApexChartsColors();

function getDurationTuple(type = "week", offset = 1) {
  let startDate, endDate;

  switch (type) {
    case "week": {
      let [firstDayTs, lastDayTs] = getWeekTuple();

      firstDayTs += offset * 86400000 * 7;
      lastDayTs += offset * 86400000 * 7;

      startDate = new Date(firstDayTs);
      endDate = new Date(lastDayTs);

      break;
    }
    case "month":
    case "month-7_days": {
      const todayDate = new Date(Date.now());
      todayDate.setMonth(todayDate.getMonth() + offset);

      startDate = new Date(todayDate.setDate(1));
      endDate = new Date(
        todayDate.setDate(
          new Date(
            todayDate.getFullYear(),
            todayDate.getMonth() + 1,
            0
          ).getDate()
        )
      );

      break;
    }
    case "year":
    case "year-1_month":
    case "year-7_days":
    default: {
      const todayDate = new Date(Date.now());
      const year = todayDate.getFullYear() + offset;

      startDate = new Date(todayDate.setFullYear(year, 0, 1));
      endDate = new Date(todayDate.setFullYear(year, 11, 31));
      break;
    }
  }
  return [startDate, endDate];
}

function getWeekTuple() {
  const date1 = new Date();
  const date2 = new Date();
  const today = date1.getDate();
  const currentDay = date1.getDay();
  const firstDay = date1.setDate(today - (currentDay || 7));
  const lastDay = date2.setDate(today - currentDay + 7);

  return [firstDay, lastDay];
}

function formatDate(date) {
  if (!date) {
    return;
  }
  return date.toISOString().split("T")[0].replaceAll("-", "/");
}

function setDefaultData() {
  if (!period?.value?.length) {
    const [startDate, endDate] = getDurationTuple(
      periodType.value,
      periodOffset.value
    );

    emit("input:period", [startDate, endDate]);
  }

  if (!periodType?.value) {
    emit("input:period-type", "month");
  }

  if (
    !periods?.value ||
    !periods.value.first?.length ||
    !periods.value.second?.length
  ) {
    const first = getDurationTuple(periodType.value, periodsFirstOffset.value);
    const second = getDurationTuple(
      periodType.value,
      periodsSecondOffset.value
    );

    emit("input:periods", { first: first, second: second });
  }
}

setDefaultData();

watch(periodType, () => {
  emit("input:period-offset", 0);
  emit("input:periods-first-offset", 0);
  emit("input:periods-second-offset", -1);
});

watch([periodType, periodOffset, comparable], () => {
  if (comparable?.value) {
    return;
  }

  const [startDate, endDate] = getDurationTuple(
    periodType.value,
    periodOffset.value
  );

  emit("input:period", [startDate, endDate]);
});

watch([periodType, periodsSecondOffset, periodsFirstOffset, comparable], () => {
  if (!comparable?.value) {
    return;
  }

  const first = getDurationTuple(periodType.value, periodsFirstOffset.value);
  const second = getDurationTuple(periodType.value, periodsSecondOffset.value);

  emit("input:periods", { first: first, second: second });
});
</script>

<style lang="scss">
.title {
  margin: 0px 20px;

  h1 {
    margin: 0px;
  }
}

.chart_container {
  margin: 0px 20px;
  width: 90% !important;

  h3 {
    margin: 0px;
  }

  .chart_options {
    display: flex;
    justify-content: space-between;

    .current_periods {
      display: flex;
      flex-direction: column;
    }

    .period_title {
      width: 50px;
      height: 20px;
    }

    .current_duration {
      display: flex;
      align-items: center;
      .current_duration_info {
        font-size: 1rem;
        text-align: center;
        margin: 0px 10px;
      }
    }
  }
}
</style>
