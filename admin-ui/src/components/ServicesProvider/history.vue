<template>
  <history-table hide-uuid table-name="sp-history" :uuid="template.uuid" />
</template>

<script>
import historyTable from "@/components/historyTable.vue";
import search from "@/mixins/search";

export default {
  props: ["template"],
  name: "sp-history",
  mixins:[search({name:'sp-history'})],
  components: { historyTable },
};
</script>

<style scoped></style>
