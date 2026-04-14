<template>
  <div class="d-flex align-center">
    <v-btn
      class="mr-1"
      :to="{ name: 'Invoice create', query: { account: accountUuid } }"
    >
      Create
    </v-btn>

    <confirm-dialog disabled>
      <v-btn class="mr-1" disabled>Merge</v-btn>
    </confirm-dialog>

    <confirm-dialog
      :disabled="isCopyDisabled"
      :loading="isCopyLoading"
      @confirm="handleCopyInvoice"
    >
      <v-btn :disabled="isCopyDisabled" :loading="isCopyLoading" class="mr-1"
        >Copy</v-btn
      >
    </confirm-dialog>

    <confirm-dialog
      v-if="isKsefEnabled"
      :disabled="isKsefDisabled"
      :loading="isKsefLoading"
      @confirm="handleKsefEnqueue"
    >
      <v-btn :disabled="isKsefDisabled" :loading="isKsefLoading" class="mr-8"
        >Ksef enqueue</v-btn
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
          button.disabled ||
          !selectedInvoices.length
        "
        :loading="isUpdateStatusLoading && updateStatusName === button.status"
        class="mr-1"
      >
        {{ button.title }}
      </v-btn>
    </confirm-dialog>
  </div>
</template>

<script setup>
import confirmDialog from "@/components/confirmDialog.vue";
import {
  BillingStatus,
  UpdateInvoiceStatusRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import { computed, ref, toRefs } from "vue";
import { useRouter } from "vue-router/composables";
import { useStore } from "@/store";

const props = defineProps({
  selectedInvoices: {
    type: Array,
    required: true,
  },
  accountUuid: {
    type: String,
    required: false,
  },
});
const { selectedInvoices, accountUuid } = toRefs(props);

const emit = defineEmits(["input", "refresh"]);

const router = useRouter();
const store = useStore();

const isCopyLoading = ref(false);
const isKsefLoading = ref(false);

const isUpdateStatusLoading = ref(false);
const updateStatusName = ref("");

const isCopyDisabled = computed(() => selectedInvoices.value.length !== 1);
const isKsefDisabled = computed(() => selectedInvoices.value.length < 1);

const changeStatusBtns = computed(() => [
  {
    title: "paid",
    status: "PAID",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "DRAFT", "RETURNED"].includes(invoice.status),
    ),
  },
  {
    title: "unpaid",
    status: "UNPAID",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "RETURNED"].includes(invoice.status),
    ),
  },
  {
    title: "cancel",
    status: "CANCELED",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "RETURNED", "DRAFT", "PAID"].includes(invoice.status),
    ),
  },
  {
    title: "terminate",
    status: "TERMINATED",
    disabled: false,
  },
]);

const isKsefEnabled = computed(() => store.getters["settings/ksefEnabled"]);

const refetchInvoices = () => {
  emit("refresh");
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
          }),
        );
      }),
    );

    emit("input", []); // Clear selection
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
    const data = await store.dispatch(
      "invoices/copy",
      selectedInvoices.value[0],
    );
    router.push({ name: "Invoice page", params: { uuid: data.uuid } });

    refetchInvoices();
    emit("input", []); // Clear selection
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isCopyLoading.value = false;
  }
};

const handleKsefEnqueue = async () => {
  isKsefLoading.value = true;

  try {
    await Promise.all(
      selectedInvoices.value.map((invoice) => {
        return store.getters["invoices/invoicesClient"].ksefEnqueue({
          invoiceUuid: invoice.uuid,
        });
      }),
    );

    refetchInvoices();
    emit("input", []); // Clear selection

    store.commit("snackbar/showSnackbarSuccess", {
      message: "KSeF enqueue request sent",
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: "Failed to send KSeF enqueue requests",
    });
  } finally {
    isKsefLoading.value = false;
  }
};
</script>
