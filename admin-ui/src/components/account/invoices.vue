<template>
  <div>
    <div class="d-flex justify-end">
      <invoices-actions
        :selected-invoices="selectedInvoices"
        @input="selectedInvoices = $event"
        @refresh="onRefresh"
        :account-uuid="account.uuid"
      />
    </div>
    <invoices-table
      table-name="account-invoices"
      no-search
      v-model="selectedInvoices"
      :custom-filter="{ account: [account.uuid] }"
      :refetch="refetch"
    />
  </div>
</template>

<script setup>
import { ref, toRefs } from "vue";
import InvoicesTable from "../invoicesTable.vue";
import InvoicesActions from "../invoicesActions.vue";

const props = defineProps(["account"]);
const { account } = toRefs(props);

const selectedInvoices = ref([]);
const refetch = ref(false);

const onRefresh = () => {
  console.log(2);
  
  refetch.value = !refetch.value;
};
</script>

<script>
export default {
  name: "account-reports",
};
</script>

<style scoped lang="scss"></style>
