<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Accounts' }">{{ navTitle('Accounts') }}</router-link>
      / {{ accountTitle }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="tabs"
    >
      <v-tab>Info</v-tab>
      <v-tab>Template</v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="tabs"
    >
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <accounts-info v-if="account" :account="account" />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <template v-if="!editing">
          <accounts-template v-if="account" :template="account" @getType="changeType" />
          <v-btn
            class="ma-4 mt-0"
            @click="editing = true"
          >
            Edit
          </v-btn>
        </template>
        <template v-else>
          <json-textarea class="mx-4" v-if="type === 'JSON'" :json="account" @getTree="changeTree" />
          <yaml-editor v-else class="mx-4" :json="account" @getTree="changeTree" />
          <v-btn
            class="ma-4 mt-0"
            color="success"
            :disabled="!isValid"
            @click="editAccount"
          >
            Save
          </v-btn>
          <v-btn
            class="mb-4"
            @click="cancel"
          >
            Cancel
          </v-btn>
        </template>
      </v-tab-item>
    </v-tabs-items>

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
import yaml from 'yaml';
import config from '@/config.js';
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import AccountsInfo from '@/components/account/info.vue';
import AccountsTemplate from '@/components/account/template.vue';
import JsonTextarea from '@/components/JsonTextarea.vue';
import YamlEditor from '@/components/YamlEditor.vue';

export default {
  name: 'account-view',
  components: { AccountsInfo, AccountsTemplate, JsonTextarea, YamlEditor },
  mixins: [snackbar],
  data: () => ({
    tabs: 0,
    navTitles: config.navTitles ?? {},

    type: 'YAML',
    tree: '',
    isValid: false,
    isLoading: false,
    editing: false
  }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    changeType(value) {
      this.type = value;
    },
    changeTree(value) {
      try {
        if (this.type === 'JSON') JSON.parse(value);
        else yaml.parse(value);

        this.tree = value;
        this.isValid = true;
      } catch {
        this.isValid = false;
      }
    },
    editAccount() {
      this.isLoading = true;
      api.accounts.update(this.account.uuid, JSON.parse(this.tree))
        .then(() => {
          this.showSnackbarSuccess({
            message: 'Account edited successfully'
          });

          setTimeout(() => {
            this.$router.push({ name: 'Accounts' });
          }, 1500);
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    cancel() {
      this.editing = false;
      this.isValid = false;
      this.type = 'YAML';
    }
  },
  computed: {
    account() {
      const id = this.$route.params?.accountId;

      return this.$store.getters['accounts/all']
        .find(({ uuid }) => uuid === id);
    },
    accountTitle() {
      return this?.account?.title ?? 'not found';
    },
    accountLoading() {
      return this.$store.getters['accounts/loading'];
    },
  },
  created() {
    this.$store.dispatch('accounts/fetch')
      .then(() => {
        document.title = `${this.accountTitle} | NoCloud`;
      });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: 'accounts/fetch'
    });
  }
}
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
