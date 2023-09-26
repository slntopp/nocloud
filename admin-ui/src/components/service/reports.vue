<template>
  <reports_table
    table-name="service-reports"
    hide-account
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

const props = defineProps(["service"]);
const { service } = toRefs(props);

const duration = ref({ to: null, from: null });

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
