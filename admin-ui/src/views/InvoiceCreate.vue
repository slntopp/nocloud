<template>
  <div class="pa-4">
    <h1 v-if="!isEdit && isDraft" class="page__title">Create invoice</h1>

    <div
      v-if="isEdit"
      style="z-index: 0; position: relative; top: -15px; right: 20px"
    >
      <div class="d-flex justify-end mt-1 align-center flex-wrap">
        <v-chip class="mr-2" :color="getTotalColor(newInvoice)">
          Total:
          {{ `${newInvoice.total} ${newInvoice.currency?.title}` }}
        </v-chip>

        <v-chip :color="getInvoiceStatusColor(newInvoice.status)">
          {{ `${newInvoice.status}` }}
        </v-chip>
      </div>
    </div>

    <v-form v-model="isValid" ref="invoiceForm">
      <div class="invoice__container">
        <div class="item" style="width: 150px">
          <v-select
            :disabled="isEdit"
            item-value="id"
            item-text="title"
            label="Type"
            v-model="newInvoice.type"
            :items="types"
          >
          </v-select>
        </div>

        <div class="item d-flex">
          <accounts-autocomplete
            :loading="isInstancesLoading"
            @input="onChangeAccount"
            :disabled="isEdit"
            label="Account"
            v-model="newInvoice.account"
            fetch-value
            return-object
            :rules="requiredRule"
          />
          <v-btn
            style="margin-top: 20px"
            @click="openAccountWindow"
            icon
            v-if="isEdit && newInvoice.account && !isInstancesLoading"
          >
            <v-icon>mdi-login</v-icon>
          </v-btn>
        </div>

        <div class="item">
          <v-autocomplete
            :disabled="isEdit || isBalanceInvoice"
            :filter="defaultFilterObject"
            label="Instances"
            v-model="newInvoice.instances"
            multiple
            item-text="title"
            item-value="uuid"
            :items="instances"
            @change="onChangeInstance"
            :loading="isInstancesLoading"
          />
        </div>

        <div class="item" v-if="isBalanceInvoice">
          <v-text-field
            type="number"
            label="Total"
            :value="newInvoice.total"
            :suffix="accountCurrency?.title"
            @input="newInvoice.items[0].price = +$event"
          />
        </div>

        <div class="item date">
          <date-picker
            :min="formatSecondsToDateString(Date.now() / 1000)"
            label="Due date"
            v-model="newInvoice.deadline"
          />
        </div>

        <template v-if="isEdit">
          <div class="item date">
            <date-picker
              label="Created"
              v-model="newInvoice.created"
              :readonly="!newInvoice.created"
              :disabled="!newInvoice.created"
            />
          </div>

          <div class="item date">
            <date-picker
              :min="newInvoice.created"
              label="Payment"
              v-model="newInvoice.payment"
              :readonly="!newInvoice.payment"
              :disabled="!newInvoice.payment"
            />
          </div>

          <div class="item date">
            <date-picker
              label="Processed"
              v-model="newInvoice.processed"
              readonly
              disabled
            />
          </div>

          <div class="item date" v-if="newInvoice.returned">
            <date-picker
              :min="newInvoice.payment"
              label="Refunded"
              v-model="newInvoice.returned"
              :readonly="!newInvoice.returned"
              :disabled="!newInvoice.returned"
            />
          </div>
        </template>
      </div>

      <v-textarea
        outlined
        no-resize
        label="Admin note"
        v-model="newInvoice.meta.note"
      ></v-textarea>

      <div class="mt-2">
        <div v-if="!isBalanceInvoice" class="d-flex justify-space-between">
          <v-subheader>Invoice items</v-subheader>
          <v-btn @click="addInvoiceItem">Add</v-btn>
        </div>
        <invoice-items-table
          :show-delete="!isBalanceInvoice"
          :account="newInvoice.account"
          :items="newInvoice.items"
          @click:delete="deleteInvoiceItem"
        />
      </div>

      <nocloud-expansion-panels class="mt-4" title="Meta">
        <json-editor
          :json="newInvoice.meta"
          @changeValue="(data) => (newInvoice.meta = data)"
        />
      </nocloud-expansion-panels>

      <v-row justify="space-between" class="mt-4 mb-4">
        <div class="mt-2 ml-1">
          <v-btn
            v-if="!isEdit"
            class="mx-3"
            color="background-light"
            :loading="isSaveLoading"
            :disabled="isSaveDisabled"
            @click="saveInvoice(false, 'DRAFT')"
          >
            Draft
          </v-btn>
          <v-btn
            class="mx-3"
            color="background-light"
            :loading="isSaveLoading"
            :disabled="isSaveDisabled"
            @click="saveInvoice(false)"
          >
            {{ isEdit && !isDraft ? "Save" : "Publish" }}
          </v-btn>
          <v-btn
            class="mx-1"
            color="background-light"
            :loading="isSaveLoading"
            @click="saveInvoice(true)"
            :disabled="isEmailDisabled"
          >
            {{ isEdit && !isDraft ? "Save" : "Publish" }} + email
          </v-btn>
          <v-btn
            v-if="isEdit"
            class="mx-1"
            color="background-light"
            :loading="isSendEmailLoading"
            @click="sendEmail"
            :disabled="isEmailDisabled"
          >
            email
          </v-btn>

          <v-btn
            class="mx-1"
            v-if="isEdit"
            color="background-light"
            @click="downloadInvoice"
          >
            download
          </v-btn>

          <confirm-dialog
            v-if="isEdit"
            @confirm="copyInvoice"
            :loading="isCopyLoading"
          >
            <v-btn
              class="mx-1"
              :loading="isCopyLoading"
              color="background-light"
            >
              copy
            </v-btn>
          </confirm-dialog>
        </div>

        <div class="mt-2" v-if="isEdit">
          <confirm-dialog
            v-for="btn in changeStatusBtns"
            :key="btn.status"
            :disabled="
              (isStatusChangeLoading && btn.status !== newStatus) ||
              btn.disabled.includes(newInvoice.status) ||
              btn.status === newInvoice.status
            "
            :loading="isStatusChangeLoading && btn.status === newStatus"
            @confirm="
              btn.onClick ? btn.onClick() : changeInvoiceStatus(btn.status)
            "
          >
            <v-btn
              class="mx-1"
              :loading="isStatusChangeLoading && btn.status === newStatus"
              :disabled="
                (isStatusChangeLoading && btn.status !== newStatus) ||
                btn.disabled.includes(newInvoice.status) ||
                btn.status === newInvoice.status
              "
              color="background-light"
            >
              {{ btn.title }}
            </v-btn>
          </confirm-dialog>
        </div>
      </v-row>
    </v-form>

    <v-dialog v-model="isAddPaymentDialogOpen" width="400px" persistent>
      <v-card class="px-5 py-2">
        <v-card-title class="text-h5"> Add payment </v-card-title>

        <v-switch
          v-model="addPaymentOptions.changePaymentDate"
          label="Change payment date?"
        />

        <div style="max-width: 300px">
          <date-picker
            :min="newInvoice.created"
            label="Payment date"
            v-model="addPaymentOptions.paymentDate"
            :readonly="!addPaymentOptions.changePaymentDate"
            :disabled="!addPaymentOptions.changePaymentDate"
          />
        </div>

        <v-switch v-model="addPaymentOptions.sendEmail" label="Send email?" />

        <v-card-actions class="d-flex justify-end">
          <v-btn
            :disabled="isAddPaymentLoading"
            color="primary"
            text
            @click="isAddPaymentDialogOpen = false"
          >
            Close</v-btn
          >

          <v-btn
            color="primary"
            :loading="isAddPaymentLoading"
            text
            @click="changeStatusToPaid"
          >
            Add payment
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import JsonEditor from "@/components/JsonEditor.vue";
import { defaultFilterObject, formatSecondsToDateString } from "@/functions";
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import NocloudExpansionPanels from "@/components/ui/nocloudExpansionPanels.vue";
import datePicker from "@/components/ui/datePicker.vue";
import InvoiceItemsTable from "@/components/invoiceItemsTable.vue";
import { useRouter } from "vue-router/composables";
import {
  BillingStatus,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  UpdateInvoiceStatusRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";
import useInvoices from "../hooks/useInvoices";
import confirmDialog from "@/components/confirmDialog.vue";

const props = defineProps({
  invoice: {},
  isEdit: { type: Boolean, default: false },
});
const { invoice, isEdit } = toRefs(props);

const emit = defineEmits(["refresh"]);

const store = useStore();
const router = useRouter();
const { getInvoiceStatusColor, getTotalColor } = useInvoices();

const newInvoice = ref({
  account: null,
  status: "DRAFT",
  type: "NO_ACTION",
  total: 0,
  items: [{ price: null, description: "", amount: 1, unit: "Pcs" }],
  deadline: formatSecondsToDateString(Date.now() / 1000 + 86400 * 30),
  meta: {
    note: "",
  },
  instances: [],
});

const requiredRule = ref([(val) => !!val || "Field required"]);
const isValid = ref(false);
const invoiceForm = ref(null);
const isSaveLoading = ref(false);
const isSendEmailLoading = ref(false);
const isStatusChangeLoading = ref(false);
const newStatus = ref("");
const isAddPaymentDialogOpen = ref(false);
const isAddPaymentLoading = ref(false);
const isCopyLoading = ref(false);
const addPaymentOptions = ref({
  sendEmail: false,
  changePaymentDate: false,
  paymentDate: formatSecondsToDateString(Date.now() / 1000),
});

const isInstancesLoading = ref(true);
const instancesAccountMap = ref({});

const types = computed(() => {
  const types = [
    { id: "NO_ACTION", title: "No action" },
    { id: "INSTANCE_START", title: "Instance start" },
    { id: "INSTANCE_RENEWAL", title: "Instance renewal" },
    {
      id: "BALANCE",
      title: "Top up balance",
    },
  ];

  if (isEdit.value && !isDraft.value) {
    types.push({ id: "WHMCS_INVOICE", title: "WHMCS Invoice" });
  }

  return types;
});

const changeStatusBtns = [
  {
    title: "draft",
    status: "DRAFT",
    disabled: ["TERMINATED", "PAID", "RETURNED"],
  },
  {
    title: "Add payment",
    status: "PAID",
    onClick: () => openAddPaymentDialog(),
    disabled: ["TERMINATED", "DRAFT", "RETURNED"],
  },
  {
    title: "Unpaid",
    status: "UNPAID",
    disabled: ["TERMINATED", "RETURNED", "DRAFT", "PAID"],
  },
  {
    title: "cancel",
    status: "CANCELED",
    disabled: ["TERMINATED", "RETURNED", "DRAFT", "PAID"],
  },
  {
    title: "Refund",
    status: "RETURNED",
    disabled: ["CANCELED", "TERMINATED", "UNPAID", "DRAFT"],
  },
  { title: "terminate", status: "TERMINATED", disabled: ["TERMINATED"] },
];

onMounted(async () => {
  setInvoice();

  isInstancesLoading.value = false;
});

const isBalanceInvoice = computed(() => newInvoice.value.type === "BALANCE");
const isDraft = computed(() => newInvoice.value?.status === "DRAFT");
const instances = computed(() => {
  const account = newInvoice.value.account?.uuid;
  if (!account || isInstancesLoading.value) {
    return [];
  }

  return instancesAccountMap.value[account];
});

const topUpItemMessage = computed(() => {
  const data = store.getters["settings/all"].find(
    (s) => s.key === "billing-invoices"
  );

  return (JSON.parse(data?.value || "{}") || {}).top_up_item_message;
});

const accountCurrency = computed(
  () =>
    newInvoice.value.account?.currency || store.getters["currencies/default"]
);

const isEmailDisabled = computed(() =>
  ["TERMINATED", "CANCELED"].includes(newInvoice.value.status)
);

const isSaveDisabled = computed(() =>
  ["TERMINATED"].includes(newInvoice.value.status)
);

const setInvoice = () => {
  if (invoice.value) {
    newInvoice.value = {
      ...invoice.value,
      items: invoice.value.items || [],
      total: invoice.value.total || 0,
      deadline: invoice.value.deadline
        ? formatSecondsToDateString(invoice.value.deadline)
        : null,
      payment: formatSecondsToDateString(invoice.value.payment),
      returned: formatSecondsToDateString(invoice.value.returned),
      processed: formatSecondsToDateString(invoice.value.processed),
      created: formatSecondsToDateString(invoice.value.created),
    };

    if (isBalanceInvoice.value) {
      setTimeout(() => {
        newInvoice.value.items = invoice.value.items;
      }, 1);
    }
  }
};

const getInvoiceDateTs = (originalTs, newDate) => {
  return (
    new Date(newDate).getTime() / 1000 +
    (+originalTs -
      new Date(formatSecondsToDateString(originalTs)).getTime() / 1000)
  );
};

const saveInvoice = async (withEmail = false, status = "UNPAID") => {
  if (!(await invoiceForm.value.validate())) {
    return;
  }
  isSaveLoading.value = true;

  try {
    const data = {
      total: convertPrice(newInvoice.value.total),
      account: newInvoice.value.account.uuid || invoice.value.account,
      currency: { id: accountCurrency.value.id },
      items: newInvoice.value.items,
      meta: newInvoice.value.meta,
      status: status ? status : newInvoice.value.status,
      instances: newInvoice.value.instances || [],
      deadline: new Date(newInvoice.value.deadline).getTime() / 1000,
      type: newInvoice.value.type,
      processed: invoice.value?.processed || 0,
    };

    if (!isEdit.value && !invoice.value?.uuid) {
      await store.getters["invoices/invoicesClient"].createInvoice(
        CreateInvoiceRequest.fromJson({
          invoice: data,
          isSendEmail: !!withEmail,
        })
      );
      router.push({ name: "Invoices" });
    } else {
      data.uuid = invoice.value.uuid;

      if (newInvoice.value.created) {
        data.created = getInvoiceDateTs(
          invoice.value.created,
          newInvoice.value.created
        );
      }

      if (newInvoice.value.returned) {
        data.returned = getInvoiceDateTs(
          invoice.value.returned,
          newInvoice.value.returned
        );
      }

      if (newInvoice.value.payment) {
        data.payment = getInvoiceDateTs(
          invoice.value.payment,
          newInvoice.value.payment
        );
      }

      await store.getters["invoices/invoicesClient"].updateInvoice(
        UpdateInvoiceRequest.fromJson({
          invoice: data,
          isSendEmail: !!withEmail,
        })
      );
      store.commit("snackbar/showSnackbarSuccess", {
        message: "Invoice successfully saved",
      });

      if (!isEdit.value) {
        store.dispatch("reloadBtn/onclick");
      }
    }
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSaveLoading.value = false;
  }
};

const downloadInvoice = async () => {
  try {
    const { paymentLink } = await store.getters["invoices/invoicesClient"].pay({
      invoiceId: invoice.value.uuid,
    });
    if (!paymentLink) {
      throw new Error("No link");
    }
    window.open(paymentLink, "_blanc");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: "Document not found",
    });
  }
};

const copyInvoice = async () => {
  isCopyLoading.value = true;

  try {
    await store.dispatch("invoices/copy", invoice.value);
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isCopyLoading.value = false;
  }
};

const convertPrice = (price) => {
  return Math.abs(price);
};

const addInvoiceItem = () => {
  newInvoice.value.items.push({
    description: "",
    price: 0,
    amount: 1,
    unit: "Pcs",
  });
};

const deleteInvoiceItem = (index) => {
  if (!newInvoice.value.items.length) {
    return;
  }
  newInvoice.value.items = newInvoice.value.items.filter((_, i) => i !== index);
};

const onChangeAccount = async () => {
  if (!isEdit.value && !isDraft.value) {
    newInvoice.value.instances = [];
  }

  const account = newInvoice.value.account?.uuid;

  if (instancesAccountMap.value[account]) {
    return;
  }

  isInstancesLoading.value = true;
  try {
    instancesAccountMap.value[account] = store.dispatch("instances/fetch", {
      filters: { account: [account] },
    });

    const data = await instancesAccountMap.value[account];

    instancesAccountMap.value[account] = data.map((response) => ({
      uuid: response.uuid,
      title: response.title,
      price: response.estimate,
    }));
  } catch (e) {
    instancesAccountMap.value[account] = undefined;
  } finally {
    isInstancesLoading.value = false;
  }
};

const onChangeInstance = () => {
  if (!newInvoice.value.instances || !newInvoice.value.instances.length) {
    return;
  }
  const newItems = [];

  newInvoice.value.instances.forEach((uuid) => {
    const { price, title } = instances.value.find(
      (instance) => instance.uuid === uuid
    );
    if (!price && !title) {
      return;
    }

    const existedProduct = newInvoice.value.items.find(
      (item) => item.description === title
    );

    if (existedProduct) {
      newItems.push(existedProduct);
    } else {
      newItems.push({
        price: price,
        amount: 1,
        description: title,
        unit: "Pcs",
      });
    }
  });

  newInvoice.value.items = newItems;
};

const sendEmail = async () => {
  isSendEmailLoading.value = true;
  try {
    await new Promise((resolve) => setTimeout(resolve, 5000));
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSendEmailLoading.value = false;
  }
};

const openAddPaymentDialog = () => {
  isAddPaymentDialogOpen.value = true;
};

const openAccountWindow = () => {
  return window.open(
    "/admin/accounts/" + newInvoice.value.account.uuid,
    "_blanc"
  );
};

const changeStatusToPaid = async () => {
  try {
    isAddPaymentLoading.value = true;

    await changeInvoiceStatus("PAID", {
      paymentDate: addPaymentOptions.value.changePaymentDate
        ? new Date(addPaymentOptions.value.paymentDate).getTime() / 1000
        : null,
      isSendEmail: addPaymentOptions.value.sendEmail,
    });
  } finally {
    isAddPaymentLoading.value = false;
    isAddPaymentDialogOpen.value = false;
  }
};

const changeInvoiceStatus = async (status, params = null) => {
  isStatusChangeLoading.value = true;
  newStatus.value = status;
  try {
    await store.getters["invoices/invoicesClient"].updateInvoiceStatus(
      UpdateInvoiceStatusRequest.fromJson({
        uuid: invoice.value.uuid,
        status: BillingStatus[status],
        params,
      })
    );

    emit("refresh");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isStatusChangeLoading.value = false;
    newStatus.value = "";
  }
};

watch(isBalanceInvoice, (value) => {
  if (value) {
    newInvoice.value.items = [
      {
        description: topUpItemMessage.value,
        price: newInvoice.value.total,
        amount: 1,
        unit: "Pcs",
      },
    ];
  } else {
    newInvoice.value.items = [];
  }

  newInvoice.value.instances = [];
});

watch(
  () => newInvoice.value.items,
  () => {
    newInvoice.value.total = newInvoice.value.items?.reduce(
      (acc, i) => acc + Number(i.price || 0) * Number(i.amount || 0),
      0
    );
  },
  { deep: true }
);

watch(invoice, setInvoice);
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}

.invoice__container {
  display: flex;
  flex-wrap: wrap;

  .item {
    margin-left: 10px;
    margin-left: 10px;
    width: 200px;
    &.date {
      width: 140px;
    }
  }
}
</style>
