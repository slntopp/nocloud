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
            label="title"
            :rules="rules.req"
            :value="instance.title"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.username', newVal)"
            label="username"
            :rules="rules.req"
            :value="instance.config?.username"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.password', newVal)"
            label="password"
            :rules="rules.req"
            :value="instance.config?.password"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            :rules="rules.req"
            @change="(newVal) => setValue('config.hostname', newVal)"
            label="hostname"
            :value="instance.hostname?.password"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            @change="setValue('billing_plan', $event)"
            :items="plans.list"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="product"
            :value="instance.product"
            :rules="rules.req"
            :items="products"
            @change="setValue('product', $event)"
            item-text="title"
            item-value="key"
          />
        </v-col>
        <v-col cols="6" class="d-flex align-center">
          Existing:
          <v-switch
            class="d-inline-block ml-2"
            :input-value="instance.data?.existing"
            @change="setValue('data.existing', $event)"
          />
        </v-col>
        <v-col cols="6" class="d-flex align-center">
          <v-text-field
            v-if="instance.data?.existing"
            label="Service id"
            type="number"
            :value="instance.data?.serviceId"
            :rules="rules.req"
            @change="setValue(`data.serviceId`, +$event)"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="os"
            item-text="title"
            return-object
            :value="instance.config.configurations?.[os[0]?.type]"
            @change="
              setValue('config.configurations.' + $event.type, $event.title)
            "
            :items="os"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6" v-for="type in addonsTypes" :key="type">
          <v-autocomplete
            :label="type"
            item-text="title"
            return-object
            :value="instance.config.configurations?.[type]"
            @change="setValue('config.configurations.' + type, $event.title)"
            :items="addons.filter((a) => a.type === type)"
            :rules="planRules"
          />
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, toRefs, ref } from "vue";
import { defaultFilterObject } from "@/functions";

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    configurations: {},
  },
  resources: {},
  data: { existing: false },
  billing_plan: {},
});

const props = defineProps([
  "plans",
  "instance",
  "planRules",
  "spUuid",
  "isEdit",
]);
const { plans, instance, planRules } = toRefs(props);
const emits = defineEmits(["set-instance", "set-value"]);

const product = ref("");
const rules = ref({
  req: [(v) => !!v || "required field"],
});

const billingPlan = computed(() =>
  plans.value.list.find((p) => p.uuid === instance.value.billing_plan)
);

const fullProduct = computed(() => billingPlan.value?.products[product.value]);

const addons = computed(() => {
  const addons = [];
  fullProduct.value?.meta.addons?.forEach((a) => {
    addons.push({
      ...a,
      ...billingPlan.value.resources.find((r) => r.key === a.key),
    });
  });

  return addons;
});
const addonsTypes = computed(() => {
  return [...new Set(addons.value.map((a) => a.type))];
});

const os = computed(() => {
  const os = [];
  fullProduct.value?.meta.os?.forEach((a) => {
    os.push({
      ...a,
      ...billingPlan.value.resources.find((r) => r.key === a.key),
    });
  });

  return os;
});

const products = computed(() => {
  const products = [];
  Object.keys(billingPlan.value?.products || {}).forEach((key) => {
    products.push({ ...billingPlan.value?.products[key], key });
  });

  return products;
});

onMounted(() => {
  emits("set-instance", getDefaultInstance());
});

const setValue = (key, value) => {
  emits("set-value", { key, value });

  /* eslint-disable */
  switch (key) {
    case "product": {
      product.value = value;
      setValue(
        "config.cycle",
        value.endsWith("monthly") ? "monthly" : "annually"
      );
      setValue("config.id", fullProduct.value?.meta?.keywebId);
    }
    case 'data.existing':{
      setValue("data.serviceId", undefined);
    }
  }
  /* eslint-enable */
};
</script>

<script>
export default {
  name: "instance-keyweb-create",
};
</script>

<style scoped></style>
