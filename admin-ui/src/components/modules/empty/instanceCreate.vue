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
            @change="setValue('title', $event.target.value)"
            label="Name"
            :model-value="instance.title"
          >
          </v-text-field>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <plans-autocomplete
            :model-value="bilingPlan"
            :custom-params="{
              filters: { type: ['empty'], 'meta.isIndividual': [false] },
              anonymously: false,
            }"
            @update:model-value="setValue('billing_plan', $event)"
            return-object
            label="Price model"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Product"
            :rules="requiredRule"
            :model-value="instance.product"
            v-if="products.length > 0"
            :items="products"
            item-title="key"
            item-value="key"
            @update:model-value="changeProduct"
          >
            <template v-slot:item="{ internalItem: item, props: itemProps }">
              <v-list-item v-bind="itemProps">
                <template #title>
                  <div
                    style="width: 100%"
                    class="d-flex justify-space-between align-center"
                  >
                    <span>{{ item.raw.key }}</span>
                    <span class="ml-4">{{ item.raw.title }}</span>
                  </div>
                </template>
              </v-list-item>
            </template>
          </v-autocomplete>
        </v-col>

        <v-col cols="6" v-if="tarrifAddons.length > 0">
          <v-autocomplete
            @update:model-value="(newVal) => setValue('addons', newVal)"
            label="Addons"
            :model-value="instance.addons"
            :items="isAddonsLoading ? [] : getAvailableAddons()"
            :loading="isAddonsLoading"
            item-value="uuid"
            item-title="title"
            multiple
          />
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import useInstanceAddons from "@/hooks/useInstanceAddons";
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";

const props = defineProps(["instance", "planRules", "spUuid"]);
const { instance, planRules } = toRefs(props);

const emit = defineEmits(["set-instance", "set-value"]);

const { tarrifAddons, setTariffAddons, getAvailableAddons, isAddonsLoading } =
  useInstanceAddons(instance, (key, value) => setValue(key, value));

const getDefaultInstance = () => ({
  title: "instance",
  data: {},
  config: {},
  addons: [],
  billing_plan: {},
});

const product = ref([]);
const requiredRule = ref([(val) => !!val || "Field required"]);

onMounted(() => {
  emit("set-instance", getDefaultInstance());
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
  product.value = val;
  setValue("product", product.value);

  setTariffAddons();
};
const setValue = (key, value) => {
  emit("set-value", { key, value });
};
</script>

<script>
export default {
  name: "instance-empty-create",
};
</script>

<style scoped></style>
