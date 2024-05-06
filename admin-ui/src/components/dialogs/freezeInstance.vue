<template>
  <instance-control-btn hint="Freeze">
    <v-dialog v-model="isModalOpen" max-width="500">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          :disabled="disabled"
          :loading="loading"
          class="ma-1"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon>mdi-snowflake</v-icon>
        </v-btn>
      </template>
      <v-card class="pa-5" color="background-light">
        <date-picker
          label="Frozen until"
          :min="formatSecondsToDateString(Date.now() / 1000 + 86400)"
          v-model="freezeTo"
        />
        <v-card-actions class="d-flex justify-end">
          <v-btn class="mr-2" @click="isModalOpen = false"> Cancel </v-btn>
          <v-btn class="mr-2" @click="freeze(false)"> Freeze </v-btn>
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
const freezeTo = ref();

const freeze = async () => {
  isModalOpen.value = false;

  emit("click", new Date(freezeTo.value).getTime() / 1000);
};
</script>

<style scoped></style>
