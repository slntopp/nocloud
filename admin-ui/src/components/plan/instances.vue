<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <div class="d-flex justify-space-between">
      <span> Instances </span>
      <v-btn
        :disabled="isLoading"
        @click="updateAllServices"
        :loading="isUpdateAllLoading"
        >Update all</v-btn
      >
    </div>
    <instances-table :value="selected" :items="instances" />
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
import { computed, onMounted, defineProps, ref } from "vue";
import api from "@/api";
const props = defineProps(["template"]);

const store = useStore();

const selected = ref([]);
const isUpdateAllLoading = ref(false);

const instances = computed(() =>
  store.getters["services/getInstances"].filter(
    (i) => i.billingPlan.uuid === props.template?.uuid
  )
);
const services = computed(() => store.getters["services/all"]);
const isLoading = computed(() => store.getters["services/isLoading"]);

const updateAllServices = async () => {
  const toUpdate = services.value.filter((s) => {
    if (
      s.instancesGroups.find((ig) =>
        ig.instances.find(({ uuid }) =>
          instances.value.find((i) => i.uuid === uuid)
        )
      )
    ) {
      return s;
    }
  });

  isUpdateAllLoading.value = true;
  try {
    await Promise.all(toUpdate.map((s) => api.services._update(s)));
  } catch (err) {
    store.commit("snackbar/showSnackbarError", { message: err });
  } finally {
    isUpdateAllLoading.value = false;
  }
};

onMounted(() => {
  store.dispatch("services/fetch", { showDeleted: true });
  store.dispatch("accounts/fetch");
  store.dispatch("namespaces/fetch");
});
</script>
