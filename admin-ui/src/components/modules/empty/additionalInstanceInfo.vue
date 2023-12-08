<template>
  <v-expansion-panels class="mb-15">
    <v-expansion-panel>
      <v-expansion-panel-header color="background-light">
        <div>
          <span style="color: var(--v-primary-base)" class="text-h6">Description
          </span>
          <v-btn @click="goToPlan" small>edit</v-btn>

        </div>
        <template v-slot:actions>
          <v-icon color="primary" x-large> $expand </v-icon>
        </template>
      </v-expansion-panel-header>
      <v-expansion-panel-content color="background-light">
        <rich-editor class="pa-5" disabled :value="template.billingPlan.products[template.product].meta?.description
          " />
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
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
