<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5">
      <v-btn class="mr-2" :to="{ name: 'Invoice create' }"> Create </v-btn>
    </div>
    <invoices-table :loading="isLoading" :items="invoices" />
  </div>
</template>

<script setup>
import { computed, onMounted } from "vue";
import { useStore } from "@/store";
import InvoicesTable from "@/components/invoicesTable.vue";

const store = useStore();

onMounted(() => {
  fetchInvoices();

  store.commit("reloadBtn/setCallback", { event: fetchInvoices });
});

const isLoading = computed(() => store.getters["invoices/isLoading"]);
const invoices = computed(() => store.getters["invoices/all"]);

const fetchInvoices = () => {
  store.dispatch("invoices/fetch");
  store.dispatch("accounts/fetch");
};
</script>

<script>
export default { name: "invoices-view" };
</script>
<style scoped></style>
