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

    <div :class="{ 'editor-readonly': !isEditing }" class="editor-container">
      <json-editor-new 
        ref="editorRef"
        :value="currentEditorValue"
        :read-only="!isEditing"
        @input="handleEditorInput"
        @change="handleEditorChange"
        @valid="handleEditorValid"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";
import JsonEditorNew from "./JsonEditor-New.vue";

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

const editorRef = ref(null);
const isEditing = ref(false);
const isSaving = ref(false);
const isValidJson = ref(false);
const currentValue = ref("");
const originalValue = ref("");
const currentEditorValue = ref(props.value);

function getFormattedText(val) {
  try {
    return JSON.stringify(val, null, 2);
  } catch (e) {
    return "";
  }
}

function handleEditorInput(value) {
  try {
    currentValue.value = getFormattedText(value);
    emit("input", value);
  } catch (e) {
    console.error("Input error:", e);
  }
}

function handleEditorChange(value) {
  try {
    currentValue.value = getFormattedText(value);
  } catch (e) {
    console.error("Change error:", e);
  }
}

function handleEditorValid(isValid) {
  isValidJson.value = isValid;
  emit("valid", isValid);
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
  currentEditorValue.value = JSON.parse(originalValue.value);
}

watch(isEditing, () => {
  if (isEditing.value) {
    currentEditorValue.value = props.value;
  }
});

watch(
  () => props.value,
  (newVal) => {
    if (!isEditing.value) {
      currentEditorValue.value = newVal;
      originalValue.value = getFormattedText(newVal);
    }
  },
  { deep: true }
);

originalValue.value = getFormattedText(props.value);
currentEditorValue.value = props.value;
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

.editor-container {
  border-radius: 12px;
  overflow: hidden;
  flex: 1;
}

.editor-readonly {
  position: relative;
  cursor: not-allowed;
  
  &::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.2);
    cursor: not-allowed;
    z-index: 10;
    border-radius: 12px;
    pointer-events: none;
  }
}
</style>