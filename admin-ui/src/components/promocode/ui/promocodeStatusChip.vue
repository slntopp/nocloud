<template>
  <v-chip :color="promocodeStatusColor">{{ promocodeStatusTitle }}</v-chip>
</template>

<script setup>
import { PromocodeStatus } from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";
import { computed, toRefs } from "vue";

const props = defineProps(["item"]);
const { item } = toRefs(props);

const status = computed(() => PromocodeStatus[item.value.status || "ACTIVE"]);

const promocodeStatusColor = computed(() => {
  switch (status.value) {
    case PromocodeStatus.ACTIVE:
      return "success";
    case PromocodeStatus.SUSPENDED:
      return "warning";
    case PromocodeStatus.DELETED:
    default:
      return "blue-grey darken-2";
  }
});

const promocodeStatusTitle = computed(() => {
  switch (status.value) {
    case PromocodeStatus.ACTIVE:
      return "ACTIVE";
    case PromocodeStatus.SUSPENDED:
      return "SUSPENDED";
    case PromocodeStatus.DELETED:
    default:
      return "DELETED";
  }
});
</script>
