<template>
  <v-container>
    <v-row>
      <v-col>
        <v-chip outlined color="primary">Due to date: {{ dueDate }}</v-chip>
      </v-col>
      <v-col>
        <confirm-dialog
          title="Do you want to renew server?"
          :text="renewTemplate"
          :width="500"
          :success-disabled="isRenewDisabled"
          @confirm="sendRenew"
        >
          <v-btn
            color="primary"
            :disabled="isRenewDisabled"
            :loading="isLoading"
          >
            Renew
          </v-btn>
        </confirm-dialog>
      </v-col>
    </v-row>
    <v-row justify="end" align="center">
      <v-col class="d-flex justify-end">
        <v-chip color="primary" outlined>Price: {{ price }}</v-chip>
      </v-col>
      <v-col class="d-flex justify-end">
        <instance-state :state="template.state.state" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { computed, ref } from "vue";
import { formatSecondsToDate, getOvhPrice } from "@/functions";
import { useStore } from "@/store";
import confirmDialog from "@/components/confirmDialog.vue";
import InstanceState from "@/components/ui/instanceState.vue";

const props = defineProps(["template"]);

const store = useStore();
const isDisabled = ref(false);

const dueDate = computed(() => {
  const { duration, planCode } = props.template.config;
  const key = `${duration} ${planCode}`;
  if (props.template.config.type === "cloud") {
    return formatSecondsToDate(
      +props.template?.data?.last_monitoring +
        +props.template.billingPlan.products[key].period
    );
  }

  return props.template.data.expiration;
});

const isLoading = computed(() => store.getters["actions/isSendActionLoading"]);

const tariffPrice = computed(() => {
  const { duration, planCode } = props.template.config;
  const key = `${duration} ${planCode}`;

  return props.template.billingPlan.products[key]?.price ?? 0;
});
const addonsPrice = ref(
  props.template.config.addons?.reduce((res, addon) => {
    const { price } =
      props.template.billingPlan.resources?.find(
        ({ key }) => key === `${props.template.config.duration} ${addon}`
      ) ?? {};
    let key = "";

    if (addon.includes("ram")) return res;
    if (addon.includes("raid")) return res;
    if (addon.includes("vrack")) key = "Vrack";
    if (addon.includes("bandwidth")) key = "Traffic";
    if (addon.includes("additional")) key = "Additional drive";
    if (addon.includes("snapshot")) key = "Snapshot";
    if (addon.includes("backup")) key = "Backup";
    if (addon.includes("windows")) key = "Windows";

    return { ...res, [key]: +price || 0 };
  }, {})
);

const currency = computed(() => ({
  code: store.getters["currencies/default"],
}));
const price = computed(() => {
  return getOvhPrice(props.template);
});

const isRenewDisabled = computed(() => {
  return (
    getAccountBalance() < getOvhPrice(props.template) ||
    props.template.data.blocked ||
    isDisabled.value
  );
});
const getAccountBalance = () => {
  const namespace = store.getters["namespaces/all"]?.find(
    (n) => n.uuid === props.template.access.namespace
  );
  const account = store.getters["accounts/all"].find(
    (a) => a.uuid === namespace.access.namespace
  );
  return account.balance;
};

function sendRenew() {
  store
    .dispatch("actions/sendVmAction", {
      action: "manual_renew",
      template: props.template,
    })
    .then(() => {
      isDisabled.value = true;
    });
}

const addonsTemplate = Object.entries(addonsPrice.value).map(
  ([key, value]) => `<li>${key}: ${value} ${currency.value.code}</li>`
);

const renewTemplate = `
      <div style="font-size: 16px; white-space: initial">
        <div>Manual renewal:</div>
        <span style="font-weight: 700">Tariff price: </span>
        ${tariffPrice.value} ${currency.value.code}
        <div>
          <span style="font-weight: 700">Addons prices:</span>
          ${
            addonsTemplate.value
              ? `<ul style="list-style: '-  '; padding-left: 25px; margin-bottom: 5px">
                ${addonsTemplate.join("")}
              </ul>`
              : `0 ${currency.value.code}
          `
          }
        </div>

        <div>
          <span style="font-weight: 700">Total: </span>
          ${price.value} ${currency.value.code}
        </div>
      </div>
    `.trim();
</script>

<style scoped></style>
