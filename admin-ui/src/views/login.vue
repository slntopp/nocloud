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
                  v-if="typesAccounts.length > 1"
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

<script>
import snackbar from "@/mixins/snackbar";

export default {
  name: "login-view",
  mixins: [snackbar],
  data() {
    return {
      loginFormRules: [(v) => !!v || "Required"],
      loginLoading: false,
      isLoginFailed: false,
      username: "",
      password: "",
      type: "Standard",
      typesAccounts: ["Standard"],
    };
  },
  methods: {
    tryLogin() {
      this.loginLoading = true;
      (this.isLoginFailed = false),
        this.$store
          .dispatch("auth/login", {
            login: this.username,
            password: this.password,
            type: this.type.toLowerCase(),
          })
          .then(() => {
            this.$router.push({ name: "Home" });
            this.$store.dispatch("auth/fetchUserData");
          })
          .catch((error) => {
            console.log(error);
            this.showSnackbarError({
              message: error.response.data.message || "Error during login",
            });
            if (error.response && error.response.status == 401) {
              this.isLoginFailed = true;
            }
          })
          .finally(() => {
            this.loginLoading = false;
          });
    },
  },
  computed: {
    getErrorMessages() {
      return this.isLoginFailed ? ["username or password is not correct"] : [];
    },
  },
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
