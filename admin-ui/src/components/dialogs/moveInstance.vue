<template>
  <v-dialog
    :persistent="isMoveLoading"
    @update:model-value="emit('update:modelValue', $event)"
    width="50%"
    :model-value="value"
  >
    <v-card class="pa-5">
      <v-card-title>Move instance</v-card-title>
      <accounts-autocomplete
        label="Account"
        return-object
        v-model="selectedAccount"
        advanced
      />

      <v-switch
        label="Do not transfer invoices?"
        v-model="doNotTransferInvoices"
      />

      <v-card-actions class="d-flex justify-end">
        <v-btn @click="emit('update:modelValue', false)" :disabled="isMoveLoading"
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
import { ref, toRefs } from "vue";
import { useStore } from "@/store";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";
import { TransferInstanceRequest } from "nocloud-proto/proto/es/instances/instances_pb";

const props = defineProps(["modelValue", "template"]);
const { modelValue: value, template } = toRefs(props);
const emit = defineEmits(["refresh", "update:modelValue"]);

const store = useStore();

const selectedAccount = ref("");
const isMoveLoading = ref(false);
const doNotTransferInvoices = ref(true);

const move = async () => {
  isMoveLoading.value = true;
  try {
    await store.getters["instances/instancesClient"].transferInstance(
      TransferInstanceRequest.fromJson({
        uuid: template.value.uuid,
        account: selectedAccount.value.uuid,
        doNotTransferInvoices: doNotTransferInvoices.value,
      })
    );

    emit("refresh");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e?.rawMessage || "Error during move instance",
    });
  } finally {
    isMoveLoading.value = false;
  }
};
</script>

<style scoped></style>
