<template>
  <v-container>
    <v-row justify="end">
      <v-col
        style="max-height: 50px"
        class="d-flex justify-end align-start pa-0"
      >
        <v-switch
          hide-details
          dense
          :input-value="template.config.regular_payment"
          @change="
            emit('update', { key: 'config.regular_payment', value: !!$event })
          "
          :disabled="isDeleted"
          label="Invoice based"
        />
      </v-col>
      <v-col
        style="max-height: 50px; max-width: 90px"
        class="d-flex justify-end align-start pa-0"
      >
        <v-switch
          :disabled="isDeleted || isRenewDisabled"
          hide-details
          dense
          :input-value="template.config.auto_renew"
          @change="
            emit('update', { key: 'config.auto_renew', value: !!$event })
          "
          label="Auto"
        />
      </v-col>
      <v-col class="d-flex justify-end" style="max-width: 120px">
        <confirm-dialog
          title="Do you want to renew server?"
          :text="renewTemplate"
          :disabled="isRenewDisabled || isDeleted"
          :width="500"
          :success-disabled="isRenewDisabled"
          @confirm="sendRenew"
        >
          <v-btn
            color="primary"
            :disabled="isRenewDisabled || isDeleted"
            :loading="isLoading"
          >
            Renew
          </v-btn>
        </confirm-dialog>
      </v-col>
    </v-row>
    <v-row class="mt-0" align="center" justify="end">
      <v-col class="d-flex justify-end px-1">
        <instance-state :template="template" />
      </v-col>
      <v-col class="d-flex justify-end px-1">
        <v-chip color="primary" outlined
          >Price: {{ price }} {{ accountCurrency }}</v-chip
        >
      </v-col>
      <v-col class="px-1 d-flex justify-end">
        <v-chip outlined color="primary">Due to date: {{ dueDate }}</v-chip>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { computed, toRefs } from "vue";
import { useStore } from "@/store";
import confirmDialog from "@/components/confirmDialog.vue";
import InstanceState from "@/components/ui/instanceState.vue";
import useCurrency from "@/hooks/useCurrency";

const props = defineProps({
  template: {},
  dueDate: {},
  addonsPrice: {},
  tariffPrice: {},
  account: {},
  renewDisabled: { type: Boolean, default: false },
});
const { template, addonsPrice, tariffPrice, dueDate, account, renewDisabled } =
  toRefs(props);

const emit = defineEmits(["update"]);

const store = useStore();
const { convertTo } = useCurrency();

const currency = computed(() => ({
  code: store.getters["currencies/default"],
}));

const accountCurrency = computed(() => {
  return account.value?.currency;
});

const price = computed(() => {
  const addonsSum = Object.values(addonsPrice.value).reduce((a, b) => a + b, 0);

  return convertTo(tariffPrice.value + addonsSum, accountCurrency.value);
});

const isRenewDisabled = computed(() => {
  return (
    (account.value?.balance || 0) < price.value ||
    template.value.data.blocked ||
    renewDisabled.value
  );
});

const isDeleted = computed(() => {
  return template.value.state?.state === "DELETED";
});

function sendRenew() {
  store
    .dispatch("actions/sendVmAction", {
      action: "manual_renew",
      template: template.value,
    })
    .then(() => {
      template.value.data.blocked = true;
      template.value.data = Object.assign({}, template.value.data);
    });
}

const addonsTemplate = Object.entries(addonsPrice.value).map(
  ([key, value]) =>
    `<li>${key}: ${(value || 0).toFixed(2)} ${currency.value?.code}</li>`
);

const isLoading = computed(() => store.getters["actions/isSendActionLoading"]);

const renewTemplate = `
      <div style="font-size: 16px; white-space: initial">
        <div>Manual renewal:</div>
        <span style="font-weight: 700">Tariff price: </span>
        ${tariffPrice.value} ${currency.value?.code}
        ${
          addonsPrice.value
            ? `
          <div>
            <span style="font-weight: 700">Addons prices:</span>
            <ul style="list-style: '-  '; padding-left: 25px; margin-bottom: 5px">
              ${addonsTemplate.join("")}
            </ul>
          </div>
        `
            : ""
        }

        <div>
          <span style="font-weight: 700">Total: </span>
          ${price.value} ${currency.value?.code}
        </div>
      </div>
    `.trim();
</script>

<style scoped></style>
