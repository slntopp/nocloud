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
            label="Name"
            :value="instance.title"
            @change="(newVal) => setValue('title', newVal)"
          />
        </v-col>
        <v-col cols="6">
          <plans-autocomplete
            :value="bilingPlan"
            :custom-params="{
              filters: { type: ['proxmox'], 'meta.isIndividual': [false] },
              anonymously: false,
            }"
            return-object
            label="Price model"
            :rules="planRules"
            @input="setValue('billing_plan', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            v-if="products.length > 0"
            label="Product"
            :rules="requiredRule"
            :value="instance.product"
            :items="products"
            item-text="key"
            item-value="key"
            @change="changeProduct"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            type="number"
            label="Template ID"
            :value="instance.config?.template_id"
            @change="(v) => setValue('config.template_id', Number(v))"
          />
        </v-col>
        <v-col cols="4">
          <v-text-field
            type="number"
            label="CPU"
            :value="instance.resources?.cpu"
            @change="(v) => setValue('resources.cpu', Number(v))"
          />
        </v-col>
        <v-col cols="4">
          <v-text-field
            type="number"
            label="RAM (MB)"
            :value="instance.resources?.ram"
            @change="(v) => setValue('resources.ram', Number(v))"
          />
        </v-col>
        <v-col cols="4">
          <v-text-field
            type="number"
            label="Disk (MB)"
            :value="instance.resources?.drive_size"
            @change="(v) => setValue('resources.drive_size', Number(v))"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="Username"
            :value="instance.config?.username"
            @change="(v) => setValue('config.username', v)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="Password"
            :value="instance.config?.password"
            @change="(v) => setValue('config.password', v)"
          />
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";

const props = defineProps(["instance", "planRules", "spUuid"]);
const { instance, planRules } = toRefs(props);
const emit = defineEmits(["set-instance", "set-value"]);
const requiredRule = ref([(val) => !!val || "Field required"]);

onMounted(() => {
  emit("set-instance", {
    title: "instance",
    data: {},
    config: { auto_start: true, auto_renew: true },
    resources: { cpu: 1, ram: 1024, drive_size: 20480, drive_type: "ssd" },
    addons: [],
    billing_plan: {},
  });
});

const bilingPlan = computed(() => instance.value.billing_plan);
const products = computed(() => {
  if (bilingPlan.value?.products) {
    return Object.keys(bilingPlan.value.products).map((key) => ({
      key,
      title: bilingPlan.value.products[key].title,
    }));
  }
  return [];
});

const changeProduct = (val) => {
  setValue("product", val);
  const p = bilingPlan.value?.products?.[val];
  if (p?.resources) {
    Object.keys(p.resources).forEach((k) =>
      setValue(`resources.${k}`, p.resources[k])
    );
  }
};

const setValue = (key, value) => {
  emit("set-value", { key, value });
};
</script>

<script>
export default { name: "instance-proxmox-create" };
</script>
