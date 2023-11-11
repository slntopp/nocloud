<template>
  <history-table hide-uuid table-name="service-history" :uuid="template.uuid" />
</template>

<script>
import historyTable from "@/components/historyTable.vue";
import search from "@/mixins/search";

export default {
  props: ["template"],
  mixins: [search("service-history")],
  name: "service-history",
  components: { historyTable },
};
</script>

<style scoped></style>
