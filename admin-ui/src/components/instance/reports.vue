<template>
  <reports_table
    table-name="instance-reports"
    hide-account
    hide-instance
    hide-service
    :filters="filters"
  />
</template>

<script setup>
import { toRefs, computed } from "vue";
import Reports_table from "@/components/reports_table.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const filters = computed(() => {
  return {
    instance: [template.value.uuid],
  };
});
</script>

<script>
export default {
  name: "instance-reports",
};
</script>

<style scoped lang="scss"></style>
