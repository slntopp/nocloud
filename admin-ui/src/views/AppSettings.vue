<template>
  <div class="settings pa-10">
    <div class="d-flex justify-space-between align-center">
      <div class="d-flex">
        <v-card-title> Upload settings</v-card-title>
        <v-file-input
          class="file-input"
          label="your settings.."
          accept=".json"
          @change="onFileChange"
        />
      </div>
      <download-template-button
        class="mr-10"
        :template="settings"
        name="settings"
        type="JSON"
      />
    </div>
    <v-card-title class="px-0 mb-3"> Main:</v-card-title>
    <v-row>
      <v-col cols="3">
        <v-text-field
          v-model="settings.whmcs.site_url"
          label="Whmcs site url"
          style="display: inline-block; width: 330px"
        />
      </v-col>
      <v-col cols="3">
        <v-switch
          v-model="settings.dangerModeNoSSLCheck"
          label="Danger mode no SSL check"
        />
      </v-col>
      <v-col cols="3">
        <v-select
          v-model="settings.languages"
          :items="existedLanguages"
          label="Languages"
          item-color="dark"
          multiple
        ></v-select>
      </v-col>
    </v-row>
    <v-card-title class="px-0 mb-3"> App:</v-card-title>
    <v-row>
      <v-col cols="2">
        <v-text-field v-model="settings.app.folder" label="App folder" />
      </v-col>
      <v-col cols="2">
        <v-text-field v-model="settings.app.title" label="App title" />
      </v-col>
      <v-col cols="2">
        <template>
          <v-menu open-on-hover top offset-y>
            <template v-slot:activator="{ on }">
              <v-text-field
                v-on="on"
                v-model="settings.app.logo"
                label="Logo filekey"
              />
            </template>
            <v-card-title style="font-size: medium">
              filename with extension or path from ./img, blank if don't need
            </v-card-title>
          </v-menu>
        </template>
      </v-col>
      <v-col cols="2">
        <v-select
          item-color="dark"
          v-model="settings.app.logo_position"
          :items="positionTypes"
          label="Logo position"
        />
      </v-col>
    </v-row>
    <v-card-title class="px-0 mb-3"> Colors:</v-card-title>
    <v-row>
      <v-col cols="3" v-for="(value, key) in settings.app.colors" :key="key">
        <color-picker v-model="settings.app.colors[key]" :label="key" />
      </v-col>
    </v-row>
    <v-row>
      <v-expansion-panels>
        <v-expansion-panel>
          <v-expansion-panel-header color="background-light"
            >Edit json</v-expansion-panel-header
          >
          <v-expansion-panel-content color="background-light">
            <json-editor :json="settings" @changeValue="edit" />
          </v-expansion-panel-content>
        </v-expansion-panel>
      </v-expansion-panels>
    </v-row>
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import JsonEditor from "@/components/JsonEditor.vue";
import { readJSONFile } from "@/functions.js";
import ColorPicker from "@/components/ui/colorPicker.vue";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";

export default {
  key: "app-settings",
  components: { DownloadTemplateButton, ColorPicker, JsonEditor },
  mixins: [snackbar],
  data() {
    return {
      existedLanguages: ["en", "pl", "ru", "fr", "vi"],
      positionTypes: ["top", "right", "bottom", "left", "relative"],
      settings: {
        whmcs: { site_url: "" },
        app: {
          shared_folder: "virtualHosting",
          title: "",
          logo: "",
          logo_position: "top",
          colors: {
            main: "#fff",
            success: "#fff",
            warn: "#fff",
            err: "#fff",
            gray: "#fff",
            bright_font: "#fff",
            bright_bg: "#fff",
          },
        },
        languages: ["en"],
        dangerModeNoSSLCheck: false,
      },
    };
  },
  methods: {
    onFileChange(file) {
      readJSONFile(file).then((res) => (this.settings = res));
    },
    edit(data) {
      this.settings = data;
    },
  },
};
</script>

<style scoped>
.file-input {
  width: 300px;
}
</style>
