<template>
  <div class="settings pa-10">
    <div class="d-flex mb-5">
      <h1 class="page__title">Invoices settings</h1>
    </div>

    <v-row>
      <v-col cols="6">
        <v-text-field
          :loading="isSettingsLoading"
          v-model="newSettings.template"
          label="template"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          :loading="isSettingsLoading"
          v-model="newSettings.new_template"
          label="new_template"
        />
      </v-col>

      <v-col cols="6">
        <v-text-field
          :loading="isSettingsLoading"
          v-model.number="newSettings.start_with_number"
          type="number"
          label="start_with_number"
        />
      </v-col>
      <v-col cols="6">
        <v-text-field
          :loading="isSettingsLoading"
          v-model="newSettings.reset_counter_mode"
          label="reset_counter_mode"
        />
      </v-col>
    </v-row>

    <v-row justify="end" class="mx-5">
      <v-btn
        @click="resetSettings"
        :disabled="isSettingsLoading"
        color="error"
        class="mr-3"
      >
        Reset
      </v-btn>
      <v-btn
        @click="saveSettings"
        :loading="isSaveLoading"
        :disabled="isSettingsLoading"
        color="primary"
      >
        Save
      </v-btn>
    </v-row>
  </div>
</template>

<script setup>
import { useStore } from "@/store";
import { computed, ref, watch } from "vue";
import api from "../api";

const store = useStore();

const newSettings = ref({});
const isSaveLoading = ref(false);

const settings = computed(() => store.getters["settings/all"]);
const isSettingsLoading = computed(() => store.getters["settings/isLoading"]);

const invoicesSettings = computed(() => {
  const data = settings.value.find((s) => s.key === "billing-invoices");

  return JSON.parse(data?.value || "{}") || {};
});

const setNewSettings = () => {
  newSettings.value = JSON.parse(JSON.stringify(invoicesSettings.value));
};

const saveSettings = async () => {
  isSaveLoading.value = true;
  const key = "billing-invoices";
  const data = {
    ...settings.value.find((s) => s.key === key),
    value: JSON.stringify(newSettings.value),
  };

  try {
    await api.settings.addKey(key, data);
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error on save invoices settings",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const resetSettings = () => {
  newSettings.value = invoicesSettings.value;
};

watch(invoicesSettings, setNewSettings);
</script>

<style scoped>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
