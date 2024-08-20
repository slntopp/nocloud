<template>
  <v-menu offset-y transition="slide-y-transition">
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
import { computed, onMounted, ref } from "vue";
import api from "@/api";
import { formatSecondsToDate } from "@/functions";

const notifications = ref([]);
const isFetchLoading = ref(false);

onMounted(() => {
  fetchNotifications();
});

const isUnseenNotification = computed(() =>
  notifications.value.find(
    (notification) => notification.ts > Date.now() / 1000 - 86400
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
</script>

<script>
export default {
  name: "notifications-menu",
};
</script>
