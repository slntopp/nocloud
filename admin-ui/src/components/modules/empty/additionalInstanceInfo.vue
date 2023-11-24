<template>
  <div class="mb-15">
    <v-row justify="start" align="center">
      <v-card-title>Description</v-card-title>
      <v-btn @click="goToPlan" small>edit</v-btn>
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
import { useRouter } from "vue-router/composables";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const router = useRouter();

const goToPlan = () => {
  router.push({
    name: "Plan",
    params: { planId: template.value.billingPlan.uuid },
  });
};
</script>
<style scoped></style>
