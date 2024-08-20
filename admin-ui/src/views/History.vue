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
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";

export default {
  name: "all-history",
  components: { HistoryTable },
  mixins: [searchMixin({ name: "history" })],
  data: () => ({
    account: null,
    uuid: null,
    isVariantsLoading: false,
  }),
  async mounted() {
    this.$store.commit("appSearch/setSearchName", "app-logs");

    this.isVariantsLoading = true;
    await Promise.all([
      this.$store.dispatch("services/fetch", { showDeleted: true }),
      this.$store.dispatch("servicesProviders/fetch", { anonymously: false }),
    ]);
    this.isVariantsLoading = false;

    this.$store.commit("appSearch/pushFields", [
      {
        items: this.services
          .concat(this.instances)
          .concat(this.sps)
          .concat(this.plans),
        type: "select",
        title: "Entity",
        key: "entity",
        item: { value: "uuid", title: "title" },
        single: true,
      },
      {
        type: "select",
        key: "account",
        custom: true,
        component: AccountsAutocomplete,
        label: "Accounts",
        clearable: true,
        fetchValue: true,
      },
    ]);
  },
  computed: {
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
    plans() {
      return this.$store.getters["plans/all"];
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
