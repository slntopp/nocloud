<template>
  <div class="pa-4">
    <div class="buttons__inline pb-2 mt-4">
      <v-dialog max-width="400">
        <template v-slot:activator="{ on, attrs }">
          <v-btn class="mr-2" color="background-light" v-bind="attrs" v-on="on">
            Create
          </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row dense>
            <v-col cols="12">
              <v-text-field
                dense
                label="Title"
                v-model="newCurrency.title"
                :rules="rules.required"
              />
            </v-col>
            <v-col>
              <v-btn
                :disabled="!newCurrency.title"
                :loading="isCreateCurrencyLoading"
                @click="createCurrency"
                >Create</v-btn
              >
            </v-col>
          </v-row>
        </v-card>
      </v-dialog>

      <v-text-field
        dense
        readonly
        label="Default currency"
        class="d-inline-block"
        style="width: 200px"
        :value="defaultCurrency?.title"
      />
    </div>

    <nocloud-table
      table-name="currencies-table"
      class="mt-4"
      item-key="id"
      :items="currencies"
      :headers="currenciesHeaders"
      :loading="isLoading"
      :footer-error="fetchError"
      :show-select="false"
      show-expand
      :expanded.sync="expanded"
    >
      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length" style="padding: 0">
          <v-card color="background-light">
            <div class="d-flex align-center">
              <v-card-title>Rates</v-card-title>

              <v-dialog max-width="400" v-model="isCreateRateOpen">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn v-bind="attrs" v-on="on"> Add </v-btn>
                </template>
                <v-card class="pa-4">
                  <v-row dense>
                    <v-col cols="12">
                      <v-autocomplete
                        dense
                        label="To"
                        v-model="newRate.to"
                        item-text="title"
                        item-value="id"
                        return-object
                        :items="getCurrenciesTo(item)"
                      />
                    </v-col>
                    <v-col cols="12">
                      <v-text-field
                        dense
                        type="number"
                        label="Rate"
                        v-model="newRate.rate"
                        :rules="rules.number"
                      />
                    </v-col>
                    <v-col>
                      <v-btn
                        :loading="isCreateRateLoading"
                        @click="createRate(item)"
                        >Save</v-btn
                      >
                    </v-col>
                  </v-row>
                </v-card>
              </v-dialog>
            </div>

            <nocloud-table
              table-name="rates-table"
              class="mt-4"
              :items="getRates(item)"
              :headers="ratesHeaders"
            >
              <template v-slot:[`item.from`]="{ item }">
                {{ item.from.title }}
              </template>
              <template v-slot:[`item.to`]="{ item }">
                {{ item.to.title }}
              </template>
              <template v-slot:[`item.rate`]="{ item }">
                <v-text-field
                  dense
                  type="number"
                  style="width: 200px"
                  :value="Math.round(item.rate * 100) / 100"
                  @input="item.rate = $event"
                  :rules="rules.number"
                />
              </template>

              <template v-slot:[`item.commission`]="{ item }">
                <v-text-field
                  dense
                  type="number"
                  style="width: 200px"
                  :value="item.commission"
                  @input="item.commission = $event"
                />
              </template>

              <template v-slot:[`item.actions`]="{ item }">
                <div class="d-flex justify-end">
                  <v-btn
                    class="ml-2"
                    @click="editRate(item)"
                    :loading="isEditRateLoading"
                  >
                    Save
                    <v-icon class="ml-2">
                      mdi-content-save-edit-outline
                    </v-icon>
                  </v-btn>

                  <confirm-dialog
                    :loading="isDeleteRateLoading"
                    @confirm="deleteRate(item)"
                  >
                    <v-btn class="ml-2" :loading="isDeleteRateLoading">
                      Delete
                      <v-icon class="ml-2" title="Save edit">
                        mdi-delete-outline
                      </v-icon>
                    </v-btn>
                  </confirm-dialog>
                </div>
              </template>
            </nocloud-table>
          </v-card>
        </td>
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import { useStore } from "@/store";
import { computed, ref } from "vue";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { onMounted } from "vue";
import {
  CreateCurrencyRequest,
  CreateExchangeRateRequest,
  DeleteExchangeRateRequest,
  UpdateExchangeRateRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import { watch } from "vue";

const store = useStore();

const currenciesHeaders = [{ text: "Name ", value: "title" }];
const ratesHeaders = [
  { text: "To", value: "to" },
  { text: "Rate ", value: "rate" },
  { text: "Commission ", value: "commission" },
  { text: "Actions", value: "actions", sortable: false },
];

const fetchError = ref("");
const expanded = ref([]);

const isCreateCurrencyLoading = ref(false);
const newCurrency = ref({ title: "" });

const isCreateRateOpen = ref(false);
const isCreateRateLoading = ref(false);
const isEditRateLoading = ref(false);
const isDeleteRateLoading = ref(false);
const newRate = ref({ from: "", to: "", rate: "1" });

const rules = ref({
  number: [
    (value) =>
      /^[-+]?[0-9]*[.,]?[0-9]+(?:[eE][-+]?[0-9]+)?$/.test(value) || "Invalid!",
  ],
  required: [(v) => !!v || "This field is required!"],
});

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: fetchCurrencies,
  });

  fetchCurrencies();
});

const currenciesClient = computed(
  () => store.getters["currencies/currencyClient"]
);

const currencies = computed(() => store.getters["currencies/all"]);
const defaultCurrency = computed(() => store.getters["currencies/default"]);
const isLoading = computed(() => store.getters["currencies/isLoading"]);

const rates = computed(() =>
  store.getters["currencies/rates"].map((c) => ({
    ...c,
    rate: c.rate?.toString(),
    commission: c.commission?.toString(),
  }))
);

const fetchCurrencies = async (silent = false) => {
  try {
    await store.dispatch("currencies/fetch", { silent });
  } catch (err) {
    const message = err.response?.data?.message ?? err.message ?? err;
    fetchError.value = message;
  }
};

const createCurrency = async () => {
  try {
    isCreateCurrencyLoading.value = true;

    await currenciesClient.value.createCurrency(
      CreateCurrencyRequest.fromJson({ currency: newCurrency.value })
    );

    newCurrency.value = { title: "" };

    fetchCurrencies();
  } catch (err) {
    const message = err.response?.data?.message ?? err.message ?? err;
    store.commit("snackbar/showSnackbarError", { message: message });
  } finally {
    isCreateCurrencyLoading.value = false;
  }
};

const getRates = (item) => {
  return rates.value.filter((rate) => rate.from.id === item.id);
};

const getCurrenciesTo = (item) => {
  const rates = getRates(item);

  return currencies.value.filter(
    (currency) => !rates.find((rate) => rate.to.id === currency.id)
  );
};

const createRate = async (currency) => {
  try {
    isCreateRateLoading.value = true;

    await currenciesClient.value.createExchangeRate(
      CreateExchangeRateRequest.fromJson({ ...newRate.value, from: currency })
    );

    newRate.value = {};
    fetchCurrencies(true);
  } catch (err) {
    const message = err.response?.data?.message ?? err.message ?? err;
    store.commit("snackbar/showSnackbarError", { message: message });
  } finally {
    isCreateRateLoading.value = false;
  }
};

const editRate = async (item) => {
  try {
    isEditRateLoading.value = true;

    await currenciesClient.value.updateExchangeRate(
      UpdateExchangeRateRequest.fromJson({
        rate: item.rate,
        commission: item.commission,
        from: item.from,
        to: item.to,
      })
    );

    fetchCurrencies(true);
  } catch (err) {
    const message = err.response?.data?.message ?? err.message ?? err;
    store.commit("snackbar/showSnackbarError", { message: message });
  } finally {
    isEditRateLoading.value = false;
  }
};

const deleteRate = async (item) => {
  try {
    isDeleteRateLoading.value = true;

    await currenciesClient.value.deleteExchangeRate(
      DeleteExchangeRateRequest.fromJson({ from: item.from, to: item.to })
    );

    fetchCurrencies(true);
  } catch (err) {
    const message = err.response?.data?.message ?? err.message ?? err;
    store.commit("snackbar/showSnackbarError", { message: message });
  } finally {
    isDeleteRateLoading.value = false;
  }
};

watch(isCreateRateOpen, () => {
  if (!isCreateRateOpen.value) {
    newRate.value = {};
  }
});
</script>

<script>
export default {
  name: "currencies-view",
};
</script>
