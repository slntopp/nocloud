<template>
  <v-autocomplete
      placeholder="Start typing..."
      :items="allAccounts"
      :filter="accountsFilter"
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
import api from "@/api";
import { debounce } from "@/functions";

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
});
const { value, fetchValue, multiple, loading, returnObject } = toRefs(props);

const emit = defineEmits(["input"]);

const allAccounts = ref([]);
const isLoading = ref(false);
const isInitLoading = ref(false);

onMounted(() => {
  fetchAccount();

  updateAccounts("");
});

const updateAccounts = async (value) => {
  if (allAccounts.value.find((acc) => acc.title === value) || value === null) {
    isLoading.value = false;
    return;
  }

  try {
    const data = await api.post("accounts", {
      page: 1,
      limit: 25,
      filters: {
        search_param: value,
      },
    });
    allAccounts.value.push(...data.pool);
  } finally {
    isLoading.value = false;
  }
};

const updateAccountsDebounce = debounce(updateAccounts, 300);

const accountsFilter = (item) => {
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

  allAccounts.value = allAccounts.value.filter(
      (account) =>
          value.value?.includes?.(account) ||
          value.value?.includes?.(account.uuid) ||
          value.value === account ||
          value.value?.uuid === account.uuid
  );

  updateAccountsDebounce(e);
};

const fetchAccount = async () => {
  if (fetchValue.value && value.value) {
    isInitLoading.value = true;
    try {
      if (Array.isArray(value.value)) {
        allAccounts.value = await Promise.all([
          ...value.value.map((uuid) => api.accounts.get(uuid)),
        ]);
      } else {
        allAccounts.value = [
          await api.accounts.get(value.value?.uuid || value.value),
        ];
        if (returnObject.value) {
          emit("input", allAccounts.value[0]);
        }
      }
    } finally {
      isInitLoading.value = false;
    }
  }
};

watch(loading, () => {
  if (!loading.value) {
    fetchAccount();
  }
});
</script>
