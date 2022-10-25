<template>
  <div class="servicesProviders-create pa-4">
    <div class="page__title">Create service provider</div>
    <v-container>
      <v-row>
        <v-col lg="6" cols="12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader> Provider type </v-subheader>
            </v-col>

            <v-col cols="9">
              <v-select
                v-model="provider.type"
                :items="types"
                label="Type"
              ></v-select>
            </v-col>
          </v-row>
          <v-row align="center" v-if="provider.type === 'custom'">
            <v-col cols="3">
              <v-subheader> Key </v-subheader>
            </v-col>

            <v-col cols="9">
              <v-text-field v-model="key" label="Key"></v-text-field>
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader> Provider title </v-subheader>
            </v-col>

            <v-col cols="9">
              <v-text-field
                v-model="provider.title"
                label="Title"
              ></v-text-field>
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader> Public </v-subheader>
            </v-col>

            <v-col cols="9">
              <v-switch v-model="provider.public" />
            </v-col>
          </v-row>

          <v-divider></v-divider>

          <component
            :is="templates[provider.type]"
            :secrets="provider.secrets"
            @change:secrets="(data) => handleFieldsChange('secrets', data)"
            :vars="provider.vars"
            @change:vars="(data) => handleFieldsChange('vars', data)"
            :passed="isPassed"
            @passed="(data) => (isPassed = data)"
          ></component>
        </v-col>
        <v-col lg="6" cols="12">
          <v-tabs
            v-model="tabs"
            background-color="background-light"
            class="mb-2"
          >
            <v-tab>Extentions</v-tab>
          </v-tabs>

          <v-tabs-items v-model="tabs" color="primary">
            <v-tab-item>
              <v-card color="background">
                <v-row>
                  <v-col>
                    <v-select
                      v-model="extentions.selected"
                      :items="
                        extentions.items
                          .filter(
                            (el) => !Object.keys(extentions.data).includes(el)
                          )
                          .map((el) => ({
                            value: el,
                            text: extentionsMap[el].title,
                          }))
                      "
                      label="extention"
                      no-data-text="no extentions avaliable"
                    ></v-select>
                  </v-col>
                  <v-col>
                    <v-btn
                      color="background-light"
                      class="mt-3"
                      :disabled="extentions.selected.length < 1"
                      @click="addExtention"
                    >
                      Add
                    </v-btn>
                  </v-col>
                </v-row>

                <component
                  v-for="extention in Object.keys(extentions.data)"
                  :key="extention.title"
                  :is="extentionsMap[extention].component"
                  :provider="provider"
                  :data="extentions.data[extention]"
                  @change:data="(data) => (extentions.data[extention] = data)"
                  @change:provider="
                    (data) => (provider = mergeDeep(provider, data))
                  "
                  @remove="() => removeExtention(extention)"
                />
              </v-card>
            </v-tab-item>
          </v-tabs-items>
        </v-col>
      </v-row>

      <v-row justify="end">
        <v-col col="6">
          <v-tooltip bottom :disabled="isTestSuccess">
            <template v-slot:activator="{ on, attrs }">
              <div v-bind="attrs" v-on="on" style="display: inline-block">
                <v-btn
                  color="background-light"
                  class="mr-2"
                  @click="tryToSend"
                  :loading="isLoading"
                  :disabled="!isTestSuccess"
                >
                  create
                </v-btn>
              </div>
            </template>
            <span>Test must be passed before creation.</span>
          </v-tooltip>

          <v-btn
            :color="testButtonColor"
            class="mr-2"
            @click="testConfig"
            :loading="isTestLoading"
          >
            Test
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
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
import api from "@/api.js";
import Vue from "vue";
import extentionsMap from "@/components/extentions/map.js";
import snackbar from "@/mixins/snackbar.js";

import { mergeDeep } from "@/functions.js";

export default {
  name: "servicesProviders-create",
  mixins: [snackbar],
  data: () => ({
    types: [],
    templates: {},
    key: "",
    provider: {
      type: "custom",
      title: "",
      public: true,
      secrets: {},
      vars: {},
    },

    isPassed: false,
    isLoading: false,
    isTestLoading: false,
    testButtonColor: "background-light",
    isTestSuccess: false,

    tabs: null,
    extentions: {
      loading: false,
      items: [],
      data: {},
      selected: "",
    },

    tooltipVisible: false,
  }),
  created() {
    const types = require.context(
      "@/components/modules/",
      true,
      /serviceProviders\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(
        /\.\/([A-Za-z0-9-_,\s]*)\/serviceProviders\.vue/i
      );
      if (matched && matched.length > 1) {
        const type = matched[1];
        this.types.push(type);
        this.templates[type] = () =>
          import(`@/components/modules/${type}/serviceProviders.vue`);
      }
    });

    this.fetchExtentions();
  },
  computed: {
    template() {
      return () =>
        import(`@/components/modules/${this.type}/serviceProviders.vue`);
    },
    extentionsMap() {
      return extentionsMap;
    },
    serviceProviderBody() {
      if (Object.keys(this.extentions.data).length > 0) {
        if (this.provider.type === "custom") {
          return {
            ...this.provider,
            type: this.key,
            extentions: this.extentions.data,
          };
        } else {
          return {
            ...this.provider,
            extentions: this.extentions.data,
          };
        }
      } else {
        return this.provider;
      }
    },
  },
  methods: {
    handleFieldsChange(type, data) {
      if (type == "secrets") {
        this.provider.secrets = data;
      }
      if (type == "vars") {
        this.provider.vars = data;
      }

      this.testButtonColor = "background-light";
      this.isTestSuccess = false;
    },
    fetchExtentions() {
      this.extentions.loading = true;
      api
        .get("/sp-ext")
        .then((res) => {
          this.extentions.items = res.types;
        })
        .finally(() => {
          this.extentions.loading = false;
        });
    },
    tryToSend() {
      if (!this.isPassed || !this.isTestSuccess) {
        const opts = {
          message: `Error: Test must be passed before creation.`,
        };
        this.showSnackbarError(opts);
        return;
      }
      this.isLoading = true;
      api.servicesProviders
        .create(this.serviceProviderBody)
        .then(() => {
          this.$router.push({ name: "ServicesProviders" });
        })
        .finally(() => {
          this.isLoading = false;
        })
        .catch((err) => {
          this.errorDisplay(err);
        });
    },
    testConfig() {
      this.isTestLoading = true;
      
      if (
        this.serviceProviderBody.type === "ione" &&
        this.serviceProviderBody.secrets.vlans
      ) {
        let isWrongVlans = false;

        for (const value of Object.values(
          this.serviceProviderBody.secrets.vlans
        )) {
          if (value.start && value.size) {
            isWrongVlans = !(value.start + value.size < 4096);
          }
        }

        if (isWrongVlans) {
          this.showSnackbarError({ message: "Vlans cant be more 4096" });
          this.isTestLoading = false;
          return;
        }
      }

      api.servicesProviders
        .testConfig(this.serviceProviderBody)
        .then((res) => {
          if (!res.result) {
            throw res;
          }
          this.testButtonColor = "success";
          this.isTestSuccess = true;
        })
        .catch((err) => {
          this.errorDisplay(err);
        })
        .finally(() => {
          this.isTestLoading = false;
        });
    },
    errorDisplay(err) {
      let opts;
      if (err?.response?.status) {
        if (err.response.status >= 500 || err.response.status < 600) {
          opts = {
            message: `Service Unavailable: ${
              err?.response?.data?.message ?? "Unknown"
            }.`,
            timeout: 0,
          };
        } else {
          opts = {
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          };
        }
      } else {
        opts = {
          message: `Error: ${err?.error ?? "Unknown"}.`,
        };
      }
      this.showSnackbarError(opts);
    },
    addExtention() {
      this.$set(this.extentions.data, this.extentions.selected, {});
      this.extentions.selected = "";
    },
    removeExtention(extention) {
      Vue.delete(this.extentions.data, extention);
    },
    mergeDeep(target, ...sources) {
      return mergeDeep(target, ...sources);
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

// .page__content{
// 	flex-grow: 1;
// 	max-width: 750px;
// }
</style>
