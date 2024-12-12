<template>
  <promocodes-table
    no-search
    table-name="plan-promocodes"
    :custom-filter="{ plan: [template.uuid] }"
  />
</template>

<script>
import promocodesTable from "@/components/promocodesTable.vue";

export default {
  props: ["template"],
  name: "plan-promocodes",
  components: { promocodesTable },
};
</script>

<style scoped></style>
