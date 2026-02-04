<template>
  <div class="pa-4 h-100">


    <div class="d-flex justify-space-between pb-2 mt-4 flex-wrap">
      <div class="d-flex flex-wrap align-center">
        <v-btn color="background-light" class="mr-2 mb-2 action-btn" to="/settings/app">app settings</v-btn>
        <v-btn color="background-light" class="mr-2 mb-2 action-btn" to="/settings/widget">widget settings</v-btn>
        <v-btn color="background-light" class="mr-2 mb-2 action-btn" to="/settings/plugins">plugins settings</v-btn>
        <v-btn color="background-light" class="mr-2 mb-2 action-btn" to="/settings/invoices">invoice settings</v-btn>

        <v-dialog style="height: 100%">
          <template v-slot:activator="{ on, attrs }">
            <v-btn color="background-light" class="mr-2 mb-2 action-btn" v-on="on" v-bind="attrs">
              chats settings
            </v-btn>
          </template>
          <plugin-iframe
            style="height: 80vh"
            url="/cc.ui/"
            :params="{ redirect: 'settings' }"
          />
        </v-dialog>

        <v-btn
          color="background-light"
          class="mr-2 mb-2 action-btn"
          :to="{ name: 'SettingsItem', params: { key: 'create' } }"
        >
          create
        </v-btn>
      </div>

      <confirm-dialog :disabled="selected.length < 1" @confirm="deleteSelected">
        <v-btn color="background-light" :disabled="selected.length < 1" class="mb-2 action-btn">
          delete
        </v-btn>
      </confirm-dialog>
    </div>

    <nocloud-table
      v-model="selected"
      table-name="settings"
      item-key="key"
      show-select
      :headers="headers"
      :items="filteredSettings"
      :loading="loading"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.key`]="{ item }">
        <div class="d-flex align-center">
          <router-link
            class="setting-link text-truncate mr-1"
            :to="{ name: 'SettingsItem', params: { key: item.key } }"
          >
            {{ item.key }}
          </router-link>

          <v-btn icon @click.stop="copyKey(item.key)" class="ml-1">
            <v-icon>mdi-content-copy</v-icon>
          </v-btn>
        </div>
      </template>

      <template v-slot:[`item.description`]="{ item }">
        {{ getShortName(item.description, 15) }}
      </template>

      <template v-slot:[`item.value`]="{ item }">
        {{ getShortName(item.value, 15) }}
      </template>
    </nocloud-table>

    <div class="widgets align-start mt-8 d-flex flex-wrap">
      <component
        v-for="widget in widgetComponents"
        :key="widget"
        :is="widget"
        class="mx-2 mb-4"
        style="width: 30%; min-width: 300px;"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useStore } from "@/store";
import api from "@/api.js";
import { addToClipboard, getShortName, filterArrayIncludes } from "@/functions";

import noCloudTable from "@/components/table.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import PluginIframe from "@/components/plugin/iframe.vue";
import ServicesWidget from "@/components/widgets/services";
import HealthWidget from "@/components/widgets/health";
import RoutinesWidget from "@/components/widgets/routines";

const store = useStore();

const selected = ref([]);
const fetchError = ref("");
const widgetComponents = ["ServicesWidget", "HealthWidget", "RoutinesWidget"];

const headers = [
  { text: "Key", value: "key" },
  { text: "Description", value: "description" },
  { text: "Value", value: "value" },
];

const settings = computed(() => store.getters["settings/all"]);
const loading = computed(() => store.getters["settings/isLoading"]);
const searchParam = computed(() => store.getters["appSearch/param"]);

const filteredSettings = computed(() => {
  if (searchParam.value) {
    return filterArrayIncludes(settings.value, {
      keys: ["key", "description", "value"],
      value: searchParam.value,
    });
  }
  return settings.value;
});

const fetchSettings = () => {
  store.dispatch("settings/fetch").catch((err) => {
    fetchError.value = err.response?.data?.message || "Error";
  });
};

const copyKey = (val) => addToClipboard(val);

const deleteSelected = async () => {
  try {
    const promises = selected.value.map((s) => api.settings.delete(s.key));
    await Promise.all(promises);
    selected.value = [];
    fetchSettings();
  } catch (err) {
    store.commit("snackbar/showSnackbarError", { message: "Delete failed" });
  }
};

onMounted(() => {
  fetchSettings();
  store.commit("reloadBtn/setCallback", { event: fetchSettings });
});
</script>

<script>
export default {
  name: "SettingsView",
  components: {
    "nocloud-table": noCloudTable,
    ConfirmDialog,
    PluginIframe,
    ServicesWidget,
    HealthWidget,
    RoutinesWidget,
  },
};
</script>

<style scoped lang="scss">

.page__title {
  color: #5171f1;
  font-weight: 300;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  text-transform: uppercase;
}

.action-btn {
  text-transform: uppercase;
  font-weight: 500;
}

.setting-link {
  color: var(--v-primary-base);
  text-decoration: none;
  font-weight: 500;
  &:hover {
    text-decoration: underline;
  }
}
</style>