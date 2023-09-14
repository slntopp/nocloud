<template>
  <div class="pa-4">
    <history-table
      :path="path"
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
  mixins: [searchMixin],
  data: () => ({
    account: null,
    uuid: null,
    path: null,
    isVariantsLoading: false,
  }),
  async mounted() {
    this.$store.commit("appSearch/setSearchName", "app-logs");

    this.isVariantsLoading = true;
    await Promise.all([
      this.$store.dispatch("accounts/fetch"),
      this.$store.dispatch("services/fetch"),
      this.$store.dispatch("servicesProviders/fetch"),
    ]);
    this.isVariantsLoading = false;

    this.$store.commit("appSearch/setVariants", {
      service: { items: this.services, title: "Service", key: "entity" },
      sp: { items: this.sps, title: "Service providers", key: "entity" },
      instance: { items: this.instances, title: "Instances", key: "entity" },
      account: { items: this.accounts, title: "Accounts" },
      path: { title: "Path" },
    });
  },
  computed: {
    ...mapGetters("accounts", { accounts: "all" }),
    ...mapGetters("appSearch", { customParams: "customParams" }),
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
    customParams: {
      handler(newValue) {
        this.uuid = newValue.entity?.value;
        this.account = newValue.account?.value;
        this.path = newValue.path?.value;
      },
      deep: true,
    },
  },
};
</script>

<style scoped></style>
