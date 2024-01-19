<template>
  <div class="settings pa-10">
    <nocloud-table
      :loading="isLoading"
      sort-by="id"
      item-key="id"
      table-name="plugins-settings"
      :items="localPlugins"
      :headers="headers"
      v-model="selectedPlugins"
    >
      <template v-slot:[`item.url`]="{ item }">
        <v-text-field v-model="item.url"></v-text-field>
      </template>
      <template v-slot:[`item.title`]="{ item }">
        <v-text-field v-model="item.title"></v-text-field>
      </template>
      <template v-slot:[`item.icon`]="{ item }">
        <v-autocomplete :items="icons" v-model="item.icon">
          <template v-slot:prepend>
            <v-icon class="ml-3">{{ `mdi-${item.icon}` }}</v-icon>
          </template>
          <template v-slot:item="{ item }">
            <icon-title-preview :icon="item" :title="item" type="mdi" />
          </template>
        </v-autocomplete>
      </template>
      <template v-slot:[`item.preview`]="{ item }">
        <icon-title-preview :title="item.title" :icon="item.icon" type="mdi" />
      </template>
      <template v-slot:footer.prepend>
        <v-btn @click="addPlugin" class="mx-2">Add</v-btn>
        <v-btn @click="deletePlugins" class="mx-2">Delete</v-btn>
        <v-btn :loading="saveLoading" @click="savePlugins" class="mx-2"
          >Save</v-btn
        >
      </template>
    </nocloud-table>
  </div>
</template>

<script>
import NocloudTable from "@/components/table.vue";
import api from "@/api";
import { mapGetters } from "vuex";
import snackbar from "@/mixins/snackbar";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";

export default {
  name: "PluginsSettings",
  components: { IconTitlePreview, NocloudTable },
  mixins: [snackbar],
  data: () => ({
    headers: [
      { text: "URL", value: "url" },
      { text: "Title", value: "title" },
      { text: "Icon", value: "icon" },
      { text: "Preview", value: "preview" },
    ],
    localPlugins: [],
    icons: [],
    selectedPlugins: [],
    saveLoading: false,
  }),
  computed: {
    ...mapGetters("plugins", { plugins: "all", isLoading: "isLoading" }),
    settings() {
      return this.$store.getters["settings/all"];
    },
  },
  mounted() {
    this.setLocalPlugins();
    this.fetchIcons();

    if (!this.settings.length) {
      this.$store.dispatch("settings/fetch");
    }
  },
  methods: {
    fetchIcons() {
      fetch(
        "https://raw.githubusercontent.com/Templarian/MaterialDesign/master/meta.json",
        { method: "get" }
      )
        .then((d) => d.json())
        .then((data) => {
          this.icons = data.map((icon) => icon.name);
        });
    },
    setLocalPlugins() {
      this.localPlugins = JSON.parse(
        JSON.stringify(
          this.plugins.map((p, index) => ({
            id: index.toString() + Date.now(),
            ...p,
          }))
        )
      );
    },
    addPlugin() {
      this.localPlugins.unshift({
        url: "",
        title: "",
        icon: "",
        id: Date.now(),
      });
    },
    deletePlugins() {
      this.localPlugins = this.localPlugins.filter(
        (p) => this.selectedPlugins.findIndex((sp) => sp.id === p.id) === -1
      );
    },
    setGlobalPlugins() {
      this.$store.commit("plugins/setPlugins", this.localPlugins);
    },
    savePlugins() {
      this.saveLoading = true;
      const key = "plugins";
      const data = {
        ...this.settings.find((s) => s.key === "plugins"),
        value: JSON.stringify(this.localPlugins),
      };
      api.settings
        .addKey(key, data)
        .then(() => {
          this.setGlobalPlugins();
        })
        .catch(() => {
          this.showSnackbarError("Error on save plugins");
        })
        .finally(() => {
          this.saveLoading = false;
        });
    },
  },
  watch: {
    plugins() {
      this.setLocalPlugins();
    },
  },
};
</script>

<style scoped></style>
