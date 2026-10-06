<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-4">
      <v-spacer />
      <v-btn class="mr-2" @click="openCreate">
        <v-icon left>mdi-plus</v-icon>
        Add package
      </v-btn>
      <v-btn
        color="primary"
        :loading="isSaveLoading"
        :disabled="!isDirty"
        @click="save"
      >
        Save
      </v-btn>
    </div>

    <v-alert v-if="modelsError" type="warning" text dense>
      {{ modelsError }}
    </v-alert>

    <nocloud-table
      :headers="headers"
      :items="rows"
      item-key="key"
      no-hide-uuid
      :show-select="false"
    >
      <template v-slot:[`item.key`]="{ item }">
        {{ item.key }}
        <v-tooltip v-if="item.misplaced" bottom>
          <template v-slot:activator="{ on, attrs }">
            <v-icon small color="warning" v-bind="attrs" v-on="on">
              mdi-alert
            </v-icon>
          </template>
          AI keys are in resources, where the driver does not read them. Edit
          and apply the package to move them to meta.
        </v-tooltip>
      </template>
      <template v-slot:[`item.price`]="{ item }">
        {{ item.price }} {{ currencyCode }}
      </template>
      <template v-slot:[`item.credit`]="{ item }">
        <span v-if="item.isAi">{{ item.credit }} NCU</span>
        <v-chip v-else x-small label>not an AI package</v-chip>
      </template>
      <template v-slot:[`item.models`]="{ item }">
        <span v-if="!item.isAi">—</span>
        <span v-else-if="!item.form.models.length">All models</span>
        <template v-else>
          <v-chip
            v-for="model in item.form.models.slice(0, MODELS_SHOWN)"
            :key="model"
            x-small
            class="mr-1"
            :color="modelsByKey[model] ? undefined : 'warning'"
          >
            {{ modelLabel(model) }}
          </v-chip>
          <v-chip
            v-if="item.form.models.length > MODELS_SHOWN"
            x-small
            outlined
            :title="item.form.models.slice(MODELS_SHOWN).map(modelLabel).join(', ')"
          >
            +{{ item.form.models.length - MODELS_SHOWN }}
          </v-chip>
        </template>
      </template>
      <template v-slot:[`item.public`]="{ item }">
        <v-icon small>{{ item.public ? "mdi-check" : "mdi-minus" }}</v-icon>
      </template>
      <template v-slot:[`item.site`]="{ item }">
        <v-icon small>{{ item.form.site ? "mdi-check" : "mdi-minus" }}</v-icon>
      </template>
      <template v-slot:[`item.actions`]="{ item }">
        <div class="d-flex">
          <v-btn icon small title="Edit" @click="openEdit(item.key)">
            <v-icon small>mdi-pencil</v-icon>
          </v-btn>
          <v-btn icon small title="Duplicate" @click="duplicate(item.key)">
            <v-icon small>mdi-content-copy</v-icon>
          </v-btn>
          <confirm-dialog
            title="Delete this package?"
            text="Instances bought on it keep running but lose their product: renewals and AI credit stop working for them. Save to apply."
            @confirm="remove(item.key)"
          >
            <v-btn icon small title="Delete">
              <v-icon small color="error">mdi-delete</v-icon>
            </v-btn>
          </confirm-dialog>
        </div>
      </template>
    </nocloud-table>

    <v-dialog v-model="isDialogOpen" max-width="760" persistent>
      <v-card v-if="form">
        <v-card-title>
          {{ editingKey ? "Edit package" : "New package" }}
        </v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col cols="6">
              <v-text-field
                v-model="form.key"
                label="Key"
                :disabled="!!editingKey && savedKeys.includes(editingKey)"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field v-model="form.title" label="Title" />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model.number="form.price"
                type="number"
                label="Price"
                :suffix="currencyCode"
              />
            </v-col>
            <v-col cols="6">
              <date-field
                :key="`${dialogKey}-period`"
                :period="form.period"
                :periodKind="form.periodKind"
                @changeDate="form.period = $event"
                @changePeriodKind="form.periodKind = $event"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model.number="form.credit"
                type="number"
                label="Credit"
                suffix="NCU"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field v-model="form.seats" type="number" label="Seats" />
            </v-col>
            <v-col cols="12">
              <v-autocomplete
                v-model="form.models"
                :items="modelItems"
                :loading="isModelsLoading"
                label="Models"
                multiple
                chips
                small-chips
                deletable-chips
              />
            </v-col>
            <v-col cols="6">
              <v-switch v-model="form.public" label="Public" />
            </v-col>
            <v-col cols="6">
              <v-switch v-model="form.site" label="Site constructor only" />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model.number="form.sorter"
                type="number"
                label="Sorter"
              />
            </v-col>
          </v-row>
          <v-alert v-if="formMessage" type="error" text dense class="mt-3 mb-0">
            {{ formMessage }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text @click="closeDialog">Cancel</v-btn>
          <v-btn color="primary" :disabled="!!formMessage" @click="apply">
            Apply
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import DateField from "@/components/date.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import NocloudTable from "@/components/table.vue";
import {
  toForm,
  toProduct,
  formError,
  copyKey,
  isAiPackage,
  hasMisplacedKeys,
} from "./aiPackages";

const props = defineProps(["template"]);
const { template } = toRefs(props);
const store = useStore();

const products = ref(JSON.parse(JSON.stringify(template.value.products || {})));
const savedKeys = Object.keys(template.value.products || {});
const isDirty = ref(false);
const isSaveLoading = ref(false);

const models = ref([]);
const isModelsLoading = ref(false);
const modelsError = ref("");

const isDialogOpen = ref(false);
const dialogKey = ref(0);
const editingKey = ref(null);
const form = ref(null);

const currencyCode = computed(() => store.getters["currencies/default"]?.code || "");

const headers = [
  { text: "Key", value: "key" },
  { text: "Title", value: "title" },
  { text: "Price", value: "price" },
  { text: "Period", value: "period" },
  { text: "Credit", value: "credit" },
  { text: "Seats", value: "seats" },
  { text: "Models", value: "models", sortable: false },
  { text: "Public", value: "public" },
  { text: "Site", value: "site" },
  { text: "Sorter", value: "sorter" },
  { text: "Actions", value: "actions", sortable: false },
];

const rows = computed(() =>
  Object.entries(products.value)
    .map(([key, product]) => {
      const form = toForm(key, product);
      return {
        key,
        title: product.title || key,
        price: form.price,
        period: periodLabel(product),
        credit: form.credit,
        seats: form.seats || "—",
        public: form.public,
        site: form.site,
        sorter: form.sorter,
        form,
        isAi: isAiPackage(product),
        misplaced: hasMisplacedKeys(product),
      };
    })
    .sort((a, b) => a.sorter - b.sorter || a.key.localeCompare(b.key))
);

const modelsByKey = computed(() =>
  Object.fromEntries(models.value.map((model) => [model.key, model]))
);

/** A package's row lists this many of its models and counts the rest. */
const MODELS_SHOWN = 3;

const modelLabel = (key) => {
  const model = modelsByKey.value[key];
  if (!model) return key;
  return `${model.name || key}${model.disabled ? " · disabled" : ""}`;
};

/**
 * The driver's models by provider, for the picker. A disabled model is not offered, but one the
 * package already has stays, marked, and so does a key the driver no longer lists.
 */
const modelItems = computed(() => {
  const chosen = new Set(form.value?.models || []);
  const byProvider = {};
  models.value
    .filter((model) => !model.disabled || chosen.has(model.key))
    .forEach((model) => {
      const provider = model.provider || "other";
      (byProvider[provider] = byProvider[provider] || []).push(model);
    });
  const items = [];
  Object.keys(byProvider)
    .sort()
    .forEach((provider) => {
      items.push({ header: provider });
      byProvider[provider]
        .sort((a, b) => (a.name || a.key).localeCompare(b.name || b.key))
        .forEach((model) =>
          items.push({
            text: `${model.name || model.key} (${model.key})${model.disabled ? " · disabled" : ""}`,
            value: model.key,
          })
        );
    });
  (form.value?.models || [])
    .filter((key) => !modelsByKey.value[key])
    .forEach((key) => items.push({ text: `${key} (not in the driver config)`, value: key }));
  return items;
});

const otherKeys = computed(() =>
  Object.keys(products.value).filter((key) => key !== editingKey.value)
);

const formMessage = computed(() => (form.value ? formError(form.value, otherKeys.value) : ""));



function periodLabel(product) {
  const kinds = {
    CALENDAR_MONTH: "calendar month",
    CALENDAR_QUARTER: "calendar quarter",
    CALENDAR_YEAR: "calendar year",
  };
  if (kinds[product.periodKind]) return kinds[product.periodKind];
  const days = Math.round((Number(product.period) || 0) / 86400);
  if (days && days % 365 === 0) return days === 365 ? "year" : `${days / 365} years`;
  if (days && days % 30 === 0) return days === 30 ? "month" : `${days / 30} months`;
  return days === 1 ? "day" : `${days} days`;
}

const openDialog = (key, value) => {
  editingKey.value = key;
  form.value = value;
  dialogKey.value += 1;
  isDialogOpen.value = true;
};

const openCreate = () => openDialog(null, toForm("", {}));

const openEdit = (key) => openDialog(key, toForm(key, products.value[key]));

const duplicate = (key) => {
  const copy = toForm(copyKey(key, Object.keys(products.value)), products.value[key]);
  copy.title = `${copy.title} (copy)`;
  openDialog(null, copy);
};

const closeDialog = () => {
  isDialogOpen.value = false;
  form.value = null;
  editingKey.value = null;
};

const apply = () => {
  if (formMessage.value) return;
  const original = editingKey.value ? products.value[editingKey.value] : {};
  const next = { ...products.value };
  if (editingKey.value && editingKey.value !== form.value.key) {
    delete next[editingKey.value];
  }
  next[form.value.key] = toProduct(form.value, original);
  products.value = next;
  isDirty.value = true;
  closeDialog();
};

const remove = (key) => {
  const next = { ...products.value };
  delete next[key];
  products.value = next;
  isDirty.value = true;
};

const save = async () => {
  isSaveLoading.value = true;
  try {
    await api.plans.update(template.value.uuid, {
      ...template.value,
      products: products.value,
    });
    isDirty.value = false;
    store.commit("snackbar/showSnackbarSuccess", {
      message: "AI packages saved",
    });
    store.dispatch("reloadBtn/onclick");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save AI packages",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

onMounted(async () => {
  isModelsLoading.value = true;
  try {
    const response = await api.get("/api/openai/get_config");
    models.value = Object.entries(response.cfg?.models || {}).map(([key, model]) => ({
      key,
      name: model.name,
      provider: model.provider,
      disabled: model.disabled,
    }));
  } catch (e) {
    modelsError.value =
      "Could not load the driver's models; model keys can still be kept as they are.";
  } finally {
    isModelsLoading.value = false;
  }
});
</script>

<script>
export default { name: "ai-packages" };
</script>

