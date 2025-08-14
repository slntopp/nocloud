<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <div class="actions">
      <div class="d-flex justify-end mt-1 align-center flex-wrap">
        <hint-btn hint="Create transaction">
          <v-btn
            :class="viewport < 600 ? 'ma-0' : 'ma-1'"
            :small="viewport < 600"
            :to="{
              name: 'Transactions create',
              query: {
                account: account.uuid,
              },
            }"
          >
            <v-icon>mdi-abacus</v-icon>
          </v-btn>
        </hint-btn>

        <hint-btn hint="Create instance">
          <v-btn
            :class="viewport < 600 ? 'ma-0' : 'ma-1'"
            :small="viewport < 600"
            :disabled="isLocked"
            :to="{
              name: 'Instance create',
              query: {
                accountId: account.uuid,
                type: 'ione',
              },
            }"
          >
            <v-icon>mdi-server</v-icon>
          </v-btn>
        </hint-btn>

        <hint-btn hint="Invoice based">
          <v-dialog v-model="isChangeRegularPaymentOpen" max-width="500">
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                :disabled="isChangeRegularPaymentLoading"
                :loading="isChangeRegularPaymentLoading"
                :small="viewport < 600"
                :class="viewport < 600 ? 'ma-4' : 'ma-1'"
                v-bind="attrs"
                v-on="on"
              >
                <v-icon>mdi-invoice-check-outline</v-icon>
              </v-btn>
            </template>
            <v-card color="background-light pa-5">
              <v-card-actions class="d-flex justify-center">
                <v-btn class="mr-2" @click="changeRegularPayment(false)">
                  Disable to all
                </v-btn>
                <v-btn class="mr-2" @click="changeRegularPayment(true)">
                  Enable to all</v-btn
                >
              </v-card-actions>
            </v-card>
          </v-dialog>
        </hint-btn>

        <hint-btn
          v-for="button in stateButtons"
          :key="button.title"
          :hint="button.hint"
        >
          <confirm-dialog
            @confirm="
              button.method
                ? button.method()
                : changeStatus(button.newStatusValue)
            "
          >
            <v-btn
              :loading="button.newStatusValue === statusChangeValue"
              :small="viewport < 600"
              class="mr-2"
            >
              <v-icon>{{ button.icon }}</v-icon>
            </v-btn>
          </confirm-dialog>
        </hint-btn>
        <hint-btn hint="Create invoice">
          <v-chip @click="openInvoice" class="ma-1" color="primary" outlined
            >Balance: {{ account.balance?.toFixed(2) || 0 }}
            {{ account.currency?.code }}</v-chip
          >
        </hint-btn>
      </div>
    </div>

    <v-row>
      <v-col lg="2" md="4" sm="6" cols="12">
        <v-text-field v-model="uuid" readonly label="UUID" />
      </v-col>
      <v-col lg="2" md="4" sm="6" cols="12">
        <v-text-field v-model="title" label="name">
          <template v-slot:append>
            <login-in-account-icon :uuid="account.uuid" />
          </template>
        </v-text-field>
      </v-col>
      <v-col lg="2" md="4" sm="6" cols="12">
        <v-select
          :readonly="isCurrencyReadonly"
          :items="currencies"
          v-model="currency"
          item-text="title"
          return-object
          item-value="id"
          label="currency"
        />
      </v-col>

      <v-col lg="2" md="4" sm="6" cols="12">
        <v-text-field v-model="taxRate" label="tax rate" suffix="%" />
      </v-col>

      <v-col
        lg="2"
        md="4"
        sm="6"
        cols="12"
        class="d-flex justify-start align-center"
      >
        <span class="mr-2"> Phone verified </span>
        <v-icon :color="account.isPhoneVerified ? 'green' : 'red'">{{
          account.isPhoneVerified ? "mdi-check-circle" : "mdi-close-circle"
        }}</v-icon>
      </v-col>
    </v-row>

    <v-row>
      <v-col lg="2" md="3" sm="4" cols="1">
        <v-text-field readonly :value="account.data?.email" label="Email" />
      </v-col>

      <v-col lg="2" md="3" sm="4" cols="1">
        <v-text-field readonly :value="account.data?.company" label="Company" />
      </v-col>

      <v-col lg="1" md="2" sm="3" cols="4">
        <v-text-field
          readonly
          :value="
            account.data?.phone_new?.phone_cc
              ? `${account.data?.phone_new?.phone_cc} ${account.data?.phone_new?.phone_number}`
              : account.data?.phone || ''
          "
          label="Phone"
        />
      </v-col>

      <v-col lg="1" md="2" sm="3" cols="4">
        <v-text-field
          readonly
          :value="formatSecondsToDate(account.data?.date_create || 0)"
          label="Date of create"
        />
      </v-col>

      <v-col lg="1" md="2" sm="3" cols="4">
        <v-text-field readonly :value="account.data?.country" label="Country" />
      </v-col>

      <v-col lg="1" md="2" sm="3" cols="4">
        <v-text-field readonly :value="account.data?.city" label="City" />
      </v-col>

      <v-col lg="2" md="3" sm="4" cols="1">
        <v-text-field readonly :value="account.data?.address" label="Address" />
      </v-col>

      <v-col lg="1" md="2" sm="3" cols="4">
        <v-text-field readonly :value="account.data?.whmcs_id" label="WHMCS id">
          <template v-slot:append>
            <whmcs-btn :account="account" />
          </template>
        </v-text-field>
      </v-col>
    </v-row>

    <div class="d-flex align-center">
      <v-card-title class="px-0 instances-panel">Instances:</v-card-title>
      <v-switch
        class="ml-3 mt-5"
        dense
        label="Show deleted"
        v-model="showDeletedInstances"
      />
    </div>
    <instances-table
      table-name="account-instances-table"
      no-search
      :show-select="false"
      :custom-filter="{
        account: [uuid],
        'state.state': !!showDeletedInstances ? [] : [0, 1, 2, 3, 4, 6, 7, 8],
      }"
    />

    <v-card-title class="px-0">SSH keys:</v-card-title>

    <div class="pt-4">
      <v-menu
        bottom
        offset-y
        transition="slide-y-transition"
        v-model="isVisible"
        :close-on-content-click="false"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn class="mr-2" v-bind="attrs" v-on="on"> Create </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row>
            <v-col>
              <v-text-field
                dense
                label="title"
                v-model="newKey.title"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-textarea
                dense
                label="key"
                v-model="newKey.value"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-btn @click="addKey"> Send </v-btn>
            </v-col>
          </v-row>
        </v-card>
      </v-menu>

      <v-btn class="mr-8" :disabled="selected.length < 1" @click="deleteKeys">
        Delete
      </v-btn>
    </div>

    <nocloud-table
      table-name="account-info"
      item-key="value"
      v-model="selected"
      :items="keys"
      :headers="headers"
    />

    <v-btn class="mt-4 mr-2" :loading="isEditLoading" @click="editAccount">
      Save
    </v-btn>
  </v-card>
</template>

<script setup>
import api from "@/api.js";
import nocloudTable from "@/components/table.vue";
import InstancesTable from "@/components/instancesTable.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import hintBtn from "@/components/ui/hintBtn.vue";
import { formatSecondsToDate } from "@/functions";
import whmcsBtn from "@/components/ui/whmcsBtn.vue";
import { computed, onMounted, onUnmounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { useRouter } from "vue-router/composables";

const props = defineProps(["account"]);
const { account } = toRefs(props);

const store = useStore();
const router = useRouter();

const newKey = ref({ title: "", value: "" });
const headers = ref([
  { text: "Title", value: "title" },
  { text: "Key", value: "value" },
]);
const generalRule = ref([(v) => !!v || "Required field"]);
const accountNamespace = ref(null);
const uuid = ref("");
const currency = ref("");
const title = ref("");
const taxRate = ref(0);
const keys = ref([]);
const selected = ref([]);
const isVisible = ref(false);
const isEditLoading = ref(false);
const isChangeRegularPaymentLoading = ref(false);
const isChangeRegularPaymentOpen = ref(false);
const showDeletedInstances = ref(false);
const statusChangeValue = ref("");
const viewport = ref(window.innerWidth);

onMounted(() => {
  title.value = account.value.title;
  currency.value = account.value.currency;
  uuid.value = account.value.uuid;
  keys.value = account.value.data?.ssh_keys || [];
  store.dispatch("services/fetch", { showDeleted: true });
  store.dispatch("servicesProviders/fetch", { anonymously: true });
  fetchNamespace();

  window.addEventListener("resize", setViewport);

  initTaxRate();
});

onUnmounted(() => {
  window.removeEventListener("resize", setViewport);
});

const services = computed(() => store.getters["services/all"]);
const currencies = computed(() =>
  store.getters["currencies/all"].filter((c) => c.title !== "NCU")
);

const instances = computed(() => store.getters["services/getInstances"]);
const accountsByInstance = computed(() =>
  instances.value.filter(
    (i) => i.access.namespace === accountNamespace.value?.uuid
  )
);

const isCurrencyReadonly = computed(
  () => account.value.currency && account.value.currency.code !== "NCU"
);
const isLocked = computed(() => account.value.status !== "ACTIVE");
const stateButtons = computed(() => {
  const status = account.value.status?.toLowerCase();
  const lock = {
    hint: "Delete user",
    icon: "mdi-delete",
    newStatusValue: "PERMANENT_LOCK",
    method: permanentLock,
  };

  switch (status) {
    case "lock": {
      return [
        {
          hint: "Unlock access",
          newStatusValue: "ACTIVE",
          icon: "md-lock-off",
        },
        lock,
      ];
    }
    case "active": {
      return [
        { hint: "Block access", icon: "mdi-lock", newStatusValue: "LOCK" },
        lock,
      ];
    }
    default: {
      return [];
    }
  }
});

const whmcsApi = computed(() => store.getters["settings/whmcsApi"]);

const addKey = () => {
  keys.value.push(newKey.value);
  isVisible.value = false;
  newKey.value = { title: "", value: "" };
};
const deleteKeys = () => {
  if (selected.value.length < 1) return;
  const arr = selected.value.map((el) => el.value);

  keys.value = keys.value.filter((el) => !arr.includes(el.value));
  selected.value = [];
};
const updateAccount = (newAccount) => {
  return api.accounts.update(account.value.uuid, newAccount).catch((err) => {
    store.commit("snackbar/showSnackbarError", {
      message: err,
    });
  });
};
const editAccount = async () => {
  const newAccount = {
    ...account.value,
    title: title.value,
    currency: currency.value,
    data: {
      ...(account.value.data || {}),
      tax_rate: taxRate.value / 100,
    },
  };
  if (!newAccount.data) {
    newAccount.data = {};
  }
  newAccount.data.ssh_keys = keys.value;

  isEditLoading.value = true;
  try {
    await updateAccount(newAccount);
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Account edited successfully",
    });
  } finally {
    isEditLoading.value = false;
  }
};
const changeStatus = async (newStatus) => {
  statusChangeValue.value = newStatus;
  try {
    await fetch(
      /https:\/\/(.+?\.?\/)/.exec(whmcsApi.value)[0] +
        `modules/addons/nocloud/api/index.php?run=status_user&account=${
          account.value.uuid
        }&status=${newStatus === "ACTIVE" ? "open" : "close"}`
    );
    await updateAccount({ ...account.value, status: newStatus });
    account.value.status = newStatus;
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Status change successfully",
    });
  } finally {
    statusChangeValue.value = "";
  }
};
//need remake to instances api
const permanentLock = async () => {
  const newStatus = "PERMANENT_LOCK";
  statusChangeValue.value = newStatus;
  try {
    const accountServices = services.value.filter(
      (s) => s.access.namespace === accountNamespace.value?.uuid
    );

    const servicesForDown = accountServices.filter((s) => s.status !== "INIT");
    await Promise.all(servicesForDown.map((s) => api.services.down(s.uuid)));
    await Promise.all(accountServices.map((s) => api.services.delete(s.uuid)));
    await changeStatus(newStatus);
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error while change status",
    });
  } finally {
    statusChangeValue.value = "";
  }
};
//need remake to instances api
const changeRegularPayment = async (value) => {
  isChangeRegularPaymentLoading.value = true;
  isChangeRegularPaymentOpen.value = false;
  try {
    const services = [];

    accountsByInstance.value.forEach((instance) => {
      const tempService =
        services.find((s) => s.uuid === instance.service) ||
        JSON.parse(
          JSON.stringify(
            services.value.find((s) => s.uuid === instance.service)
          )
        );
      const igIndex = tempService.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === instance.uuid)
      );
      const instanceIndex = tempService.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === instance.uuid);

      instance.config.regular_payment = value;

      tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;

      const sIndex = services.findIndex((s) => s.uuid === instance.service);
      if (sIndex !== -1) {
        services[sIndex] = tempService;
      } else {
        services.push(tempService);
      }
    });
    await Promise.all(services.map((s) => api.services._update(s)));
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error while change invoice based",
    });
  } finally {
    isChangeRegularPaymentLoading.value = false;
  }
};
const openInvoice = async () => {
  router.push({
    name: "Invoice create",
    query: { account: account.value.uuid },
  });
};
const fetchNamespace = async () => {
  try {
    const { pool } = await store.dispatch("namespaces/fetch", {
      filters: { account: account.value.uuid },
    });
    accountNamespace.value = pool[0];
  } catch (e) {
    console.log(e);
  }
};
const setViewport = () => {
  viewport.value = window.innerWidth;
};
const initTaxRate = () => {
  taxRate.value = account.value.data?.tax_rate
    ? account.value.data?.tax_rate * 100
    : 0;
};

watch(account, () => {
  initTaxRate();
});
</script>

<script>
export default {
  name: "account-info",
};
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

.actions {
  position: relative;
  top: -15px;
  right: 40px;
  z-index: 0;

  @media (max-width: 600px) {
    right: auto;
  }
}

.account-additional {
  @media (max-width: 1600px) {
    margin-top: 50px;
  }
  @media (max-width: 1250px) {
    margin-top: 0;
  }
}
</style>
