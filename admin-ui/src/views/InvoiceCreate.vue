<template>
  <div class="pa-4">
    <h1 class="page__title">Create invoice</h1>
    <v-form v-model="isValid" ref="invoiceForm">
      <v-row>
        <v-col lg="6" cols="12">
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
                v-model="newInvoice.meta.type"
                :items="types"
              >
              </v-select>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <v-row>
        <v-col lg="6" cols="12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Account</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-autocomplete
                :disabled="isEdit"
                :filter="defaultFilterObject"
                label="Account"
                v-model="newInvoice.account"
                return-object
                :rules="requiredRule"
                item-text="title"
                item-value="uuid"
                :items="accounts"
                :loading="isAccountsLoading"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Amount</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                type="number"
                label="Amount"
                :suffix="accountCurrency"
                :value="amount"
                disabled
                readonly
              />
            </v-col>
          </v-row>

          <v-row class="mx-5">
            <v-textarea
              no-resize
              label="Admin note"
              v-model="newInvoice.meta.note"
            ></v-textarea>
          </v-row>

          <div class="mt-2">
            <div class="d-flex justify-space-between">
              <v-subheader>Invoice items</v-subheader>
              <v-btn @click="addInvoiceItem">Add</v-btn>
            </div>
            <invoice-items-table
              :instances="instances"
              show-delete
              :account="newInvoice.account"
              :items="newInvoice.meta.items"
              @click:delete="deleteInvoiceItem"
            />
          </div>

          <nocloud-expansion-panels class="mt-4" title="Meta">
            <json-editor
              :json="newInvoice.meta"
              @changeValue="(data) => (newInvoice.meta = data)"
            />
          </nocloud-expansion-panels>
        </v-col>
      </v-row>

      <v-row justify="start" class="mb-4">
        <v-btn
          class="mx-3"
          color="background-light"
          :loading="isSaveLoading"
          @click="saveInvoice(false)"
        >
          Publish
        </v-btn>
        <v-btn
          class="mx-4"
          color="background-light"
          :loading="isSaveLoading"
          @click="saveInvoice(true)"
        >
          Publish + email
        </v-btn>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import JsonEditor from "@/components/JsonEditor.vue";
import { defaultFilterObject } from "@/functions";
import { computed, onMounted, ref, toRefs } from "vue";
import { useStore } from "@/store";
import NocloudExpansionPanels from "@/components/ui/nocloudExpansionPanels.vue";
import InvoiceItemsTable from "@/components/invoiceItemsTable.vue";
import api from "@/api";
import { useRouter } from "vue-router/composables";

const props = defineProps({
  invoice: {},
  isEdit: { type: Boolean, default: false },
});
const { invoice, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const newInvoice = ref({
  account: null,
  total: 0,
  meta: {
    note: "",
    items: [{ amount: null, type: "default", title: "" }],
    type: "payment",
  },
});

const requiredRule = ref([(val) => !!val || "Field required"]);
const isValid = ref(false);
const invoiceForm = ref(null);
const isSaveLoading = ref(false);

const types = [
  {
    id: "payment",
    title: "Payment invoice (no balance change)",
  },
  {
    id: "top-up",
    title: "Top-up invoice (with balance change)",
  },
];

onMounted(async () => {
  if (isEdit.value) {
    newInvoice.value = {
      ...invoice.value,
      meta: {
        ...invoice.value.meta,
        items: invoice.value.meta.items.map((i) => ({ ...i, type: "default" })),
      },
    };
  }

  await Promise.all([
    store.dispatch("accounts/fetch"),
    store.dispatch("services/fetch"),
    store.dispatch("namespaces/fetch"),
  ]);

  if (isEdit.value) {
    newInvoice.value.account = accounts.value.find(
      (a) => a.uuid === invoice.value.account
    );
  }
});

const accounts = computed(() => store.getters["accounts/all"]);
const isAccountsLoading = computed(() => store.getters["accounts/isLoading"]);

const namespaces = computed(() => store.getters["namespaces/all"]);

const services = computed(() => store.getters["services/all"]);
const instances = computed(() => {
  if (!newInvoice.value.account) {
    return;
  }

  const namespace = namespaces.value.find(
    (n) => n.access.namespace === newInvoice.value.account.uuid
  );
  const servicesByAccount = services.value.filter(
    (s) => s.access.namespace === namespace.uuid
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
const amount = computed(() =>
  newInvoice.value.meta.items.reduce((acc, i) => acc + i.amount, 0)
);

const saveInvoice = async (withEmail = false) => {
  console.log(withEmail, newInvoice.value);
  if (!(await invoiceForm.value.validate())) {
    return;
  }
  isSaveLoading.value = true;

  try {
    const data = {
      total: convertPrice(amount.value),
      account: newInvoice.value.account.uuid,
      meta: {
        ...newInvoice.value.meta,
        items: newInvoice.value.meta.items.map((item) => ({
          title: item.title,
          amount: convertPrice(item.amount),
        })),
      },
    };
    if (!isEdit.value) {
      await api.put("/billing/invoices", data);
      router.push({ name: "Invoices" });
    } else {
      await api.patch("/billing/invoices/" + invoice.value.uuid, data);
      store.commit("snackbar/showSnackbarSuccess", {
        message: "Invoice successfully saved",
      });
    }
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSaveLoading.value = false;
  }
};

const convertPrice = (price) => {
  return newInvoice.value.meta.type === "payment"
    ? Math.abs(price)
    : -Math.abs(price);
};

const addInvoiceItem = () => {
  newInvoice.value.meta.items.push({ title: "", amount: 0, type: "default" });
};

const deleteInvoiceItem = (index) => {
  if (!newInvoice.value.meta.items.length) {
    return;
  }
  newInvoice.value.meta.items = newInvoice.value.meta.items.filter(
    (_, i) => i !== index
  );
};
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
