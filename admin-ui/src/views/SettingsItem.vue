<template>
  <div class="pa-4 h-100">

    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Settings' }">Settings</router-link>
      <span class="page__title-separator"> / </span>
      <span class="page__title-current">
        {{ isEdit ? setting.key : 'Create' }}
      </span>
    </h1>

    <v-card class="settings-card" elevation="0">
      <v-form v-model="isValid">
        <v-row>
          <v-col cols="12" md="4">
            <v-text-field
              v-model="setting.key"
              label="Key"
              outlined dense
              :disabled="isEdit"
              :rules="[requiredRule]"
            />
          </v-col>

          <v-col cols="12" md="8">
            <v-text-field
              v-model="setting.description"
              label="Description"
              outlined dense
              :rules="[requiredRule]"
            />
          </v-col>
        </v-row>


        <v-row>
          <v-col cols="12">
            <div class="editor-header">
              <v-icon small class="mr-2">mdi-code-json</v-icon>
              <span>Value configuration</span>
            </div>

            <json-editor-new 
              v-if="!isDataLoading"
              v-model="setting.value" 
            />
            
            <div v-else class="d-flex justify-center pa-10">
              <v-progress-circular indeterminate color="primary" />
            </div>
          </v-col>
        </v-row>

        <v-row justify="end" class="mt-6 mx-0">
          <confirm-dialog v-if="isEdit" @confirm="deleteSetting">
            <v-btn text color="error" class="mr-3 action-btn-text">Delete</v-btn>
          </confirm-dialog>

          <confirm-dialog :disabled="!isValid" @confirm="saveSetting">
            <v-btn class="px-8 action-btn-main" :loading="isSaveLoading" :disabled="!isValid" elevation="0">
              {{ isEdit ? "Save" : "Create" }}
            </v-btn>
          </confirm-dialog>
        </v-row>
      </v-form>
    </v-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useStore } from "@/store";
import { useRouter, useRoute } from "vue-router/composables";
import api from "@/api";
import ConfirmDialog from "@/components/confirmDialog.vue";
import JsonEditorNew from "@/components/JsonEditor-New.vue";

const store = useStore();
const router = useRouter();
const route = useRoute();

const isValid = ref(false);
const isSaveLoading = ref(false);
const isDataLoading = ref(true); 
const requiredRule = v => !!v || "Required";

const isEdit = computed(() => route.params.key !== "create");
const allSettings = computed(() => store.getters["settings/all"]);

const setting = ref({
  key: "",
  description: "",
  value: {}
});



async function init() {
  isDataLoading.value = true;
  try {
    if (isEdit.value) {
      if (!allSettings.value.length) {
        await store.dispatch("settings/fetch");
      }

      const item = allSettings.value.find(s => s.key === route.params.key);

      if (item) {
        setting.value.key = item.key;
        setting.value.description = item.description;
        setting.value.value = item.value;
      }
    }
  } catch (e) {
    console.error("Failed to load setting:", e);
  } finally {
    isDataLoading.value = false;
  }


}

async function saveSetting() {
  isSaveLoading.value = true;
  try {
    await api.settings.addKey(setting.value.key, {
      description: setting.value.description,
      value: JSON.stringify(setting.value.value)
    });
    store.commit("snackbar/showSnackbarSuccess", { message: "Saved" });
    await store.dispatch("settings/fetch");

    if (!isEdit.value) {
      router.push({ name: "SettingsItem", params: { key: setting.value.key } });
    }
  } finally {
    isSaveLoading.value = false;
  }
}

async function deleteSetting() {
  await api.settings.delete(setting.value.key);
  router.push({ name: "Settings" });
}

onMounted(init);


</script>

<style scoped lang="scss">
.settings-card {
  padding: 24px;
  border-radius: 14px;
  background: var(--v-background-light-base);
  border: 1px solid rgba(0, 0, 0, 0.06);
}

.page__title {
  font-family: "Quicksand", sans-serif;
  font-size: 32px;
  font-weight: 400;
  color: var(--v-primary-base);
  text-transform: none;
  a { color: inherit; text-decoration: none; border-bottom: 2px solid currentColor; padding-bottom: 2px; }
  &-separator { margin: 0 10px; }
  &-current { text-transform: uppercase; }
}


.editor-header {
  display: flex;
  align-items: center;
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: rgba(0, 0, 0, 0.6);
  margin-bottom: 12px;
  i { color: var(--v-primary-base) !important; }
}

.action-btn-main { text-transform: uppercase; font-weight: 600; letter-spacing: 1px; }
.action-btn-text { text-transform: uppercase; font-weight: 600; color: #adb0b7; }

.v-application.theme--dark .editor-header { color: #adb0b7; }
</style>
