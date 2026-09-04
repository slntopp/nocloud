<template>
  <div class="pa-5">
    <v-expansion-panels>
      <v-expansion-panel>
        <v-expansion-panel-header color="background">
          Margin rules:
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background">
          <plan-opensrs
            :fee="fee"
            :isEdit="true"
            @changeFee="changeFee"
            @onValid="(data) => (isValid = data)"
          />
          <confirm-dialog
            text="This will apply the rules markup parameters to all prices"
            @confirm="setFee"
          >
            <v-btn class="mt-4" color="secondary">Set rules</v-btn>
          </confirm-dialog>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
    <v-tabs
      class="rounded-t-lg"
      v-model="tabsIndex"
      background-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab">{{ tab }}</v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabsIndex"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item v-for="tab in tabs" :key="tab">
        <nocloud-table
          v-if="tab === 'Old prices'"
          :headers="oldPricesHeaders"
          :items="oldPricesResources"
          :show-select="false"
        >
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field type="number" dense v-model.number="item.price" />
          </template>
        </nocloud-table>

        <div v-else-if="tab === 'Prices'">
          <div class="toolbar">
            <v-text-field
              class="toolbar__search"
              dense
              outlined
              hide-details
              clearable
              prepend-inner-icon="mdi-magnify"
              v-model="searchParam"
              label="Search"
            />
            <v-select
              class="toolbar__select"
              dense
              outlined
              hide-details
              multiple
              clearable
              label="Types"
              :items="availableTypes"
              v-model="typesFilter"
            />
            <v-select
              class="toolbar__select"
              dense
              outlined
              hide-details
              multiple
              clearable
              label="Providers"
              :items="availableProviders"
              v-model="providersFilter"
            />
            <v-switch
              class="toolbar__switch"
              dense
              hide-details
              label="Only enabled"
              v-model="onlyEnabled"
            />
            <v-switch
              class="toolbar__switch"
              dense
              hide-details
              label="Hide broken"
              v-model="hideBroken"
            />
            <span class="toolbar__counter"
              >{{ newPricesResourcesFiltred.length }} /
              {{ newPricesResources.length }}</span
            >
            <v-btn small outlined :loading="isFillLoading" @click="fillConfig"
              >Fill config</v-btn
            >
            <v-btn small outlined :loading="isExportLoading" @click="exportToXlsx"
              >Export to XLSX</v-btn
            >
          </div>
          <nocloud-table
            :headers="newPricesHeaders"
            :items="newPricesResourcesFiltred"
            :show-select="false"
          >
            <template v-slot:[`item.name`]="{ item }">
              <v-text-field
                class="cell-input"
                dense
                hide-details
                v-model="item.name"
              />
            </template>

            <template v-slot:[`item.disabled`]="{ item }">
              <v-switch
                class="cell-switch"
                :key="item.key"
                dense
                hide-details
                :input-value="!item.disabled"
                @change="changeDisabled(item, $event)"
              />
            </template>

            <template v-slot:[`item.key`]="{ item }">
              <v-tooltip top :disabled="item.key === getShortName(item.key)">
                <template v-slot:activator="{ on, attrs }">
                  <span v-bind="attrs" v-on="on">{{
                    getShortName(item.key)
                  }}</span>
                </template>
                <span>{{ item.key }}</span>
              </v-tooltip>
            </template>

            <template v-slot:[`item.visibility`]="{ item }">
              <v-select
                class="cell-input"
                dense
                hide-details
                :items="['api_only', 'public', 'private']"
                v-model="item.visibility"
              />
            </template>

            <template v-slot:[`item.billing`]="{ item }">
              <v-btn icon small @click="openBillingSettings(item)">
                <v-icon small>mdi-tune-variant</v-icon>
              </v-btn>
            </template>

            <template v-slot:[`item.types`]="{ item }">
              <template v-if="item.types?.length">
                <v-chip
                  v-for="type in item.types"
                  :key="type"
                  class="mr-1 my-1"
                  style="color: white"
                  x-small
                  :color="typesColorMap[type]"
                  >{{ type }}</v-chip
                >
              </template>
              <span v-else>None</span>
            </template>

            <template v-slot:[`item.state.state`]="{ item }">
              <v-tooltip left :disabled="!item.state?.error_messages?.length">
                <template v-slot:activator="{ on, attrs }">
                  <v-chip
                    x-small
                    v-bind="attrs"
                    v-on="on"
                    style="color: white"
                    :color="stateColor(item)"
                    >{{ item.state?.state || "unknown" }}</v-chip
                  >
                </template>
                <div
                  v-for="message in item.state?.error_messages || []"
                  :key="message"
                >
                  {{ message }}
                </div>
              </v-tooltip>
            </template>

            <template v-slot:[`item.provider`]="{ item }">
              <v-chip x-small style="color: white" color="blue">{{
                item.provider
              }}</v-chip>
            </template>
          </nocloud-table>

          <v-dialog max-width="840px" scrollable v-model="isBillingSettingsOpen">
            <v-card class="billing-card" color="background-light">
              <div class="billing-card__head">
                <div class="billing-card__title">
                  <div class="billing-card__name">
                    {{ currentBillingSettings.name }}
                  </div>
                  <div class="billing-card__key">
                    {{ currentBillingSettings.key }}
                  </div>
                </div>
                <v-select
                  class="billing-card__types"
                  dense
                  outlined
                  hide-details
                  multiple
                  label="Types"
                  :items="availableTypes"
                  v-model="currentBillingSettings.types"
                />
              </div>

              <v-card-text class="billing-card__body">
                <div v-if="!Object.keys(fieldsForAdd).length" class="hint">
                  Select at least one type to configure prices.
                </div>

                <div v-else class="price-grid">
                  <div class="price-grid__head"></div>
                  <div
                    class="price-grid__head price-grid__head--num"
                    v-for="{ key: priceKey, label } in priceKeys"
                    :key="priceKey"
                  >
                    {{ label }}
                  </div>

                  <template v-for="group in Object.keys(fieldsForAdd)">
                    <div class="price-grid__group" :key="group">
                      {{ keyLabelMap[group] || group }}
                    </div>

                    <template v-for="field in fieldsForAdd[group]">
                      <template v-if="field.type === 'number'">
                        <div class="price-grid__label" :key="field.subkey">
                          <span>{{ fieldLabel(field) }}</span>
                          <v-tooltip right max-width="320">
                            <template v-slot:activator="{ on, attrs }">
                              <v-icon
                                class="price-grid__hint"
                                small
                                v-bind="attrs"
                                v-on="on"
                                >mdi-help-circle-outline</v-icon
                              >
                            </template>
                            <span>{{ fieldHint(field) }}</span>
                          </v-tooltip>
                        </div>
                        <v-text-field
                          v-for="{ key: priceKey } in priceKeys"
                          :key="`${field.subkey}-${priceKey}`"
                          class="price-field"
                          dense
                          outlined
                          hide-details
                          type="number"
                          :suffix="defaultCurrency?.code"
                          :readonly="priceKey === 'amount'"
                          :filled="priceKey === 'amount'"
                          :value="priceOf(field)[priceKey]"
                          @input="setPrice(priceOf(field), $event)"
                        />
                      </template>

                      <template v-else>
                        <div class="price-grid__group" :key="field.subkey">
                          {{ fieldLabel(field) }}
                          <v-tooltip right max-width="320">
                            <template v-slot:activator="{ on, attrs }">
                              <v-icon
                                class="price-grid__hint"
                                small
                                v-bind="attrs"
                                v-on="on"
                                >mdi-help-circle-outline</v-icon
                              >
                            </template>
                            <span>{{ fieldHint(field) }}</span>
                          </v-tooltip>
                        </div>

                        <template
                          v-for="key in Object.keys(
                            currentBillingSettings.billing[field.key][
                              field.subkey
                            ]
                          )"
                        >
                          <template
                            v-for="subkey in Object.keys(
                              currentBillingSettings.billing[field.key][
                                field.subkey
                              ][key]
                            )"
                          >
                            <div
                              class="price-grid__label"
                              :key="`${key}-${subkey}`"
                            >
                              <v-btn
                                icon
                                x-small
                                @click="deleteFromMap(field, key, subkey)"
                                ><v-icon small>mdi-close</v-icon></v-btn
                              >
                              <span>{{ key }} / {{ subkey }}</span>
                            </div>
                            <v-text-field
                              v-for="{ key: priceKey } in priceKeys"
                              :key="`${key}-${subkey}-${priceKey}`"
                              class="price-field"
                              dense
                              outlined
                              hide-details
                              type="number"
                              :suffix="defaultCurrency?.code"
                              :readonly="priceKey === 'amount'"
                              :filled="priceKey === 'amount'"
                              :value="
                                currentBillingSettings.billing[field.key][
                                  field.subkey
                                ][key][subkey][priceKey]
                              "
                              @input="
                                setPrice(
                                  currentBillingSettings.billing[field.key][
                                    field.subkey
                                  ][key][subkey],
                                  $event
                                )
                              "
                            />
                          </template>
                        </template>

                        <div class="price-grid__add" :key="`add-${field.subkey}`">
                          <v-text-field
                            dense
                            outlined
                            hide-details
                            :label="mapKeyLabels(field)[0]"
                            v-model="newKeysForMaps[field.subkey]"
                          />
                          <v-text-field
                            dense
                            outlined
                            hide-details
                            :label="mapKeyLabels(field)[1]"
                            v-model="newSubkeysForMaps[field.subkey]"
                          />
                          <v-btn
                            small
                            outlined
                            :disabled="
                              !newKeysForMaps[field.subkey] ||
                              !newSubkeysForMaps[field.subkey] ||
                              isSaveModelLoading
                            "
                            @click="addToMap(field)"
                            >Add</v-btn
                          >
                        </div>
                      </template>
                    </template>
                  </template>
                </div>
              </v-card-text>

              <v-card-actions class="billing-card__actions">
                <v-alert
                  v-if="billingSettinfsMessages.length"
                  class="billing-card__errors"
                  dense
                  text
                  type="error"
                >
                  {{ billingSettinfsMessages.join(", ") }}
                </v-alert>
                <v-spacer />
                <v-btn
                  small
                  text
                  :disabled="isSaveModelLoading"
                  @click="isBillingSettingsOpen = false"
                  >Close</v-btn
                >
                <v-btn
                  small
                  color="primary"
                  :loading="isSaveModelLoading"
                  @click="saveBillingSettings"
                  >Save changes</v-btn
                >
              </v-card-actions>
            </v-card>
          </v-dialog>
        </div>

        <div class="os-tab__card" v-else>
          <plan-addons-table
            @change:addons="addons = $event"
            :addons="template.addons"
          />
        </div>
      </v-tab-item>
    </v-tabs-items>
    <div class="d-flex justify-end">
      <v-btn :loading="isSaveLoading" @click="save">Save</v-btn>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import NocloudTable from "@/components/table.vue";
import planAddonsTable from "@/components/planAddonsTable.vue";
import api from "@/api";
import { useStore } from "@/store";
import planOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import { getMarginedValue, getShortName } from "@/functions";
import confirmDialog from "@/components/confirmDialog.vue";
import XlsxService from "@/services/XlsxService";
import { getTodayFullDate } from "../../functions";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const tabs = ref(["Prices", "Old prices", "Addons"]);
const tabsIndex = ref(0);

const isFillLoading = ref(false);
const isExportLoading = ref(false);

const isValid = ref(false);
const fee = ref(template.value.fee || {});

const addons = ref([]);

const oldPricesResources = ref([
  {
    key: "input_kilotoken",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Input kilotoken",
    public: true,
  },
  {
    key: "output_kilotoken",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Output kilotoken",
    public: true,
  },
  {
    key: "image_size_1024x1024_quality_standard",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1024",
    public: true,
  },
  {
    key: "image_size_1024x1024_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1024 HD",
    public: true,
  },
  {
    key: "image_size_1024x1792_quality_standard",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024",
    public: true,
  },
  {
    key: "image_size_1024x1792_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024 HD",
    public: true,
  },
  {
    key: "image_size_1792x1024_quality_standard",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024",
    public: true,
  },
  {
    key: "image_size_1792x1024_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024 HD",
    public: true,
  },
]);
const newPricesResources = ref([]);

const oldPricesHeaders = [
  { text: "Key", value: "key" },
  { text: "Title", value: "title" },
  {
    text: "Price",
    value: "price",
    width: 200,
  },
];

const newPricesHeaders = [
  { text: "Key", value: "key", width: 150 },
  { text: "Name", value: "name" },
  { text: "Provider", value: "provider", width: 100 },
  { text: "Types", value: "types", sortable: false },
  { text: "Visibility", value: "visibility", width: 150 },
  { text: "Enabled", value: "disabled", width: 100 },
  { text: "Billing", value: "billing", sortable: false, width: 100 },
  { text: "State", value: "state.state", width: 100 },
];

const priceKeys = [
  { key: "raw_amount", label: "Supplier price" },
  { key: "amount", label: "Final price" },
];

const searchParam = ref("");
const onlyEnabled = ref(true);
const hideBroken = ref(true);
const typesFilter = ref([]);
const providersFilter = ref([]);

const isSaveLoading = ref(false);
const isSaveModelLoading = ref(false);
const isBillingSettingsOpen = ref(false);
const billingSettinfsMessages = ref([]);
const currentBillingSettings = ref({});
const newKeysForMaps = ref({});
const newSubkeysForMaps = ref({});
const currentSerial = ref();

const availableTypes = [
  "text",
  "text_to_audio",
  "audio_to_text",
  "image",
  "video",
  "vision",
];
const typesColorMap = {
  text: "indigo darken-4",
  text_to_audio: "purple",
  audio_to_text: "pink darken-4",
  image: "red darken-4",
  video: "orange darken-3",
  vision: "blue darken-3",
};

const keyLabelMap = {
  tokens: "Tokens",
  other: "Other",
  media_duration: "Media",
  images: "Images",
};

const fieldMeta = {
  "tokens.text_input": {
    label: "Text input, per 1M tokens",
    hint: "Price for 1 000 000 tokens sent to the model (prompt).",
  },
  "tokens.text_output": {
    label: "Text output, per 1M tokens",
    hint: "Price for 1 000 000 tokens generated by the model (answer).",
  },
  "tokens.image_input": {
    label: "Image input, per 1M tokens",
    hint: "Price for 1 000 000 tokens of images sent to the model.",
  },
  "tokens.image_output": {
    label: "Image output, per 1M tokens",
    hint: "Price for 1 000 000 tokens of images generated by the model.",
  },
  "media_duration.duration_price": {
    label: "Media, per 60 seconds",
    hint: "Price for 60 seconds of audio or video. Charged proportionally to the real duration.",
  },
  "other.web_search_price": {
    label: "Web search, per 1000 requests",
    hint: "Price for 1000 web searches performed by the model.",
  },
  "other.sampling_step_price": {
    label: "Generation step",
    hint: "Price for one generation (sampling) step.",
  },
  "other.characters_price": {
    label: "Characters, per 1M",
    hint: "Price for 1 000 000 input characters (used by TTS models).",
  },
  "other.pages_count_price": {
    label: "Pages, per 1000",
    hint: "Price for 1000 recognized pages.",
  },
  "images.res_to_quality": {
    label: "Price per image (resolution / quality)",
    hint: "Price for one generated image for every resolution and quality pair, e.g. 1024x1024 / standard.",
  },
};

const fieldPath = (field) => `${field.key}.${field.subkey}`;
const fieldLabel = (field) =>
  fieldMeta[fieldPath(field)]?.label || field.subkey;
const fieldHint = (field) =>
  fieldMeta[fieldPath(field)]?.hint || "No description";
const mapKeyLabels = (field) =>
  field.subkey === "res_to_quality"
    ? ["New resolution", "New quality"]
    : ["New key", "New subkey"];

const priceOf = (field) =>
  currentBillingSettings.value.billing[field.key][field.subkey].price;

const setPrice = (price, value) => {
  const raw = +value || 0;

  price.raw_amount = raw;
  price.amount = getMarginedValue(fee.value, raw);
};

onMounted(async () => {
  oldPricesResources.value = oldPricesResources.value.map((resource) => {
    const realResource = template.value.resources.find(
      (realResource) => realResource.key === resource.key
    );

    return { ...resource, price: realResource?.price || 0 };
  });

  try {
    const response = await api.get("/api/openai/get_config");

    Object.keys(response.cfg.models).forEach((key) => {
      newPricesResources.value.push({ ...response.cfg.models[key], key });
    });
    currentSerial.value = response.serial;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch keyweb prices",
    });
  }
  addons.value = template.value.addons;
});

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const changeFee = (value) => {
  fee.value = JSON.parse(JSON.stringify(value));
};

const fieldsForTypes = {
  text: {
    type: "default",
    fields: [
      {
        "tokens.text_output": "number",
        "tokens.text_input": "number",
        "other.web_search_price": "number",
      },
    ],
  },
  text_to_audio: {
    type: "default",
    fields: [
      {
        "tokens.text_output": "number",
        "tokens.text_input": "number",
        "media_duration.duration_price": "number",
        "other.sampling_step_price": "number",
        "other.characters_price": "number",
      },
    ],
  },
  audio_to_text: {
    type: "default",
    fields: [
      {
        "tokens.text_output": "number",
        "tokens.text_input": "number",
        "media_duration.duration_price": "number",
      },
    ],
  },
  vision: {
    type: "default",
    fields: [{ "other.pages_count_price": "number" }],
  },
  image: {
    type: "variant",
    fields: [
      { "images.res_to_quality": "map-map-number" },
      {
        "tokens.text_input": "number",
        "tokens.image_input": "number",
        "tokens.image_output": "number",
      },
    ],
  },
  video: {
    type: "variant",
    fields: [
      {
        "media_duration.duration_price": "number",
      },
    ],
  },
};

const applyMargin = (model) => {
  for (const type of Object.keys(fieldsForTypes)) {
    for (const fields of fieldsForTypes[type].fields) {
      for (const field of Object.keys(fields)) {
        const [key, subkey] = field.split(".");

        if (model.billing[key] == null) {
          model.billing[key] = {};
        }

        if (model.billing[key][subkey] == null) {
          continue;
        }

        if (fields[field] === "map-map-number") {
          for (const fieldKey of Object.keys(model.billing[key][subkey])) {
            for (const fieldSubkey of Object.keys(
              model.billing[key][subkey][fieldKey]
            )) {
              setPrice(
                model.billing[key][subkey][fieldKey][fieldSubkey],
                model.billing[key][subkey][fieldKey][fieldSubkey].raw_amount
              );
            }
          }
        } else {
          setPrice(
            model.billing[key][subkey].price,
            model.billing[key][subkey].price.raw_amount
          );
        }
      }
    }
  }

  return model;
};

const setFee = () => {
  newPricesResources.value = newPricesResources.value.map(applyMargin);
};

const availableProviders = computed(() => [
  ...new Set(newPricesResources.value.map((r) => r.provider).filter(Boolean)),
]);

const newPricesResourcesFiltred = computed(() => {
  const param = (searchParam.value || "").toLowerCase().trim();

  return newPricesResources.value.filter((r) => {
    if (onlyEnabled.value && r.disabled) return false;
    if (hideBroken.value && r.state?.state === "broken") return false;
    if (
      typesFilter.value.length &&
      !typesFilter.value.some((type) => (r.types || []).includes(type))
    ) {
      return false;
    }
    if (
      providersFilter.value.length &&
      !providersFilter.value.includes(r.provider)
    ) {
      return false;
    }
    if (!param) return true;

    return [r.name, r.key, r.provider, ...(r.types || [])]
      .filter(Boolean)
      .some((v) => v.toLowerCase().includes(param));
  });
});

const fieldsForAdd = computed(() => {
  if (!Object.keys(currentBillingSettings.value).length) {
    return {};
  }

  const resultMap = {};

  for (const type of currentBillingSettings.value.types || []) {
    for (const fields of fieldsForTypes[type]?.fields || []) {
      for (const field of Object.keys(fields)) {
        resultMap[field] = {
          type: fields[field],
          key: field.split(".")[0],
          subkey: field.split(".")[1],
        };
      }
    }
  }

  const result = Object.values(resultMap);

  result.sort((a) => (a.type === "number" ? -1 : 1));

  const resultByTypes = result.reduce((acc, v) => {
    if (!acc[v.key]) {
      acc[v.key] = [];
    }
    acc[v.key].push(v);
    acc[v.key].sort((a, b) => a.subkey.localeCompare(b.subkey));
    return acc;
  }, {});

  return resultByTypes;
});

const openBillingSettings = (item) => {
  const temp = JSON.parse(JSON.stringify(item));
  if (!temp.types) {
    temp.types = [];
  }

  for (const type of Object.keys(fieldsForTypes)) {
    for (const fields of fieldsForTypes[type].fields) {
      for (const field of Object.keys(fields)) {
        const [key, subkey] = field.split(".");

        if (temp.billing[key] == null) {
          temp.billing[key] = {};
        }

        if (temp.billing[key][subkey] != null) {
          continue;
        }

        if (fields[field] === "map-map-number") {
          temp.billing[key][subkey] = {};
        } else {
          temp.billing[key][subkey] = {};
          temp.billing[key][subkey].price = {
            amount: 0,
            raw_amount: 0,
            currency: defaultCurrency.value.code,
          };
        }
      }
    }
  }

  currentBillingSettings.value = applyMargin(temp);
  isBillingSettingsOpen.value = true;
};

const saveBillingSettings = async () => {
  isSaveModelLoading.value = true;

  const configModel = JSON.parse(
    JSON.stringify({ ...currentBillingSettings.value, billing: {} })
  );
  delete configModel.key;

  for (const type of configModel.types) {
    for (const fields of fieldsForTypes[type].fields) {
      for (const field of Object.keys(fields)) {
        const [key, subkey] = field.split(".");
        if (!configModel.billing[key]) {
          configModel.billing[key] = {};
        }

        if (fields[field] === "number") {
          configModel.billing[key][subkey] = currentBillingSettings.value
            .billing[key][subkey].price.raw_amount
            ? currentBillingSettings.value.billing[key][subkey]
            : null;
        }

        if (fields[field] === "map-map-number") {
          for (const fieldKey of Object.keys(
            currentBillingSettings.value.billing[key][subkey]
          )) {
            if (!configModel.billing[key][subkey]) {
              configModel.billing[key][subkey] = {};
            }
            configModel.billing[key][subkey][fieldKey] = {};

            for (const fieldSubKey of Object.keys(
              currentBillingSettings.value.billing[key][subkey][fieldKey]
            )) {
              if (
                currentBillingSettings.value.billing[key][subkey][fieldKey][
                  fieldSubKey
                ].raw_amount
              ) {
                configModel.billing[key][subkey][fieldKey][fieldSubKey] =
                  currentBillingSettings.value.billing[key][subkey][fieldKey][
                    fieldSubKey
                  ];
              }
            }
          }
        }
      }
    }
  }

  isSaveModelLoading.value = true;
  try {
    const response = await api.post("/api/openai/test_config", {
      model: currentBillingSettings.value.key,
      cfg: configModel,
    });
    billingSettinfsMessages.value = response.error_messages || [];
    if (billingSettinfsMessages.value.length) {
      return;
    }

    const { cfg: resultModel, new_serial } = await api.post(
      "/api/openai/save_model_config",
      {
        model: currentBillingSettings.value.key,
        cfg: configModel,
      }
    );
    currentSerial.value = new_serial;

    resultModel.key = currentBillingSettings.value.key;

    const index = newPricesResources.value.findIndex(
      (item) => item.key === resultModel.key
    );
    newPricesResources.value[index] = { ...resultModel };

    newPricesResources.value = [...newPricesResources.value];
    isBillingSettingsOpen.value = false;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save prices",
    });
  } finally {
    isSaveModelLoading.value = false;
  }
};

const addToMap = (field) => {
  if (
    !currentBillingSettings.value.billing[field.key][field.subkey][
      newKeysForMaps.value[field.subkey]
    ]
  ) {
    currentBillingSettings.value.billing[field.key][field.subkey][
      newKeysForMaps.value[field.subkey]
    ] = {};
  }

  currentBillingSettings.value.billing[field.key][field.subkey][
    newKeysForMaps.value[field.subkey]
  ][newSubkeysForMaps.value[field.subkey]] = {
    amount: 0,
    raw_amount: 0,
    currency: defaultCurrency.value.code,
  };

  currentBillingSettings.value.billing[field.key][field.subkey] = {
    ...currentBillingSettings.value.billing[field.key][field.subkey],
  };

  newKeysForMaps.value[field.subkey] = "";
  newSubkeysForMaps.value[field.subkey] = "";
};

const deleteFromMap = (field, key, subkey) => {
  const map = currentBillingSettings.value.billing[field.key][field.subkey];

  delete map[key][subkey];
  if (!Object.keys(map[key]).length) {
    delete map[key];
  }

  currentBillingSettings.value.billing[field.key][field.subkey] = {
    ...currentBillingSettings.value.billing[field.key][field.subkey],
  };
};

const fillConfig = async () => {
  try {
    isFillLoading.value = true;

    await api.post("/api/openai/fill_config", {});

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Fill config success",
    });
    store.dispatch("reloadBtn/onclick");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fill config",
    });
  } finally {
    isFillLoading.value = false;
  }
};

const exportToXlsx = async () => {
  try {
    isExportLoading.value = true;

    const baseHeaders = [
      { title: "Name", key: "name" },
      { title: "Price per 1M input tokens", key: "tokens.text_input" },
      { title: "Price per 1M output tokens", key: "tokens.text_output" },
      { title: "Price per 1M image input tokens", key: "tokens.image_input" },
      { title: "Price per 1M image output tokens", key: "tokens.image_output" },
      { title: "Price per generation step", key: "other.sampling_step_price" },
      { title: "Price per 1M input characters", key: "other.characters_price" },
      {
        title: "Price per 60 seconds of content",
        key: "media_duration.duration_price",
      },
      { title: "Price for 1000 pages", key: "other.pages_count_price" },
      { title: "Price for 1000 Web Searches", key: "other.web_search_price" },
      { title: "Price per image", key: "images.res_to_quality" },
    ];

    const allowedModels = newPricesResourcesFiltred.value.filter(
      (r) => !r.disabled && r.state?.state !== "broken"
    );

    return XlsxService.downloadXlsx(template.value.title + getTodayFullDate(), [
      {
        name: template.value.title,
        headers: baseHeaders,
        items: allowedModels.map((model) => {
          const result = {};

          baseHeaders.forEach(({ key }) => {
            if (key === "name") {
              result[key] = model.name;
              return;
            }

            const [mainKey, subKey] = key.split(".");

            if (model.billing[mainKey] == null) {
              result[key] = "";
              return;
            }

            if (subKey == null) {
              result[key] = JSON.stringify(model.billing[mainKey]);
              return;
            }

            if (model.billing[mainKey][subKey] == null) {
              result[key] = "";
              return;
            }

            if (subKey === "res_to_quality") {
              result[key] = Object.keys(model.billing[mainKey][subKey])
                .map((key) =>
                  Object.keys(model.billing[mainKey][subKey][key])
                    .map(
                      (subkey) =>
                        `${key} ${subkey} ${model.billing[mainKey][subKey][key][subkey].amount} ${defaultCurrency.value.code}`
                    )
                    .join(", ")
                )
                .join(", ");
              return;
            }

            result[key] = [
              model.billing[mainKey][subKey].price.amount,
              defaultCurrency.value.code,
            ].join(" ");
          });
          return result;
        }),
      },
    ]);
  } catch (e) {
    console.log(e);

    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during export config",
    });
  } finally {
    isExportLoading.value = false;
  }
};

const save = async () => {
  isSaveLoading.value = true;
  try {
    const oldPricesResult = JSON.parse(
      JSON.stringify(oldPricesResources.value)
    );
    const imageSize1792x1024 = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1792x1024_quality_standard")
    );
    const imageSize1792x1024HD = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1792x1024_quality_hd")
    );

    const imageSize1024x1792 = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1024x1792_quality_standard")
    );
    const imageSize1024x1792HD = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1024x1792_quality_hd")
    );

    imageSize1792x1024.price = imageSize1024x1792.price;
    imageSize1792x1024HD.price = imageSize1024x1792HD.price;

    const { new_serial } = await api.post("/api/openai/save_config", {
      serial: currentSerial.value,
      cfg: {
        models: newPricesResources.value.reduce((acc, r) => {
          acc[r.key] = { ...r, key: undefined };
          return acc;
        }, {}),
      },
    });
    currentSerial.value = new_serial;

    await api.plans.update(props.template.uuid, {
      ...props.template,
      products: {},
      addons: addons.value,
      resources: [
        ...template.value.resources.filter(
          (r) => oldPricesResult.findIndex((old) => old.key === r.key) === -1
        ),
        ...oldPricesResult,
      ],
      fee: fee.value,
    });

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Price model edited successfully",
    });
    store.dispatch("reloadBtn/onclick");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save prices",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const stateColor = (item) => {
  if (item.state?.state === "broken") return "red";
  return item.state?.state ? "green" : "grey";
};

const changeDisabled = (item, value) => {
  item.disabled = !value;
};

watch(fee, () => {
  if (isBillingSettingsOpen.value) {
    applyMargin(currentBillingSettings.value);
  }
});

watch(isBillingSettingsOpen, (value) => {
  if (!value) {
    currentBillingSettings.value = {};
    billingSettinfsMessages.value = [];
  }
});
</script>

<style scoped>
/* --- toolbar --- */
.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  padding: 12px 12px 14px;
}

.toolbar__search {
  max-width: 260px;
}

.toolbar__select {
  max-width: 200px;
}

.toolbar__switch {
  margin: 0;
  padding: 0;
  flex: 0 0 auto;
}

.toolbar__counter {
  margin-left: auto;
  font-size: 0.75rem;
  opacity: 0.6;
  white-space: nowrap;
}

.toolbar ::v-deep .v-label {
  font-size: 0.8rem;
}

.toolbar__select ::v-deep .v-select__selections {
  flex-wrap: nowrap;
  overflow: hidden;
}

.toolbar__select ::v-deep .v-select__selections > input {
  min-width: 0;
}

.toolbar__select ::v-deep .v-select__selection {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8rem;
}

/* --- table cells --- */
.cell-input {
  font-size: 0.85rem;
}

.cell-input ::v-deep .v-select__selections {
  flex-wrap: nowrap;
}

.cell-switch {
  margin: 0;
  padding: 0;
}

/* --- billing dialog --- */
.billing-card__head {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 16px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.25);
}

.billing-card__title {
  min-width: 0;
}

.billing-card__name {
  font-size: 1rem;
  font-weight: 500;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.billing-card__key {
  font-size: 0.7rem;
  opacity: 0.6;
}

.billing-card__types {
  margin-left: auto;
  max-width: 320px;
}

.billing-card__types ::v-deep .v-select__selections {
  flex-wrap: nowrap;
  overflow: hidden;
}

.billing-card__types ::v-deep .v-select__selection {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8rem;
}

.billing-card__body {
  padding: 12px 16px 16px;
  max-height: 62vh;
}

.billing-card__actions {
  padding: 8px 16px;
  border-top: 1px solid rgba(128, 128, 128, 0.25);
}

.billing-card__errors {
  margin: 0;
  font-size: 0.75rem;
  flex: 1 1 auto;
}

.hint {
  font-size: 0.8rem;
  opacity: 0.7;
}

.price-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 150px 150px;
  align-items: center;
  column-gap: 12px;
  row-gap: 8px;
}

.price-grid__head {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  opacity: 0.6;
}

.price-grid__head--num {
  text-align: center;
}

.price-grid__group {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  padding-bottom: 4px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.25);
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  opacity: 0.8;
}

.price-grid__hint {
  opacity: 0.55;
  cursor: help;
}

.price-grid__hint:hover {
  opacity: 1;
}

.price-grid__label {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  font-size: 0.8rem;
}

.price-grid__label span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.price-grid__add {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 150px 150px;
  column-gap: 12px;
  align-items: center;
  margin-top: 4px;
}

.price-grid__add ::v-deep .v-label {
  font-size: 0.75rem;
}

.price-field ::v-deep input {
  font-size: 0.8rem;
  text-align: right;
  /* ponytail: native spinners only add visual noise on price inputs */
  -moz-appearance: textfield;
}

.price-field ::v-deep input::-webkit-outer-spin-button,
.price-field ::v-deep input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.price-field ::v-deep .v-text-field__suffix {
  font-size: 0.7rem;
  opacity: 0.6;
}
</style>
