<template>
  <v-menu v-model="isOpen" offset-y transition="slide-y-transition">
    <template v-slot:activator="{ on, attrs }">
      <v-btn
        class="mx-2"
        :loading="isFetchLoading"
        fab
        color="background-light"
        v-bind="attrs"
        v-on="on"
      >
        <v-icon dark>
          {{ !isUnseenNotification ? "mdi-bell" : "mdi-bell-badge" }}
        </v-icon>
      </v-btn>
    </template>
    <v-card color="background-light" style="max-height: 30vh">
      <v-list color="background-light" dence min-width="250px">
        <v-list-item>
          <v-list-item-content>
            <v-list-item-title class="text-h6">
              Last notifications</v-list-item-title
            >
          </v-list-item-content>
        </v-list-item>

        <v-list-item
          v-for="notification in notifications"
          :key="notification.uuid + notification.ts"
        >
          <v-list-item-content>
            <v-list-item-title> {{ notification.action }}</v-list-item-title>
            <v-list-item-subtitle>
              {{ formatSecondsToDate(notification.ts, true) }}
            </v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import api from "@/api";
import { formatSecondsToDate } from "@/functions";

const isOpen = ref(false);
const notifications = ref([]);
const seen = ref([]);
const isFetchLoading = ref(false);

onMounted(() => {
  fetchNotifications();
});

const isUnseenNotification = computed(() =>
  notifications.value.some(
    (notification) => !seen.value.includes(getLogId(notification))
  )
);

const fetchNotifications = async () => {
  isFetchLoading.value = true;
  try {
    notifications.value = (
      await api.logging.list({
        page: 1,
        limit: 20,
        field: "ts",
        sort: "DESC",
        filters: {
          //   priority: 1,
        },
      })
    ).events;
  } finally {
    isFetchLoading.value = false;
  }
};

const getLogId = (log) => `${log.uuid}$${log.ts}`;

const getSeenNotifcations = () => {
  seen.value = JSON.parse(
    localStorage.getItem("nocloud-notifications") || `[]`
  );
};

const setSeenNotifcations = (seen) => {
  localStorage.setItem(
    "nocloud-notifications",
    JSON.stringify(
      [...new Set(seen)]
        .sort((a, b) => {
          const bTs = +b.split("$")[1] || 0;
          const aTs = +a.split("$")[1] || 0;

          return bTs - aTs;
        })
        .slice(0, 20)
    )
  );
  getSeenNotifcations();
};

watch(isOpen, () => {
  setSeenNotifcations([
    ...seen.value,
    ...notifications.value.map((n) => getLogId(n)),
  ]);
});

getSeenNotifcations();
</script>

<script>
export default {
  name: "notifications-menu",
};
</script>
