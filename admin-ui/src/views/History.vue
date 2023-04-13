<template>
  <v-container>
    <history-table />
  </v-container>
</template>

<script>
import HistoryTable from "@/components/historyTable.vue";

export default {
  name: "all-history",
  components: { HistoryTable },
};
</script>

<style scoped></style>
