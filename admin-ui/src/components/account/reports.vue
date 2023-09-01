<template>
  <reports_table hide-account :filters="filters" hide-service />
</template>

<script setup>
import { toRefs, computed } from "vue";
import Reports_table from "@/components/reports_table.vue";

const props = defineProps(["account"]);
const { account } = toRefs(props);

const filters = computed(() => {
  return {
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
