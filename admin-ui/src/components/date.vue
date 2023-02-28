<template>
  <v-row dense align="center">
    <v-col cols="10">
      <v-select
        dense
        v-model="date"
        :items="typesDate"
        :rules="rules.general"
      />
    </v-col>
    <v-col cols="2" v-if="date !== 'Custom'">
      <v-text-field
        dense
        v-model="amountDate"
        v-if="date === 'Time'"
        :rules="rules.time"
      />
      <v-text-field
        dense
        v-else
        type="number"
        v-model="amountDate"
        :rules="rules.number"
      />
    </v-col>
    <v-col cols="2" v-else>
      <v-menu left v-model="menuVisible" :close-on-content-click="false">
        <template v-slot:activator="{ on, attrs }">
          <v-icon v-bind="attrs" v-on="on"> mdi-playlist-edit </v-icon>
        </template>

        <v-card>
          <v-list class="columns-2">
            <v-list-item v-for="item of items" :key="item.title">
              <v-list-item-title>{{ item.title }}</v-list-item-title>
              <v-list-item-action>
                <v-text-field
                  dense
                  v-model="fullDate[item.model]"
                  :type="item.model === 'time' ? 'text' : 'number'"
                  :rules="item.model === 'time' ? rules.time : rules.customNumber"
                />
              </v-list-item-action>
            </v-list-item>
          </v-list>

          <v-card-actions>
            <v-spacer />
            <v-btn text @click="resetDate(fullDate)">Reset</v-btn>
            <v-btn text color="primary" @click="menuVisible = false">
              Save
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-menu>
    </v-col>
  </v-row>
</template>

<script>
export default { name: "date-field" }
</script>

<script setup>
import { onMounted, ref, toRefs, watch } from 'vue';

const props = defineProps({ period: Object });
const emits = defineEmits(["changeDate"]);
const { period } = toRefs(props);

const date = ref("");
const amountDate = ref("0");
const menuVisible = ref(false);

let fullDate = ref({
  day: "0",
  month: "0",
  year: "0",
  quarter: "0",
  week: "0",
  time: "00:01:00",
});
const typesDate = [
  "Day",
  "Week",
  "Month",
  "Quarter",
  "Year",
  "Time",
  "Hour",
  "Minute",
  "Custom",
];

const items = [
  { title: "Day", model: "day" },
  { title: "Week", model: "week" },
  { title: "Month", model: "month" },
  { title: "Quarter", model: "quarter" },
  { title: "Year", model: "year" },
  { title: "Time", model: "time" },
];

const rules = {
  general: [(v) => !!v || "This field is required!"],
  number: [
    (value) => !!value || "Is required!",
    (value) => /^[1-9][0-9]{0,1}$/.test(value) || "Invalid!",
  ],
  customNumber: [
    (value) => !!value || "Is required!",
    (value) => /^[1-9][0-9]{0,1}|0$/.test(value) || "Invalid!",
  ],
  time: [
    (value) => !!value || "Is required!",
    (value) =>
      /^([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/.test(value) ||
      "Invalid!",
  ],
};

function resetDate(date) {
  Object.keys(date).forEach((key) => {
    date[key] = key === "time" ? "00:00:00" : "0";
  });
}

onMounted(() => {
  if (period.value) {
    date.value = "Custom";
    fullDate.value = period.value;
  }
});

watch(date, (value) => {
  if (value === "Custom") return;

  const key = value.toLowerCase();
  const amount = key === "time" ? "00:00:00" : "1";

  resetDate(fullDate.value);

  amountDate.value = amount;
});

watch(amountDate, (value) => {
  if (date.value === "") return;

  let key = date.value.toLowerCase();
  const newValue = value.length < 2 ? `0${value}` : value;

  switch (key) {
    case "hour":
      key = "time";
      value = `${newValue}:00:00`;
      break;
    case "minute":
      key = "time";
      value = `00:${newValue}:00`;
  }

  resetDate(fullDate.value);

  fullDate.value[key] = value;
});

watch(
  () => fullDate,
  (value) => { emits("changeDate", value) },
  { deep: true }
);

watch(period, (value) => {
  if (value) {
    date.value = "Custom";
    fullDate.value = value;
  }
});
</script>

<style scoped lang="scss">
.columns-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
}
</style>
