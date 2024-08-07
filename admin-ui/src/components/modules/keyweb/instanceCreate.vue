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
            :rules="rules.req"
            :value="instance.title"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.username', newVal)"
            label="Username"
            :rules="rules.req"
            :value="instance.config?.username"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.password', newVal)"
            label="Password"
            :rules="rules.req"
            :value="instance.config?.password"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            :rules="rules.req"
            @change="(newVal) => setValue('config.hostname', newVal)"
            label="Hostname"
            :value="instance.hostname?.password"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="Price model"
            item-text="title"
            item-value="uuid"
            return-object
            :value="instance.billing_plan"
            @change="setValue('billing_plan', $event)"
            :items="plans"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Product"
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
            label="OS"
            item-text="title"
            return-object
            :value="instance.config.configurations?.[os[0]?.type]"
            @change="
              setValue(
                'config.configurations.' + $event.type,
                $event.key?.split('$')?.[0]
              )
            "
            :items="os"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6" v-for="type in addonsTypes" :key="type">
          <v-autocomplete
            :label="Type"
            item-text="title"
            return-object
            :value="instance.config.configurations?.[type]"
            @change="
              setValue(
                'config.configurations.' + $event.type,
                $event.key?.split('$')?.[0]
              )
            "
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
  instance.value.billing_plan.uuid
    ? instance.value.billing_plan
    : plans.value.find((p) => p.uuid === instance.value.billing_plan)
);

const fullProduct = computed(() => billingPlan.value?.products[product.value]);

const addons = computed(() => {
  const addons = [];
  fullProduct.value?.meta.addons?.forEach((addonKey) => {
    const addon = billingPlan.value.resources.find((r) => r.key === addonKey);
    addons.push({
      title: addon.title,
      type: addon.meta.type,
      key: addon.key,
    });
  });

  return addons;
});
const addonsTypes = computed(() => {
  return [...new Set(addons.value.map((a) => a.type))];
});

const os = computed(() => {
  const oss = [];
  fullProduct.value?.meta.os?.forEach((osKey) => {
    const os = billingPlan.value.resources.find((r) => r.key === osKey);
    oss.push({
      title: os.title,
      type: os.meta.type,
      key: os.key,
    });
  });

  return oss;
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
    case "data.existing": {
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
