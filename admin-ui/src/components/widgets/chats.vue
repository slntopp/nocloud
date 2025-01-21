<template>
  <widget title="Chats" :loading="isLoading" class="pa-0 ma-0">
    <v-card color="background-light" flat>
      <div class="d-flex justify-space-between">
        <v-btn-toggle
          class="mt-2"
          dense
          :value="data.status"
          @change="
            emit('update:key', { value: $event || data.status, key: 'status' })
          "
          borderless
        >
          <v-btn x-small :value="item" :key="item" v-for="item in statuses">
            {{ item }}
          </v-btn>
        </v-btn-toggle>

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
              {{ getShortName(chat.topic) }}
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
import api from "@/api";
import { formatSecondsToDate, getShortName } from "@/functions";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const periods = ref(["day", "week", "month"]);
const statuses = ref(["all", "closed"]);
const accounts = ref({});
const isAccountsLoading = ref(false);

onMounted(() => fetchAccounts());

const dayChats = computed(() => store.getters["chats/dayChats"]);
const weekChats = computed(() => store.getters["chats/weekChats"]);
const monthChats = computed(() => store.getters["chats/monthChats"]);
const isLoading = computed(() => store.getters["chats/loading"]);

const lastActivityChats = computed(() => {
  return dayChats.value.slice(0, 3);
});

const countForPeriod = computed(() => {
  if (!data.value.period) {
    return 0;
  }

  let chats;

  switch (data.value.period) {
    case "day": {
      chats = dayChats.value;
      break;
    }
    case "month": {
      chats = monthChats.value;
      break;
    }
    case "week": {
      chats = weekChats.value;
      break;
    }
  }

  return chats.filter((chat) => {
    return !(data.value.status === "closed" && chat.status !== 3);
  }).length;
});

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week", starus: "all" });
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
