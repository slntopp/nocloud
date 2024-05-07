<template>
  <div class="pa-4">
    <h1 class="page__title">Create invoice</h1>
    <v-form v-model="isValid" ref="invoiceForm">
      <v-row>
        <v-col cols="6">
          <v-row v-if="!isEdit" align="center">
            <v-col cols="3">
              <v-subheader>Draft</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-switch
                :input-value="newInvoice.status === 'DRAFT'"
                @change="onChangeDraft"
                :items="types"
              >
              </v-switch>
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Type</v-subheader>
            </v-col>
            <v-col cols="9">
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
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Deadline</v-subheader>
            </v-col>
            <v-col cols="9">
              <date-picker
                :min="formatSecondsToDateString(Date.now() / 1000)"
                label="Deadline"
                v-model="newInvoice.deadline"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Account</v-subheader>
            </v-col>
            <v-col cols="9">
              <accounts-autocomplete
                :loading="isInstancesLoading"
                @change="onChangeAccount"
                :disabled="isEdit"
                label="Account"
                v-model="newInvoice.account"
                fetch-value
                return-object
                :rules="requiredRule"
              />
            </v-col>
          </v-row>

          <v-row v-if="!isBalanceInvoice" align="center">
            <v-col cols="3">
              <v-subheader>Instances</v-subheader>
            </v-col>
            <v-col cols="9">
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
          </v-row>
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Total</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                type="number"
                label="Total"
                :suffix="accountCurrency?.title"
                v-model="newInvoice.total"
                :disabled="!isBalanceInvoice"
              />
            </v-col>
          </v-row>
        </v-col>
        <v-col v-if="isEdit" cols="6">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Created date</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                :value="formatSecondsToDate(newInvoice.created, true) || '-'"
                readonly
                disabled
              />
            </v-col>
          </v-row>
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Payment date</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                :value="formatSecondsToDate(newInvoice.payment, true) || '-'"
                readonly
                disabled
              />
            </v-col>
          </v-row>
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Processed date</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                :value="formatSecondsToDate(newInvoice.processed, true) || '-'"
                readonly
                disabled
              />
            </v-col>
          </v-row>

          <v-row align="center" v-if="newInvoice.returned">
            <v-col cols="3">
              <v-subheader>Returned date</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                :value="formatSecondsToDate(newInvoice.returned, true) || '-'"
                readonly
                disabled
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Status</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-chip>{{ invoice.status }}</v-chip>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <v-textarea
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

      <v-row justify="start" class="mt-4 mb-4">
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

        <template v-if="isEdit">
          <v-btn
            class="mx-4"
            color="background-light"
            :loading="isSendEmailLoading"
            @click="sendEmail"
            :disabled="isEmailDisabled"
          >
            email
          </v-btn>

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
            @click="changeInvoiceStatus(btn.status)"
          >
            {{ btn.title }}
          </v-btn>
        </template>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import JsonEditor from "@/components/JsonEditor.vue";
import {
  defaultFilterObject,
  formatSecondsToDate,
  formatSecondsToDateString,
} from "@/functions";
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import NocloudExpansionPanels from "@/components/ui/nocloudExpansionPanels.vue";
import datePicker from "@/components/ui/datePicker.vue";
import InvoiceItemsTable from "@/components/invoiceItemsTable.vue";
import { useRouter } from "vue-router/composables";
import {
  BillingStatus,
  Invoice,
  UpdateInvoiceStatusRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";

const props = defineProps({
  invoice: {},
  isEdit: { type: Boolean, default: false },
});
const { invoice, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const selectedInstances = ref([]);

const newInvoice = ref({
  account: null,
  status: "DRAFT",
  type: "INSTANCE_RENEWAL",
  total: 0,
  items: [{ price: null, title: "", amount: 1, unit: "Pcs" }],
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

const types = [
  { id: "INSTANCE_START", title: "Instance start" },
  { id: "INSTANCE_RENEWAL", title: "Instance renewal" },
  {
    id: "BALANCE",
    title: "Top up balance",
  },
];

const changeStatusBtns = [
  {
    title: "pay",
    status: "PAID",
    disabled: ["TERMINATED", "CANCELED", "DRAFT", "RETURNED"],
  },
  {
    title: "cancel",
    status: "CANCELED",
    disabled: ["TERMINATED", "RETURNED", "DRAFT", "UNPAID"],
  },
  {
    title: "return",
    status: "RETURNED",
    disabled: ["CANCELED", "TERMINATED", "UNPAID", "DRAFT"],
  },
  { title: "terminate", status: "TERMINATED", disabled: ["TERMINATED"] },
];

onMounted(async () => {
  if (invoice.value) {
    newInvoice.value = {
      ...invoice.value,
      deadline: formatSecondsToDateString(invoice.value.deadline),
    };

    selectedInstances.value = invoice.value.items?.map((item) => item.instance);
  }

  await Promise.all([
    store.dispatch("services/fetch"),
    store.dispatch("namespaces/fetch"),
  ]);

  // if (isEdit.value) {
  //   newInvoice.value.account = accounts.value.find(
  //     (a) => a.uuid === invoice.value.account
  //   );
  // }
});

const namespaces = computed(() => store.getters["namespaces/all"]);

const isInstancesLoading = computed(() => store.getters["services/isLoading"]);
const isBalanceInvoice = computed(() => newInvoice.value.type === "BALANCE");
const services = computed(() => store.getters["services/all"]);
const instances = computed(() => {
  if (
    newInvoice.value.account === undefined ||
    newInvoice.value.account === null
  ) {
    return;
  }

  const namespace = namespaces.value.find(
    (n) => n.access.namespace === newInvoice.value.account?.uuid
  );
  const servicesByAccount = services.value.filter(
    (s) => s.access.namespace === namespace?.uuid
  );
  const instances = [];

  servicesByAccount.forEach((s) => {
    s?.instancesGroups.forEach((ig) => {
      ig.instances.forEach((i) => instances.push({ ...i, type: ig.type }));
    });
  });

  return instances;
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

const saveInvoice = async (withEmail = false) => {
  console.log(withEmail, newInvoice.value);
  if (!(await invoiceForm.value.validate())) {
    return;
  }
  isSaveLoading.value = true;

  try {
    const data = {
      total: convertPrice(newInvoice.value.total),
      account: newInvoice.value.account.uuid,
      items: newInvoice.value.items,
      meta: newInvoice.value.meta,
      status: newInvoice.value.status,
      deadline: new Date(newInvoice.value.deadline).getTime() / 1000,
      type: newInvoice.value.type,
    };
    if (!isEdit.value && !invoice.value?.uuid) {
      await store.getters["invoices/invoicesClient"].createInvoice(
        Invoice.fromJson(data)
      );
      router.push({ name: "Invoices" });
    } else {
      await store.getters["invoices/invoicesClient"].updateInvoice(
        Invoice.fromJson({
          ...data,
          uuid: invoice.value.uuid,
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

const convertPrice = (price) => {
  return Math.abs(price);
};

const addInvoiceItem = () => {
  newInvoice.value.items.push({
    title: "",
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

const onChangeAccount = () => {
  selectedInstances.value = null;
};

const onChangeInstance = () => {
  if (!selectedInstances.value || !selectedInstances.value.length) {
    return;
  }
  const newItems = [];

  selectedInstances.value.forEach((instance) => {
    const { price: productPrice, title: productTitle } =
      instance.billingPlan.products[instance.product];

    const existedProduct = newInvoice.value.items.find(
      (item) =>
        item.title.includes(productTitle) && item.instance === instance.uuid
    );

    if (existedProduct) {
      newItems.push(existedProduct);
    } else {
      newItems.push({
        price: productPrice,
        amount: 1,
        title: productTitle,
        instance: instance.uuid,
        unit: "Pcs",
      });
    }
  });

  newInvoice.value.items = newItems;
};

const sendEmail = async () => {
  isSendEmailLoading.value = true;
  try {
    console.log(newInvoice.value.account);
    await new Promise((resolve) => setTimeout(resolve, 5000));
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSendEmailLoading.value = false;
  }
};

const onChangeDraft = (isDraft) => {
  newInvoice.value.status = isDraft ? "DRAFT" : "UNPAID";
};

const changeInvoiceStatus = async (status) => {
  isStatusChangeLoading.value = true;
  newStatus.value = status;
  try {
    console.log("new BillingStatus", BillingStatus[status], status);
    await store.getters["invoices/invoicesClient"].updateInvoiceStatus(
      UpdateInvoiceStatusRequest.fromJson({
        uuid: invoice.value.uuid,
        status: BillingStatus[status],
      })
    );
    newInvoice.value.status = status;
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
