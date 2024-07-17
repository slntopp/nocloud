<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <div class="d-flex justify-space-between">
      <span> Instances </span>
    </div>
    <instances-table
      :show-select="false"
      :custom-filter="{ billing_plan: [template.uuid] }"
      table-name="plan-instances-table"
    />
  </v-card>
</template>

<script>
import InstancesTable from "@/components/instancesTable.vue";

export default {
  name: "instances-tab",
  components: { InstancesTable },
};
</script>

<script setup>
import { defineProps } from "vue";

defineProps(["template"]);
</script>
