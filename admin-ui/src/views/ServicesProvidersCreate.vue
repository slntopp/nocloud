<template>
  <div class="servicesProviders-create pa-4">
    <div class="page__title">
      {{ $route.params.uuid ? "Edit" : "Create" }} service provider
    </div>
    <v-container>
      <v-row>
        <v-col lg="6" cols="12">
          <v-row class="align-center">
            <v-col cols="3">
              <v-list-subheader> Provider type </v-list-subheader>
            </v-col>

            <v-col cols="9">
              <v-autocomplete
                v-model="provider.type"
                :items="types"
                label="Type"
              />
              <v-text-field
                ref="type"
                label="Type name"
                v-if="provider.type === 'custom'"
                v-model="customTitle"
                :rules="[(v) => !!v || 'This field is required!']"
              />
            </v-col>
          </v-row>

          <v-row class="align-center">
            <v-col cols="3">
              <v-list-subheader> Provider title </v-list-subheader>
            </v-col>

            <v-col cols="9">
              <v-text-field
                v-model="provider.title"
                label="Title"
              ></v-text-field>
            </v-col>
          </v-row>

          <v-row
            v-if="!isProxyHide"
            :class="{ 'align-center': !provider.vars.console?.value }"
          >
            <v-col cols="3">
              <v-list-subheader> Proxy </v-list-subheader>
            </v-col>

            <v-col cols="9">
              <v-text-field v-model="provider.proxy.socket" label="Socket" />
              <v-select
                label="Console"
                v-if="provider.vars.console?.value"
                v-model="provider.vars.console.value"
                :items="consoleTypes"
              />
            </v-col>
          </v-row>

          <v-expansion-panels>
            <v-expansion-panel>
              <v-expansion-panel-title color="background-light">
                <span class="text-h6">Meta</span>
                <template v-slot:actions>
                  <v-icon size="x-large"> $expand </v-icon>
                </template>
              </v-expansion-panel-title>
              <v-expansion-panel-text color="background-light">
                <json-editor
                  :json="provider.meta"
                  @changeValue="(data) => (provider.meta = data)"
                />
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>

          <v-row class="align-center">
            <v-col cols="3">
              <v-list-subheader> Public </v-list-subheader>
            </v-col>

            <v-col cols="9">
              <v-switch v-model="provider.public" />
            </v-col>
          </v-row>

          <v-divider></v-divider>

          <component
            :is="templates[provider.type]"
            :secrets="provider.secrets"
            :key="providerKey"
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
            bg-color="background-light"
            class="mb-2"
          >
            <v-tab>Extentions</v-tab>
          </v-tabs>

          <v-window v-model="tabs" color="primary">
            <v-window-item>
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
            </v-window-item>
          </v-window>
        </v-col>
      </v-row>

      <v-row class="justify-end">
        <v-col cols="6">
          <v-btn class="mr-2" :loading="isLoading" @click="tryToSend">
            Save
          </v-btn>
        </v-col>
        <v-col cols="6">
          <div class="d-flex align-start justify-center">
            <download-template-button
              :type="isJson ? 'JSON' : 'YAML'"
              class="mr-2"
              :template="serviceProviderBody"
              :name="downloadedFileName"
            />
            <v-switch
              class="mr-2"
              style="margin-top: 5px; padding-top: 5px"
              v-model="isJson"
              :label="!isJson ? 'YAML' : 'JSON'"
            />
            <v-file-input
              class="file-input"
              :label="`upload ${isJson ? 'json' : 'yaml'} sp...`"
              :accept="isJson ? '.json' : '.yaml'"
              @update:model-value="onJsonInputChange"
            />
          </div>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script>
import { defineAsyncComponent } from "vue";
import { moduleComponent, moduleTypes } from "@/modules";
import api from "@/api.js";
import extentionsMap from "@/components/extentions/map.js";
import snackbar from "@/mixins/snackbar.js";
import JsonEditor from "@/components/JsonEditor.vue";

import {
  mergeDeep,
  readJSONFile,
  readYAMLFile,
  toKebabCase,
  toPascalCase,
} from "@/functions.js";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";

export default {
  name: "servicesProviders-create",
  components: { DownloadTemplateButton, IconTitlePreview, JsonEditor },
  mixins: [snackbar],
  data: () => ({
    types: [],
    templates: {},
    customTitle: "",
    provider: {
      type: "custom",
      title: "",
      public: true,
      proxy: { socket: "" },
      secrets: {},
      vars: { console: {} },
      meta: {},
    },

    originalSecrets: {},
    providerKey: "",

    isPassed: false,
    isLoading: false,

    tabs: null,
    extentions: {
      loading: false,
      items: [],
      data: {},
      selected: "",
    },
    consoleTypes: [
      { text: "VNC", value: { default: "vnc" } },
      { text: "VMRC", value: { default: "vmrc" } },
    ],

    tooltipVisible: false,

    isJson: true,
  }),
  created() {
    const id = this.$route.params.uuid;
    moduleTypes("serviceProviders").forEach((type) => {
      this.types.push(type);
      this.templates[type] = moduleComponent("serviceProviders", type);
    });
    this.types.push("bitrix24");

    this.providerKey = this.generateComponentId();
    this.fetchExtentions();

    if (id) {
      this.$store.dispatch("servicesProviders/fetchById", id).then((res) => {
        if (!this.types.includes(res.type)) {
          this.customTitle = res.type;
          res.type = "custom";
        }
        this.originalSecrets = JSON.parse(JSON.stringify(res.secrets));

        this.provider = res;
      });
    }
  },
  computed: {
    template() {
      return defineAsyncComponent(() =>
        import(`@/components/modules/${this.type}/serviceProviders.vue`));
    },
    extentionsMap() {
      return extentionsMap;
    },
    serviceProviderBody() {
      if (Object.keys(this.extentions.data).length > 0) {
        return { ...this.provider, extentions: this.extentions.data };
      } else {
        return this.provider;
      }
    },
    isProxyHide() {
      return ["cpanel", "proxmox"].includes(this.serviceProviderBody.type);
    },
    downloadedFileName() {
      return this.serviceProviderBody.title
        ? this.serviceProviderBody.title.replaceAll(" ", "_")
        : "unknown_sp";
    },
  },
  methods: {
    generateComponentId() {
      return "id" + Math.random().toString(16).slice(2);
    },
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
    async tryToSend() {
      const action = this.$route.params.uuid ? "edit" : "create";

      this.serviceProviderBody.secrets = {
        ...this.originalSecrets,
        ...this.serviceProviderBody.secrets,
      };

      if (
        !this.isPassed ||
        (this.customTitle === "" && this.provider.type === "custom")
      ) {
        const opts = {
          message: `Error: Test must be passed before creation.`,
        };
        this.showSnackbarError(opts);
        this.$refs.type.focus();
        return;
      }
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
          return;
        }
      }

      if (action === "create") delete this.provider.uuid;
      if (this.provider.type === "custom") {
        this.provider.type = this.customTitle;
      }

      this.isLoading = true;
      try {
        const id = this.$route.params.uuid;
        const promise =
          action === "create"
            ? api.servicesProviders.create(this.serviceProviderBody)
            : api.servicesProviders.update(id, this.serviceProviderBody);
        const response = await promise;
        if (response.errors) {
          throw response;
        }
        this.showSnackbarSuccess({
          message:
            action === "create"
              ? "Service provider created successfully"
              : "Service provider updated successfully",
        });
        if (action === "create") {
          this.$router.push({ name: "ServicesProviders" });
        }
      } catch (err) {
        this.errorDisplay(err);
      } finally {
        this.isLoading = false;
      }
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
      this.extentions.data[this.extentions.selected] = {};
      this.extentions.selected = "";
    },
    removeExtention(extention) {
      delete this.extentions.data[extention];
    },
    mergeDeep(target, ...sources) {
      return mergeDeep(target, ...sources);
    },
    setSP(res) {
      const requiredKeys = ["vars", "secrets", "title", "public", "type"];

      for (const key of requiredKeys) {
        if (res[key] === undefined) {
          throw new Error("JSON need keys:" + requiredKeys.join(", "));
        }
      }

      if (!this.types.includes(res.type)) {
        throw new Error(`Type ${res.type} not exists!`);
      }

      this.providerKey = this.generateComponentId();

      this.provider = res;
    },
    onJsonInputChange(file) {
      if (this.isJson) {
        readJSONFile(file)
          .then((res) => this.setSP(res))
          .catch(({ message }) => {
            this.showSnackbarError({ message });
          });
      } else {
        readYAMLFile(file)
          .then((res) => this.setSP(res))
          .catch(({ message }) => {
            this.showSnackbarError({ message });
          });
      }
    },
    setIconToKebabCase(icon) {
      this.provider.meta.service.icon = toKebabCase(icon);
    },
    toPascalCase,
  },
};
</script>

<style scoped lang="scss">
.page__title {
  color: rgb(var(--v-theme-primary));
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}

.file-input {
  max-width: 200px;
  min-width: 200px;
  margin-top: 0;
  padding-top: 0;
}

// .page__content{
// 	flex-grow: 1;
// 	max-width: 750px;
// }
</style>
