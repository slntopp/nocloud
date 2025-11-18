<template>
  <div v-if="showPinned && pinnedNotes.length" class="pinned-notes-container">
    <v-card
      color="background-light"
      elevation="8"
      class="pinned-notes-card"
    >
      <v-card-title class="d-flex justify-space-between align-center pa-3">
        <span class="text-subtitle-1">
          <v-icon small class="mr-1">mdi-pin</v-icon>
          Pinned Notes ({{ pinnedNotes.length }})
        </span>
        <v-btn icon small @click="closePinned">
          <v-icon small>mdi-close</v-icon>
        </v-btn>
      </v-card-title>

      <v-divider></v-divider>

      <div class="pinned-notes-content">
        <v-card
          v-for="(note, index) in pinnedNotes"
          :key="index"
          color="background"
          class="ma-2 pa-2"
          outlined
        >
          <div class="d-flex justify-space-between align-center mb-2">
            <v-chip x-small color="primary" v-if="!isAccountsLoading">
              {{ getAccount(note.admin)?.title || note.admin }}
            </v-chip>
            <v-skeleton-loader type="chip" v-else />

             <v-chip x-small class="mr-1 mb-1">
              Created: {{ formatDate(note.created) }}
            </v-chip>
          </div>
          <EditorContainer
            :value="note.msg"
            class="pinned-note-content"
          ></EditorContainer>
        </v-card>
      </div>
    </v-card>
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { EditorContainer } from "nocloud-ui";
import { formatSecondsToDate } from "@/functions";
import api from "@/api";

const props = defineProps({
  notes: {
    type: Array,
    default: () => [],
  },
});

const showPinned = ref(true);
const accounts = ref({});
const isAccountsLoading = ref(false);

const pinnedNotes = computed(() => {
  return props.notes.filter((note) => note.pinned);
});

const closePinned = () => {
  showPinned.value = false;
};

const getAccount = (uuid) => {
  return accounts.value[uuid];
};

const formatDate = (timestamp) => {
  return formatSecondsToDate(timestamp, true);
};

watch(
  pinnedNotes,
  () => {
    pinnedNotes.value.forEach(async ({ admin: uuid }) => {
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
  { deep: true, immediate: true }
);
</script>

<style scoped>
.pinned-notes-container {
  position: fixed;
  bottom: 50px;
  right: 20px;
  z-index: 100;
  max-width: 400px;
  width: 100%;
}

.pinned-notes-card {
  max-height: 500px;
  display: flex;
  flex-direction: column;
  border-radius: 10px !important;
}

.pinned-notes-content {
  overflow-y: auto;
  max-height: calc(100vh - 180px);
  padding: 8px 0;
}

.pinned-notes-content::-webkit-scrollbar {
  width: 8px;
}

.pinned-notes-content::-webkit-scrollbar-track {
  background: var(--v-background-base);
  border-radius: 4px;
}

.pinned-notes-content::-webkit-scrollbar-thumb {
  background: var(--v-primary-base);
  border-radius: 4px;
}

.pinned-notes-content::-webkit-scrollbar-thumb:hover {
  background: var(--v-primary-darken1);
}

.pinned-note-content {
  font-size: 0.875rem;
  max-height: 200px;
  overflow-y: auto;
}

@media (max-width: 960px) {
  .pinned-notes-container {
    max-width: 320px;
    right: 10px;
    top: 70px;
  }
}

@media (max-width: 600px) {
  .pinned-notes-container {
    max-width: calc(100vw - 20px);
    right: 10px;
    left: 10px;
    top: 60px;
  }
}
</style>
