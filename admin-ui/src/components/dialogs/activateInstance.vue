<template>
  <instance-control-btn hint="Activate">
    <v-dialog v-model="isModalOpen" max-width="500">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          :disabled="disabled"
          :loading="loading || isInstanceSaveLoading"
          class="ma-1"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon>mdi-checkbox-marked-circle-auto-outline</v-icon>
        </v-btn>
      </template>
      <v-card color="background-light">
        <v-card-title>
          This action will start the server and begin billing immediately.
          Please confirm that you want to proceed with resource activation.
        </v-card-title>

        <v-card-title>
          The user's balance will be reduced by the amount of the purchase of
          this service {{ formatPrice(template.estimate) }}
          {{ defaultCurrency.code }}
        </v-card-title>
        <v-card-actions class="d-flex justify-end">
          <v-btn class="mr-2" @click="start(false)"> No </v-btn>
          <v-btn class="mr-2" @click="start(true)"> Yes </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </instance-control-btn>
</template>

<script setup>
import { toRefs, ref } from "vue";
import { useStore } from "@/store";
import InstanceControlBtn from "@/components/ui/hintBtn.vue";
import { formatPrice } from "@/functions";
import useCurrency from "@/hooks/useCurrency";

const props = defineProps(["disabled", "loading", "template"]);
const emit = defineEmits(["click"]);

const store = useStore();
const { defaultCurrency } = useCurrency();

const { disabled, loading, template } = toRefs(props);

const isInstanceSaveLoading = ref(false);
const isModalOpen = ref(false);

const start = async (proceed) => {
  isModalOpen.value = false;

  if (!proceed) {
    return;
  }

  try {
    isInstanceSaveLoading.value = true;
    await store.getters["instances/instancesClient"].start({
      id: template.value.uuid,
    });

    emit("click");
  } catch (err) {
    store.commit("snackbar/showSnackbarError", { message: err });
  } finally {
    isInstanceSaveLoading.value = false;
  }
};
</script>

<style scoped></style>
