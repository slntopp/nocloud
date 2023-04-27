<template>
  <v-dialog
    :persistent="isMoveLoading"
    @input="emit('input', $event)"
    width="50%"
    :value="value"
  >
    <v-card class="pa-5">
      <v-card-title>Move instance</v-card-title>
      <v-select
        label="Account"
        :items="filtredAccounts"
        item-text="title"
        item-value="uuid"
        v-model="selectedAccount"
      />
      <v-select
        v-if="accountsServices?.length > 1"
        :items="accountsServices"
        item-text="title"
        item-value="uuid"
        label="Service"
        return-object
        v-model="selectedService"
      />
      <v-select
        v-if="servicesInstanceGroups?.length > 1"
        :items="servicesInstanceGroups"
        item-text="title"
        item-value="uuid"
        label="Instance group"
        return-object
        v-model="selectedIG"
      />
      <v-card-actions class="d-flex justify-end">
        <v-btn @click="emit('input', false)" :disabled="isMoveLoading"
          >Close</v-btn
        >
        <v-btn
          :loading="isMoveLoading"
          @click="move"
          :disabled="!newInstanceGroup"
          >Move</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, defineProps, defineEmits, computed, toRefs } from "vue";
import api from "@/api";

const props = defineProps([
  "value",
  "account",
  "accounts",
  "namespaces",
  "services",
  "template",
]);
const { namespaces, account, services, value, accounts, template } =
  toRefs(props);
const emit = defineEmits(["refresh", "input"]);

const selectedAccount = ref("");
const selectedService = ref("");
const selectedIG = ref("");
const isMoveLoading = ref(false);

const filtredAccounts = computed(() =>
  accounts.value.filter(
    (a) => a.uuid !== account.value?.uuid && getServicesByAccount(a).length > 0
  )
);
const namespace = computed(() => {
  return namespaces.value?.find(
    (n) => n.access.namespace == selectedAccount.value
  );
});

const getServicesByAccount = (account) => {
  const namespace = namespaces.value?.find(
    (n) => n.access.namespace == account.uuid
  );
  return (
    services.value?.filter((s) => s.access.namespace == namespace.uuid) || []
  );
};

const accountsServices = computed(() => {
  if (!namespace.value) {
    return null;
  }

  return services.value?.filter(
    (s) => s.access.namespace == namespace.value.uuid
  );
});

const servicesInstanceGroups = computed(() => {
  return (
    selectedService.value || accountsServices.value?.[0]
  )?.instancesGroups.filter((ig) => ig.type === template.value.type);
});

const newInstanceGroup = computed(() => {
  if (servicesInstanceGroups.value?.length < 2) {
    return servicesInstanceGroups.value[0];
  } else {
    return selectedIG.value;
  }
});

const move = async () => {
  isMoveLoading.value = true;
  try {
    await api.instances.move(template.value.uuid, newInstanceGroup.value.uuid);
    emit('refresh')
  } finally {
    isMoveLoading.value = false;
  }
};
</script>

<style scoped></style>
