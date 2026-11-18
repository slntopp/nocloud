<template>
  <div v-if="editable || modelValue.length" class="note-attachments">
    <template v-for="uuid in modelValue" :key="uuid">
      <a
        v-if="files[uuid] && isImage(files[uuid].title)"
        class="note-attachments__image"
        :href="files[uuid].url"
        target="_blank"
        rel="noopener"
        :title="files[uuid].title"
      >
        <img :src="files[uuid].url" :alt="files[uuid].title" />
      </a>

      <v-chip
        v-else
        class="note-attachments__file"
        :href="files[uuid]?.url"
        target="_blank"
        rel="noopener"
        prepend-icon="mdi-paperclip"
        size="small"
      >
        {{ files[uuid]?.title || "file" }}
      </v-chip>

      <v-btn
        v-if="editable"
        icon
        variant="text"
        size="x-small"
        title="Remove from the note"
        @click="remove(uuid)"
      >
        <v-icon>mdi-close</v-icon>
      </v-btn>
    </template>

    <template v-if="editable">
      <input
        ref="input"
        type="file"
        multiple
        hidden
        @change="onFilesSelected"
      />
      <v-btn
        size="small"
        variant="tonal"
        prepend-icon="mdi-paperclip"
        :loading="isUploading"
        @click="input.click()"
      >
        Attach files
      </v-btn>
    </template>
  </div>
</template>

<script setup>
import { ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import {
  getAttachments,
  isImage,
  uploadAttachment,
} from "@/hooks/useAttachments";

const props = defineProps({
  // uuids of the attached files
  modelValue: { type: Array, default: () => [] },
  editable: { type: Boolean, default: false },
});
const emit = defineEmits(["update:modelValue", "uploading"]);
const { modelValue } = toRefs(props);

const store = useStore();

const input = ref(null);
const files = ref({});
const isUploading = ref(false);

const load = async (uuids) => {
  if (!uuids.length) return;
  try {
    Object.assign(files.value, await getAttachments(uuids));
  } catch (e) {
    console.error("Failed to load attachments", e);
  }
};

const onFilesSelected = async (event) => {
  const selected = [...(event.target.files ?? [])];
  event.target.value = "";
  if (!selected.length) return;

  isUploading.value = true;
  emit("uploading", true);
  const uploaded = [];
  try {
    for (const file of selected) {
      const attachment = await uploadAttachment(file);
      files.value[attachment.uuid] = attachment;
      uploaded.push(attachment.uuid);
    }
  } catch (e) {
    console.error(e);
    store.commit("snackbar/showSnackbarError", {
      message: "Error during file upload",
    });
  } finally {
    if (uploaded.length) {
      emit("update:modelValue", [...modelValue.value, ...uploaded]);
    }
    isUploading.value = false;
    emit("uploading", false);
  }
};

// only detaches the file from the note: it stays in the storage, so
// cancelling the edit doesn't lose it
const remove = (uuid) => {
  emit(
    "update:modelValue",
    modelValue.value.filter((u) => u !== uuid),
  );
};

watch(modelValue, (uuids) => load(uuids ?? []), { immediate: true });
</script>

<style scoped>
.note-attachments {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
}

.note-attachments__image img {
  display: block;
  height: 64px;
  max-width: 160px;
  object-fit: cover;
  border-radius: 6px;
}
</style>
