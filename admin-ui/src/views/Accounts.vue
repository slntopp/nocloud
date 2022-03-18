<template>
  <div class="namespaces pa-4 flex-wrap">
    <div class="buttons__inline pb-4">
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
          <v-from ref="form" v-model="newAccount.formValid">
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  v-model="newAccount.data.title"
                  placeholder="title"
                  :rules="newAccount.rules.title"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  hide-details
                  v-model="newAccount.data.auth.data[0]"
                  placeholder="username"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  hide-details
                  v-model="newAccount.data.auth.data[1]"
                  placeholder="password"
                  type="password"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-select
                  :items="namespacesForSelect"
                  v-model="newAccount.data.namespace"
                  label="namespace"
                  :rules="newAccount.rules.selector"
                ></v-select>
              </v-col>
            </v-row>
            <v-row>
              <v-col>
                <v-select
                  :items="accessLevels"
                  v-model="newAccount.data.access"
                  label="access"
                ></v-select>
              </v-col>
            </v-row>
            <v-row justify="end">
              <v-col md="5">
                <v-btn :loading="newAccount.loading" @click="createAccount">
                  send
                </v-btn>
              </v-col>
            </v-row>
          </v-from>
        </v-card>
      </v-menu>

      <v-btn
        color="background-light"
        class="mr-8"
        :disabled="selected.length < 1"
        @click="deleteSelectedAccount"
        :loading="deletingLoading"
      >
        delete
      </v-btn>
    </div>

    <accounts-table v-model="selected"> </accounts-table>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script>
import accountsTable from "@/components/accounts_table.vue";
import api from "@/api.js";

import snackbar from "@/mixins/snackbar.js";

export default {
  name: "accounts-view",
  components: {
    "accounts-table": accountsTable,
  },
  mixins: [snackbar],
  data() {
    return {
      createMenuVisible: false,
      selected: [],
      newAccount: {
        data: {
          title: "",
          auth: {
            type: "standard",
            data: ["", ""],
          },
          namespace: "",
          access: 1,
        },
        rules: {
          title: [
            (value) => !!value || "Title is required",
            (value) => (value || "").length >= 3 || "Min 3 characters",
          ],
          selector: [(value) => !!value || "Namespace is required"],
        },
        formValid: true,
        loading: false,
      },
      deletingLoading: false,
      accessLevels: [0, 1, 2, 3],
    };
  },
  methods: {
    createAccount() {
      if (!this.newAccount.formValid) return;
      this.newAccount.loading = true;
      api.accounts
        .create(this.newAccount.data)
        .then(() => {
          this.newAccount.title = "";
          this.createMenuVisible = false;
          this.$store.dispatch("accounts/fetch");

          this.newAccount.data = {
            title: "",
            auth: {
              type: "standard",
              data: ["", ""],
            },
            namespace: "",
            access: 1,
          };
        })
        .catch((error) => {
          console.error(error);
          this.snackbar.message = "Something went wrong... Try later.";
          this.snackbar.visibility = true;
        })
        .finally(() => {
          this.newAccount.loading = false;
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
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", { type: "accounts/fetch" });
  },
};
</script>

<style></style>
