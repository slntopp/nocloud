<template>
  <v-dialog :value="value" @input="emit('input', $event)" max-width="60%">
    <v-card>
      <v-tabs>
        <v-tab>Base</v-tab>
        <v-tab>Individual</v-tab>
        <v-tab-item>
          <v-card class="pa-5">
            <v-row>
              <v-col cols="3">
                <v-card-title>Tarrif:</v-card-title>
                <v-select
                  v-model="selectedTarrif"
                  :items="availableTarrifs"
                  item-text="title"
                  return-object
                ></v-select>
              </v-col>
              <v-col cols="9">
                <v-card-title>Tarrif resources:</v-card-title>
                <v-text-field
                  readonly
                  v-for="resource in Object.keys(
                    selectedTarrif?.resources || {}
                  )"
                  :key="resource"
                  :value="selectedTarrif?.resources?.[resource]"
                  :label="resource"
                ></v-text-field>
              </v-col>
            </v-row>
            <v-row justify="end">
              <v-btn class="mx-3" @click="emit('input', false)">Close</v-btn>
              <v-btn
                class="mx-3"
                @click="changeTarrif"
                :disabled="selectedTarrif.title === template.product"
                :loading="changeTarrifLoading"
                >Change tarrif</v-btn
              >
            </v-row>
          </v-card>
        </v-tab-item>
        <v-tab-item>
          <v-card class="pa-5">
            <v-row>
              <v-col cols="5">
                <v-text-field
                  type="number"
                  label="price"
                  v-model.number="individualPlan.product.price"
                />
              </v-col>
              <v-col cols="5">
                <date-field
                  class="mt-3"
                  :period="individualPlan.product.period"
                  @change-date="individualPlan.product.period = $event"
                ></date-field
              ></v-col>
            </v-row>
            <v-card-title>Product resources</v-card-title>
            <v-row>
              <v-col
                v-for="key in Object.keys(
                  individualPlan.product.resources || {}
                )"
                :key="key"
              >
                <v-text-field
                  type="number"
                  v-model.number="individualPlan.product.resources[key]"
                  :label="key"
                /> </v-col
            ></v-row>
            <v-card-title>Plan resources</v-card-title>
            <v-row>
              <v-col
                v-for="resource in individualPlan.resources"
                :key="resource.key"
              >
                <v-text-field
                  :label="`${resource.key}(price)`"
                  v-model.number="resource.price"
                ></v-text-field>
              </v-col>
            </v-row>
            <v-row justify="end">
              <v-btn class="mx-3" @click="$emit('input', false)">Close</v-btn>
              <v-btn
                class="mx-3"
                @click="createIndividual"
                :loading="createIndividualLoading"
                >Create individual</v-btn
              >
            </v-row>
          </v-card>
        </v-tab-item>
      </v-tabs>
    </v-card>
  </v-dialog>
</template>

<script setup>
import dateField from "@/components/date.vue";
import { onMounted, toRefs, ref } from "vue";
import api from "@/api";
import { getTimestamp } from "@/functions";

const props = defineProps([
  "template",
  "service",
  "value",
  "billingPlan",
  "availableTarrifs",
]);
const emit = defineEmits(["refresh", "input"]);

const { template, service, billingPlan, value, availableTarrifs } =
  toRefs(props);
const selectedTarrif = ref({});
const individualPlan = ref({ product: {}, resources: {} });
const changeTarrifLoading = ref(false);
const createIndividualLoading = ref(false);

const changeTarrif = () => {
  const tempService = JSON.parse(JSON.stringify(service.value));
  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  tempService.instancesGroups[igIndex].instances[instanceIndex].product =
    selectedTarrif.value.title;
  Object.keys(selectedTarrif.value.resources).forEach((k) => {
    tempService.instancesGroups[igIndex].instances[instanceIndex].resources[k] =
      selectedTarrif.value.resources[k];
  });

  tempService.instancesGroups[igIndex].instances[instanceIndex].billingPlan =
    billingPlan;

  changeTarrifLoading.value = true;

  api.services
    ._update(tempService)
    .then(() => {
      emit("refresh");
    })
    .finally(() => {
      changeTarrifLoading.value = false;
      emit("input", false);
    });
};

const createIndividual = () => {
  const product = individualPlan.value.product;
  const resources = individualPlan.value.resources;

  product.period = getTimestamp(product.period);
  const plan = {
    title: template.value.title + " (Individual)",
    public: false,
    kind: billingPlan.value.kind,
  };
  plan.resources = resources;
  plan.products = { [template.value.product]: product };
  createIndividualLoading.value = true;

  api.plans.create(plan).then((data) => {
    api.servicesProviders.bindPlan(template.value.sp, data.uuid).then(() => {
      const tempService = JSON.parse(JSON.stringify(service.value));
      const igIndex = tempService.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === template.value.uuid)
      );
      const instanceIndex = tempService.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === template.value.uuid);

      tempService.instancesGroups[igIndex].instances[instanceIndex].product =
        template.value.product;
      tempService.instancesGroups[igIndex].instances[
        instanceIndex
      ].billingPlan = data;
      Object.keys(individualPlan.value.product.resources).forEach((k) => {
        tempService.instancesGroups[igIndex].instances[instanceIndex].resources[
          k
        ] = individualPlan.value.product.resources[k];
      });

      api.services._update(tempService).then(() => {
          createIndividualLoading.value = false;
          emit('refresh')
      });
    });
  });
};

const setIndividualPlan = () => {
  individualPlan.value.product = JSON.parse(
    JSON.stringify(template.value.billingPlan.products[template.value.product])
  );
  const date = new Date(individualPlan.value.product.period * 1000);
  const time = date.toUTCString().split(" ");
  individualPlan.value.product.period = {
    day: `${date.getUTCDate() - 1}`,
    month: `${date.getUTCMonth()}`,
    year: `${date.getUTCFullYear() - 1970}`,
    quarter: "0",
    week: "0",
    time: time.at(-2),
  };
  individualPlan.value.resources = JSON.parse(
    JSON.stringify(template.value.billingPlan?.resources)
  );
};

onMounted(() => {
  selectedTarrif.value = {
    title: template.value.product,
    resources:
      template.value.billingPlan.products[template.value.product]?.resources,
  };
  setIndividualPlan();
});
</script>

<style scoped></style>
