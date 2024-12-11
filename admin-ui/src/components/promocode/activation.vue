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
        {{ account(value) && accounts[value]?.title }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:[`item.instance`]="{ value }">
      <router-link
        v-if="!isInstancesLoading"
        :to="{ name: 'Instance', params: { instanceId: value } }"
      >
        {{ instance(value) && instances[value]?.instance?.title }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { ref, toRefs } from "vue";
import api from "@/api";
import { formatSecondsToDate } from "../../functions";
import { useStore } from "@/store";

const props = defineProps(["promocode"]);
const { promocode } = toRefs(props);

const store = useStore();

const isAccountsLoading = ref(false);
const accounts = ref({});

const isInstancesLoading = ref(false);
const instances = ref({});

const headers = ref([
  { text: "Account", value: "account" },
  { text: "Instance", value: "instance" },
  { text: "Execution", value: "exec" },
]);

const account = async (uuid) => {
  if (accounts.value[uuid]) {
    return accounts.value[uuid] === 1 ? {} : accounts.value[uuid];
  }
  isAccountsLoading.value = true;

  try {
    accounts.value[uuid] = 1;

    accounts.value[uuid] = await api.accounts.get(uuid);
  } finally {
    setTimeout(() => {
      isAccountsLoading.value = Object.keys(accounts.value).every(
        (key) => accounts.value[key] === 1
      );
    }, 0);
  }
};

const instance = async (uuid) => {
  if (instances.value[uuid]) {
    return instances.value[uuid] === 1 ? {} : instances.value[uuid];
  }
  isInstancesLoading.value = true;

  try {
    instances.value[uuid] = 1;

    instances.value[uuid] = await store.dispatch("instances/get", uuid);
  } catch {
    instances.value[uuid] = "no";
  } finally {
    setTimeout(() => {
      isInstancesLoading.value = Object.keys(instances.value).every(
        (key) => instances.value[key] === 1
      );
    }, 0);
  }
};
</script>

<script>
export default {
  name: "promocode-activation",
};
</script>

<style scoped></style>
