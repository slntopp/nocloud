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
import { computed, ref } from "vue";
import { getOvhPrice } from "@/functions";
import { useStore } from "@/store";
import sendVmAction from "@/mixins/sendVmAction";

export default {
  mixins: [sendVmAction],
  props: ["template"],
  setup(props) {
    const store = useStore();
    const isDisabled = ref(false);
    const isLoading = ref(false);

    const dueDate = computed(() => {
      return props.template.data.expiration;
    });

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
