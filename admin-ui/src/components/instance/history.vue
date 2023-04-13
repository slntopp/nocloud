<template>
  <history-table table-name="instance-history" :uuid="template.uuid" />
</template>

<script>
import historyTable from "@/components/historyTable.vue";

export default {
  props: ["template"],
  name: "invoice-history",
  components: { historyTable },
};
</script>

<style scoped></style>
