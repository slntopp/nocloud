<template>
  <promocodes-table
    no-search
    table-name="showcase-promocodes"
    :custom-filter="{ showcase: [template.uuid] }"
  />
</template>

<script>
import promocodesTable from "@/components/promocodesTable.vue";

export default {
  props: ["template"],
  name: "showcase-promocodes",
  components: { promocodesTable },
};
</script>

<style scoped></style>
