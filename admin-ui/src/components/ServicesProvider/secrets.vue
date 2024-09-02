<template>
  <v-card color="background-light" class="pa-5">
    <component v-if="component" :is="component" :template="template" />
    <json-editor :json="newSecrets" @changeValue="newSecrets = $event" />

    <div class="d-flex jsutify-end align-center">
      <v-btn :loading="isSaveLoading" @click="updateSp">Update</v-btn>
    </div>
  </v-card>
</template>

<script setup>
import { defineAsyncComponent, ref, toRefs } from "vue";
import JsonEditor from "@/components/JsonEditor.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const newSecrets = ref({ ...template.value.secrets });
const isSaveLoading = ref(false);

const component = defineAsyncComponent(() =>
  import(
    `@/components/modules/${template.value?.type}/serviceProviderSecrets.vue`
  )
);

const updateSp = async () => {
  console.log(newSecrets.value);
  isSaveLoading.value = true;
  try {
    await api.servicesProviders.update(template.value.uuid, {
      ...template.value,
      secrets: { ...newSecrets.value },
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save offering items",
    });
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<script>
export default {
  name: "sp-secrets",
};
</script>
