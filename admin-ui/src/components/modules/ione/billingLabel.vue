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
        <v-btn
          color="primary"
          :disabled="isRenewDisabled"
          :loading="isLoading"
          @click="sendRenew"
        >
          Renew
        </v-btn>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { computed, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import { useStore } from "@/store";
import sendVmAction from "@/mixins/sendVmAction";

export default {
  mixins: [sendVmAction],
  props: ["template"],
  setup(props) {
    const store = useStore();

    const { template } = toRefs(props);
    const isLoading = ref(false);

    const price = computed(() => {
      const initialPrice =
        template.value.billingPlan.products[template.value.product]?.price ?? 0;
      return +template.value.billingPlan.resources
        .reduce((prev, curr) => {
          if (
            curr.key ===
            `drive_${template.value.resources.drive_type.toLowerCase()}`
          ) {
            return (
              prev + (curr.price * template.value.resources.drive_size) / 1024
            );
          } else if (curr.key === "ram") {
            return prev + (curr.price * template.value.resources.ram) / 1024;
          } else if (template.value.resources[curr.key]) {
            return prev + curr.price * template.value.resources[curr.key];
          }
          return prev;
        }, initialPrice)
        ?.toFixed(2);
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
      return formatSecondsToDate(template.value?.data?.last_monitoring);
    });

    function sendRenew() {
      isLoading.value = true;
      sendVmAction.methods.sendVmAction('manual_renew', template.value)
        .then(() => {
          template.value.data.blocked = true;
          template.value.data = Object.assign({}, template.value.data);
        })
        .finally(() => { isLoading.value = false });
    }

    return {
      isRenewDisabled,
      isLoading,
      price,
      dueDate,
      sendRenew
    };
  },
};
</script>

<style scoped></style>
