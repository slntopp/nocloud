<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5">
      <v-btn class="mr-2" :to="{ name: 'Invoice create' }"> Create </v-btn>
      <v-btn
        :disabled="isBtnDisabled"
        :loading="isCancelLoading"
        class="mr-2"
        @click="cancelInvoices"
      >
        Cancel
      </v-btn>
      <v-btn
        :disabled="isBtnDisabled"
        :loading="isTerminateLoading"
        class="mr-2"
        @click="terminateInvoices"
      >
        Terminate
      </v-btn>
    </div>
    <invoices-table
      v-model="selectedInvoices"
      :loading="isLoading"
      :items="invoices"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import InvoicesTable from "@/components/invoicesTable.vue";
import api from "@/api";

const selectedInvoices = ref([]);
const isCancelLoading = ref(false);
const isTerminateLoading = ref(false);

const store = useStore();

onMounted(() => {
  fetch();

  store.commit("reloadBtn/setCallback", { event: fetch });
});

const isBtnDisabled = computed(
  () => isCancelLoading.value || isTerminateLoading.value
);

const isLoading = computed(() => store.getters["invoices/isLoading"]);
const invoices = computed(() => store.getters["invoices/all"]);

const fetch = () => {
  store.dispatch("accounts/fetch");
};

const cancelInvoices = async () => {
  try {
    isCancelLoading.value = true;
    await Promise.all(
      selectedInvoices.value.map((invoice) =>
        changeInvoiceStatus(invoice, "CANCELED")
      )
    );
    selectedInvoices.value = [];
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isCancelLoading.value = false;
  }
};

const terminateInvoices = async () => {
  try {
    isTerminateLoading.value = true;
    await Promise.all(
      selectedInvoices.value.map((invoice) =>
        changeInvoiceStatus(invoice, "TERMINATED")
      )
    );
    selectedInvoices.value = [];
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isTerminateLoading.value = false;
  }
};

const changeInvoiceStatus = async (invoice, status) => {
  return api.patch("/billing/invoices/" + invoice.uuid, {
    ...invoice,
    status,
  });
};
</script>

<script>
export default { name: "invoices-view" };
</script>
<style scoped></style>
