<template>
  <div class="pa-4">
    <h1 v-if="!isEdit && !isDraft" class="page__title">Create invoice</h1>

    <div
      v-if="isEdit"
      style="z-index: 0; position: relative; top: -15px; right: 20px"
    >
      <div class="d-flex justify-end mt-1 align-center flex-wrap">
        <v-chip class="mr-2" :color="getTotalColor(newInvoice)">
          Total:
          {{ `${newInvoice.total} ${accountCurrency?.title}` }}
        </v-chip>

        <v-chip :color="getInvoiceStatusColor(newInvoice.status)">
          {{ `${newInvoice.status}` }}
        </v-chip>
      </div>
    </div>

    <v-form v-model="isValid" ref="invoiceForm">
      <v-row>
        <v-col cols="2">
          <v-select
            :disabled="isEdit"
            item-value="id"
            item-text="title"
            label="Type"
            v-model="newInvoice.type"
            :items="types"
          >
          </v-select>
        </v-col>

        <v-col cols="3">
          <div class="d-flex align-center">
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
              @click="openAccountWindow"
              icon
              v-if="isEdit && newInvoice.account && !isInstancesLoading"
            >
              <v-icon>mdi-login</v-icon>
            </v-btn>
          </div>
        </v-col>

        <v-col cols="3">
          <v-autocomplete
            :disabled="isEdit"
            :filter="defaultFilterObject"
            label="Instances"
            v-model="selectedInstances"
            return-object
            multiple
            item-text="title"
            item-value="uuid"
            :items="instances"
            @change="onChangeInstance"
            :loading="isInstancesLoading"
          />
        </v-col>

        <v-col cols="2" v-if="isBalanceInvoice">
          <v-text-field
            type="number"
            label="Total"
            :suffix="accountCurrency?.title"
            v-model="newInvoice.total"
          />
        </v-col>

        <v-col cols="2">
          <date-picker
            :min="formatSecondsToDateString(Date.now() / 1000)"
            label="Due date"
            v-model="newInvoice.deadline"
          />
        </v-col>

        <template v-if="isEdit">
          <v-col cols="2">
            <date-picker
              label="Created"
              v-model="newInvoice.created"
              :readonly="!newInvoice.created"
              :disabled="!newInvoice.created"
            />
          </v-col>

          <v-col cols="2">
            <date-picker
              :min="newInvoice.created"
              label="Payment"
              v-model="newInvoice.payment"
              :readonly="!newInvoice.payment"
              :disabled="!newInvoice.payment"
            />
          </v-col>

          <v-col cols="2">
            <date-picker
              label="Processed"
              v-model="newInvoice.processed"
              readonly
              disabled
            />
          </v-col>

          <v-col v-if="newInvoice.returned" cols="2">
            <date-picker
              :min="newInvoice.payment"
              label="Refunded"
              v-model="newInvoice.returned"
              :readonly="!newInvoice.returned"
              :disabled="!newInvoice.returned"
            />
          </v-col>
        </template>
      </v-row>
      <v-textarea
        outlined
        no-resize
        label="Admin note"
        v-model="newInvoice.meta.note"
      ></v-textarea>

      <div class="mt-2" v-if="!isBalanceInvoice">
        <div class="d-flex justify-space-between">
          <v-subheader>Invoice items</v-subheader>
          <v-btn @click="addInvoiceItem">Add</v-btn>
        </div>
        <invoice-items-table
          show-delete
          :account="newInvoice.account"
          :items="newInvoice.items"
          :instances="instances"
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
        <div>
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
            Publish
          </v-btn>
          <v-btn
            class="mx-4"
            color="background-light"
            :loading="isSaveLoading"
            @click="saveInvoice(true)"
            :disabled="isEmailDisabled"
          >
            Publish + email
          </v-btn>
          <v-btn
            v-if="isEdit"
            class="mx-4"
            color="background-light"
            :loading="isSendEmailLoading"
            @click="sendEmail"
            :disabled="isEmailDisabled"
          >
            email
          </v-btn>

          <v-btn
            class="mx-4"
            v-if="isEdit"
            color="background-light"
            @click="downloadInvoice"
          >
            download
          </v-btn>
        </div>

        <div v-if="isEdit">
          <v-btn
            v-for="btn in changeStatusBtns"
            class="mx-4"
            :key="btn.status"
            :loading="isStatusChangeLoading && btn.status === newStatus"
            :disabled="
              (isStatusChangeLoading && btn.status !== newStatus) ||
              btn.disabled.includes(newInvoice.status) ||
              btn.status === newInvoice.status
            "
            color="background-light"
            @click="
              btn.onClick ? btn.onClick() : changeInvoiceStatus(btn.status)
            "
          >
            {{ btn.title }}
          </v-btn>
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

const props = defineProps({
  invoice: {},
  isEdit: { type: Boolean, default: false },
});
const { invoice, isEdit } = toRefs(props);

const emit = defineEmits(["refresh"]);

const store = useStore();
const router = useRouter();
const { getInvoiceStatusColor, getTotalColor } = useInvoices();

const selectedInstances = ref([]);

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
const addPaymentOptions = ref({
  sendEmail: false,
  changePaymentDate: false,
  paymentDate: formatSecondsToDateString(Date.now() / 1000),
});

const isInstancesLoading = ref(true);
const instancesAccountMap = ref({});

const types = [
  { id: "NO_ACTION", title: "No action" },
  { id: "INSTANCE_START", title: "Instance start" },
  { id: "INSTANCE_RENEWAL", title: "Instance renewal" },
  {
    id: "BALANCE",
    title: "Top up balance",
  },
];

const changeStatusBtns = [
  {
    title: "draft",
    status: "DRAFT",
    disabled: ["TERMINATED", "CANCELED", "PAID", "RETURNED"],
  },
  {
    title: "Add payment",
    status: "PAID",
    onClick: () => openAddPaymentDialog(),
    disabled: ["TERMINATED", "CANCELED", "DRAFT", "RETURNED"],
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
const isDraft = computed(() => invoice.value?.status !== "DRAFT");
const instances = computed(() => {
  const account = newInvoice.value.account?.uuid;
  if (!account || isInstancesLoading.value) {
    return [];
  }

  return instancesAccountMap.value[account];
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
      deadline: formatSecondsToDateString(invoice.value.deadline),
      payment: formatSecondsToDateString(invoice.value.payment),
      returned: formatSecondsToDateString(invoice.value.returned),
      processed: formatSecondsToDateString(invoice.value.processed),
      created: formatSecondsToDateString(invoice.value.created),
    };

    selectedInstances.value = invoice.value.items?.map((item) => item.instance);
  }
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
      items: newInvoice.value.items,
      meta: newInvoice.value.meta,
      status: status ? status : newInvoice.value.status,
      deadline: new Date(newInvoice.value.deadline).getTime() / 1000,
      type: newInvoice.value.type,
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
        data.created = new Date(newInvoice.value.created).getTime() / 1000;
      }

      if (newInvoice.value.returned) {
        data.returned = new Date(newInvoice.value.returned).getTime() / 1000;
      }

      if (newInvoice.value.payment) {
        data.payment = new Date(newInvoice.value.payment).getTime() / 1000;
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
        router.push({ name: "Invoices" });
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

const convertPrice = (price) => {
  return Math.abs(price);
};

const addInvoiceItem = () => {
  newInvoice.value.items.push({
    description: "",
    price: 0,
    instance: "",
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
  if (!isEdit.value) {
    selectedInstances.value = null;
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
  if (!selectedInstances.value || !selectedInstances.value.length) {
    return;
  }
  const newItems = [];

  selectedInstances.value.forEach((instance) => {
    const { price, title, uuid } = instance;

    const existedProduct = newInvoice.value.items.find(
      (item) => item.description.includes(title) && item.instance === uuid
    );

    if (existedProduct) {
      newItems.push(existedProduct);
    } else {
      newItems.push({
        price: price,
        amount: 1,
        description: title,
        instance: uuid,
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

watch(instances, (instances) => {
  if (isEdit.value) {
    const instancesUuid = [
      ...new Set(invoice.value.items?.map((item) => item.instance)),
    ];
    selectedInstances.value = instances.filter((instance) =>
      instancesUuid.includes(instance.uuid)
    );
  }
});

watch(isBalanceInvoice, () => {
  selectedInstances.value = [];
  newInvoice.value.items = [];
});

watch(
  () => newInvoice.value.items,
  () => {
    if (invoice.value?.uuid && invoice.value.type === "BALANCE") {
      return;
    }
    newInvoice.value.total = newInvoice.value.items?.reduce(
      (acc, i) => acc + Number(i.price) * Number(i.amount),
      0
    );
  },
  { deep: true }
);

watch(invoice, setInvoice);
</script>

<style scoped>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
