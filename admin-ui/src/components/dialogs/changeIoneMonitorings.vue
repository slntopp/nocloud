<template>
  <v-dialog :value="value" @input="emit('input', value)" max-width="60%">
    <v-card class="pa-5">
      <v-card-title class="text-center">Change monitoring dates</v-card-title>
      <v-row v-for="key in Object.keys(lastMonitorings || {})" :key="key">
        <v-col cols="4">
          <v-card-title>{{ lastMonitorings[key].title }}</v-card-title>
        </v-col>
        <v-col cols="8">
          <v-menu
            :close-on-content-click="false"
            transition="scale-transition"
            min-width="auto"
          >
            <template v-slot:activator="{ on, attrs }">
              <v-text-field
                v-bind="attrs"
                v-on="on"
                prepend-inner-icon="mdi-calendar"
                :value="lastMonitorings[key].value"
                readonly
              />
            </template>
            <v-date-picker
              scrollable
              :min="lastMonitorings[key].firstValue"
              v-model="lastMonitorings[key].value"
            ></v-date-picker>
          </v-menu>
        </v-col>
      </v-row>
      <v-row justify="end">
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

const props = defineProps(["template", "service", "value"]);
const emit = defineEmits(["refresh", "input"]);

const { template, service } = toRefs(props);
const changeDatesLoading = ref(false);
const lastMonitorings = ref({});

const setLastMonitorings = () => {
  const data = JSON.parse(JSON.stringify(template.value.data));

  const monitorings = {};

  Object.keys(data).forEach((key) => {
    if (key.includes("last_monitoring") && data[key]) {
      monitorings[key] = {
        value: formatSecondsToDate(data[key]),
        firstValue: formatSecondsToDate(data[key]),
        title: key
          .replace("_last_monitoring", "")
          .replace("last_monitoring", "product"),
      };
    }
  });

  lastMonitorings.value = monitorings;
};

const changeDates = () => {
  const tempService = JSON.parse(JSON.stringify(service.value));

  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  const changedDates = {};

  Object.keys(lastMonitorings.value).forEach((key) => {
    if (lastMonitorings.value[key].firstValue != lastMonitorings.value[key].value) {
      changedDates[key] = new Date(lastMonitorings.value[key].value).getTime() / 1000;
    }
  });

  tempService.instancesGroups[igIndex].instances[instanceIndex].data = {
    ...tempService.instancesGroups[igIndex].instances[instanceIndex].data,
    ...changedDates,
  };

  changeDatesLoading.value = true;
  api.services
    ._update(tempService)
    .then(() => {
      emit("refresh");
    })
    .finally(() => {
      changeDatesLoading.value = false;
      emit("input", false);
    });
};

onMounted(() => {
  setLastMonitorings();
});
</script>

<style scoped></style>
