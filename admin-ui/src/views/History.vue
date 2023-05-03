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
        <v-text-field label="Path" :value="path" @change="path=$event"/>
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
  mounted() {
    this.$store.dispatch("accounts/fetch");
    this.$store.dispatch("services/fetch");
    this.$store.dispatch("servicesProviders/fetch");
  },
  computed: {
    ...mapGetters("accounts", { accounts: "all" }),
    entitys() {
      const instances = this.$store.getters["services/getInstances"].map(
        (s) => {
          s.entity = "Instance";
          return s;
        }
      );
      const sps = this.$store.getters["servicesProviders/all"].map((s) => {
        s.entity = "Service provider";
        return s;
      });
      const services = this.$store.getters["services/all"].map((s) => {
        s.entity = "Service";
        return s;
      });
      return instances.concat(sps).concat(services);
    },
  },
};
</script>

<style scoped></style>
