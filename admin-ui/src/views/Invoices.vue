<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5 justify-space-between">
      <div class="d-flex align-center">
        <v-btn
          v-for="layout in Object.keys(defaultLayouts)"
          class="mr-1"
          :key="layout"
          :disabled="defaultLayouts[layout].id === currentSearchLayout"
          @click="setInvoicesLayout(layout)"
        >
          <v-icon small class="mr-1">mdi-filter</v-icon>
          {{ layout }}</v-btn
        >
      </div>
      <div class="d-flex align-center">
        <v-btn class="mr-1" :to="{ name: 'Invoice create' }"> Create </v-btn>

        <confirm-dialog>
          <v-btn class="mr-1" disabled>Merge</v-btn>
        </confirm-dialog>

        <confirm-dialog
          :disabled="isCopyDisabled"
          :loading="isCopyLoading"
          @confirm="handleCopyInvoice"
        >
          <v-btn
            :disabled="isCopyDisabled"
            :loading="isCopyLoading"
            class="mr-8"
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
              button.disabled ||
              !selectedInvoices.length
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
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import InvoicesTable from "@/components/invoicesTable.vue";
import {
  BillingStatus,
  UpdateInvoiceStatusRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import confirmDialog from "@/components/confirmDialog.vue";

const selectedInvoices = ref([]);
const refetch = ref(false);

const isCopyLoading = ref(false);

const isUpdateStatusLoading = ref(false);
const updateStatusName = ref("");

const store = useStore();

const isLoading = computed(() => store.getters["invoices/isLoading"]);
const invoices = computed(() => store.getters["invoices/all"]);

const isCopyDisabled = computed(() => selectedInvoices.value.length !== 1);

const changeStatusBtns = computed(() => [
  {
    title: "paid",
    status: "PAID",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "DRAFT", "RETURNED"].includes(invoice.status)
    ),
  },
  {
    title: "unpaid",
    status: "UNPAID",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "RETURNED"].includes(invoice.status)
    ),
  },
  {
    title: "cancel",
    status: "CANCELED",
    disabled: selectedInvoices.value.some((invoice) =>
      ["TERMINATED", "RETURNED", "DRAFT", "PAID"].includes(invoice.status)
    ),
  },
  {
    title: "terminate",
    status: "TERMINATED",
    disabled: false,
  },
]);

const defaultLayouts = computed(() => ({
  unpaid: {
    title: "Unpaid",
    filter: { status: [2] },
    fields: ["status"],
    id: "id-unpaid",
  },
  overdue: {
    id: "id-overdue",
    title: "Overdue",
    fields: ["deadline", "status"],
    filter: {
      status: [2],
      deadline: [
        new Date(Date.now() - 86000 * 365 * 1000).toISOString().split("T")[0],
        new Date(Date.now()).toISOString().split("T")[0],
      ],
    },
  },
}));

const currentSearchLayout = computed(
  () => store.getters["appSearch/currentLayout"]
);

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
    await store.dispatch["invoices/copy"](selectedInvoices.value[0]);

    refetchInvoices();
    selectedInvoices.value = [];
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isCopyLoading.value = false;
  }
};

const setDefaultLayouts = () => {
  const defaults = Object.values(defaultLayouts.value);
  const layouts = JSON.parse(
    JSON.stringify(store.getters["appSearch/layouts"])
  );

  defaults.forEach((layout) => {
    const index = layouts.findIndex((l) => l.title === layout.title);
    if (index == -1) {
      layouts.push(layout);
    } else {
      layouts[index] = layout;
    }
  });

  store.commit("appSearch/setLayouts", layouts);
};

const setInvoicesLayout = (key) => {
  store.commit("appSearch/setCurrentLayout", defaultLayouts.value[key].id);
};

onMounted(() => {
  setTimeout(() => setDefaultLayouts(), 500);
});
</script>

<script>
export default { name: "invoices-view" };
</script>
<style scoped></style>
