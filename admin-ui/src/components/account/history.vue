<template>
  <history-table hide-requestor table-name="accounts-history" :account-id="account.uuid" />
</template>

<script>
import HistoryTable from "@/components/historyTable.vue";
import search from "@/mixins/search";

export default {
  props: ["account"],
  name: "account-history",
  mixins:[search('accounts-history')],
  components: { HistoryTable },
};
</script>

<style scoped></style>
