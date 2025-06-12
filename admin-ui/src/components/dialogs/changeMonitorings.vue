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
        <v-row v-for="key in Object.keys(nextPaymentDates || {})" :key="key">
          <v-col cols="4">
            <v-card-title>{{ nextPaymentDates[key].title }}</v-card-title>
          </v-col>
          <v-col cols="8">
            <date-picker v-model="nextPaymentDates[key].value" />
          </v-col>
        </v-row>
      </div>
      <v-row v-else>
        <v-col cols="4">
          <v-card-title>product</v-card-title>
        </v-col>
        <v-col cols="8">
          <date-picker v-model="newAllDate" />
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
import { useStore } from "@/store";
import DatePicker from "@/components/ui/dateTimePicker.vue";

const props = defineProps(["template", "service", "value"]);
const emit = defineEmits(["refresh", "input"]);

const store = useStore();

const { template, service } = toRefs(props);
const changeDatesLoading = ref(false);
const isChangeAll = ref(true);
const nextPaymentDates = ref({});
const newAllDate = ref();

function toDateTimeLocal(timestamp) {
  const date = new Date(timestamp * 1000);

  const pad = (n) => String(n).padStart(2, "0");

  const yyyy = date.getFullYear();
  const mm = pad(date.getMonth() + 1);
  const dd = pad(date.getDate());
  const hh = pad(date.getHours());
  const min = pad(date.getMinutes());

  return `${yyyy}-${mm}-${dd}T${hh}:${min}`;
}

const setNextPaymentDate = () => {
  const data = JSON.parse(JSON.stringify(template.value.data));

  const monitorings = {};

  Object.keys(data).forEach((key) => {
    if (key.includes("next_payment_date") && data[key]) {
      const title = key
        .replace("_next_payment_date", "")
        .replace("next_payment_date", "product");
      let value = +data[key];
      value = toDateTimeLocal(value);
      monitorings[key] = {
        value: value,
        firstValue: value,
        title: title,
      };
    }
  });

  nextPaymentDates.value = monitorings;

  newAllDate.value = monitorings["next_payment_date"].value;
};

const changeDates = async () => {
  changeDatesLoading.value = true;

  const tempService = JSON.parse(JSON.stringify(service.value));

  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  const changedDates = {};

  Object.keys(nextPaymentDates.value).forEach((key) => {
    if (
      isChangeAll.value ||
      nextPaymentDates.value[key].firstValue !=
        nextPaymentDates.value[key].value
    ) {
      const { value } = nextPaymentDates.value[key];

      let newVal =
        new Date(isChangeAll.value ? newAllDate.value : value).getTime() / 1000;
      let oldVal = new Date(template.value.data[key]).getTime();

      changedDates[key.replace("next_payment_date", "last_monitoring")] =
        oldVal + (newVal - oldVal);
      changedDates[key] = newVal;
    }
  });

  tempService.instancesGroups[igIndex].instances[instanceIndex].data = {
    ...tempService.instancesGroups[igIndex].instances[instanceIndex].data,
    ...changedDates,
  };

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
  setNextPaymentDate();
});
</script>

<style scoped></style>
