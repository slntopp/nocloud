<template>
  <div class="chart_container">
    <div class="chart_options">
      <div class="d-flex align-center">
        <v-card-title> {{ description }} </v-card-title>

        <div class="current_duration">
          <v-btn small @click="periodOffset--">
            <v-icon>mdi-minus</v-icon>
          </v-btn>
          <span class="current_duration_info"
            >{{ formatDate(period[0]) }} - {{ formatDate(period[1]) }}</span
          >
          <v-btn small @click="periodOffset++">
            <v-icon>mdi-plus</v-icon>
          </v-btn>
        </div>
      </div>

      <div class="d-flex alingn-center">
        <slot name="options" />
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
import { ref } from "vue";

const props = defineProps([
  "loading",
  "period",
  "periodType",
  "description",
  "type",
]);
const { period, periodType, loading } = toRefs(props);

const emit = defineEmits(["input:period", "input:period-type", "input:type"]);

const durationOptions = [
  { label: "Week", value: "week" },
  { label: "Month", value: "month" },
  { label: "Year", value: "year" },
];

const typeOptions = [
  { label: "Bar", value: "bar" },
  { label: "Line", value: "line" },
  { label: "Area", value: "area" },
];

const periodOffset = ref(0);

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
    case "month": {
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
  if (!period?.value) {
    emit("input:period", []);
  }

  if (!periodType?.value) {
    emit("input:period-type", "month");
  }

  const [startDate, endDate] = getDurationTuple(
    periodType.value,
    periodOffset.value
  );

  emit("input:period", [startDate, endDate]);
}

setDefaultData();

watch(periodType, () => {
  periodOffset.value = 0;
});

watch([periodType, periodOffset], () => {
  const [startDate, endDate] = getDurationTuple(
    periodType.value,
    periodOffset.value
  );

  emit("input:period", [startDate, endDate]);
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

  .chart {
    max-width: 1400px;
  }

  .chart_options {
    max-width: 1400px;
    display: flex;
    justify-content: space-between;
    .current_duration {
      display: flex;
      align-items: center;
      .current_duration_info {
        font-size: 1.3rem;
        text-align: center;
        margin: 0px 10px;
      }
    }
  }
}
</style>
