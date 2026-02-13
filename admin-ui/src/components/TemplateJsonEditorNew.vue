<template>
  <div class="editor-wrapper">
    <div class="editor-top">
      <div class="editor-left">
        {{ title }}
        
        <template v-if="isEditing">
          <v-btn
            class="mr-2"
            :loading="isSaving"
            :disabled="!isValidJson"
            @click="handleSave"
          >
            Save
          </v-btn>
          <v-btn @click="handleCancel">Cancel</v-btn>
        </template>

        <template v-else>
          <v-btn @click="isEditing = true">Edit</v-btn>
        </template>
      </div>
    </div>

    <div class="editor-shell">
      <div ref="editorContainer" class="json-editor"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, shallowRef, watch } from "vue";
import { JSONEditor } from "vanilla-jsoneditor";

const props = defineProps({
  value: {
    type: Object,
    required: true,
  },
  title: {
    type: String,
    default: "Template JSON",
  },
  onSave: {
    type: Function,
    default: null,
  },
});

const emit = defineEmits(["input", "valid", "save"]);

const editorContainer = ref(null);
const editor = shallowRef(null);
const isEditing = ref(false);
const isSaving = ref(false);
const isValidJson = ref(false);
const currentValue = ref("");
const originalValue = ref("");

function getFormattedText(val) {
  try {
    return JSON.stringify(val, null, 2);
  } catch (e) {
    return "";
  }
}

function initEditor() {
  if (editor.value) editor.value.destroy();

  editor.value = new JSONEditor({
    target: editorContainer.value,
    props: {
      mode: "text",
      modes: ["text"], 
      mainMenuBar: true,
      navigationBar: false,
      statusBar: false,
      askToFormat: false,
      readOnly: !isEditing.value,
      content: {
        text: getFormattedText(props.value),
      },
      onChangeText: (text) => {
        try {
          JSON.parse(text);
          
          isValidJson.value = true;
          currentValue.value = text;
          emit("valid", true);
          emit("input", text);
        } catch (e) {
          isValidJson.value = false;
          emit("valid", false);
        }
      },
    },
  });
}

async function handleSave() {
  try {
    isSaving.value = true;
    const parsedValue = JSON.parse(currentValue.value);
    
    if (props.onSave) {
      await props.onSave(parsedValue);
    }
    
    emit("save", parsedValue);
    isEditing.value = false;
    originalValue.value = currentValue.value;
  } catch (err) {
    console.error("Save error:", err);
  } finally {
    isSaving.value = false;
  }
}

function handleCancel() {
  isEditing.value = false;
  isValidJson.value = false;
  currentValue.value = originalValue.value;
  if (editor.value) {
    editor.value.update({
      content: { text: getFormattedText(props.value) },
    });
  }
}

watch(isEditing, () => {
  initEditor();
});

watch(
  () => props.value,
  (newVal) => {
    if (editor.value && !isEditing.value) {
      editor.value.update({
        content: { text: getFormattedText(newVal) },
      });
    }
  },
  { deep: true }
);

onMounted(() => {
  originalValue.value = getFormattedText(props.value);
  initEditor();
});

onUnmounted(() => {
  if (editor.value) editor.value.destroy();
});
</script>

<style scoped lang="scss">
.editor-wrapper {
  display: flex;
  flex-direction: column;
  height: 75vh;
}

.editor-top {
  position: sticky;
  top: 0px; 
  
  z-index: 5;
  display: flex;
  justify-content: space-between;
  align-items: center;

  background: var(--v-background-light-base) !important;
  
  padding: 10px 0;
  margin-bottom: 10px;

  border-bottom: 1px solid rgba(0,0,0,0.05);
}

.editor-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-shell {
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(0, 0, 0, 0.12);
  background: var(--v-background-light-base);
  height: 70vh;
  flex: 1;
}

.json-editor {
  height: 100%;
  width: 100%;
  font-family: "JetBrains Mono", monospace;

  --jse-background-color: #ffffff;
  --jse-panel-background: #f5f5f5;
  --jse-text-color: #1a1a1a;
  --jse-menu-background: #f0f0f0;
  --jse-menu-color: #5c5c5c;
  --jse-menu-button-size: 32px;
  
  --jse-menu-button-color-selected: #ffffff;
  --jse-menu-button-background-highlight: #e0e0e0;
  --jse-menu-button-background-selected: #aeaeae;
  
  --jse-key-color: #1a01cc;
  --jse-string-color: #00802b;
  --jse-number-color: #d13600;
  --jse-delimiter-color: #333333;
  --jse-theme-color: #b4b4b4;
}
</style>

<style lang="scss">
.jse-main, .jse-contents, .jse-text-mode {
  border: none !important;
  box-shadow: none !important;
}

.jse-menu {
  border-bottom: 1px solid rgba(0, 0, 0, 0.05) !important;
  background-color: var(--jse-menu-background) !important;
}

.jse-menu button {
  transition: all 0.2s ease !important;
  
  &:hover:not([disabled]) {
    background-color: #a80bae !important;
    color: white !important;
  }
}

.jse-menu button {
  &[title="table"] {
    display: none !important;
  }
}

button[title*="table"] {
  display: none !important;
}

.v-application.theme--dark {
  .json-editor {
    --jse-background-color: #1a1c2a;
    --jse-panel-background: #1a1c2a;
    --jse-text-color: #e5e7eb;
    --jse-menu-background: #0e061d;
    --jse-menu-color: #ffffff;
    --jse-menu-button-color: #cccccc;
    --jse-menu-button-background-highlight: #251b3d;
    --jse-menu-button-color-selected: #ffffff;
    --jse-menu-button-background-selected: #a80bae;
    
    --jse-key-color: #74b1f6;
    --jse-string-color: #34d399;
    --jse-number-color: #c4b5fd;
    --jse-delimiter-color: #94a3b8;
  }

  .jse-message.jse-info {
    background-color: #2e1a47 !important;
    border: 1px solid #a80bae !important;
    .jse-message-text { color: white !important; }
  }
}

.jse-text-mode {
  border: none !important;
}
</style>