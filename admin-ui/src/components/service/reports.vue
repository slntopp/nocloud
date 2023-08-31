<template>
  <reports_table hide-account hide-service :filters="filters" />
</template>

<script setup>
import { toRefs, computed } from "vue";
import Reports_table from "@/components/reports_table.vue";

const props = defineProps(["service"]);
const { service } = toRefs(props);

const filters = computed(() => {
  return {
    service: [service.value.uuid],
  };
});
</script>

<script>
export default {
  name: "service-reports",
};
</script>

<style scoped lang="scss"></style>
