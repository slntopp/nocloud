<template>
  <v-autocomplete
    placeholder="Start typing..."
    :items="allPlans"
    :filter="plansFilter"
    :value="value"
    @input="emit('input', $event)"
    item-text="title"
    item-value="uuid"
    :rules="rules"
    :disabled="disabled"
    :label="label"
    :return-object="returnObject"
    @update:search-input="onSearchInput"
    :loading="isLoading || isInitLoading || loading"
    :multiple="multiple"
    :clearable="clearable"
    :dense="dense"
  />
</template>

<script setup>
import { toRefs, ref, onMounted, watch } from "vue";
import { debounce } from "@/functions";
import { useStore } from "@/store";
import { ListRequest } from "nocloud-proto/proto/es/billing/billing_pb";

const props = defineProps({
  value: {},
  label: {},
  rules: {},
  returnObject: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  multiple: { type: Boolean, default: false },
  clearable: { type: Boolean, default: false },
  fetchValue: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  customParams: { type: Object, default: () => ({ filters: {} }) },
});
const { value, fetchValue, multiple, loading, returnObject, customParams } =
  toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();

const allPlans = ref([]);
const isLoading = ref(false);
const isInitLoading = ref(false);

onMounted(() => {
  fetchPlan();

  updatePlans("");
});

const updatePlans = async (value) => {
  if (allPlans.value.find((plan) => plan.title === value) || value === null) {
    isLoading.value = false;
    return;
  }

  try {
    const data = await store.getters["plans/plansClient"].listPlans(
      ListRequest.fromJson({
        ...customParams.value,
        page: 1,
        limit: 5,
        filters: {
          ...(customParams.value.filters || {}),
          search_param: value,
        },
      })
    );
    allPlans.value.push(...data.pool);
  } finally {
    isLoading.value = false;
  }
};

const updatePlansDebounce = debounce(updatePlans, 300);

const plansFilter = (item) => {
  if (multiple.value) {
    return !(value.value?.includes(item.uuid) || value.value?.includes(item));
  }
  return true;
};

const onSearchInput = async (e) => {
  if (isInitLoading.value) {
    return;
  }

  isLoading.value = true;

  allPlans.value = allPlans.value.filter(
    (plan) =>
      value.value?.includes?.(plan) ||
      value.value?.includes?.(plan.uuid) ||
      value.value === plan ||
      value.value?.uuid === plan.uuid
  );

  updatePlansDebounce(e);
};

const fetchPlan = async () => {
  if (
    fetchValue.value &&
    value.value &&
    (typeof value.value === "string" || Object.keys(value.value).length)
  ) {
    isInitLoading.value = true;
    try {
      if (Array.isArray(value.value)) {
        allPlans.value = await Promise.all([
          ...value.value.map((uuid) =>
            store.getters["plans/plansClient"].getPlan({
              uuid,
            })
          ),
        ]);
      } else {
        allPlans.value = [
          store.getters["plans/plansClient"].getPlan({
            uuid: value.value?.uuid || value.value,
          }),
        ];
        if (returnObject.value) {
          emit("input", allPlans.value[0]);
        }
      }
    } finally {
      isInitLoading.value = false;
    }
  }
};

watch(loading, () => {
  if (!loading.value) {
    fetchPlan();
  }
});
</script>
