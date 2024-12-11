<template>
  <v-chip :color="promocodeConditionColor">{{
    promocodeConditionTitle
  }}</v-chip>
</template>

<script setup>
import { PromocodeCondition } from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";
import { computed, toRefs } from "vue";

const props = defineProps(["item"]);
const { item } = toRefs(props);

const condition = computed(
  () => PromocodeCondition[item.value.condition || "USABLE"]
);

const promocodeConditionColor = computed(() => {
  switch (condition.value) {
    case PromocodeCondition.USABLE:
      return "success";
    case PromocodeCondition.ALL_TAKEN:
    case PromocodeCondition.USED:
      return "warning";
    case PromocodeCondition.EXPIRED:
    default:
      return "blue-grey darken-2";
  }
});

const promocodeConditionTitle = computed(() => {
  switch (condition.value) {
    case PromocodeCondition.USABLE:
      return "USABLE";
    case PromocodeCondition.EXPIRED:
      return "EXPIRED";
    case PromocodeCondition.ALL_TAKEN:
    default:
      return "ALL TAKEN";
  }
});
</script>
