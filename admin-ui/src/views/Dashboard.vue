<template>
  <div class="widgets align-start gg-15px pa-4 d-flex flex-wrap">
    <component
      v-for="{ key, component } of widgets"
      :key="key"
      :is="component"
      style="width: 33%"
      :data="widgetsData[key] || {}"
      @update="updateWidgetData(key, $event)"
      @update:key="updateWidgetDataKey(key, $event)"
    />
  </div>
</template>

<script setup>
import AccountsWidget from "@/components/widgets/accounts";
import TransactionsWidget from "@/components/widgets/transactions";
import InstancesWidget from "@/components/widgets/instances";
import InstancesTypesWidget from "@/components/widgets/instances-types.vue";
import InstancesPricesWidget from "@/components/widgets/instances-prices.vue";
import ChatsWidget from "@/components/widgets/chats.vue";
import ChatsResponsibles from "@/components/widgets/chats-responsibles.vue";
import { onBeforeUnmount, onMounted, ref } from "vue";
import { useStore } from "@/store/";

const store = useStore();

const widgets = [
  { component: ChatsWidget, key: "chats" },
  { component: AccountsWidget, key: "accounts" },
  { component: InstancesWidget, key: "instances" },
  { component: TransactionsWidget, key: "transactions" },
  { component: ChatsResponsibles, key: "chats-responsibles" },
  { component: InstancesTypesWidget, key: "instances-types" },
  { component: InstancesPricesWidget, key: "instances-prices" },
];

const widgetsData = ref({
  chats: {},
  accounts: {},
  instances: {},
  transactions: {},
  "instances-types": {},
  "instances-prices": {},
});

onMounted(() => {
  loadWidgetsData();

  store.commit("reloadBtn/setCallback", {
    event: () => this.$router.go(),
  });
  window.addEventListener("beforeunload", saveWidgetsData);
});

onBeforeUnmount(() => {
  saveWidgetsData();
  window.removeEventListener("beforeunload", saveWidgetsData);
});

const updateWidgetData = (key, data) => {
  widgetsData.value[key] = data;
};

const updateWidgetDataKey = (key, { key: subkey, value }) => {
  widgetsData.value[key][subkey] = value;
};

const saveWidgetsData = () => {
  localStorage.setItem("nocloud-widgets", JSON.stringify(widgetsData.value));
};

const loadWidgetsData = () => {
  try {
    const data = JSON.parse(localStorage.getItem("nocloud-widgets"));
    widgetsData.value = { ...widgetsData.value, ...data };
  } catch (e) {
    console.log(e);
  }
};
</script>

<script>
export default {
  name: "dashboard-view",
};
</script>

<style scoped lang="scss">
.d-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;

  @media (max-width: 1010px) {
    grid-template-columns: 1fr;
    justify-items: center;
  }
}
.grid-row {
  grid-row: 1 / 3;
  justify-self: end;
}
</style>
