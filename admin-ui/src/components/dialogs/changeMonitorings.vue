<template>
  <v-dialog
    persistent
    :value="value"
    @input="emit('input', value)"
    max-width="60%"
  >
    <v-card class="pa-5">
      <v-card-title class="text-center">Change next payment dates</v-card-title>
      <div v-if="!isChangeAll">
        <v-row v-for="key in Object.keys(lastMonitorings || {})" :key="key">
          <v-col cols="4">
            <v-card-title>{{ lastMonitorings[key].title }}</v-card-title>
          </v-col>
          <v-col cols="8">
            <date-picker :min="min" v-model="lastMonitorings[key].value" />
          </v-col>
        </v-row>
      </div>
      <v-row v-else>
        <v-col cols="4">
          <v-card-title>product</v-card-title>
        </v-col>
        <v-col cols="8">
          <date-picker :min="min" v-model="newAllDate" />
        </v-col>
      </v-row>

      <v-row justify="end">
        <v-switch
          class="mx-3 mt-0"
          v-model="isChangeAll"
          label="All at the same time"
        />
        <v-btn class="mx-3" @click="emit('input', false)">Close</v-btn>
        <v-btn class="mx-3" :loading="changeDatesLoading" @click="changeDates"
          >Change dates</v-btn
        >
      </v-row>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { onMounted, toRefs, ref } from "vue";
import api from "@/api";
import { formatSecondsToDate } from "@/functions";
import { useStore } from "@/store";
import DatePicker from "@/components/ui/datePicker.vue";

const props = defineProps(["template", "service", "value"]);
const emit = defineEmits(["refresh", "input"]);

const store = useStore();

const { template, service } = toRefs(props);
const changeDatesLoading = ref(false);
const isChangeAll = ref(true);
const lastMonitorings = ref({});
const newAllDate = ref();
const min = ref();

const setLastMonitorings = () => {
  const data = JSON.parse(JSON.stringify(template.value.data));

  const monitorings = {};

  Object.keys(data).forEach((key) => {
    if (key.includes("last_monitoring") && data[key]) {
      const title = key
        .replace("_last_monitoring", "")
        .replace("last_monitoring", "product");
      let value = +data[key];

      if (title === "product") {
        value =
          value +
          +template.value.billingPlan.products[template.value.product].period;
      } else {
        value =
          value +
          +template.value.billingPlan.resources.find((r) => r.key === title)
            ?.period;
      }

      value = formatSecondsToDate(value);
      monitorings[key] = {
        value: value,
        firstValue: value,
        title: title,
      };
    }
  });

  lastMonitorings.value = monitorings;

  newAllDate.value = monitorings["last_monitoring"].value;
};

const changeDates = async () => {
  const tempService = JSON.parse(JSON.stringify(service.value));

  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  const changedDates = {};

  Object.keys(lastMonitorings.value).forEach((key) => {
    if (
      isChangeAll.value ||
      lastMonitorings.value[key].firstValue != lastMonitorings.value[key].value
    ) {
      const { value, title } = lastMonitorings.value[key];

      let baseVal =
        new Date(isChangeAll.value ? newAllDate.value : value).getTime() / 1000;

      if (title === "product") {
        baseVal =
          baseVal -
          +template.value.billingPlan.products[template.value.product].period;
      } else {
        baseVal =
          baseVal -
          +template.value.billingPlan.resources.find((r) => r.key === title)
            ?.period;
      }

      changedDates[key] = baseVal;
    }
  });

  tempService.instancesGroups[igIndex].instances[instanceIndex].data = {
    ...tempService.instancesGroups[igIndex].instances[instanceIndex].data,
    ...changedDates,
  };

  changeDatesLoading.value = true;
  try {
    await api.services._update(tempService);
    emit("refresh");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during change monitoring",
    });
  } finally {
    changeDatesLoading.value = false;
    emit("input", false);
  }
};

onMounted(() => {
  setLastMonitorings();
  //tommoraw
  min.value = new Date(new Date().getTime() + 24 * 60 * 60 * 1000)
    .toISOString()
    .slice(0, 10);
});
</script>

<style scoped></style>
