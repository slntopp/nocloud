<template>
  <v-card min-height="50vh" color="background-light">
    <v-card
      class="px-5 pt-3"
      outlined
      color="background-light"
      style="position: relative"
    >
      <rich-editor v-model="newNote.msg" />
      <div style="position: absolute; bottom: 30px; right: 25px">
        <v-btn class="mx-1" @click="saveNewNote" :loading="isAddLoading"
          >Add</v-btn
        >
      </div>
    </v-card>
    <div class="mt-3 mx-4">
      <v-card
        color="background-light"
        class="my-3 pa-3"
        style="
          border: 2px solid;
          border-color: var(--v-primary-base) !important;
          border-radius: 10px;
        "
        outlined
        v-for="(note, index) in filteredNotes"
        :key="index"
      >
        <div class="d-flex justify-space-between mb-3">
          <div class="d-flex">
            <v-chip class="mx-1" color="primary" v-if="!isAccountsLoading">{{
              getAccount(note.admin)?.title || note.admin
            }}</v-chip>
            <v-skeleton-loader type="chip" v-else />

            <v-chip class="mx-1">Created at: {{ note.created }}</v-chip>
            <v-chip class="mx-1">Updated at: {{ note.updated }}</v-chip>
          </div>
          <div
            v-if="currentUserUuid === note.admin && !isEditMode"
            class="d-flex"
          >
            <v-btn
              :disabled="isRemoveLoading"
              class="mx-1"
              icon
              color="primary"
              @click="startEdit(index)"
            >
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
            <v-btn
              :loading="isRemoveLoading && removedNoteIndex === index"
              :disabled="isRemoveLoading && removedNoteIndex !== index"
              @click="removeNote(index)"
              class="mx-1"
              icon
              color="primary"
            >
              <v-icon>mdi-delete</v-icon>
            </v-btn>
          </div>
        </div>
        <EditorContainer
          v-if="!isEditMode || editedNoteIndex !== index"
          :value="note.msg"
        ></EditorContainer>
        <div v-else style="position: relative">
          <rich-editor v-model="note.msg"> </rich-editor>
          <div style="position: absolute; bottom: 30px; right: 25px">
            <v-btn class="mx-1" @click="cancelEdit" :disabled="isEditLoading"
              >Close</v-btn
            >
            <v-btn class="mx-1" @click="saveEditedNote" :loading="isEditLoading"
              >Save</v-btn
            >
          </div>
        </div>
      </v-card>
    </div>
  </v-card>
</template>

<script setup>
import RichEditor from "@/components/ui/richEditor.vue";
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { EditorContainer } from "nocloud-ui";
import { formatSecondsToDate } from "@/functions";
import api from "@/api";

const props = defineProps(["template", "onUpdate", "onAdd", "onDelete"]);
const { template, onAdd, onUpdate, onDelete } = toRefs(props);

const store = useStore();

const notes = ref([]);
const newNote = ref({ msg: "" });

const isAddLoading = ref(false);
const isRemoveLoading = ref(false);
const removedNoteIndex = ref("");
const isEditMode = ref(false);
const isEditLoading = ref(false);
const accounts = ref([]);
const editedNoteIndex = ref("");
const isAccountsLoading = ref(false);

onMounted(() => {
  setNotes();
});

const currentUserUuid = computed(() => store.getters["auth/userdata"]?.uuid);
const filteredNotes = computed(() =>
  notes.value.map((n) => ({
    ...n,
    created: formatSecondsToDate(n.created, true),
    updated: formatSecondsToDate(+n.updated || n.created, true),
  }))
);

const getAccount = (uuid) => {
  return accounts.value[uuid];
};

const setNotes = (data) => {
  if (!data) {
    data = template.value.adminNotes;
  }

  notes.value = data;
};

const saveNewNote = async () => {
  try {
    if (!newNote.value.msg) {
      return store.commit("snackbar/showSnackbarError", {
        message: "Not valid message",
      });
    }

    isAddLoading.value = true;

    const { adminNotes } = await onAdd.value(
      template.value.uuid,
      newNote.value
    );
    setNotes(adminNotes);
    newNote.value = { msg: "" };
  } catch (err) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during create note",
    });
  } finally {
    isAddLoading.value = false;
  }
};

const removeNote = async (index) => {
  try {
    isRemoveLoading.value = true;
    removedNoteIndex.value = index;

    const { adminNotes } = await onDelete.value(template.value.uuid, {
      index,
    });
    setNotes(adminNotes);
  } catch (err) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during remove note",
    });
  } finally {
    removedNoteIndex.value = "";
    isRemoveLoading.value = false;
  }
};

const startEdit = (index) => {
  isEditMode.value = true;
  editedNoteIndex.value = index;
};

const cancelEdit = () => {
  isEditMode.value = false;
  editedNoteIndex.value = "";
};

const saveEditedNote = async () => {
  try {
    const note = filteredNotes.value[editedNoteIndex.value];
    if (!note.msg) {
      return store.commit("snackbar/showSnackbarError", {
        message: "Not valid message",
      });
    }

    isEditLoading.value = true;

    const { adminNotes } = await onUpdate.value(template.value.uuid, {
      msg: note.msg,
      index: editedNoteIndex.value,
    });
    setNotes(adminNotes);
  } catch (err) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during update note",
    });
  } finally {
    cancelEdit();
    isEditLoading.value = false;
  }
};

watch(
  template,
  () => {
    setNotes();
  },
  { deep: true }
);

watch(
  notes,
  () => {
    notes.value.forEach(async ({ admin: uuid }) => {
      isAccountsLoading.value = true;
      try {
        if (!accounts.value[uuid]) {
          accounts.value[uuid] = api.accounts.get(uuid);
          accounts.value[uuid] = await accounts.value[uuid];
        }
      } finally {
        isAccountsLoading.value = Object.values(accounts.value).some(
          (acc) => acc instanceof Promise
        );
      }
    });
  },
  { deep: true }
);
</script>

<style scoped></style>
