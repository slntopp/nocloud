<template>
  <v-container>
    <v-row>
      <v-col>
        <v-autocomplete
          label="Account (Requestor)"
          :items="accounts"
          item-value="uuid"
          item-text="title"
          v-model="account"
        />
      </v-col>
      <v-col>
        <v-autocomplete
          label="Entity"
          :items="entitys"
          item-value="uuid"
          item-text="title"
          v-model="uuid"
        >
          <template v-slot:item="{ item }">
            <div class="d-flex justify-space-between" style="width: 100%">
              <span class="text-start">{{ item.title }}</span>
              <span class="ml-2 text-end">{{ item.entity }}</span>
            </div>
          </template>
        </v-autocomplete>
      </v-col>
      <v-col>
        <v-text-field label="Path" :value="path" @change="path = $event" />
      </v-col>
      <v-col cols="4"></v-col>
    </v-row>
    <history-table
      :path="path"
      :account-id="account"
      :uuid="uuid"
      table-name="all-logs"
    />
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
    path: null,
  }),
  async mounted() {
    await Promise.all([
      this.$store.dispatch("accounts/fetch"),
      this.$store.dispatch("services/fetch"),
      this.$store.dispatch("servicesProviders/fetch"),
    ]);

    this.$store.commit("appSearch/setAdvancedSearch", true);
    this.$store.commit("appSearch/setVariants", {
      service: { items: this.services, title: "Service" },
      sp: { items: this.sps, title: "Service providers" },
      instance: { items: this.instances, title: "Instances" },
      account: { items: this.accounts, title: "Accounts" },
    });
  },
  computed: {
    ...mapGetters("accounts", { accounts: "all" }),
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
};
</script>

<style scoped></style>
