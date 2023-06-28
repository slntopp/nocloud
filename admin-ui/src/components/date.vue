<template>
  <v-row class="ma-0 pa-0" dense align="center">
    <v-col cols="10">
      <v-select
        dense
        :value="date"
        @change="changeDateValue"
        :items="typesDate"
        :rules="rules.general"
      />
    </v-col>
    <v-col cols="2" v-if="date !== 'Custom'">
      <v-text-field
        dense
        type="number"
        :value="fullDate[dateKey]"
        @change="changeAmountValue"
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
                  :rules="
                    item.model === 'time' ? rules.time : rules.customNumber
                  "
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
export default { name: "date-field" };
</script>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";

const props = defineProps({ period: Object });
const emits = defineEmits(["changeDate"]);
const { period } = toRefs(props);

const date = ref("");
// const amountDate = ref("0");
const menuVisible = ref(false);

let fullDate = ref({
  day: "0",
  month: "0",
  year: "0",
  quarter: "0",
  week: "0",
  time: "00:01:00",
});
const typesDate = ["Day", "Week", "Month", "Quarter", "Year", "Custom"];

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
      /^([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/.test(value) || "Invalid!",
  ],
};

function resetDate(date) {
  Object.keys(date).forEach((key) => {
    date[key] = key === "time" ? "00:00:00" : "0";
  });
}

onMounted(() => {
  if (period.value) {
    fullDate.value = period.value;
    setDateAndAmount();
  }
});

const changeAmountValue = (value) => {
  if (!dateKey.value) {
    return;
  }

  resetDate(fullDate.value);

  fullDate.value[dateKey.value] = value;
};

const changeDateValue = (value) => {
  date.value = value;
  
  if (value === "Custom") {
    return;
  }


  resetDate(fullDate.value);

  fullDate.value[dateKey.value] = 1;
};

watch(
  () => fullDate,
  (value) => {
    emits("changeDate", value);
  },
  { deep: true }
);

const setDateAndAmount = () => {
  let newDate = "";
  let newAmount = 0;
  for (const key of Object.keys(fullDate.value)) {
    const numberValue = +fullDate.value[key];
    if (numberValue && newDate) {
      newDate = "";
      newAmount = 0;
      break;
    } else if (numberValue && key === "day") {
      if (numberValue % 30 === 0) {
        newDate = "Month";
        newAmount = numberValue / 30;
      } else if (numberValue % 7 === 0) {
        newDate = "Week";
        newAmount = numberValue / 7;
      } else {
        newDate = "Day";
        newAmount = numberValue;
      }
    } else if (numberValue) {
      newDate = key.slice(0, 1).toUpperCase() + key.slice(1);
      newAmount = numberValue;
    }
  }
  if (newDate) {
    date.value = newDate;
    resetDate(fullDate.value);
    fullDate.value[dateKey.value] = newAmount;
  } else {
    date.value = "Custom";
  }
};

watch(period, (value) => {
  if (value) {
    fullDate.value = value;
    setDateAndAmount();
  }
});

const dateKey = computed(() => {
  if (date.value === "Custom") {
    return;
  }
  return date.value.toLowerCase();
});
</script>

<style scoped lang="scss">
.columns-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
}
</style>
