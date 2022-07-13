<template>
  <div class="service pa-4 h-100">
    <div class="page__title mb-5">
      <router-link :to="{ name: 'Services' }">{{ navTitle('Services') }}</router-link>
      /
      {{ serviceTitle }}
      <v-chip x-small :color="chipColor"> </v-chip>
    </div>

    <v-tabs
      class="rounded-t-lg"
      v-model="tabs"
      background-color="background-light"
    >
      <v-tab>Info</v-tab>
      <!-- <v-tab>Control</v-tab> -->
      <v-tab>Template</v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabs"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <service-info v-if="service" :service="service" :chipColor="chipColor" />
      </v-tab-item>

      <!-- <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <service-control
          v-if="service"
          :service="service"
          :chip-color="chipColor"
        />
      </v-tab-item> -->

      <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <template v-if="!editing">
          <service-template v-if="service" :service="service" @getType="changeType" />
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
import serviceTemplate from "@/components/service/template.vue";
// import serviceControl from "@/components/service/control.vue";
import serviceInfo from "@/components/service/info.vue";
import JsonTextarea from '@/components/JsonTextarea.vue';
import YamlEditor from '@/components/YamlEditor.vue';

export default {
  name: "service-view",
  components: {
    "service-template": serviceTemplate,
    // "service-control": serviceControl,
    "service-info": serviceInfo,
    JsonTextarea,
    YamlEditor
  },
  mixins: [snackbar],
  data: () => ({
    found: false,
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
    service() {
      const items = this.$store.getters["services/all"];
      const item = items.find((el) => el.uuid == this.serviceId);

      if (item) return item;

      return null;
    },
    serviceId() {
      return this.$route.params.serviceId;
    },
    chipColor() {
      const dict = {
        init: "orange darken-2",
        up: "green darken-2",
        del: "gray darken-2",
      };
      return dict?.[this?.service?.status] ?? "blue-grey darken-2";
    },
    serviceTitle() {
      return this?.service?.title ?? "not found";
    },
    servicesLoading() {
      return this.$store.getters["services/loading"];
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetch");
    this.$store.dispatch("services/fetchById", this.serviceId).then(() => {
      this.found = !!this.service;
      document.title = `${this.serviceTitle} | NoCloud`;
    });
  },
  mounted() {
    document.title = `${this.serviceTitle} | NoCloud`;
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetchById",
      params: this.serviceId,
    });

    const url = 'wss://api.nocloud.ione-cloud.net/services';
    const socket = new WebSocket(`${url}/${this.serviceId}/stream`);

    socket.onmessage = (msg) => {
      const response = JSON.parse(msg.data).result;
      if (!response) {
        this.showSnackbarError({
          message: `Empty response, ${msg}`
        });
        return;
      }

      try {
        this.$store.commit('services/updateInstance', {
          value: response, uuid: this.serviceId
        });
      } catch {
        socket.close(1000, 'работа закончена');
      }
    };
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
