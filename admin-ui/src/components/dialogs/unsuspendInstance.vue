<template>
  <instance-control-btn hint="unsuspend">
    <v-dialog v-model="isModalOpen" max-width="500">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          :disabled="disabled"
          :loading="loading"
          class="ma-1"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon>mdi-weather-sunny</v-icon>
        </v-btn>
      </template>
      <v-card class="pa-5" color="background-light">
        <date-picker
          label="Unsuspend until"
          :min="formatSecondsToDateString(Date.now() / 1000 + 86400)"
          v-model="unsuspendTo"
          clearable
        />
        <v-card-actions class="d-flex justify-end">
          <v-btn class="mr-2" @click="isModalOpen = false"> Cancel </v-btn>
          <v-btn class="mr-2" @click="unsuspend(false)"> unsuspend </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </instance-control-btn>
</template>

<script setup>
import { toRefs, ref } from "vue";
import { formatSecondsToDateString } from "@/functions";
import InstanceControlBtn from "@/components/ui/hintBtn.vue";
import DatePicker from "@/components/ui/datePicker.vue";

const props = defineProps(["disabled", "loading", "template"]);
const emit = defineEmits(["click"]);

const { disabled, loading } = toRefs(props);

const isModalOpen = ref(false);
const unsuspendTo = ref();

const unsuspend = async () => {
  isModalOpen.value = false;

  emit("click", new Date(unsuspendTo.value).getTime() / 1000);
};
</script>

<style scoped></style>
