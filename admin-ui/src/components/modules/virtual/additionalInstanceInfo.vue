<template>
  <div class="mb-15">
    <v-row>
      <v-card-title>Description</v-card-title>
    </v-row>
    <v-row>
      <rich-editor
        class="pa-5"
        disabled
        :value="
          template.billingPlan.products[template.product].meta?.description
        "
      />
    </v-row>
  </div>
</template>

<script setup>
import { toRefs } from "vue";
import RichEditor from "@/components/ui/richEditor.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);
</script>
<style scoped></style>
