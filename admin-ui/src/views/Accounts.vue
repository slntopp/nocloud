<template>
  <div class="namespaces pa-4 flex-wrap">
    <div class="buttons__inline pb-8 pt-4">
      <v-menu
        offset-y
        transition="slide-y-transition"
        bottom
        :close-on-content-click="false"
        v-model="createMenuVisible"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            color="background-light"
            class="mr-2"
            v-bind="attrs"
            v-on="on"
            @click="openCreateAccountMenuHandler"
          >
            create
          </v-btn>
        </template>
        <v-card class="pa-4">
          <v-form ref="form" v-model="formValid">
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  v-model="newAccount.title"
                  placeholder="title"
                  :rules="rules.title"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  v-model="newAccount.auth.data[0]"
                  placeholder="username"
                  :rules="rules.required"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  v-model="newAccount.data.email"
                  placeholder="email"
                  :rules="rules.email"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  v-model="newAccount.auth.data[1]"
                  placeholder="password"
                  type="password"
                  :rules="rules.required"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-autocomplete
                  dense
                  :items="namespacesForSelect"
                  v-model="newAccount.namespace"
                  label="namespace"
                  :rules="rules.required"
                ></v-autocomplete>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-select
                  dense
                  :items="accessLevels"
                  item-value="id"
                  item-text="title"
                  v-model="newAccount.access"
                  label="access"
                ></v-select>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-select
                  dense
                  :items="currencies"
                  v-model="newAccount.currency"
                  label="currency"
                ></v-select>
              </v-col>
            </v-row>
            <v-row justify="end">
              <v-col md="5">
                <v-btn :loading="loading" @click="createAccount"> send </v-btn>
              </v-col>
            </v-row>
          </v-form>
        </v-card>
      </v-menu>
      <confirm-dialog
        :disabled="selected.length < 1"
        @confirm="deleteSelectedAccount"
      >
        <v-btn
          color="background-light"
          class="mr-8"
          :disabled="selected.length < 1"
          :loading="deletingLoading"
        >
          delete
        </v-btn>
      </confirm-dialog>
    </div>

    <accounts-table :searchParam="searchParam" v-model="selected">
    </accounts-table>
  </div>
</template>

<script>
import accountsTable from "@/components/accounts_table.vue";
import api from "@/api.js";

import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";

import ConfirmDialog from "../components/confirmDialog.vue";

export default {
  name: "accounts-view",
  components: {
    "accounts-table": accountsTable,
    ConfirmDialog,
  },
  mixins: [snackbar, search],
  data() {
    return {
      createMenuVisible: false,
      selected: [],
      newAccount: {},
      rules: {
        title: [
          (value) => !!value || "Title is required",
          (value) => (value || "").length >= 3 || "Min 3 characters",
        ],
        required: [(value) => !!value || "Namespace is required"],
        email: [
          (value) =>
            !!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.exec(value) || "Wrong email",
        ],
      },
      formValid: true,
      loading: false,
      deletingLoading: false,
      accessLevels: [
        { id: 0, title: "NONE" },
        { id: 1, title: "READ" },
        { id: 2, title: "MGMT" },
        { id: 3, title: "ADMIN" },
      ],
    };
  },
  created() {
    this.setDefaultAccount();
  },
  methods: {
    setDefaultAccount() {
      this.newAccount = {
        title: "",
        auth: {
          type: "standard",
          data: ["", ""],
        },
        namespace: "",
        access: 1,
        currency: this.defaultCurrency,
        data: {
          email: "",
        },
      };
    },
    createAccount() {
      if (!this.formValid) return;
      this.loading = true;
      api.accounts
        .create(this.newAccount)
        .then(() => {
          this.newAccount.title = "";
          this.createMenuVisible = false;
          this.$store.dispatch("accounts/fetch");

          this.setDefaultAccount();
        })
        .catch((error) => {
          console.error(error);
          this.snackbar.message = "Something went wrong... Try later.";
          this.snackbar.visibility = true;
        })
        .finally(() => {
          this.loading = false;
        });
    },
    deleteSelectedAccount() {
      if (this.selected.length > 0) {
        const deletePromices = this.selected.map((el) =>
          api.accounts.delete(el.uuid)
        );
        this.deletingLoading = true;
        Promise.all(deletePromices)
          .then((res) => {
            if (res.every((el) => el.result)) {
              this.snackbar.message = `Account${
                deletePromices.length == 1 ? "" : "s"
              } deleted successfully.`;
              this.snackbar.visibility = true;
            }

            this.selected = [];
            this.$store.dispatch("accounts/fetch");
          })
          .catch((err) => {
            console.error(err);
          })
          .finally(() => {
            this.deletingLoading = false;
          });
      }
    },
    openCreateAccountMenuHandler() {
      this.$store.dispatch("namespaces/fetch");
    },
  },
  computed: {
    namespacesForSelect() {
      let namespaces = this.$store.getters["namespaces/all"] ?? [];
      namespaces = namespaces.map((namespace) => ({
        text: namespace.title,
        value: namespace.uuid,
      }));
      return namespaces;
    },
    searchParam() {
      return this.$store.getters["appSearch/param"];
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    currencies() {
      return this.$store.getters["currencies/all"];
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "accounts/fetch",
    });

    this.$store.commit("appSearch/setSearchName", "all-accounts");
    this.$store.dispatch("currencies/fetch");
  },
  watch: {
    defaultCurrency(newVal) {
      if (!this.newAccount.currency) {
        this.newAccount.currency = newVal;
      }
    },
  },
};
</script>

<style></style>
