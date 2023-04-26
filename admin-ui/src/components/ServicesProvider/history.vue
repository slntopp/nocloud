<template>
  <history-table hide-uuid table-name="service-history" :uuid="template.uuid" />
</template>

<script>
import historyTable from "@/components/historyTable.vue";

export default {
  props: ["template"],
  name: "sp-history",
  components: { historyTable },
};
</script>

<style scoped></style>
