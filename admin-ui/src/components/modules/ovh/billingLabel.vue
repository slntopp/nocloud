<template>
  <v-container>
    <v-row>
      <v-col>
        <v-chip outlined color="primary">Due to date: {{ dueDate }}</v-chip>
      </v-col>
      <v-col>
        <v-chip color="primary" outlined>Price: {{ price }}</v-chip>
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
  </v-container>
</template>

<script>
import { computed, ref } from "vue";
import { getOvhPrice } from "@/functions";
import { useStore } from "@/store";
import sendVmAction from "@/mixins/sendVmAction";
import confirmDialog from "@/components/confirmDialog.vue";

export default {
  components: { confirmDialog },
  mixins: [sendVmAction],
  props: ["template"],
  setup(props) {
    const store = useStore();
    const isDisabled = ref(false);
    const isLoading = ref(false);

    const dueDate = computed(() => {
      return props.template.data.expiration;
    });

    const tariffPrice = computed(() => {
      const { duration, planCode } = props.template.config;
      const key = `${duration} ${planCode}`;

      return props.template.billingPlan.products[key]?.price ?? 0;
    });
    const addonsPrice = ref(props.template.config.addons?.reduce((res, addon) => {
      const { price } = props.template.billingPlan.resources?.find(
        ({ key }) => key === `${props.template.config.duration} ${addon}`
      ) ?? {};
      let key = '';

      if (addon.includes('ram')) return res;
      if (addon.includes('raid')) return res;
      if (addon.includes('vrack')) key = 'Vrack';
      if (addon.includes('bandwidth')) key = 'Traffic';
      if (addon.includes('additional')) key = 'Additional drive';
      if (addon.includes('snapshot')) key = 'Snapshot';
      if (addon.includes('backup')) key = 'Backup';
      if (addon.includes('windows')) key = 'Windows';

      return { ...res, [key]: +price || 0 };
    }, {}));

    const currency = computed(() => ({
      code: this.$store.getters["currencies/default"]
    }))
    const price = computed(() => {
      return getOvhPrice(props.template);
    });

    const isRenewDisabled = computed(() => {
      return getAccountBalance() < getOvhPrice(props.template) ||
        props.template.data.blocked ||
        isDisabled.value;
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
      isLoading.value = true;
      sendVmAction.methods.sendVmAction('manual_renew', props.template)
        .then(() => { isDisabled.value = true })
        .finally(() => { isLoading.value = false });
    }

    const addonsTemplate = Object.entries(addonsPrice.value).map(([key, value]) =>
      `<li>${key}: ${value} ${currency.value.code}</li>`
    );

    const renewTemplate = `
      <div style="font-size: 16px; white-space: initial">
        <div>Manual renewal:</div>
        <span style="font-weight: 700">Tariff price: </span>
        ${tariffPrice.value} ${currency.value.code}
        <div>
          <span style="font-weight: 700">Addons prices:</span>
          ${(addonsTemplate.value)
            ? `<ul style="list-style: '-  '; padding-left: 25px; margin-bottom: 5px">
                ${ addonsTemplate.join('') }
              </ul>`
            : `0 ${currency.value.code}
          `}
        </div>

        <div>
          <span style="font-weight: 700">Total: </span>
          ${price.value} ${currency.value.code}
        </div>
      </div>
    `.trim();

    return {
      isRenewDisabled,
      isLoading,
      price,
      dueDate,
      sendRenew,
      renewTemplate
    };
  },
};
</script>

<style scoped></style>
