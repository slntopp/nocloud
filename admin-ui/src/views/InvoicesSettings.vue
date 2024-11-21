<template>
  <div class="settings pa-10">
    <v-row>
      <v-col cols="6">
        <div class="d-flex">
          <h2 class="page__title">Invoices settings</h2>
        </div>

        <div class="d-flex align-center">
          <v-text-field
            :loading="isSettingsLoading"
            v-model="newSettings.template"
            label="Template"
          />
          <v-tooltip bottom>
            <template v-slot:activator="{ on, attrs }">
              <v-btn icon v-bind="attrs" v-on="on">
                <v-icon>mdi-help</v-icon>
              </v-btn>
            </template>
            <span
              >Available auto-formatted parameters: {YEAR}, {MONTH}, {DAY},
              {NUMBER}.</span
            >
          </v-tooltip>
        </div>

        <v-text-field
          :loading="isSettingsLoading"
          v-model="newSettings.new_template"
          label="New template"
        />

        <v-text-field
          type="number"
          :loading="isSettingsLoading"
          v-model.number="newSettings.issue_renewal_invoice_after"
          label="Issue renewal invoice after"
        />

        <v-text-field
          :loading="isSettingsLoading"
          v-model.number="newSettings.start_with_number"
          type="number"
          label="Start with number"
        />

        <v-select
          :loading="isSettingsLoading"
          v-model="newSettings.reset_counter_mode"
          :items="resetCounterModes"
          label="Reset counter mode"
        />

        <v-textarea
          :loading="isSettingsLoading"
          v-model.number="newSettings.top_up_item_message"
          label="Top up item message"
        />
      </v-col>
      <v-divider vertical></v-divider>
      <v-col cols="6">
        <div class="d-flex">
          <h2 class="page__title">Preview</h2>
        </div>
        <v-text-field
          readonly
          :loading="isSettingsLoading || isPreviewLoading"
          v-model="settingsPreview.templateExample"
          label="Preview Template"
        />

        <v-text-field
          readonly
          :loading="isSettingsLoading || isPreviewLoading"
          v-model="settingsPreview.newTemplateExample"
          label="Preview New template"
        />

        <v-text-field
          readonly
          :loading="isSettingsLoading || isPreviewLoading"
          v-model="settingsPreview.issueRenewalInvoiceAfterExample"
          label="Preview Issue renewal invoice after"
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
import { computed, onMounted, ref, watch } from "vue";
import api from "../api";
import { debounce } from "@/functions";

const store = useStore();

const newSettings = ref({});
const settingsPreview = ref({});
const isSaveLoading = ref(false);
const isPreviewLoading = ref(false);

const invoicesClient = computed(() => store.getters["invoices/invoicesClient"]);

const resetCounterModes = ["DAILY", "MONTHLY", "YEARLY", "NO RESET"];

onMounted(() => {
  setNewSettings();
});

const settings = computed(() => store.getters["settings/all"]);
const isSettingsLoading = computed(() => store.getters["settings/isLoading"]);

const invoicesSettings = computed(() => {
  const data = settings.value.find((s) => s.key === "billing-invoices");

  return JSON.parse(data?.value || "{}") || {};
});

const setNewSettings = () => {
  newSettings.value = JSON.parse(JSON.stringify(invoicesSettings.value));
};

const setSettingsTemplatePreview = async () => {
  try {
    isPreviewLoading.value = true;

    const data = await invoicesClient.value.getInvoiceSettingsTemplateExample({
      newTemplate: newSettings.value.new_template,
      template: newSettings.value.template,
      issueRenewalInvoiceAfter: newSettings.value.issue_renewal_invoice_after,
    });
    settingsPreview.value = data;
  } finally {
    isPreviewLoading.value = false;
  }
};

const setSettingsTemplatePreviewDebounce = debounce(
  setSettingsTemplatePreview,
  3000
);

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
watch(
  [
    () => newSettings.value.template,
    () => newSettings.value.new_template,
    () => newSettings.value.issue_renewal_invoice_after,
  ],
  () => {
    isPreviewLoading.value = true;
    setSettingsTemplatePreviewDebounce();
  }
);
</script>

<style scoped>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
