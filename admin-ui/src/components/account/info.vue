<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <div style="position: absolute; top: 0; right: 75px">
      <div>
        <v-chip color="primary" outlined
          >Balance: {{ account.balance?.toFixed(2) }}
          {{ account.currency }}</v-chip
        >
      </div>
    </div>

    <v-row>
      <v-col cols="3">
        <v-text-field v-model="uuid" readonly label="UUID" />
      </v-col>
      <v-col cols="3">
        <v-text-field v-model="title" label="name" style="width: 330px" />
      </v-col>
      <v-col cols="3">
        <v-select :items="currencies" v-model="currency" label="currency" style="width: 330px" />
      </v-col>
    </v-row>
    <v-card-title class="px-0">Instances:</v-card-title>

    <instances-table
      :value="null"
      :items="accountInstances"
      :show-select="false"
    />

    <v-card-title class="px-0">SSH keys:</v-card-title>

    <div class="pt-4">
      <v-menu
        bottom
        offset-y
        transition="slide-y-transition"
        v-model="isVisible"
        :close-on-content-click="false"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn class="mr-2" v-bind="attrs" v-on="on"> Create </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row>
            <v-col>
              <v-text-field
                dense
                label="title"
                v-model="newKey.title"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-textarea
                dense
                label="key"
                v-model="newKey.value"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-btn @click="addKey"> Send </v-btn>
            </v-col>
          </v-row>
        </v-card>
      </v-menu>

      <v-btn class="mr-8" :disabled="selected.length < 1" @click="deleteKeys">
        Delete
      </v-btn>
    </div>

    <nocloud-table
      table-name="account-info"
      class="mt-4"
      item-key="value"
      v-model="selected"
      :items="keys"
      :headers="headers"
    />

    <v-btn class="mt-4 mr-2" :loading="isEditLoading" @click="editAccount">
      Save
    </v-btn>
  </v-card>
</template>

<script>
import config from "@/config.js";
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import InstancesTable from "@/components/instances_table.vue";

export default {
  name: "account-info",
  components: { InstancesTable, nocloudTable },
  mixins: [snackbar],
  props: ["account"],
  data: () => ({
    newKey: { title: "", value: "" },
    headers: [
      { text: "Title", value: "title" },
      { text: "Key", value: "value" },
    ],
    generalRule: [(v) => !!v || "Required field"],
    navTitles: config.navTitles ?? {},
    uuid: "",
    title: "",
    currency:"",
    keys: [],
    selected: [],
    isVisible: false,
    isEditLoading: false,
  }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    addKey() {
      this.keys.push(this.newKey);
      this.isVisible = false;
      this.newKey = { title: "", value: "" };
    },
    deleteKeys() {
      if (this.selected.length < 1) return;
      const arr = this.selected.map((el) => el.value);

      this.keys = this.keys.filter((el) => !arr.includes(el.value));
      this.selected = [];
    },
    editAccount() {
      const newAccount = { ...this.account, title: this.title ,currency:this.currency};
      if(!newAccount.data){
        newAccount.data={}
      }
      newAccount.data.ssh_keys = this.keys;

      this.isEditLoading = true;
      api.accounts
        .update(this.account.uuid, newAccount)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Account edited successfully",
          });

          setTimeout(() => {
            this.$router.push({ name: "Accounts" });
          }, 1500);
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isEditLoading = false;
        });
    },
  },
  mounted() {
    this.title = this.account.title;
    this.currency = this.account.currency;
    this.uuid = this.account.uuid;
    this.keys = this.account.data?.ssh_keys || [];
    if (this.namespaces.length < 2) {
      this.$store.dispatch("namespaces/fetch");
    }
    if (this.services.length < 2) {
      this.$store.dispatch("services/fetch");
    }
    if (this.servicesProviders.length < 2) {
      this.$store.dispatch("servicesProviders/fetch");
    }
  },
  computed: {
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    services() {
      return this.$store.getters["services/all"];
    },
    currencies() {
      return this.$store.getters["currencies/all"].filter((c) => c !== "NCU");
    },
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
    instances() {
      return this.$store.getters["services/getInstances"];
    },
    accountInstances() {
      const accountNamespace = this.namespaces.find(
        (n) => n.access.namespace === this.account.uuid
      );
      return this.instances.filter(
        (i) => i.access.namespace === accountNamespace.uuid
      );
    },
  },
};
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
