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
          <plans-autocomplete
            :value="instance.billing_plan"
            :custom-params="{
              filters: { type: ['keyweb'] },
              anonymously: true,
            }"
            @input="setValue('billing_plan', $event)"
            return-object
            label="Price model"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            v-model="duration"
            label="Duration"
            :items="Object.keys(durations)"
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
        <v-col cols="6">
          <v-autocomplete
            label="OS"
            item-text="title"
            return-object
            v-model="selectedOs"
            :items="os"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6" v-for="type in addonsTypes" :key="type">
          <v-autocomplete
            :label="type || 'Custom'"
            v-model="selectedAddons"
            :items="configurationAddons.filter((a) => a?.meta?.type === type)"
            :loading="isAddonsLoading"
            multiple
            item-text="title"
            return-object
            :rules="planRules"
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
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, toRefs, ref, watch } from "vue";
import useInstanceAddons from "@/hooks/useInstanceAddons";
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    configurations: {},
  },
  resources: {},
  data: { existing: false },
  billing_plan: {},
  addons: [],
});

const props = defineProps(["instance", "planRules", "spUuid", "isEdit"]);
const { instance, planRules } = toRefs(props);
const emits = defineEmits(["set-instance", "set-value"]);

const osTypeKey = "VM Template|OS";

const { setTariffAddons, getAvailableAddons, isAddonsLoading } =
  useInstanceAddons(instance, (key, value) => setValue(key, value));

const product = ref("");
const rules = ref({
  req: [(v) => !!v || "required field"],
});
const addons = ref([]);
const selectedAddons = ref([]);
const selectedOs = ref();
const duration = ref("Monthly");
const durations = { Monthly: 2592000, Yearly: 31536000 };

const billingPlan = computed(() => instance.value.billing_plan);

const fullProduct = computed(() => billingPlan.value?.products[product.value]);
const addonsTypes = computed(() => {
  return [...new Set(configurationAddons.value.map((a) => a.meta?.type))];
});

const configurationAddons = computed(() => {
  return addons.value.filter((a) => a.meta?.type !== osTypeKey);
});

const os = computed(() => {
  return addons.value.filter((a) => a.meta?.type === osTypeKey);
});

const products = computed(() => {
  const products = [];
  Object.keys(billingPlan.value?.products || {}).forEach((key) => {
    if (
      Number(billingPlan.value?.products[key]?.period) !==
      durations[duration.value]
    ) {
      return;
    }

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

const setAddons = () => {
  const addons = [];

  selectedAddons.value.concat(selectedOs.value).map((addon) => {
    if (!addon) {
      return;
    }

    if (addon?.meta?.type) {
      setValue(
        "config.configurations." + addon?.meta?.type,
        addon.meta.key?.split("$")?.[0]
      );
    }

    addons.push(addon.uuid);

    setValue("addons", addons);
  });
};

watch(product, () => {
  setTariffAddons();
});

watch(isAddonsLoading, (value) => {
  if (!value) {
    addons.value = getAvailableAddons() || [];
  }
});

watch(selectedOs, setAddons);
watch(selectedAddons, setAddons);
</script>

<script>
export default {
  name: "instance-keyweb-create",
};
</script>

<style scoped></style>
