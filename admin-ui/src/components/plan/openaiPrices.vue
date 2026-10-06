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
          <template v-slot:[`item.header`]="{ item }">
            <v-text-field type="number" dense v-model.number="item.price" />
          </template>
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field type="number" dense v-model.number="item.price" />
          </template>
        </nocloud-table>

        <div v-else-if="tab === 'Prices'">
          <div class="d-flex justify-space-between align-center">
            <v-text-field
              style="max-width: 400px"
              v-model="searchParam"
              label="Search"
            ></v-text-field>
            <v-btn :loading="isFillLoading" @click="fillConfig"
              >Fill config</v-btn
            >

            <v-btn :loading="isExportLoading" @click="exportToXlsx"
              >Export to XLSX</v-btn
            >
          </div>
          <div class="d-flex align-center">
            <v-select
              style="max-width: 200px"
              class="mr-2"
              dense
              label="Provider"
              :items="providers"
              v-model="bulkGroupProvider"
            />
            <v-text-field
              style="max-width: 200px"
              class="mr-2"
              dense
              label="Group"
              v-model="bulkGroup"
            />
            <v-btn :disabled="!bulkGroupProvider" @click="setGroupForProvider"
              >Set group for provider</v-btn
            >
          </div>
          <nocloud-table
            :headers="newPricesHeaders"
            :items="newPricesResourcesFiltred"
            :show-select="false"
          >
            <template v-slot:[`item.name`]="{ item }">
              <v-text-field dense v-model="item.name" />
            </template>

            <template v-slot:[`item.meta.group`]="{ item }">
              <v-text-field dense v-model="item.meta.group" />
            </template>

            <template v-slot:[`item.meta.sorter`]="{ item }">
              <v-text-field dense type="number" v-model.number="item.meta.sorter" />
            </template>

            <template v-slot:[`item.disabled`]="{ item }">
              <v-switch
                :key="item.key"
                dense
                :input-value="!item.disabled"
                @change="changeDisabled(item, $event)"
              />
            </template>

            <template v-slot:[`item.key`]="{ item }">
              {{ getShortName(item.key) }}
            </template>

            <template v-slot:[`item.visibility`]="{ item }">
              <v-select
                dense
                :items="['api_only', 'public', 'private']"
                v-model="item.visibility"
              />
            </template>

            <template v-slot:[`item.billing`]="{ item }">
              <v-btn icon @click="openBillingSettings(item)">
                <v-icon> mdi-menu-open </v-icon>
              </v-btn>
            </template>

            <template v-slot:[`item.types`]="{ item }">
              <template v-if="item.types?.length">
                <v-chip
                  v-for="type in item.types"
                  :key="type"
                  class="ml-1"
                  style="color: white"
                  small
                  dense
                  :color="typesColorMap[type]"
                  >{{ type }}</v-chip
                >
              </template>
              <span v-else>None</span>
            </template>

            <template v-slot:[`item.state.state`]="{ item }">
              <v-tooltip left>
                <template v-slot:activator="{ on, attrs }">
                  <v-chip
                    dense
                    v-bind="attrs"
                    v-on="on"
                    style="color: white"
                    y
                    :color="item.state?.state === 'broken' ? 'red' : 'green'"
                    >{{ item.state?.state }}</v-chip
                  >
                </template>
                <span
                  v-for="message in item.state.error_messages"
                  :key="message"
                  ><tr />
                  {{ message }}</span
                >
              </v-tooltip>
            </template>

            <template v-slot:[`item.provider`]="{ item }">
              <v-chip dense color="blue">{{ item.provider }}</v-chip>
            </template>
          </nocloud-table>

          <v-dialog width="70%" v-model="isBillingSettingsOpen">
            <v-card
              class="pa-5"
              color="background-light"
              style="min-height: 60vh"
            >
              <v-card-title class="key_title"
                >{{ currentBillingSettings.name }}
              </v-card-title>
              <v-select
                outlined
                chips
                label="Types"
                :items="availableTypes"
                multiple
                v-model="currentBillingSettings.types"
              />

              <v-row v-for="key in Object.keys(fieldsForAdd)" :key="key">
                <v-col cols="12">
                  <v-card-title class="key_title" style="padding: 0px">{{
                    keyLabelMap[key]
                  }}</v-card-title>
                </v-col>
                <v-col cols="12">
                  <v-row
                    align="center"
                    v-for="field in fieldsForAdd[key]"
                    :key="field.key"
                  >
                    <template v-if="field.type === 'number'">
                      <v-col cols="2" :key="field.subkey">
                        <span class="key_text">{{ fieldLabel(field) }}</span>
                      </v-col>

                      <v-col
                        v-for="{
                          key: priceKey,
                          label: priceLabel,
                        } in priceKeys"
                        cols="5"
                        :key="`${field.subkey}-${priceKey}`"
                      >
                        <v-text-field
                          hide-details
                          dense
                          outlined
                          type="number"
                          :label="priceLabel"
                          :value="
                            currentBillingSettings.billing[field.key][
                              [field.subkey]
                            ].price[priceKey]
                          "
                          @input="
                            currentBillingSettings.billing[field.key][
                              [field.subkey]
                            ].price[priceKey] = +$event
                          "
                        />
                      </v-col>
                    </template>
                    <template v-else-if="field.type === 'resolution-table'">
                      <v-col cols="12">
                        <span class="key_text">Price / 1 second</span>
                        <div class="resolution-prices-wrap">
                          <table class="resolution-prices">
                            <colgroup>
                              <col class="resolution-prices__label-col" />
                              <col
                                v-for="index in resolutionColumns(field).length *
                                2"
                                :key="index"
                                class="resolution-prices__value-col"
                              />
                            </colgroup>
                            <thead>
                              <tr class="resolution-prices__group">
                                <th>Audio</th>
                                <th
                                  v-for="resolution in resolutionColumns(field)"
                                  :key="resolution"
                                  colspan="2"
                                >
                                  <span>{{ resolution }}</span>
                                  <v-btn
                                    icon
                                    x-small
                                    @click="removeResolution(field, resolution)"
                                    ><v-icon small>mdi-close</v-icon></v-btn
                                  >
                                </th>
                              </tr>
                              <tr class="resolution-prices__sub">
                                <th></th>
                                <th
                                  v-for="cell in resolutionSubheads(field)"
                                  :key="cell.key"
                                >
                                  {{ cell.label }}
                                </th>
                              </tr>
                            </thead>
                            <tbody>
                              <tr v-for="row in audioPriceRows" :key="row.key">
                                <th>{{ row.label }}</th>
                                <template
                                  v-for="resolution in resolutionColumns(field)"
                                >
                                  <td
                                    v-for="{
                                      key: priceKey,
                                      label: priceLabel,
                                    } in priceKeys"
                                    :key="`${row.key}-${resolution}-${priceKey}`"
                                  >
                                    <v-text-field
                                      hide-details
                                      dense
                                      outlined
                                      type="number"
                                      :label="priceLabel"
                                      :value="
                                        audioCell(
                                          field,
                                          resolution,
                                          row.key
                                        )[priceKey]
                                      "
                                      @input="
                                        audioCell(field, resolution, row.key)[
                                          priceKey
                                        ] = +$event
                                      "
                                    />
                                  </td>
                                </template>
                              </tr>
                            </tbody>
                          </table>
                        </div>
                        <v-row justify="end" class="mt-2">
                          <v-col cols="3">
                            <v-text-field
                              dense
                              outlined
                              hide-details
                              label="New resolution"
                              v-model="newResolution"
                            />
                          </v-col>
                          <v-col cols="2">
                            <v-btn
                              :disabled="!newResolution || isSaveModelLoading"
                              @click="addResolution(field)"
                              >Add new value</v-btn
                            >
                          </v-col>
                        </v-row>
                      </v-col>
                    </template>
                    <template v-else-if="field.type === 'map-number'">
                      <v-col cols="12" :key="field.subkey">
                        <span class="key_text">{{ fieldLabel(field) }}</span>
                        <v-row
                          v-for="resolution in Object.keys(
                            currentBillingSettings.billing[field.key][
                              field.subkey
                            ] || {}
                          )"
                          :key="resolution"
                          align="center"
                        >
                          <v-col cols="2">
                            <span class="key_text">{{ resolution }}</span>
                          </v-col>
                          <v-col
                            v-for="{
                              key: priceKey,
                              label: priceLabel,
                            } in priceKeys"
                            cols="5"
                            :key="`${resolution}-${priceKey}`"
                          >
                            <v-text-field
                              hide-details
                              dense
                              outlined
                              type="number"
                              :label="priceLabel"
                              :value="
                                currentBillingSettings.billing[field.key][
                                  field.subkey
                                ][resolution][priceKey]
                              "
                              @input="
                                currentBillingSettings.billing[field.key][
                                  field.subkey
                                ][resolution][priceKey] = +$event
                              "
                            />
                          </v-col>
                        </v-row>
                      </v-col>
                    </template>
                    <v-col v-else :key="field.subkey" cols="12">
                      <div>
                        <span class="key_text"
                          >Table for {{ field.subkey }}</span
                        >
                        <template
                          v-for="key in Object.keys(
                            currentBillingSettings.billing[field.key][
                              [field.subkey]
                            ]
                          )"
                        >
                          <v-row
                            v-for="subkey in Object.keys(
                              currentBillingSettings.billing[field.key][
                                [field.subkey]
                              ][key]
                            )"
                            :key="`${key}-${subkey}`"
                          >
                            <v-col
                              cols="2"
                              class="d-flex justify-start align-center"
                              ><span class="key_text"
                                >{{ key }} {{ subkey }}</span
                              >

                              <v-btn icon @click="deleteFromMap(field, key)"
                                ><v-icon>mdi-delete</v-icon></v-btn
                              >
                            </v-col>
                            <v-col
                              cols="5"
                              v-for="{
                                key: priceKey,
                                label: priceLabel,
                              } in priceKeys"
                              :key="`${field.subkey}-${priceKey}`"
                            >
                              <v-text-field
                                dense
                                outlined
                                type="number"
                                :label="priceLabel"
                                :value="
                                  currentBillingSettings.billing[field.key][
                                    [field.subkey]
                                  ][key][subkey][priceKey]
                                "
                                hide-details
                                @input="
                                  currentBillingSettings.billing[field.key][
                                    [field.subkey]
                                  ][key][subkey][priceKey] = +$event
                                "
                              />
                            </v-col>
                          </v-row>
                        </template>

                        <v-row justify="end">
                          <v-col cols="3">
                            <v-text-field
                              dense
                              outlined
                              hide-details
                              label="New key"
                              v-model="newKeysForMaps[field.subkey]"
                            />
                          </v-col>
                          <v-col cols="3">
                            <v-text-field
                              dense
                              outlined
                              hide-details
                              label="New subkey"
                              v-model="newSubkeysForMaps[field.subkey]"
                            />
                          </v-col>
                          <v-col cols="2">
                            <v-btn
                              :disabled="
                                !newKeysForMaps[field.subkey] ||
                                !newSubkeysForMaps[field.subkey] ||
                                isSaveModelLoading
                              "
                              @click="addToMap(field)"
                              >Add new value</v-btn
                            >
                          </v-col>
                        </v-row>
                      </div>
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>

              <div
                class="d-flex justify-center align-center mt-2 mb-2"
                v-if="billingSettinfsMessages.length"
              >
                <v-card-title style="color: red; text-align: center">
                  Errors: {{ billingSettinfsMessages.join(", ") }}
                </v-card-title>
              </div>

              <v-card-actions class="d-flex justify-end mt-3">
                <v-btn
                  :disabled="isSaveModelLoading"
                  @click="isBillingSettingsOpen = false"
                  >Close</v-btn
                >
                <v-btn
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
  { text: "Group", value: "meta.group", width: 150 },
  { text: "Sorter", value: "meta.sorter", width: 100 },
  { text: "Types", value: "types", sortable: false },
  { text: "Visibility", value: "visibility", width: 150 },
  { text: "Enabled", value: "disabled", width: 100 },
  { text: "Billing", value: "billing", sortable: false, width: 100 },
  { text: "State", value: "state.state", width: 100 },
];

const priceKeys = [
  { key: "raw_amount", label: "Price" },
  { key: "amount", label: "Margined price" },
];

const fieldLabel = (field) => {
  const isVideo = (currentBillingSettings.value.types || []).includes("video");
  if (
    isVideo &&
    (field.subkey === "duration_price" || field.subkey === "resolution_prices")
  ) {
    return "Price / 1 second";
  }
  return field.subkey;
};

const searchParam = ref("");

const isSaveLoading = ref(false);
const isSaveModelLoading = ref(false);
const isBillingSettingsOpen = ref(false);
const billingSettinfsMessages = ref([]);
const currentBillingSettings = ref({});
const newKeysForMaps = ref({});
const newSubkeysForMaps = ref({});
const newResolution = ref("");
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
      newPricesResources.value.push(
        withMeta({ ...response.cfg.models[key], key })
      );
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

const audioPriceRows = [
  { key: "with_audio", label: "With audio" },
  { key: "without_audio", label: "Without audio" },
];

const emptyAudioPrice = () => ({
  amount: 0,
  raw_amount: 0,
  currency: defaultCurrency.value.code,
});

const normalizeAudioPrice = (value) => {
  const blank = () => ({
    with_audio: emptyAudioPrice(),
    without_audio: emptyAudioPrice(),
  });
  if (!value || typeof value !== "object") {
    return blank();
  }
  if (value.with_audio || value.without_audio) {
    return {
      with_audio: { ...emptyAudioPrice(), ...value.with_audio },
      without_audio: { ...emptyAudioPrice(), ...value.without_audio },
    };
  }
  if (value.raw_amount || value.amount) {
    const legacy = { ...emptyAudioPrice(), ...value };
    return { with_audio: { ...legacy }, without_audio: { ...legacy } };
  }
  return blank();
};

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
        "tokens.text_output": "number",
        "tokens.image_input": "number",
        "tokens.image_output": "number",
      },
    ],
  },
  video: {
    type: "variant",
    fields: [
      {
        "media_duration.resolution_prices": "resolution-table",
      },
    ],
  },
};

const setFee = () => {
  newPricesResources.value = newPricesResources.value.map((temp) => {
    for (const type of Object.keys(fieldsForTypes)) {
      for (const fields of fieldsForTypes[type].fields) {
        for (const field of Object.keys(fields)) {
          const [key, subkey] = field.split(".");

          if (temp.billing[key] == null) {
            temp.billing[key] = {};
          }

          if (temp.billing[key][subkey] == null) {
            continue;
          }

          if (fields[field] === "map-map-number") {
            for (const fieldKey of Object.keys(temp.billing[key][subkey])) {
              for (const fieldSubkey of Object.keys(
                temp.billing[key][subkey][fieldKey]
              )) {
                temp.billing[key][subkey][fieldKey][fieldSubkey].amount =
                  getMarginedValue(
                    fee.value,
                    temp.billing[key][subkey][fieldKey][fieldSubkey].raw_amount
                  );
              }
            }
          } else if (fields[field] === "resolution-table") {
            for (const resolution of Object.keys(temp.billing[key][subkey])) {
              const rate = temp.billing[key][subkey][resolution];
              for (const side of ["with_audio", "without_audio"]) {
                if (!rate?.[side]) {
                  continue;
                }
                rate[side].amount = getMarginedValue(
                  fee.value,
                  rate[side].raw_amount
                );
              }
            }
          } else if (fields[field] === "map-number") {
            for (const fieldKey of Object.keys(temp.billing[key][subkey])) {
              temp.billing[key][subkey][fieldKey].amount = getMarginedValue(
                fee.value,
                temp.billing[key][subkey][fieldKey].raw_amount
              );
            }
          } else {
            temp.billing[key][subkey].price.amount = getMarginedValue(
              fee.value,
              temp.billing[key][subkey].price.raw_amount
            );
          }
        }
      }
    }

    return temp;
  });
};

const newPricesResourcesFiltred = computed(() => {
  const param = searchParam.value.toLowerCase();
  return newPricesResources.value.filter(
    (r) =>
      r.name.includes(param) ||
      r.key.includes(param) ||
      (r?.types || []).includes(param) ||
      r.provider === param
  );
});

const fieldsForAdd = computed(() => {
  if (!Object.keys(currentBillingSettings.value).length) {
    return [];
  }

  const resultMap = {};

  for (const type of currentBillingSettings.value.types) {
    for (const fields of fieldsForTypes[type].fields) {
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

        if (fields[field] === "resolution-table") {
          const current = temp.billing[key][subkey] || {};
          const fromEnum =
            temp.meta?.request_parameters?.resolution?.enum || [];
          const names = [];
          for (const resolution of [...fromEnum, ...Object.keys(current)]) {
            if (resolution && !names.includes(resolution)) {
              names.push(resolution);
            }
          }
          const next = {};
          for (const resolution of names) {
            next[resolution] = normalizeAudioPrice(current[resolution]);
          }
          temp.billing[key][subkey] = next;
          continue;
        }

        if (fields[field] === "map-number") {
          if (temp.billing[key][subkey] == null) {
            temp.billing[key][subkey] = {};
          }
          const resolutions =
            temp.meta?.request_parameters?.resolution?.enum || [];
          for (const resolution of resolutions) {
            if (temp.billing[key][subkey][resolution] != null) {
              continue;
            }
            temp.billing[key][subkey][resolution] = {
              amount: 0,
              raw_amount: 0,
              currency: defaultCurrency.value.code,
            };
          }
          continue;
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

  currentBillingSettings.value = temp;
  isBillingSettingsOpen.value = true;
};

const resolutionColumns = (field) => {
  const stored =
    currentBillingSettings.value.billing?.[field.key]?.[field.subkey] || {};
  const fromEnum =
    currentBillingSettings.value.meta?.request_parameters?.resolution?.enum;
  const names = [];
  const push = (name) => {
    if (!name || names.includes(name)) {
      return;
    }
    names.push(name);
  };
  if (Array.isArray(fromEnum)) {
    fromEnum.forEach(push);
  }
  Object.keys(stored).forEach(push);
  return names;
};

const setResolutionColumns = (field, names) => {
  const settings = currentBillingSettings.value;
  const current = settings.billing[field.key][field.subkey] || {};
  const next = {};
  for (const name of names) {
    next[name] = current[name] || normalizeAudioPrice(null);
  }
  settings.billing[field.key][field.subkey] = next;

  const meta = { ...(settings.meta || {}) };
  const params = { ...(meta.request_parameters || {}) };
  params.resolution = { ...(params.resolution || {}), enum: [...names] };
  meta.request_parameters = params;
  settings.meta = meta;
};

const addResolution = (field) => {
  const name = (newResolution.value || "").trim();
  if (!name) {
    return;
  }
  const names = resolutionColumns(field);
  if (!names.includes(name)) {
    names.push(name);
  }
  setResolutionColumns(field, names);
  newResolution.value = "";
};

const removeResolution = (field, name) => {
  setResolutionColumns(
    field,
    resolutionColumns(field).filter((item) => item !== name)
  );
};

const resolutionSubheads = (field) =>
  resolutionColumns(field).flatMap((resolution) => [
    { key: `${resolution}-cost`, label: "Price" },
    { key: `${resolution}-sell`, label: "Margined" },
  ]);

const audioCell = (field, resolution, side) =>
  currentBillingSettings.value.billing[field.key][field.subkey][resolution][
    side
  ];

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

        if (fields[field] === "resolution-table") {
          configModel.billing[key][subkey] = {};
          const rates =
            currentBillingSettings.value.billing[key][subkey] || {};
          for (const resolution of Object.keys(rates)) {
            const saved = {};
            for (const side of ["with_audio", "without_audio"]) {
              if (rates[resolution]?.[side]?.raw_amount) {
                saved[side] = rates[resolution][side];
              }
            }
            if (Object.keys(saved).length) {
              configModel.billing[key][subkey][resolution] = saved;
            }
          }
          if (!configModel.meta) {
            configModel.meta = {};
          }
          const params = {
            ...(configModel.meta.request_parameters || {}),
          };
          params.resolution = {
            ...(params.resolution || {}),
            enum: Object.keys(rates),
          };
          configModel.meta.request_parameters = params;
        }

        if (fields[field] === "map-number") {
          configModel.billing[key][subkey] = {};
          for (const resolution of Object.keys(
            currentBillingSettings.value.billing[key][subkey] || {}
          )) {
            if (
              currentBillingSettings.value.billing[key][subkey][resolution]
                .raw_amount
            ) {
              configModel.billing[key][subkey][resolution] =
                currentBillingSettings.value.billing[key][subkey][resolution];
            }
          }
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
    newPricesResources.value[index] = withMeta(resultModel);

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

const deleteFromMap = (field, key) => {
  delete currentBillingSettings.value.billing[field.key][field.subkey][key];

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

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Export success",
    });
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

    await api.post("/api/openai/save_config", {
      serial: currentSerial.value,
      cfg: {
        models: newPricesResources.value.reduce((acc, r) => {
          acc[r.key] = { ...r, key: undefined, meta: withoutEmptySorter(r.meta) };
          return acc;
        }, {}),
      },
    });

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

const changeDisabled = (item, value) => {
  item.disabled = !value;
};

// meta is omitted by backend when empty; group and sorter must exist up front for Vue 2 reactivity
const withMeta = (model) => ({
  ...model,
  meta: { group: "", sorter: "", ...model.meta },
});

// sorter orders the chat's models, 0 first; a blank one is not stored
const withoutEmptySorter = ({ sorter, ...meta }) =>
  typeof sorter === "number" && Number.isFinite(sorter)
    ? { ...meta, sorter }
    : meta;

const bulkGroupProvider = ref("");
const bulkGroup = ref("");

const providers = computed(() => [
  ...new Set(newPricesResources.value.map((r) => r.provider)),
]);

const setGroupForProvider = () => {
  newPricesResources.value
    .filter((r) => r.provider === bulkGroupProvider.value)
    .forEach((r) => (r.meta.group = bulkGroup.value));
};

watch(isBillingSettingsOpen, (value) => {
  if (!value) {
    currentBillingSettings.value = {};
    billingSettinfsMessages.value = [];
  }
});
</script>

<style scoped>
.key_text {
  font-size: 1rem;
}
.key_title {
  font-size: 1.4em;
}

.resolution-prices-wrap {
  overflow-x: auto;
}

.resolution-prices {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  border: 1px solid rgba(128, 128, 128, 0.35);
  border-radius: 8px;
}

.resolution-prices__label-col {
  width: 148px;
}

.resolution-prices__value-col {
  width: 108px;
}

.resolution-prices th,
.resolution-prices td {
  height: auto;
  padding: 8px;
  text-align: center;
  vertical-align: middle;
  border-right: 1px solid rgba(128, 128, 128, 0.25);
  border-bottom: 1px solid rgba(128, 128, 128, 0.25);
}

.resolution-prices tr > :last-child {
  border-right: 0;
}

.resolution-prices tbody tr:last-child > * {
  border-bottom: 0;
}

.resolution-prices__group th {
  height: 40px;
  font-size: 14px;
  font-weight: 600;
}

.resolution-prices__sub th {
  height: 32px;
  font-size: 12px;
  font-weight: 500;
  opacity: 0.7;
}

.resolution-prices tbody th {
  font-size: 13px;
  font-weight: 500;
  text-align: left;
}
</style>
