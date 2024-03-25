<template>
  <div class="widgets align-start gg-15px pa-4 d-flex flex-wrap">
    <component
      v-for="widget of widgets"
      :key="widget"
      :is="widget"
      style="width: 33%"
    />
  </div>
</template>

<script>
import AccountsWidget from "@/components/widgets/accounts";
import TransactionsWidget from "@/components/widgets/transactions";
import InstancesWidget from "@/components/widgets/instances";
import InstancesTypesWidget from "@/components/widgets/instances-types.vue";
import InstancesPricesWidget from "@/components/widgets/instances-prices.vue";
import ChatsWidget from "@/components/widgets/chats.vue";

export default {
  name: "dashboard-view",
  components: {
    AccountsWidget,
    TransactionsWidget,
    InstancesWidget,
    InstancesTypesWidget,
    InstancesPricesWidget,
    ChatsWidget,
  },
  data: () => ({
    widgets: [
      "ChatsWidget",
      "AccountsWidget",
      "InstancesWidget",
      "TransactionsWidget",
      "InstancesTypesWidget",
      "InstancesPricesWidget",
    ],
  }),
  computed: {
    width() {
      return document.documentElement.clientWidth;
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      event: () => this.$router.go(),
    });
  },
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
