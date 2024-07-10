<template>
  <history-table hide-uuid table-name="plan-history" :uuid="template.uuid" />
</template>

<script>
import historyTable from "@/components/historyTable.vue";
import search from "@/mixins/search";

export default {
  props: ["template"],
  name: "plan-history",
  mixins:[search({name:'plan-history'})],
  components: { historyTable },
};
</script>

<style scoped></style>
