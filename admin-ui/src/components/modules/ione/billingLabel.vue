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
    <v-row class="mt-0" align="center" justify="end">
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
import { computed, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import { useStore } from "@/store";
import confirmDialog from "@/components/confirmDialog.vue";
import InstanceState from "@/components/ui/instanceState.vue";

const props = defineProps(["template"]);

const store = useStore();

const { template } = toRefs(props);

const tariffPrice = ref(
  template.value.billingPlan.products[template.value.product]?.price ?? 0
);
const addonsPrice = ref(
  template.value.billingPlan.resources.reduce((prev, curr) => {
    if (
      curr.key === `drive_${template.value.resources.drive_type.toLowerCase()}`
    ) {
      const key = "drive";

      return {
        ...prev,
        [key]: (curr.price * template.value.resources.drive_size) / 1024,
      };
    } else if (curr.key === "ram") {
      const key = "ram";

      return {
        ...prev,
        [key]: (curr.price * template.value.resources.ram) / 1024,
      };
    } else if (template.value.resources[curr.key]) {
      const key = curr.key.replace("_", " ");

      return {
        ...prev,
        [key]: curr.price * template.value.resources[curr.key],
      };
    }
    return prev;
  }, {})
);

const currency = computed(() => ({
  code: store.getters["currencies/default"],
}));
const price = computed(() => {
  const addonsSum = Object.values(addonsPrice.value).reduce((a, b) => a + b);

  return (tariffPrice.value + addonsSum)?.toFixed(2);
});

const isRenewDisabled = computed(() => {
  return getAccountBalance() < price.value || template.value.data.blocked;
});
const getAccountBalance = () => {
  const namespace = store.getters["namespaces/all"]?.find(
    (n) => n.uuid === template.value.access.namespace
  );
  const account = store.getters["accounts/all"].find(
    (a) => a.uuid === namespace.access.namespace
  );
  return account?.balance;
};

const dueDate = computed(() => {
  return formatSecondsToDate(
    +template.value?.data?.last_monitoring +
      +template.value.billingPlan.products[template.value.product].period
  );
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
  ([key, value]) => `<li>${key}: ${value} ${currency.value.code}</li>`
);

const isLoading = computed(() => store.getters["actions/isSendActionLoading"]);

const renewTemplate = `
      <div style="font-size: 16px; white-space: initial">
        <div>Manual renewal:</div>
        <span style="font-weight: 700">Tariff price: </span>
        ${tariffPrice.value} ${currency.value.code}
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
          ${price.value} ${currency.value.code}
        </div>
      </div>
    `.trim();
</script>

<style scoped></style>
