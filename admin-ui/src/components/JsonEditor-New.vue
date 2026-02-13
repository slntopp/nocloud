<template>
  <div class="editor-wrapper">
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
    type: [Object, Array, String],
    default: () => ({})
  },
  readOnly: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(["input", "change"]);

const editorContainer = ref(null);
const editor = shallowRef(null);

function safeJson(val) {
  if (!val) return {};
  if (typeof val === "object") return val;
  try {
    let parsed = JSON.parse(val);
    if (typeof parsed === "string") parsed = JSON.parse(parsed);
    return parsed;
  } catch (e) {
    return { value: val };
  }
}

function initEditor() {
  if (editor.value) editor.value.destroy();

  editor.value = new JSONEditor({
    target: editorContainer.value,
    props: {
      mode: 'text',
      modes: ['text', 'tree'],
      mainMenuBar: true,
      navigationBar: false,
      statusBar: false,
      readOnly: props.readOnly,
      content: {
        json: safeJson(props.value)
      },
      onChange: (content) => {
        let updated;
        if (content.json !== undefined) {
          updated = content.json;
        } else if (content.text !== undefined) {
          try {
            updated = JSON.parse(content.text);
          } catch (e) { return; }
        }
        emit("input", updated);
        emit("change", updated);
      }
    }
  });
}

watch(() => props.value, (newVal) => {
  if (editor.value) {
    const newData = safeJson(newVal);
    const currentContent = editor.value.get();
    const currentJson = currentContent.json || (currentContent.text ? JSON.parse(currentContent.text) : null);

    if (JSON.stringify(currentJson) !== JSON.stringify(newData)) {
      editor.value.update({ content: { json: newData } });
    }
  }
}, { deep: true });

watch(() => props.readOnly, () => {
  if (editor.value) {
    editor.value.update({ readOnly: props.readOnly });
  }
});

onMounted(initEditor);
onUnmounted(() => editor.value?.destroy());
</script>

<style scoped lang="scss">
.editor-wrapper {
  display: flex;
  flex-direction: column;
  height: 500px;
}

.editor-shell {
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(0, 0, 0, 0.12);
  background: var(--v-background-light-base);
  height: 100%;
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
    --jse-text-color: #ffffffff;
    --jse-menu-background: #0e061d;
    --jse-menu-color: #4d4e71ff;
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

  .jse-mode-button {
    background-color: #251b3d !important;
    color: #cccccc !important;
    
    &.jse-selected {
      background-color: #a80bae !important;
      color: #ffffffff !important;
    }
    
    &:hover:not(.jse-selected) {
      background-color: #3d2a5c !important;
    }
  }
}


  
.jse-text-mode {
  border: none !important;
}
</style>