<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5 justify-space-between">
      <div class="d-flex align-center">
        <v-btn class="mr-1" :to="{ name: 'Invoice create' }"> Create </v-btn>

        <v-btn class="mr-1 ml-3">overdue</v-btn>
        <v-btn class="mr-1">unpaid</v-btn>
      </div>
      <div class="d-flex align-center">
        <confirm-dialog>
          <v-btn class="mr-1">Merge</v-btn>
        </confirm-dialog>

        <confirm-dialog
          :disabled="isCopyDisabled"
          :loading="isCopyLoading"
          @confirm="handleCopyInvoice"
        >
          <v-btn
            :disabled="isCopyDisabled"
            :loading="isCopyLoading"
            class="mr-5"
            >Copy</v-btn
          >
        </confirm-dialog>

        <confirm-dialog
          v-for="button in changeStatusBtns"
          :key="button.status"
          :disabled="
            (isUpdateStatusLoading && updateStatusName !== button.status) ||
            button.disabled
          "
          @confirm="handleUpdateStatus(button.status)"
        >
          <v-btn
            :disabled="
              (isUpdateStatusLoading && updateStatusName !== button.status) ||
              button.disabled
            "
            :loading="
              isUpdateStatusLoading && updateStatusName === button.status
            "
            class="mr-1"
          >
            {{ button.title }}
          </v-btn>
        </confirm-dialog>
      </div>
    </div>
    <invoices-table
      v-model="selectedInvoices"
      :loading="isLoading"
      :refetch="refetch"
      :items="invoices"
    />
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useStore } from "@/store";
import InvoicesTable from "@/components/invoicesTable.vue";
import {
  BillingStatus,
  CreateInvoiceRequest,
  UpdateInvoiceStatusRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import confirmDialog from "@/components/confirmDialog.vue";

const selectedInvoices = ref([]);
const refetch = ref(false);

const isCopyLoading = ref(false);

const isUpdateStatusLoading = ref(false);
const updateStatusName = ref("");

const store = useStore();

const invoicesClient = computed(() => store.getters["invoices/invoicesClient"]);

const isLoading = computed(() => store.getters["invoices/isLoading"]);
const invoices = computed(() => store.getters["invoices/all"]);

const isCopyDisabled = computed(() => selectedInvoices.value.length !== 1);

const changeStatusBtns = computed(() => [
  {
    title: "paid",
    status: "PAID",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "CANCELED", "DRAFT", "RETURNED"].includes(invoice.status)
    ),
  },
  {
    title: "unpaid",
    status: "UNPAID",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "CANCELED", "RETURNED"].includes(invoice.status)
    ),
  },
  {
    title: "canceled",
    status: "CANCELED",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "RETURNED", "DRAFT", "PAID"].includes(invoice.status)
    ),
  },
  {
    title: "terminated",
    status: "TERMINATED",
    disabled: false,
  },
]);

const refetchInvoices = () => {
  refetch.value = !refetch.value;
};

const handleUpdateStatus = async (newStatus) => {
  isUpdateStatusLoading.value = true;
  updateStatusName.value = newStatus;

  try {
    await Promise.all(
      selectedInvoices.value.map((invoice) => {
        if (invoice.status === newStatus) {
          return;
        }

        return store.getters["invoices/invoicesClient"].updateInvoiceStatus(
          UpdateInvoiceStatusRequest.fromJson({
            uuid: invoice.uuid,
            status: BillingStatus[newStatus],
          })
        );
      })
    );
    selectedInvoices.value = [];
    refetchInvoices();
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isUpdateStatusLoading.value = false;
    updateStatusName.value = "";
  }
};

const handleCopyInvoice = async () => {
  isCopyLoading.value = true;

  try {
    const data = {
      items: selectedInvoices.value[0].items,
      total: selectedInvoices.value[0].total,
      account: selectedInvoices.value[0].account,
      type: selectedInvoices.value[0].type,
      deadline: selectedInvoices.value[0].deadline,
      meta: selectedInvoices.value[0].meta,
      status: "DRAFT",
    };

    await invoicesClient.value.createInvoice(
      CreateInvoiceRequest.fromJson({
        invoice: data,
        isSendEmail: false,
      })
    );

    refetchInvoices();
    selectedInvoices.value = [];
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isCopyLoading.value = false;
  }
};
</script>

<script>
export default { name: "invoices-view" };
</script>
<style scoped></style>
