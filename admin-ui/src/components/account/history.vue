<template>
  <history-table hide-requestor table-name="accounts-history" :account-id="account.uuid" />
</template>

<script>
import HistoryTable from "@/components/historyTable.vue";

export default {
  props: ["account"],
  name: "account-history",
  components: { HistoryTable },
};
</script>

<style scoped></style>
