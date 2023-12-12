<template>
  <history-table hide-uuid table-name="instance-history" :uuid="template.uuid" />
</template>

<script>
import historyTable from "@/components/historyTable.vue";
import search from "@/mixins/search";

export default {
  props: ["template"],
  name: "invoice-history",
  mixins:[search({name:'instance-history'})],
  components: { historyTable },
};
</script>

<style scoped></style>
