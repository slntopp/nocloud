<template>
  <div class="accounts pa-4 flex-wrap">
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
            class="mr-2 mt-2"
            v-bind="attrs"
            v-on="on"
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
            <v-row justify="end">
              <v-btn :loading="isAccountCreateLoading" @click="createAccount">
                create
              </v-btn>
            </v-row>
          </v-form>
        </v-card>
      </v-menu>
    </div>

    <accounts-table
      v-if="account.subaccounts.length"
      table-name="account-subaccounts"
      no-search
      :custom-filter="customFilter"
    />
  </div>
</template>

<script setup>
import AccountsTable from "@/components/accounts_table.vue";
import { computed, ref, toRefs } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import { Level } from "nocloud-proto/proto/es/access/access_pb";

const props = defineProps(["account"]);
const { account } = toRefs(props);

const store = useStore();

const formValid = ref(false);
const createMenuVisible = ref(false);
const isAccountCreateLoading = ref(false);
const newAccount = ref({
  auth: {
    data: {},
  },
});
const rules = ref({
  title: [
    (value) => !!value || "Title is required",
    (value) => (value || "").length >= 3 || "Min 3 characters",
  ],
  email: [
    (value) => !!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.exec(value) || "Wrong email",
  ],
});

const customFilter = computed(() => ({
  uuid: account.value.subaccounts || [],
}));

const namespace = computed(() => store.getters["namespaces/all"][0]);

const setDefaultAccount = () => {
  newAccount.value = {
    title: "",
    auth: {
      type: "standard",
      data: ["", ""],
    },
    namespace: namespace.value?.uuid,
    access: Level[account.value?.access.level],
    currency: account.value?.currency,
    accountOwner: account.value?.uuid,
    data: {
      email: "",
    },
  };
};

const createAccount = async () => {
  if (!formValid.value) return;
  isAccountCreateLoading.value = true;
  try {
    await api.accounts.create({
      ...newAccount.value,
    });
    setDefaultAccount();
    createMenuVisible.value = false;
    store.dispatch("accounts/fetchById", account.value.uuid);
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Something went wrong... Try later.",
    });
  } finally {
    isAccountCreateLoading.value = false;
  }
};

setDefaultAccount();
</script>

<script>
export default {
  name: "account-subaccounts",
};
</script>

<style scoped></style>
