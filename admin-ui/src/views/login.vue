<template>
  <div class="login-view">
    <v-container fluid fill-height>
      <v-layout align-center justify-center>
        <v-flex xs12 sm8 md4 xl3>
          <v-card class="elevation-12">
            <v-toolbar dark color="primary">
              <v-toolbar-title>Login</v-toolbar-title>
            </v-toolbar>
            <v-card-text>
              <v-form>
                <v-text-field
                  v-model.trim="username"
                  prepend-icon="mdi-account"
                  name="login"
                  label="Login"
                  type="text"
                  :rules="loginFormRules"
                  :error="isLoginFailed"
                ></v-text-field>
                <v-text-field
                  v-model="password"
                  uuid="password"
                  prepend-icon="mdi-lock"
                  name="password"
                  label="Password"
                  type="password"
                  :rules="loginFormRules"
                  :error="isLoginFailed"
                  :error-messages="getErrorMessages"
                  @keypress.enter="tryLogin"
                ></v-text-field>
                <v-select
                  v-model="type"
                  class="type-select"
                  :items="typesAccounts"
                  label="Type"
                ></v-select>
              </v-form>
            </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn color="primary" @click="tryLogin" :loading="loginLoading">
                Login
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRouter } from "vue-router/composables";
import { useStore } from "../store";

const loginFormRules = ref([(v) => !!v || "Required"]);
const loginLoading = ref(false);
const isLoginFailed = ref(false);
const username = ref("");
const password = ref("");
const type = ref("Standard");
const typesAccounts = ref(["Standard", "WHMCS"]);

const router = useRouter();
const store = useStore();

const tryLogin = async () => {
  loginLoading.value = true;

  try {
    const { token } = await store.dispatch("auth/login", {
      login: username.value,
      password: password.value,
      type: type.value.toLowerCase(),
    });
    store.commit("auth/setToken", "");

    await store.dispatch("namespaces/fetchById", 0);

    store.commit("auth/setToken", token);

    await router.push({ name: "Home" });
    store.dispatch("auth/fetchUserData");
  } catch (error) {
    store.dispatch("auth/logout");
    store.commit("snackbar/showSnackbarError", {
      message: error.response.data.message || "Error during login",
    });
    if (error.response && error.response.status == 401) {
      isLoginFailed.value = true;
    }
  } finally {
    loginLoading.value = false;
  }
};

const getErrorMessages = computed(() => {
  return isLoginFailed.value ? ["username or password is not correct"] : [];
});
</script>

<script>
export default {
  name: "LoginView",
};
</script>

<style>
.login-view {
  height: 100%;
}

.type-select {
  margin-left: 30px;
}
</style>
