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
              outlined
              dense
              :disabled="isEdit"
              :rules="[requiredRule]"
            />
          </v-col>

          <v-col cols="12" md="8">
            <v-text-field
              v-model="setting.description"
              label="Description"
              outlined
              dense
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

            <div class="editor-shell">
              <div ref="editorContainer" class="json-editor"></div>
            </div>
          </v-col>
        </v-row>


        <v-row justify="end" class="mt-6">
          <confirm-dialog v-if="isEdit" @confirm="deleteSetting">
            <v-btn text class="mr-3">Delete</v-btn>
          </confirm-dialog>

          <confirm-dialog :disabled="!isValid" @confirm="saveSetting">
            <v-btn
              color="primary"
              :loading="isSaveLoading"
              :disabled="!isValid"
            >
              {{ isEdit ? "Save" : "Create" }}
            </v-btn>
          </confirm-dialog>
        </v-row>
      </v-form>
    </v-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, shallowRef } from "vue";
import { useStore } from "@/store";
import { useRouter, useRoute } from "vue-router/composables";
import { JSONEditor } from "vanilla-jsoneditor";
import api from "@/api";
import ConfirmDialog from "@/components/confirmDialog.vue";

const store = useStore();
const router = useRouter();
const route = useRoute();

const isValid = ref(false);
const isSaveLoading = ref(false);
const requiredRule = v => !!v || "Required";

const isEdit = computed(() => route.params.key !== "create");
const allSettings = computed(() => store.getters["settings/all"]);

const setting = ref({
  key: "",
  description: "",
  value: {}
});

const editorContainer = ref(null);
const editor = shallowRef(null);

function safeJson(val) {
  if (typeof val === "object" && val !== null) return val;
  try {
    return JSON.parse(val);
  } catch {
    return {};
  }
}

function initEditor(json) {
  editor.value?.destroy();

  editor.value = new JSONEditor({
    target: editorContainer.value,
    props: {
      mode: "tree",
      mainMenuBar: true,
      navigationBar: false,
      statusBar: false,
      content: { json },
      onChange: ({ json }) => {
        if (json !== undefined) {
          setting.value.value = json;
        }
      }
    }
  });
}

async function init() {
  let json = {};

  if (isEdit.value) {
    if (!allSettings.value.length) {
      await store.dispatch("settings/fetch");
    }

    const item = allSettings.value.find(
      s => s.key === route.params.key
    );

    if (item) {
      setting.value.key = item.key;
      setting.value.description = item.description;
      json = safeJson(item.value);
      setting.value.value = json;
    }
  }

  initEditor(json);
}

async function saveSetting() {
  isSaveLoading.value = true;
  try {
    await api.settings.addKey(setting.value.key, {
      description: setting.value.description,
      value: JSON.stringify(setting.value.value)
    });

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Saved"
    });

    await store.dispatch("settings/fetch");

    if (!isEdit.value) {
      router.push({
        name: "SettingsItem",
        params: { key: setting.value.key }
      });
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

onUnmounted(() => {
  editor.value?.destroy();
});
</script>

<style scoped lang="scss">
.settings-card {
  padding: 24px;
  border-radius: 14px;
  background: var(--v-background-light-base);
  border: 1px solid rgba(0, 0, 0, 0.06);
}

.page__title {
  font-size: 32px;
  font-weight: 400;
  color: var(--v-primary-base);

  a {
    color: inherit;
    text-decoration: none;
    border-bottom: 2px solid currentColor;
  }

  &-separator {
    margin: 0 10px;
  }

  &-current {
    text-transform: uppercase;
  }
}


.editor-header {
  display: flex;
  align-items: center;
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #adb0b7;
  margin-bottom: 8px;
}


.editor-shell {
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
  border: 1px solid rgba(0, 0, 0, 0.08);
}


.json-editor {
  height: 460px;
  font-family: "JetBrains Mono", monospace;
  font-size: 13px;


  --jse-theme-color: #3b5bfd;

  --jse-menu-background: #f4f6fb;
  --jse-menu-color: #4b5563;
  --jse-menu-border: rgba(0, 0, 0, 0.06);


  --jse-background-color: #ffffff;
  --jse-panel-background-color: #ffffff;
  --jse-text-color: #1f2937;

  --jse-key-color: #2563eb;
  --jse-string-color: #047857;
  --jse-number-color: #7c3aed;

  --jse-theme-color:  #dadbdf; 

}

/* Dark mode */
.v-application.theme--dark .json-editor {
  --jse-menu-background: #020617;
  --jse-menu-color: #cbd5f5;

  --jse-background-color: #020617;
  --jse-panel-background-color: #020617;
  --jse-text-color: #e5e7eb;

  --jse-key-color: #60a5fa;
  --jse-string-color: #34d399;
  --jse-number-color: #a78bfa;
  --jse-theme-color:  #850a8c; 
    border-radius: 12px;
    overflow: hidden;
    border: 0px solid rgb(0, 0, 0);
}

:deep(.jse-main) {
  border: none !important;
}
</style>
