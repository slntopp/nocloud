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
          <v-btn color="background-light" class="mr-2 mt-2" v-bind="attrs" v-on="on">
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
        @confirm="changeInvoiceBased(true)"
      >
        <v-btn
          color="background-light"
          class="mr-2 mt-2"
          :disabled="
            selected.length < 1 ||
            changeInvoiceBasedAction === false ||
            !!changeAccountStatusAction
          "
          :loading="changeInvoiceBasedAction === true"
        >
          Enabled invoice based
        </v-btn>
      </confirm-dialog>
      <confirm-dialog
        :disabled="selected.length < 1"
        @confirm="changeInvoiceBased(false)"
      >
        <v-btn
          color="background-light"
          class="mr-2 mt-2"
          :disabled="
            selected.length < 1 ||
            changeInvoiceBasedAction === true ||
            !!changeAccountStatusAction
          "
          :loading="changeInvoiceBasedAction === false"
        >
          Disabled invoice based
        </v-btn>
      </confirm-dialog>
      <confirm-dialog
        v-for="btn in changeStateButtons"
        :key="btn.value"
        :disabled="
          selected.length < 1 ||
          (!!changeAccountStatusAction &&
            changeAccountStatusAction !== btn.value)
        "
        @confirm="changeAccountsStatus(btn.value)"
      >
        <v-btn
          color="background-light"
          class="mr-2 mt-2"
          :disabled="
            selected.length < 1 ||
            (!!changeAccountStatusAction &&
              changeAccountStatusAction !== btn.value)
          "
          :loading="changeAccountStatusAction === btn.value"
        >
          {{ btn.title }}
        </v-btn>
      </confirm-dialog>
    </div>

    <accounts-table v-model="selected"> </accounts-table>
  </div>
</template>

<script>
import accountsTable from "@/components/accounts_table.vue";
import api from "@/api.js";

import snackbar from "@/mixins/snackbar.js";

import ConfirmDialog from "../components/confirmDialog.vue";

export default {
  name: "accounts-view",
  components: {
    "accounts-table": accountsTable,
    ConfirmDialog,
  },
  mixins: [snackbar],
  data() {
    return {
      createMenuVisible: false,
      selected: [],
      newAccount: {},
      changeInvoiceBasedAction: undefined,
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
      changeAccountStatusAction: "",
      formValid: true,
      loading: false,
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
          this.showSnackbarError({
            message: "Something went wrong... Try later.",
          });
        })
        .finally(() => {
          this.loading = false;
        });
    },
    async changeAccountsStatus(newStatus) {
      this.changeAccountStatusAction = newStatus;
      try {
        const servicesForDown = [];
        const accountServices = [];

        const accounts = this.selected.filter((account) => {
          if (account.status === newStatus) {
            return false;
          }

          switch (newStatus) {
            case "PERMANENT_LOCK": {
              const accountNamespace = this.namespaces.find(
                (n) => n.access.namespace === account?.uuid
              );

              accountServices.push(
                ...this.services.filter(
                  (s) => s.access.namespace === accountNamespace?.uuid
                )
              );

              servicesForDown.push(
                ...accountServices.filter((s) => s.status !== "INIT")
              );
              return true;
            }
            case "LOCK": {
              return account.status !== "PERMANENT_LOCK";
            }
            case "ACTIVE": {
              return account.status === "LOCK";
            }
          }
        });

        if (servicesForDown.length) {
          await Promise.all(
            servicesForDown.map((s) => api.services.down(s.uuid))
          );
          await Promise.all(
            accountServices.map((s) => api.services.delete(s.uuid))
          );
        }
        await Promise.all(
          accounts.map((account) =>
            fetch(
              /https:\/\/(.+?\.?\/)/.exec(this.whmcsApi)[0] +
                `modules/addons/nocloud/api/index.php?run=status_user&account=${
                  account.uuid
                }&status=${newStatus === "ACTIVE" ? "open" : "close"}`
            )
          )
        );

        await Promise.all(
          accounts.map((account) =>
            api.accounts.update(account.uuid, { ...account, status: newStatus })
          )
        );

        if (accounts.length) {
          await this.$store.dispatch("accounts/fetch");
        }
      } catch (e) {
        this.showSnackbarError({
          message: "Error while change accounts statuses",
        });
      } finally {
        this.changeAccountStatusAction = "";
        this.selected = [];
      }
    },
    async changeInvoiceBased(value) {
      this.changeInvoiceBasedAction = value;
      try {
        await Promise.all(
          this.selected.map((el) => {
            if (el.data?.regular_payment === value) {
              return Promise.resolve();
            }
            if (!el.data) {
              el.data = {};
            }
            el.data.regular_payment = value;
            return api.accounts.update(el.uuid, el);
          })
        );
        this.$store.dispatch("accounts/fetch");
        this.showSnackbarSuccess({ message: "Success" });
      } finally {
        this.changeInvoiceBasedAction = undefined;
      }
    },
  },
  computed: {
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    namespacesForSelect() {
      let namespaces = this.namespaces ?? [];
      namespaces = namespaces.map((namespace) => ({
        text: namespace.title,
        value: namespace.uuid,
      }));
      return namespaces;
    },
    services() {
      return this.$store.getters["services/all"];
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    currencies() {
      return this.$store.getters["currencies/all"].filter((c) => c !== "NCU");
    },
    changeStateButtons() {
      return [
        { title: "Unlock", value: "ACTIVE" },
        { title: "Lock", value: "LOCK" },
        { title: "PERMANENT LOCK", value: "PERMANENT_LOCK" },
      ];
    },
    whmcsApi() {
      return this.$store.getters["settings/whmcsApi"];
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "accounts/fetch",
    });
    this.$store.dispatch("currencies/fetch");
    this.$store.dispatch("services/fetch",{showDeleted:true});
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
