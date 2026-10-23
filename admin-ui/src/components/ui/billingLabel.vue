<template>
  <v-container>
    <!-- one wrapping flex row: when it runs out of width whole controls move
         down together, instead of the switch label breaking in half -->
    <div class="d-flex flex-wrap justify-end align-center billing-actions">
      <template v-if="isAiPackage">
        <v-progress-circular
          v-if="isUsageLoading"
          indeterminate
          size="16"
          width="2"
        />
        <v-tooltip v-else bottom>
          <template v-slot:activator="{ on, attrs }">
            <v-chip small outlined color="primary" v-bind="attrs" v-on="on">{{
              creditText
            }}</v-chip>
          </template>
          <span>{{ creditHint }}</span>
        </v-tooltip>
        <v-tooltip bottom>
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              icon
              color="primary"
              v-bind="attrs"
              v-on="on"
              :loading="isChatLoading"
              @click="openChat"
            >
              <v-icon>mdi-chat-outline</v-icon>
            </v-btn>
          </template>
          <span>AI chat as client</span>
        </v-tooltip>
      </template>
      <v-switch
        class="ma-0 pa-0"
        hide-details
        dense
        :input-value="template.meta?.autoRenew"
        @change="updateAutomaticDebit"
        :disabled="isDeleted"
        label="Auto debit"
      />
      <confirm-dialog
        title="Do you want to renew server?"
        :text="renewTemplate"
        :disabled="isDeleted || isRenewDisabled"
        :width="500"
        :success-disabled="isRenewDisabled"
        @confirm="sendRenew"
      >
        <v-btn color="primary" :disabled="isRenewDisabled" :loading="isLoading">
          Renew
        </v-btn>
      </confirm-dialog>
    </div>
    <v-row class="mt-0" align="center" justify="end">
      <v-col class="d-flex justify-end px-1">
        <instance-state :template="template" />
      </v-col>
      <v-col class="d-flex justify-end px-1">
        <v-chip color="primary" outlined
          >Price: {{ formatPrice(price, accountCurrency) }}
          {{ accountCurrency?.code }}</v-chip
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
import useAiPackage from "@/hooks/useAiPackage";
import { formatPrice } from "../../functions";

const props = defineProps({
  template: {},
  dueDate: {},
  addonsPrice: {},
  tariffPrice: {},
  account: {},
  renewDisabled: { type: Boolean, default: false },
});
const { template, addonsPrice, tariffPrice, dueDate, account,  } =
  toRefs(props);

const emit = defineEmits(["update"]);

const store = useStore();
const { convertTo } = useCurrency();

const currency = computed(() => ({
  code: store.getters["currencies/default"].title,
}));

const accountCurrency = computed(() => {
  return account.value?.currency;
});

// empty for anything that is not an AI package, so the chip and the button stay hidden
const {
  isAiPackage,
  credit,
  usage,
  usageError,
  isUsageLoading,
  amount,
  openChat,
  isChatLoading,
} = useAiPackage(template, account);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const creditText = computed(() => {
  const left = usage.value ? amount(usage.value.left) : "-";
  const total = amount(usage.value ? usage.value.total : credit.value);
  return `${left} / ${total} ${defaultCurrency.value?.code}`;
});

const creditHint = computed(() =>
  usage.value
    ? `Credit left, used ${amount(usage.value.total - usage.value.left)}`
    : usageError.value
);

const priceForAddons = computed(() => {
  return convertTo(
    Object.values(addonsPrice.value).reduce((a, b) => a + b, 0),
    accountCurrency.value
  );
});

const priceForTarrif = computed(() => {
  return convertTo(tariffPrice.value, accountCurrency.value);
});

const price = computed(() => {
  return +priceForAddons.value + +priceForTarrif.value;
});

const isRenewDisabled = computed(() => {
  return (
    isDeleted.value ||
    template.value.data?.blocked ||
    (account.value?.balance || 0) < price.value
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
      if (!template.value.data) {
        template.value.data = {};
      }

      template.value.data.blocked = true;
      template.value.data = Object.assign({}, template.value.data);
    });
}

const addonsTemplate = computed(() =>
  Object.entries(addonsPrice.value).map(
    ([key, value]) =>
      `<li>${key}: ${(value || 0).toFixed(2)} ${currency.value?.code}</li>`
  )
);

const isLoading = computed(() => store.getters["actions/isSendActionLoading"]);

const renewTemplate = computed(() =>
  `
      <div style="font-size: 16px; white-space: initial">
        <div>Manual renewal:</div>
        <span style="font-weight: 700">Tariff price: </span>
        ${formatPrice(priceForTarrif.value, accountCurrency.value)} ${
    accountCurrency.value?.code
  }
        ${
          priceForAddons.value
            ? `
          <div>
            <span style="font-weight: 700">Addons prices:</span>
            <ul style="list-style: '-  '; padding-left: 25px; margin-bottom: 5px">
              ${addonsTemplate.value.join("")}
            </ul>
          </div>
        `
            : ""
        }

        <div>
          <span style="font-weight: 700">Total: </span>
          ${formatPrice(price.value, accountCurrency.value)} ${
    accountCurrency.value?.code
  }
        </div>
      </div>
    `.trim()
);

const updateAutomaticDebit = (value) => {
  emit("update", { key: "meta.autoRenew", value: !!value });
};
</script>

<style scoped>
.billing-actions {
  gap: 8px;
}

/* "Automatic debit" is the first thing to break when the row gets tight */
.billing-actions ::v-deep .v-label {
  white-space: nowrap;
}
</style>
