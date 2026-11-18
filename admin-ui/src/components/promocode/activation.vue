<template>
  <nocloud-table :items="promocode.uses" :headers="headers">
    <template v-slot:[`item.exec`]="{ item }">
      {{ formatSecondsToDate(item.exec, true) }}
    </template>

    <template v-slot:[`item.account`]="{ value }">
      <router-link
        v-if="!isAccountsLoading"
        :to="{ name: 'Account', params: { accountId: value } }"
      >
        {{ accounts[value]?.title }}
      </router-link>
      <account-grade v-if="!isAccountsLoading" :account="accounts[value]" />
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:[`item.instance`]="{ value }">
      <router-link
        v-if="!isInstancesLoading"
        :to="{ name: 'Instance', params: { instanceId: value } }"
      >
        {{ instances[value]?.title }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { ref, toRefs, watch } from "vue";
import useEntityMap from "@/hooks/useEntityMap";
import { formatSecondsToDate } from "../../functions";
import AccountGrade from "@/components/ui/accountGrade.vue";

const props = defineProps(["promocode"]);
const { promocode } = toRefs(props);

const {
  map: accounts,
  isLoading: isAccountsLoading,
  load: loadAccounts,
} = useEntityMap("accounts");
const {
  map: instances,
  isLoading: isInstancesLoading,
  load: loadInstances,
} = useEntityMap("instances");

const headers = ref([
  { text: "Account", value: "account" },
  { text: "Instance", value: "instance" },
  { text: "Execution", value: "exec" },
]);

watch(
  () => promocode.value?.uses,
  (uses = []) => {
    loadAccounts(uses.map(({ account }) => account));
    loadInstances(uses.map(({ instance }) => instance));
  },
  { immediate: true }
);
</script>

<script>
export default {
  name: "promocode-activation",
};
</script>

<style scoped></style>
