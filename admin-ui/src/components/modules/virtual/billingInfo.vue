<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="price model"
          :value="template.billingPlan.title"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Tarif (product plan)"
          :value="template.product"
        />
      </v-col>
      <v-col>
        <v-text-field readonly label="Price instance total" :value="price" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Date (create)"
          :value="template.data.creation"
        />
      </v-col>

      <v-col
        v-if="
          template.billingPlan.title.toLowerCase() !== 'payg' ||
          isMonitoringsEmpty
        "
      >
        <v-text-field readonly label="Due to date/next payment" :value="date" />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { computed, defineProps, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";

const props = defineProps(["template", "plans", "service", "sp"]);

const { template } = toRefs(props);

const date = computed(() =>
  formatSecondsToDate(template.value?.data?.last_monitoring)
);
const isMonitoringsEmpty = computed(() => date.value === "-");

const price = computed(() => {
  return template.value.billingPlan.products[template.value.product]?.price;
});
</script>

<style scoped></style>