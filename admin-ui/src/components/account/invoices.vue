<template>
  <invoices-table
    table-name="account-invoices"
    no-search
    hide-select
    :custom-filter="{ account: [account.uuid] }"
  />
</template>

<script setup>
import { toRefs } from "vue";
import InvoicesTable from "../invoicesTable.vue";

const props = defineProps(["account"]);
const { account } = toRefs(props);
</script>

<script>
export default {
  name: "account-reports",
};
</script>

<style scoped lang="scss"></style>
