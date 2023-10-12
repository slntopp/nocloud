<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="price model"
          :value="template.billingPlan.title"
          @click:append="priceModelDialog = true"
          append-icon="mdi-pencil"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Product name"
          :value="
            template.billingPlan.products[template.product]?.title ||
            template.product
          "
          append-icon="mdi-pencil"
          @click:append="changeTariffDialog = true"
        />
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
          isMonitoringEmpty
        "
      >
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="date"
          :append-icon="!isMonitoringEmpty ? 'mdi-pencil' : null"
          @click:append="changeDatesDialog = true"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { defineProps, toRefs, computed } from "vue";
import { formatSecondsToDate } from "@/functions";

const props = defineProps(["template", "plans", "service", "sp"]);

const { template } = toRefs(props);

const date = computed(() =>
  formatSecondsToDate(+template.value?.data?.next_payment_date)
);
const isMonitoringEmpty = computed(() => date.value === "-");
</script>
