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
    exec: {
      to: duration.value.to
        ? new Date(duration.value.to).getTime() / 1000
        : undefined,
      from: duration.value.from
        ? new Date(duration.value.from).getTime() / 1000
        : undefined,
    },
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
