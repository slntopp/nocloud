<template>
  <reports_table
    table-name="instance-reports"
    hide-account
    hide-instance
    hide-service
    :filters="filters"
    show-dates
    :duration="duration"
    @input:duration="duration = $event"
  />
</template>

<script setup>
import { toRefs, computed, ref } from "vue";
import Reports_table from "@/components/reports_table.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const duration = ref({ to: null, from: null });

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
