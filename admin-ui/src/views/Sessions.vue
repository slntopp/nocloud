<template>
  <v-card elevation="0" color="background" class="pa-4">
    <confirm-dialog :disabled="selected.length < 1" @confirm="deleteSelectedSession">
      <v-btn
        color="background-light"
        :disabled="selected.length < 1"
        :loading="isDeleteLoading"
      >
        Delete
      </v-btn>
    </confirm-dialog>

    <v-autocomplete
      :filter="defaultFilterObject"
      label="Account"
      item-text="title"
      item-value="uuid"
      class="d-inline-block ml-2"
      :items="accounts"
      @change="getActivity($event)"
    />

    <nocloud-table
      table-name="sessions"
      :headers="sessionsHeaders"
      :items="sessions"
      :loading="isLoading"
      :footer-error="fetchError"
      v-model="selected"
    >
      <template v-slot:[`item.client`]="{ value }">
        <router-link :to="{ name: 'Account', params: { accountId: value } }">
          {{ getAccount(value)?.title ?? value }}
        </router-link>
      </template>

      <template v-slot:[`item.created`]="{ value }">
        {{ getDate(value) }}
      </template>

      <template v-slot:[`item.expires`]="{ value }">
        {{ getDate(value) }}
      </template>

      <template v-slot:[`item.lastSeen`]="{ value }">
        {{ getDate(value) }}
      </template>
    </nocloud-table>
  </v-card>
</template>

<script setup>
import { ref, computed, watch, reactive } from "vue";
import { useStore } from "@/store";
import { formatSecondsToDate, defaultFilterObject } from "@/functions.js";
import api from "@/api";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";

const store = useStore();

const allSessions = reactive({});
const sessions = ref([]);
const isLoading = ref(false);
const fetchError = ref("");

const sessionsHeaders = [
  { text: "ID", value: "uuid" },
  { text: "Account", value: "client" },
  { text: "Creation date", value: "created" },
  { text: "Expiration date", value: "expires" },
  { text: "Current", value: "current" },
  { text: "Last seen", value: "lastSeen" }
];

const accounts = computed(() =>
  store.getters["accounts/all"]
);

function getAccount(uuid) {
  return accounts.value.find((account) => account.uuid === uuid);
}

watch(sessions, () => {
  fetchError.value = "";
});

isLoading.value = true;
Promise.all([
  store.dispatch("accounts/fetch"),
  api.get('/sessions')
])
  .then((response) => {
    sessions.value = response[1].sessions
      .map((session) => ({ ...session, uuid: session.id }));

    allSessions.all = JSON.parse(JSON.stringify(sessions.value));
  })
  .catch((error) => {
    console.error(error);
    fetchError.value = "Can't reach the server";
    if (error.response?.data.message) {
      fetchError.value += `: [ERROR]: ${error.response.data.message}`;
    } else {
      fetchError.value += `: [ERROR]: ${error.toJSON().message}`;
    }
  })
  .finally(() => {
    isLoading.value = false;
  });

const selected = ref([]);
const isDeleteLoading = ref(false);

async function deleteSelectedSession() {
  try {
    isDeleteLoading.value = true;
    const promises = selected.value.map(({ id }) =>
      api.delete(`/sessions/${id}`)
    );

    await Promise.all(promises);
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Sessions removed successfully."
    });
  } catch (error) {
    if (error.response.status >= 500 || error.response.status < 600) {
      store.commit("snackbar/showSnackbarError", {
        message: `Sessions Unavailable: ${
          error.response?.data.message ?? "Unknown"
        }.`,
        timeout: 0
      });
    } else {
      store.commit("snackbar/showSnackbarError", {
        message: `Error: ${error.response?.data.message ?? "Unknown"}.`,
        timeout: 0
      });
    }
  } finally {
    isDeleteLoading.value = false;
  }
}

async function getActivity(uuid) {
  if (allSessions[uuid]) {
    sessions.value = JSON.parse(JSON.stringify(allSessions[uuid]));
    return;
  }

  try {
    isLoading.value = true;
    const { lastSeen } = await api.get(`/sessions/activity/${uuid}`);

    sessions.value = [];

    allSessions.all.forEach((session) => {
      if (!lastSeen[session.uuid]) return;
      sessions.value.push({ ...session, lastSeen: lastSeen[session.uuid] });
    });
    allSessions[uuid] = JSON.parse(JSON.stringify(sessions.value));
  } catch (error) {
    console.error(error);
    store.commit("snackbar/showSnackbarError", {
      message: `Error: ${error.response?.data.message ?? "Unknown"}.`,
      timeout: 0
    });
  } finally {
    isLoading.value = false;
  }
}

function getDate(date) {
  return (typeof date === 'string')
    ? formatSecondsToDate(new Date(date).getTime() / 1000, true, '.')
    : formatSecondsToDate(date, true, '.')
}
</script>

<script>
export default { name: "sessions-page" };
</script>
