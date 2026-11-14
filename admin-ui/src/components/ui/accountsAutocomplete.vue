<template>
  <v-autocomplete
    placeholder="Start typing..."
    :items="allAccounts"
    :customFilter="accountsFilter"
    :model-value="fieldValue"
    @update:model-value="emit('update:modelValue', $event)"
    item-title="title"
    item-value="uuid"
    :rules="rules"
    :disabled="disabled"
    :label="label"
    :return-object="returnObject"
    @update:search="onSearchInput"
    :loading="isLoading || isInitLoading || loading"
    :multiple="multiple"
    :clearable="clearable"
    :density="dense ? 'compact' : undefined"
  >
    <template v-if="advanced" v-slot:selection="{ internalItem: item }">
      {{ getAdvancedItem(item.raw) }}
    </template>

    <template v-if="advanced" v-slot:item="{ internalItem: item, props: itemProps }">
      <v-list-item v-bind="itemProps" :title="getAdvancedItem(item.raw)" />
    </template>
  </v-autocomplete>
</template>

<script setup>
import { toRefs, ref, onMounted, watch, computed } from "vue";
import api from "@/api";
import { debounce } from "@/functions";

const props = defineProps({
  modelValue: {},
  label: {},
  rules: {},
  returnObject: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  multiple: { type: Boolean, default: false },
  clearable: { type: Boolean, default: false },
  fetchValue: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  advanced: { type: Boolean, default: false },
});
const { modelValue: value, fetchValue, multiple, loading, returnObject } = toRefs(props);

// Vuetify 3+ return-object does not match a uuid string against the item
// objects (Vuetify 2 did), so map uuid strings to the loaded items for display.
const fieldValue = computed(() => {
  if (!returnObject.value) return value.value;
  const resolve = (v) =>
    typeof v === "string" ? (allAccounts.value.find((i) => i.uuid === v) ?? v) : v;
  return Array.isArray(value.value) ? value.value.map(resolve) : resolve(value.value);
});

const emit = defineEmits(["update:modelValue"]);

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
    // skip accounts already in the list (searches can overlap)
    const known = new Set(allAccounts.value.map((acc) => acc.uuid));
    allAccounts.value.push(...data.pool.filter((acc) => !known.has(acc.uuid)));
  } finally {
    isLoading.value = false;
  }
};

const updateAccountsDebounce = debounce(updateAccounts, 300);

// Vuetify 3 custom-filter signature: (value, query, internalItem)
const accountsFilter = (_, __, internalItem) => {
  const item = internalItem?.raw;
  if (multiple.value) {
    return !(value.value?.includes(item?.uuid) || value.value?.includes(item));
  }
  return true;
};

// Vuetify 3+ emits update:search also on focus/selection, not only while
// typing (Vuetify 2's update:search-input); ignore repeats of the same text.
let lastSearch = "";
const onSearchInput = async (e) => {
  if (isInitLoading.value || (e ?? "") === lastSearch) {
    return;
  }
  lastSearch = e ?? "";

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
  if (
    fetchValue.value &&
    value.value &&
    (typeof value.value === "string" || Object.keys(value.value).length)
  ) {
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
          emit("update:modelValue", allAccounts.value[0]);
        }
      }
    } finally {
      isInitLoading.value = false;
    }
  }
};

// account may be a plain uuid: Vuetify 3+ also renders values missing from items
const getAdvancedItem = (account) =>
  typeof account !== "object" || !account ? String(account ?? "") : `${account.title} ${
    account.data?.email ? "- " + account.data?.email : ""
  } (${account.uuid.slice(0, 6)})`;

watch(loading, () => {
  if (!loading.value) {
    fetchAccount();
  }
});
</script>
