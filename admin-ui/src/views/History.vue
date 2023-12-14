<template>
  <div class="pa-4">
    <history-table
      :account-id="account"
      :uuid="uuid"
      :loading="isVariantsLoading"
      table-name="all-logs"
    />
  </div>
</template>

<script>
import HistoryTable from "@/components/historyTable.vue";
import { mapGetters } from "vuex";
import searchMixin from "@/mixins/search";

export default {
  name: "all-history",
  components: { HistoryTable },
  mixins: [searchMixin({name:"history"})],
  data: () => ({
    account: null,
    uuid: null,
    isVariantsLoading: false,
  }),
  async mounted() {
    this.$store.commit("appSearch/setSearchName", "app-logs");

    this.isVariantsLoading = true;
    await Promise.all([
      this.$store.dispatch("accounts/fetch"),
      this.$store.dispatch("services/fetch", { showDeleted: true }),
      this.$store.dispatch("servicesProviders/fetch",{anonymously:false}),
    ]);
    this.isVariantsLoading = false;

    this.$store.commit("appSearch/pushFields", [
      {
        items: this.services.concat(this.instances).concat(this.sps),
        type: "select",
        title: "Entity",
        key: "entity",
        item: { value: "uuid", title: "title" },
        single: true,
      },
      {
        items: this.accounts,
        title: "Accounts",
        type: "select",
        single: true,
        item: { value: "uuid", title: "title" },
        key: "account",
      },
    ]);
  },
  computed: {
    ...mapGetters("accounts", { accounts: "all" }),
    ...mapGetters("appSearch", { filter: "filter" }),
    instances() {
      return this.$store.getters["services/getInstances"];
    },
    sps() {
      return this.$store.getters["servicesProviders/all"];
    },
    services() {
      return this.$store.getters["services/all"];
    },
    accounts() {
      return this.$store.getters["accounts/all"];
    },
  },
  watch: {
    filter: {
      handler(newValue) {
        this.uuid = newValue.entity;
        this.account = newValue.account;
      },
      deep: true,
    },
  },
};
</script>

<style scoped></style>
