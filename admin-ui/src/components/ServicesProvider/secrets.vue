<template>
  <v-card color="background-light" class="pa-5">
    <div class="d-flex flex-wrap align-center mb-4" style="gap: 12px">
      <template v-if="!revealed">
        <span>
          Secret values are hidden. A key set to
          <code>{{ REDACTED_SECRET }}</code> keeps its stored value; change it
          to replace the value, delete the key to remove it.
        </span>
        <v-btn :loading="isRevealLoading" @click="reveal">Show values</v-btn>
      </template>
      <template v-else>
        <span>Showing the stored secret values.</span>
        <v-btn @click="hide">Hide values</v-btn>
      </template>
    </div>

    <component
      v-if="component && revealed"
      :is="component"
      :template="revealedTemplate"
    />
    <json-editor
      :key="editorKey"
      :json="newSecrets"
      @changeValue="newSecrets = $event"
    />

    <div class="d-flex jsutify-end align-center">
      <v-btn :loading="isSaveLoading" @click="updateSp">Update</v-btn>
    </div>
  </v-card>
</template>

<script setup>
import { computed, defineAsyncComponent, ref, toRefs } from "vue";
import JsonEditor from "@/components/JsonEditor.vue";
import api from "@/api";
import { useStore } from "@/store";
import {
  REDACTED_SECRET,
  fetchSpSecrets,
  secretsPatch,
} from "@/spSecrets.js";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

// What the editor started from: the redacted secrets the provider came with,
// or their real values once shown.
const originalSecrets = ref({ ...(template.value.secrets ?? {}) });
const newSecrets = ref({ ...originalSecrets.value });
const revealed = ref(false);
const editorKey = ref(0);
const isRevealLoading = ref(false);
const isSaveLoading = ref(false);

const revealedTemplate = computed(() => ({
  ...template.value,
  secrets: originalSecrets.value,
}));

const component = defineAsyncComponent(() =>
  import(
    `@/components/modules/${template.value?.type}/serviceProviderSecrets.vue`
  )
);

const setSecrets = (secrets) => {
  originalSecrets.value = { ...secrets };
  newSecrets.value = { ...secrets };
  editorKey.value++;
};

const reveal = async () => {
  isRevealLoading.value = true;
  try {
    setSecrets(await fetchSpSecrets(template.value.uuid));
    revealed.value = true;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message:
        e.response?.status === 403
          ? "Only root can read provider secrets"
          : e.response?.data?.message || "Error reading secrets",
    });
  } finally {
    isRevealLoading.value = false;
  }
};

const hide = () => {
  revealed.value = false;
  setSecrets(template.value.secrets ?? {});
};

const updateSp = async () => {
  isSaveLoading.value = true;
  try {
    await api.servicesProviders.update(template.value.uuid, {
      ...template.value,
      secrets: secretsPatch(originalSecrets.value, newSecrets.value),
    });
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Secrets saved",
    });
    // reloads the provider (redacted again); the page remounts this tab
    await store.dispatch("servicesProviders/fetchById", template.value.uuid);
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during secrets save",
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
