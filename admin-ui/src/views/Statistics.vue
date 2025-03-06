<template>
  <div class="statistics pa-5">
    <v-tabs v-model="tab" background-color="background-light">
      <v-tab v-for="{ title, key } of widgets" :key="key">
        {{ title }}
      </v-tab>
    </v-tabs>

    <v-tabs-items v-model="tab">
      <v-tab-item v-for="{ key, component } of widgets" :key="key">
        <v-card color="background-light">
          <component
            :key="key"
            :is="component"
            style="width: 33%"
            :data="widgetsData[key] || {}"
            @update="updateWidgetData(key, $event)"
            @update:key="updateWidgetDataKey(key, $event)"
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
import { onMounted, ref } from "vue";
import { useStore } from "@/store/";
import { useRoute } from "vue-router/composables";

const store = useStore();
const route = useRoute();

const widgets = [
  { component: InstancesStatistic, title: "Instances", key: "instances" },
  {
    component: InstancesIncomeStatistic,
    title: "Instances income",
    key: "instances-income",
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
    component: TrasactionsStatistic,
    title: "Trasactions",
    key: "transactions",
  },
];

const widgetsData = ref({});

const tab = ref(null);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => this.$router.go(),
  });

  if (route.query.tab) {
    const index = widgets.findIndex((w) => w.key === route.query.tab);

    tab.value = index == -1 ? 0 : index;
  }
});
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
