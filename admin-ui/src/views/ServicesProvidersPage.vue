<template>
  <div class="servicesProviders pa-4 flex-wrap">
    <div class="page__title mb-5">
      <router-link :to="{ name: 'ServicesProviders' }"
        >{{ navTitle('Services Providers') }}</router-link
      >
      /
      {{ title }}
    </div>

    <v-tabs
      class="rounded-t-lg"
      v-model="tabsIndex"
      background-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab.title">
        {{ tab.title }}
      </v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabsIndex"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item v-for="tab in tabs" :key="tab.title">
        <v-progress-linear v-if="loading" indeterminate class="pt-2" />
        <component v-if="!loading && item" :is="tab.component" :template="item" />
        <template v-if="!editing && tab.title === 'Template'">
          <component v-if="!loading && item" :is="tab.component" :template="item" @getType="changeType" />
          <v-btn
            class="ma-4 mt-0"
            @click="editing = true"
          >
            Edit
          </v-btn>
        </template>
        <template v-if-else="tab.title === 'Template'">
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
import JsonTextarea from '@/components/JsonTextarea.vue';
import YamlEditor from '@/components/YamlEditor.vue';

export default {
  name: 'service-providers-view',
  components: { JsonTextarea, YamlEditor  },
  mixins: [snackbar],
  data: () => ({
    found: false,
    tabsIndex: 0,
    navTitles: config.navTitles ?? {},
    tabs: [
      {
        title: "Info",
        component: () => import("@/components/ServicesProvider/info.vue"),
      },
      {
        title: "Template",
        component: () => import("@/components/ServicesProvider/template.vue"),
      },
    ],

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
    uuid() {
      return this.$route.params.uuid;
    },
    item() {
      const items = this.$store.getters["servicesProviders/all"];
      const item = items.find((el) => el.uuid == this.uuid);

      if (item) return item;

      return null;
    },
    title() {
      return this?.item?.title ?? "not found";
    },
    loading() {
      return this.$store.getters["servicesProviders/isLoading"];
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetchById", this.uuid).then(() => {
      this.found = !!this.service;
      document.title = `${this.title} | NoCloud`;
    });
  },
  mounted() {
    document.title = `${this.title} | NoCloud`;
    this.$store.commit("reloadBtn/setCallback", {
      type: "servicesProviders/fetchById",
      params: this.uuid,
    });
  },
};
</script>

<style>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
