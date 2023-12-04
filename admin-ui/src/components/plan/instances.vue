<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    Instances
    <instances-table :value="selected" :items ="instances" />
  </v-card>
</template>

<script>
import InstancesTable from "@/components/instances_table.vue";

export default {
  name: "instances-tab",
  components: { InstancesTable },
};
</script>

<script setup>
import { useStore } from "@/store";
import { computed, onMounted ,defineProps,ref} from "vue";
const props = defineProps(["template"]);

const store = useStore();
const selected=ref([])

const instances = computed(() =>
  store.getters["services/getInstances"].filter(
    (i) => i.billingPlan.uuid === props.template?.uuid
  )
);

onMounted(async () => {
  store.dispatch("services/fetch",{showDeleted:true});
});
</script>