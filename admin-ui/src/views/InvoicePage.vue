<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Invoices' }">{{
        navTitle("Invoices")
      }}</router-link>
      /
      <span v-if="!isEditingTitle">{{ invoiceTitle || "Not found" }}</span>
      <v-text-field
        v-else
        v-model="editedTitle"
        dense
        style="display: inline-block; width: 300px; margin: 0 10px"
        class="page__title-input"
        hide-details
      />
      <template v-if="invoiceTitle">
        <v-btn v-if="!isEditingTitle" icon class="ml-2" @click="startEditTitle">
          <v-icon size="large">mdi-pencil</v-icon>
        </v-btn>
        <template v-else>
          <v-btn
            icon
            class="ml-1"
            :loading="isSavingTitle"
            @click="saveTitle"
            :disabled="!editedTitle"
          >
            <v-icon size="large">mdi-content-save</v-icon>
          </v-btn>
          <v-btn
            icon
            size="large"
            class="ml-1"
            @click="cancelEditTitle"
            :disabled="isSavingTitle"
          >
            <v-icon size="large">mdi-close</v-icon>
          </v-btn>
        </template>
      </template>
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="selectedTab"
    >
      <v-tab v-for="tab in tabItems" :key="tab.title">{{ tab.title }} </v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="selectedTab"
    >
      <v-tab-item v-for="(tab, index) in tabItems" :key="tab.title">
        <v-progress-linear indeterminate class="pt-2" v-if="isInvoiceLoading" />
        <component
          v-else-if="invoice && index === selectedTab"
          :is="tab.component"
          :invoice="invoice"
          :is-edit="invoice.status !== 'DRAFT'"
          @refresh="refreshInvoice"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router/composables";
import { useStore } from "@/store";
import config from "@/config.js";
import InvoiceCreate from "@/views/InvoiceCreate.vue";
import InvoiceTemplate from "@/components/invoice/template.vue";
import { ChangeInvoiceNumberRequest } from "nocloud-proto/proto/es/billing/billing_pb";

const store = useStore();
const route = useRoute();

const selectedTab = ref(0);
const navTitles = ref(config.navTitles ?? {});
const isEditingTitle = ref(false);
const editedTitle = ref("");
const isSavingTitle = ref(false);

const invoice = computed(() => store.getters["invoices/one"]);
const invoiceTitle = computed(() =>
  isInvoiceLoading.value ? "..." : invoice.value?.number
);

const startEditTitle = () => {
  editedTitle.value = invoiceTitle.value || "";
  isEditingTitle.value = true;
};

const cancelEditTitle = () => {
  isEditingTitle.value = false;
  editedTitle.value = "";
};

const saveTitle = async () => {
  try {
    isSavingTitle.value = true;
    invoice.value.number = editedTitle.value;

    await store.getters["invoices/invoicesClient"].changeInvoiceNumber(
      ChangeInvoiceNumberRequest.fromJson({
        invoice: invoice.value.uuid,
        newNumber: editedTitle.value,
      })
    );

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Invoice title updated",
    });

    isEditingTitle.value = false;

    refreshInvoice();
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.message || "Error updating title",
    });
  } finally {
    isSavingTitle.value = false;
  }
};

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    type: "invoices/get",
    params: invoiceId.value,
  });
  selectedTab.value = route.query.tab || 0;

  refreshInvoice();
});

const invoiceId = computed(() => {
  return route.params.uuid;
});

const isInvoiceLoading = computed(() => {
  return store.getters["invoices/isLoading"];
});

const tabItems = computed(() => [
  {
    component: InvoiceCreate,
    title: "info",
  },
  {
    component: InvoiceTemplate,
    title: "template",
  },
]);

function navTitle(title) {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
}

const refreshInvoice = async () => {
  try {
    await store.dispatch("invoices/get", invoiceId.value);
    document.title = `${invoiceTitle.value} | NoCloud`;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  }
};

watch(invoiceId, () => {
  refreshInvoice();
});
</script>

<script>
export default { name: "invoice-view" };
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;

  .page__title-input {
    vertical-align: middle;
  }
}

::v-deep .page__title-input .v-input__control {
  margin-top: 0;
}

::v-deep .page__title-input input {
  font-size: 32px;
  font-weight: 400;
  font-family: "Quicksand", sans-serif;
}
</style>
