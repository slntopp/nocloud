<template>
  <v-dialog
      :persistent="isMoveLoading"
      @input="emit('input', $event)"
      width="50%"
      :value="value"
  >
    <v-card class="pa-5">
      <v-card-title>Move instance</v-card-title>
      <v-autocomplete
          label="Account"
          :items="filtredAccounts"
          item-text="title"
          item-value="uuid"
          return-object
          v-model="selectedAccount"
      />
      <v-autocomplete
          v-if="accountsServices?.length > 1"
          :items="accountsServices"
          item-text="title"
          item-value="uuid"
          label="Service"
          return-object
          v-model="selectedService"
      />
      <v-card-actions class="d-flex justify-end">
        <v-btn @click="emit('input', false)" :disabled="isMoveLoading"
        >Close
        </v-btn>
        <v-btn
            :loading="isMoveLoading"
            @click="move"
            :disabled="!selectedAccount"
        >Move
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { computed, defineEmits, defineProps, ref, toRefs } from "vue";
import api from "@/api";
import { useStore } from "@/store";

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

const store = useStore();

const selectedAccount = ref("");
const selectedService = ref("");
const isMoveLoading = ref(false);
const newIg = ref({});

const filtredAccounts = computed(() =>
    accounts.value.filter((a) => a?.uuid !== account.value?.uuid)
);
const namespace = computed(() => {
  return namespaces.value?.find(
      (n) => n.access.namespace == selectedAccount.value?.uuid
  );
});

const accountsServices = computed(() => {
  if (!namespace.value) {
    return null;
  }

  return services.value?.filter(
      (s) => s.access.namespace == namespace.value?.uuid
  );
});

const servicesInstanceGroups = computed(() => {
  return (
      service.value?.instancesGroups.filter(
          (ig) => ig.type === template.value.type
      ) || []
  );
});

const instancesGroups = computed(() => {
  if (newIg.value) {
    return [...servicesInstanceGroups.value, newIg.value];
  }
  return servicesInstanceGroups.value;
});

const service = computed(
    () => selectedService.value || accountsServices.value?.[0]
);

const move = async () => {
  isMoveLoading.value = true;
  try {
    if (!service.value) {
      selectedService.value = await api.services.create({
        namespace: namespaces.value.find(
            (n) => n.access.namespace == selectedAccount.value.uuid
        )?.uuid,
        service: {
          version: "1",
          title: selectedAccount.value.title,
          context: {},
          instancesGroups: [],
        },
      });
      await api.services.up(service.value.uuid);
    }
    
    let newIg = instancesGroups.value.find(
        (ig) => ig.type === template.value.type && ig.sp===template.value.sp
    );

    if (!newIg) {
      const newService = JSON.parse(JSON.stringify(service.value));
      newService.instancesGroups.push({
        type: template.value.type,
        sp: template.value.sp,
        title: selectedAccount.value.title + Date.now(),
        instances: [],
      });
      const data = await api.services._update(newService);

      newIg = data.instancesGroups.find(
          (ig) => ig.type === template.value.type
      );
    }

    await api.instances.move(template.value.uuid, newIg.uuid);
    emit("refresh");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during move instance",
    });
  } finally {
    isMoveLoading.value = false;
  }
};
</script>

<style scoped></style>
