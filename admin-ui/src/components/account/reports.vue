<template>
  <reports_table
    table-name="reports"
    hide-account
    :duration="duration"
    @input:duration="duration = $event"
    :filters="filters"
    hide-service
    show-dates
  />
</template>

<script setup>
import { toRefs, computed, ref } from "vue";
import Reports_table from "@/components/reports_table.vue";

const props = defineProps(["account"]);
const { account } = toRefs(props);

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
    account: [account.value.uuid],
  };
});
</script>

<script>
export default {
  name: "account-reports",
};
</script>

<style scoped lang="scss"></style>
