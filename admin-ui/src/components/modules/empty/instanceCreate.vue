<template>
  <div class="module">
    <v-card
      v-if="Object.keys(instance).length > 1"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('title', newVal)"
            label="Name"
            :value="instance.title"
          >
          </v-text-field>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-autocomplete
            label="Price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="plans"
            :rules="planRules"
            @change="changeBilling"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Product"
            :rules="requiredRule"
            :value="instance.product"
            :items="products"
            item-text="key"
            item-value="key"
            @change="changeProduct"
          >
            <template v-slot:item="{ item }">
              <div
                style="width: 100%"
                class="d-flex justify-space-between align-center"
              >
                <span>{{ item.key }}</span>
                <span class="ml-4">{{ item.title }}</span>
              </div>
            </template>
          </v-autocomplete>
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { onMounted, ref, toRefs, watch } from "vue";

const props = defineProps([
  "plans",
  "instance",
  "planRules",
  "spUuid",
  "isEdit",
]);
const { isEdit, instance, planRules, plans } = toRefs(props);

const emit = defineEmits(["set-instance", "set-value"]);

const getDefaultInstance = () => ({
  title: "instance",
  resources: {},
  data: {},
  config: {},
  billing_plan: {},
});

const bilingPlan = ref(null);
const products = ref([]);
const product = ref([]);
const requiredRule = ref([(val) => !!val || "Field required"]);

onMounted(() => {
  if (!isEdit.value) {
    emit("set-instance", getDefaultInstance());
  } else {
    changeBilling(instance.value.billing_plan);
  }
});

const changeBilling = (val) => {
  bilingPlan.value = plans.value.find((p) => p.uuid === val);
  if (bilingPlan.value) {
    products.value = Object.keys(bilingPlan.value.products).map((key) => ({
      key,
      title: bilingPlan.value.products[key].title,
    }));
  }
  setValue("billing_plan", bilingPlan.value);
};
const changeProduct = (val) => {
  product.value = val;

  const resources = {};
  if (bilingPlan.value.products[val]?.meta?.resources) {
    bilingPlan.value.products[val]?.meta?.resources.forEach((resource) => {
      resources[resource.key] = resource.title;
    });
  }
  setValue("resources", resources);

  setValue("product", product.value);
};
const setValue = (key, value) => {
  emit("set-value", { key, value });
};

watch(plans, () => {
  changeBilling(instance.value.billing_plan);
});
</script>

<script>
export default {
  name: "instance-empty-create",
};
</script>

<style scoped></style>
