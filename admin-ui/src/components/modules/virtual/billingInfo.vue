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
          label="Tarif (product plan)"
          :value="template.product"
          @click:append="priceModelDialog = true"
          append-icon="mdi-pencil"
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
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="date"
          :append-icon="!isMonitoringsEmpty ? 'mdi-pencil' : null"
          @click:append="changeDatesDialog = true"
        />
      </v-col>
    </v-row>
    <change-monitorings
      :template="template"
      :service="service"
      v-model="changeDatesDialog"
      @refresh="emit('refresh')"
    />
    <edit-price-model
      v-model="priceModelDialog"
      :template="template"
      :plans="filtredPlans"
      @refresh="emit('refresh')"
      :service="service"
    />
  </div>
</template>

<script setup>
import { computed, defineProps, toRefs, ref } from "vue";
import { formatSecondsToDate } from "@/functions";
import ChangeMonitorings from "@/components/dialogs/changeMonitorings.vue";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";

const props = defineProps(["template", "plans", "service", "sp"]);
const emit = defineEmits(["refresh"]);

const { template, plans, service } = toRefs(props);

const changeDatesDialog = ref(false);
const priceModelDialog = ref(false);

const date = computed(() =>
  formatSecondsToDate(template.value?.data?.last_monitoring)
);
const isMonitoringsEmpty = computed(() => date.value === "-");

const price = computed(() => {
  return template.value.billingPlan.products[template.value.product]?.price;
});

const filtredPlans = computed(() =>
  plans.value.filter((p) => p.type === "virtual")
);
</script>

<style scoped></style>
