<template>
  <accounts-table :custom-filter="customFilter" no-search />
</template>

<script setup>
import AccountsTable from "@/components/accounts_table.vue";
import { computed, toRefs } from "vue";

const props = defineProps(["accountGroup"]);
const { accountGroup } = toRefs(props);

const customFilter = computed(() => {
  return {
    account_groups: [accountGroup.value?.uuid],
  };
});
</script>

<script>
export default {
  name: "account-group-accounts",
};
</script>

<style scoped></style>
