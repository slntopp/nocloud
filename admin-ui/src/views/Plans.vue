<template>
  <div class="pa-4">
    <div class="d-flex align-center">
      <v-btn class="mr-2" :to="{ name: 'Plans create' }"> Create </v-btn>
      <delete-plans-dialog @delete="refetchPlans" :plans="selected" />

      <v-switch label="Individual" v-model="isIndividual" />
    </div>

    <plans-table
      :refetch="refetch"
      show-select
      :custom-params="{
        showDeleted: true,
        anonymously: false,
        filters: { 'meta.isIndividual': [isIndividual] },
      }"
      v-model="selected"
      @fetch:plans="fetchPlans"
      :plans="plans"
      :total="total"
      :isLoading="isLoading"
    />
  </div>
</template>

<script setup>
import plansTable from "@/components/plansTable.vue";
import deletePlansDialog from "@/components/deletePlansDialog.vue";
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";

const store = useStore();

const isIndividual = ref(false);
const refetch = ref(false);
const selected = ref([]);

const refetchPlans = () => {
  refetch.value = !refetch.value;
  selected.value = [];
};

const fetchPlans = (options) => {
  return store.dispatch("plans/fetch", options);
};

const plans = computed(() => store.getters["plans/all"]);
const total = computed(() => store.getters["plans/total"]);
const isLoading = computed(() => store.getters["plans/loading"]);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: refetchPlans,
  });
});
</script>

<script>
export default {
  name: "plans-view",
};
</script>

<style scoped>
.file-input {
  max-width: 300px;
  min-width: 200px;
  margin-top: 0;
  padding-top: 0;
}
</style>
