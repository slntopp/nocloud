<template>
  <div class="statistics pa-5">
    <v-tabs v-model="tab" background-color="background-light">
      <v-tab v-for="{ title, key } of widgets" :key="key">
        {{ title }}
      </v-tab>
    </v-tabs>

    <v-tabs-items v-model="tab">
      <v-tab-item v-for="({ key, component }, index) of widgets" :key="key">
        <v-card color="background-light">
          <component
            v-if="tab === index"
            :key="key"
            :is="component"
            :period="period"
            @update:period="period = $event"
            :periods="periods"
            @update:periods="periods = $event"
            :period-type="periodType"
            @update:period-type="periodType = $event"
            :type="type"
            @update:type="type = $event"
            :period-offset="periodOffset"
            @update:period-offset="periodOffset = $event"
            :periods-first-offset="periodsFirstOffset"
            @update:periods-first-offset="periodsFirstOffset = $event"
            :periods-second-offset="periodsSecondOffset"
            @update:periods-second-offset="periodsSecondOffset = $event"
          />
        </v-card>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import AccountsStatistic from "@/components/statistics/accountsStatistic.vue";
import TrasactionsStatistic from "@/components/statistics/transactionsStatistic.vue";
import InstancesStatistic from "@/components/statistics/instancesStatistic.vue";
import ChatsResponsesStatistic from "@/components/statistics/chatsResponsesStatistic.vue";
import InstancesIncomeStatistic from "@/components/statistics/instancesIncomeStatistic.vue";
import ChatsStatistics from "@/components/statistics/chatsStatistics.vue";
import RevenueStatistics from "@/components/statistics/revenueStatistics.vue";
import aiStatistic from "../components/statistics/aiStatistic.vue";
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useStore } from "@/store/";
import { useRoute } from "vue-router/composables";
import router from "../router";

const store = useStore();
const route = useRoute();

const widgets = [
  { component: InstancesStatistic, title: "Instances", key: "instances" },
  {
    component: RevenueStatistics,
    title: "Revenue",
    key: "revenue",
  },
  {
    component: aiStatistic,
    title: "AI",
    key: "ai",
  },
  {
    component: ChatsStatistics,
    title: "Chats",
    key: "chats",
  },
  {
    component: ChatsResponsesStatistic,
    title: "Chats responses",
    key: "chats-responses",
  },
  { component: AccountsStatistic, title: "Accounts", key: "accounts" },
  {
    component: InstancesIncomeStatistic,
    title: "Instances income",
    key: "instances-income",
  },
  {
    component: TrasactionsStatistic,
    title: "Trasactions",
    key: "transactions",
  },
];

const period = ref([]);
const periods = ref({ first: [], second: [] });
const periodType = ref("month");
const type = ref("bar");
const periodOffset = ref(0);
const periodsFirstOffset = ref(0);
const periodsSecondOffset = ref(-1);

const tab = ref(null);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => this.$router.go(),
  });

  if (route.query.tab) {
    const index = widgets.findIndex((w) => w.key === route.query.tab);

    tab.value = index == -1 ? 0 : index;
  }

  loadWidgetsData();

  window.addEventListener("beforeunload", saveWidgetsData);
});

onBeforeUnmount(() => {
  store.commit("statistic/clearCache");
});

watch(tab, (newValue) => {
  router.replace({ query: { ...route.query, tab: widgets[newValue].key } });
  store.commit("statistic/clearCache");
});

const saveWidgetsData = () => {
  localStorage.setItem(
    "nocloud-statistic",
    JSON.stringify({
      type: type.value,
      periodType: periodType.value,
      periodsFirstOffset: periodsFirstOffset.value,
      periodsSecondOffset: periodsSecondOffset.value,
      periodOffset: periodOffset.value,
    }),
  );
};

const loadWidgetsData = () => {
  try {
    const data = JSON.parse(localStorage.getItem("nocloud-statistic"));

    type.value = data.type ?? type.value;
    periodType.value = data.periodType ?? periodType.value;
    periodsFirstOffset.value =
      data.periodsFirstOffset ?? periodsFirstOffset.value;
    periodsSecondOffset.value =
      data.periodsSecondOffset ?? periodsSecondOffset.value;
    periodOffset.value = data.periodOffset ?? periodOffset.value;
  } catch (e) {
    console.log(e);
  }
};
</script>

<script>
export default {
  name: "statistics-view",
};
</script>

<style lang="scss">
.apexcharts-legend-text {
  font-size: 1rem !important;
}
</style>
