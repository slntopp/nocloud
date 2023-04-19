<template>
  <v-container>
    <v-row>
      <v-col cols="2">
        <v-select
          label="Account"
          :items="accounts"
          item-value="uuid"
          item-text="title"
          clearable
          v-model="account"
        />
      </v-col>
      <v-col cols="2">
        <v-select
          label="Service"
          :items="services"
          item-value="uuid"
          item-text="title"
          v-model="uuid"
          clearable
        />
      </v-col>
      <v-col cols="2">
        <v-select
          label="Instance"
          :items="instances"
          item-value="uuid"
          item-text="title"
          v-model="uuid"
          clearable
        />
      </v-col>
    </v-row>
    <history-table :account-id="account" :uuid="uuid" table-name="all-logs" />
  </v-container>
</template>

<script>
import HistoryTable from "@/components/historyTable.vue";
import { mapGetters } from "vuex";

export default {
  name: "all-history",
  components: { HistoryTable },
  data: () => ({
    account: null,
    uuid: null,
  }),
  mounted() {
    this.$store.dispatch("accounts/fetch");
    this.$store.dispatch("services/fetch");
  },
  computed: {
    ...mapGetters("accounts", { accounts: "all" }),
    ...mapGetters("services", { services: "all" }),
    ...mapGetters("services", { instances: "getInstances" }),
  },
};
</script>

<style scoped></style>
