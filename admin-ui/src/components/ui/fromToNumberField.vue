<script setup>
import { computed, toRefs } from "vue";

const props = defineProps({
  value: {},
  label: {},
  disabled: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
});
const { disabled, dense, value, label } = toRefs(props);

const emit = defineEmits(["input"]);

const fromRules = computed(() => {
  if (!value.value?.from) {
    return [];
  }

  return [(v) => !value.value.to || +v < +value.value.to];
});

const toRules = computed(() => {
  if (!value.value?.to) {
    return [];
  }

  return [(v) => !value.value.from || +v > +value.value.from];
});

const onInput = (key, e) => {
  let newValue = { ...value.value };
  if (!newValue?.from && !newValue?.to) {
    newValue = { from: null, to: null };
  }

  newValue[key] = e || undefined;
  emit("input", newValue);
};
</script>

<template>
  <div class="d-flex" style="width: 100%">
    <v-text-field
      class="mr-1"
      @input="onInput('from', $event)"
      type="number"
      :disabled="disabled"
      :dense="dense"
      :label="`${label} from`"
      :value="value?.from"
      :rules="fromRules"
    ></v-text-field>
    <v-text-field
      class="ml-1"
      @input="onInput('to', $event)"
      type="number"
      :disabled="disabled"
      :rules="toRules"
      :dense="dense"
      :label="`${label} to`"
      :value="value?.to"
    ></v-text-field>
  </div>
</template>

<style scoped></style>
