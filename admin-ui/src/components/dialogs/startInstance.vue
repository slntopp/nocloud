<template>
  <instance-control-btn hint="Start">
    <v-dialog v-model="isModalOpen" max-width="500">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          :disabled="disabled"
          :loading="loading || isInstanceSaveLoading"
          class="ma-1"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon>mdi-power</v-icon>
        </v-btn>
      </template>
      <v-card color="background-light">
        <v-card-title
          >Make a payment now (balance will be debited)?</v-card-title
        >
        <v-card-actions class="d-flex justify-end">
          <v-btn class="mr-2" @click="start(true)"> No </v-btn>
          <v-btn class="mr-2" @click="start(false)"> Yes </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </instance-control-btn>
</template>

<script setup>
import { toRefs, ref, computed } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import InstanceControlBtn from "@/components/ui/hintBtn.vue";

const props = defineProps(["disabled", "loading", "template"]);
const emit = defineEmits(["click"]);

const store = useStore();

const { disabled, loading, template } = toRefs(props);

const isInstanceSaveLoading = ref(false);
const isModalOpen = ref(false);

const service = computed(() =>
  store.getters["services/all"]?.find((s) => s.uuid == template.value.service)
);

const start = async (notSkip) => {
  isModalOpen.value = false;

  if (!notSkip) {
    return emit("click");
  }

  const tempService = JSON.parse(JSON.stringify(service.value));
  const instance = JSON.parse(JSON.stringify(template.value));
  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  const skipped = [];
  skipped.push(template.value.product);
  switch (template.value.type) {
    case "keyweb": {
      skipped.push(
        ...Object.values(template.value?.config?.configurations || {})
      );
      break;
    }
    default: {
      skipped.push(...(template.value?.config?.addons || []));
    }
  }
  instance.config = { ...instance.config, skip_next_payment: skipped };

  tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;

  try {
    isInstanceSaveLoading.value = true;
    await api.services._update(tempService);
    emit("click", instance);
  } catch (err) {
    store.commit("snackbar/showSnackbarError", { message: err });
  } finally {
    isInstanceSaveLoading.value = false;
  }
};
</script>

<style scoped></style>
