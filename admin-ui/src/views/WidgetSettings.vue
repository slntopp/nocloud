<template>
  <div class="settings pa-10">
    <v-row justify="center">
      <v-card-title>Map settings</v-card-title>
    </v-row>
    <v-card-title>Map widget script:</v-card-title>
    <v-text-field
      filled
      color="background-light"
      :value="script"
      readonly
      append-icon="mdi-content-copy"
      @click:append="copyScript"
    />
    <div ref="script" id="script"></div>
    <div
      ref="preview-widget"
      id="preview-widget"
      style="width: 1000px; height: 800px; margin: 50px auto"
    ></div>
  </div>
</template>

<script>
import api from "@/api";
export default {
  data: () => ({
    script:
      "https://cdn.jsdelivr.net/npm/nocloud-map-widget@latest/dist/map-widget.js",
  }),
  mounted() {
    let script = document.createElement("script");
    script.setAttribute("src", this.script);
    script.setAttribute("type", "module");
    this.$refs.script.appendChild(script);
    script.onload = () => {
      const apiURL = api.axios.defaults.baseURL;
      // eslint-disable-next-line
      const previewWidget = new SupportMap({
        container: document.getElementById("preview-widget"),
        apiPort: 8624,
        apiURL,
        appURL: apiURL.replace("api", "app") + "/#/",
      });

      previewWidget.init();
    };
  },
  methods: {
    copyScript() {
      navigator.clipboard.writeText(this.script);
    },
  },
  computed: {},
};
</script>
