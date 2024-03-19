<template>
  <widget title="Chats" :loading="isLoading" class="pa-0 ma-0">
    <v-card color="background-light" flat>
      <div class="d-flex justify-end">
        <v-btn-toggle
          class="mt-2"
          dense
          :value="period"
          @change="period = $event || period"
          borderless
        >
          <v-btn x-small :value="item" :key="item" v-for="item in periods">
            {{ item }}
          </v-btn>
        </v-btn-toggle>
      </div>

      <div class="d-flex justify-space-between align-center">
        <v-card-subtitle class="ma-0 my-2 pa-0"
          >Created in last {{ period }}</v-card-subtitle
        >
        <v-card-subtitle class="ma-0 pa-0">
          {{ countForPeriod }}
        </v-card-subtitle>
      </div>

      <div class="d-flex justify-space-between align-center mb-2">
        <v-card-subtitle class="ma-0 my-2 pa-0">Total created</v-card-subtitle>
        <v-card-subtitle class="ma-0 pa-0">
          {{ chats.length }}
        </v-card-subtitle>
      </div>

      <v-divider></v-divider>
      <v-list dense color="transparent">
        <v-list-item
          v-for="chat in lastActivityChats"
          :key="chat.uuid"
          class="px-0"
        >
          <v-list-item-content class="ma-0 pa-0">
            <div class="chats_list-item">
              <router-link
                target="_blank"
                :to="{
                  name: 'Chat',
                  params: { uuid: chat.uuid },
                }"
              >
                {{ chat.topic }}
              </router-link>

              <router-link
                v-if="!isAccountsLoading"
                target="_blank"
                :to="{
                  name: 'Account',
                  params: { accountId: chat.owner },
                }"
              >
                {{ accounts[chat.owner]?.title }}
              </router-link>
              <v-skeleton-loader type="text" v-else />

              <router-link
                target="_blank"
                :to="{
                  name: 'Chat',
                  params: { uuid: chat.uuid },
                }"
              >
                Department:{{ chat.department || "none" }}
              </router-link>
            </div>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-card>
  </widget>
</template>

<script setup>
import widget from "@/components/widgets/widget.vue";
import { computed, ref, watch } from "vue";
import { useStore } from "@/store";
import {
  endOfDay,
  endOfMonth,
  endOfWeek,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";
import api from "@/api";

const store = useStore();

const period = ref("day");
const periods = ref(["day", "week", "month"]);
const accounts = ref({});
const isAccountsLoading = ref(false);

const chats = computed(() => store.getters["chats/all"]);
const isLoading = computed(() => store.getters["chats/loading"]);

const lastActivityChats = computed(() => {
  console.log([...chats.value]);
  const sorted = [...chats.value].sort(
    (a, b) =>
      (Number(b.meta?.lastMessage?.sent || b.created) || 0) -
      (Number(a.meta?.lastMessage?.sent || a.created) || 0)
  );

  return sorted.slice(0, 5);
});

const countForPeriod = computed(() => {
  const dates = { from: null, to: null };

  switch (period.value) {
    case "day": {
      dates.from = startOfDay(new Date());
      dates.to = endOfDay(new Date());
      break;
    }
    case "month": {
      dates.from = startOfMonth(new Date());
      dates.to = endOfMonth(new Date());
      break;
    }
    case "week": {
      dates.from = startOfWeek(new Date());
      dates.to = endOfWeek(new Date());
      break;
    }
  }

  dates.from = dates.from.getTime() / 1000;
  dates.to = dates.to.getTime() / 1000;

  return chats.value.filter((chat) => {
    const createDate = (Number(chat.created) || 0) / 1000;
    return dates.from <= createDate && dates.to >= createDate;
  }).length;
});

watch(lastActivityChats, () => {
  lastActivityChats.value.forEach(async ({ owner: uuid }) => {
    isAccountsLoading.value = true;
    try {
      if (!accounts.value[uuid]) {
        accounts.value[uuid] = api.accounts.get(uuid);
        accounts.value[uuid] = await accounts.value[uuid];
      }
    } catch {
      accounts.value[uuid] = undefined;
    } finally {
      isAccountsLoading.value = Object.values(accounts.value).some(
        (acc) => acc instanceof Promise
      );
    }
  });
});
</script>

<script>
export default {
  name: "chats-widget",
};
</script>

<style scoped>
.chats_list-item {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
}
</style>
