<template>
  <widget title="Chats" :loading="isLoading" class="pa-0 ma-0">
    <v-card color="background-light" flat>
      <div class="d-flex justify-end">
        <v-btn-toggle
          class="mt-2"
          dense
          :value="data.period"
          @change="
            emit('update:key', { value: $event || data.period, key: 'period' })
          "
          borderless
        >
          <v-btn x-small :value="item" :key="item" v-for="item in periods">
            {{ item }}
          </v-btn>
        </v-btn-toggle>
      </div>

      <div class="d-flex justify-space-between align-center">
        <v-card-subtitle class="ma-0 my-2 pa-0"
          >Created in last {{ data.period }}</v-card-subtitle
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
      <v-card
        v-for="chat in lastActivityChats"
        :key="chat.uuid"
        dense
        color="transparent"
        class="d-flex justify-space-between pa-3"
        style="margin: 3px"
      >
        <div class="d-flex flex-column">
          <div>
            <span>Topic: </span>
            <router-link
              target="_blank"
              :to="{
                name: 'Chat',
                params: { uuid: chat.uuid },
              }"
            >
              {{
                chat.topic.length > 15
                  ? chat.topic.slice(0, 30) + "..."
                  : chat.topic
              }}
            </router-link>
          </div>

          <div class="d-flex">
            <span>Account: </span>
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
          </div>
        </div>
        <div class="d-flex flex-column">
          <span class="d-flex justify-end">
            {{
              formatSecondsToDate(
                Number(
                  chat.meta.lastMessage?.edited ||
                    chat.meta.lastMessage?.created ||
                    chat.created
                ) / 1000,
                true
              )
            }}
          </span>
          <span class="d-flex justify-end">
            {{ chat.department || "none" }}
          </span>
        </div>
      </v-card>
    </v-card>
  </widget>
</template>

<script setup>
import widget from "@/components/widgets/widget.vue";
import { computed, onMounted, ref, toRefs, watch } from "vue";
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
import { formatSecondsToDate } from "@/functions";
import { Status as ChatStatus } from "core-chatting/plugin/src/connect/cc/cc_pb";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const periods = ref(["day", "week", "month"]);
const accounts = ref({});
const isAccountsLoading = ref(false);

onMounted(() => fetchAccounts());

const chats = computed(() => store.getters["chats/all"]);
const isLoading = computed(() => store.getters["chats/loading"]);

const lastActivityChats = computed(() => {
  const sorted = [...chats.value]
    .filter((chat) => [ChatStatus.NEW, ChatStatus.OPEN].includes(chat.status))
    .sort(
      (a, b) =>
        (Number(b.meta?.lastMessage?.sent || b.created) || 0) -
        (Number(a.meta?.lastMessage?.sent || a.created) || 0)
    );

  return sorted.slice(0, 5);
});

const countForPeriod = computed(() => {
  if (!data.value.period) {
    return 0;
  }

  const dates = { from: null, to: null };

  switch (data.value.period) {
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

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week" });
  }
};

const fetchAccounts = () => {
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
};

watch(lastActivityChats, fetchAccounts);

setDefaultData();
</script>

<script>
export default {
  name: "chats-widget",
};
</script>
