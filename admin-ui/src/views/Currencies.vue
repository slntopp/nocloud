<template>
  <div class="pa-4">
    <div class="buttons__inline pb-2 mt-4">
      <v-dialog v-model="isCreateCurrencyOpen" max-width="400">
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

            <v-col cols="12">
              <v-text-field
                dense
                label="Title"
                v-model="newCurrency.code"
                :rules="rules.required"
              />
            </v-col>

            <v-col class="d-flex justify-end">
              <v-btn
                :disabled="!newCurrency.title || !newCurrency.code"
                :loading="isCreateCurrencyLoading"
                @click="createCurrency"
                >Create</v-btn
              >
            </v-col>
          </v-row>
        </v-card>
      </v-dialog>

      <v-select
        dense
        label="Default currency"
        class="d-inline-block"
        style="width: 200px"
        :items="currencies"
        item-text="code"
        item-value="id"
        v-model="newDefaultCurrency"
      />

      <template v-if="newDefaultCurrency !== defaultCurrency.id">
        <confirm-dialog
          @confirm="changeDefaultCurrency"
          :loading="isChangeDefaultCurrencyLoading"
        >
          <v-btn class="ml-3" :loading="isChangeDefaultCurrencyLoading"
            >Change default currency</v-btn
          >
        </confirm-dialog>
      </template>
    </div>

    <nocloud-table
      table-name="currencies-table"
      class="mt-4"
      item-key="id"
      :items="currenciesItems"
      :headers="currenciesHeaders"
      :loading="isLoading"
      :footer-error="fetchError"
      :show-select="false"
    >
      <template v-slot:[`item.code`]="{ item }">
        <div>
          <v-text-field
            :loading="updatingCurrencyId === item.id"
            dense
            hide-details
            :disabled="!!updatingCurrencyId"
            v-model="item.code"
          />
        </div>
      </template>

      <template v-slot:[`item.title`]="{ item }">
        <v-text-field
          :loading="updatingCurrencyId === item.id"
          dense
          hide-details
          :disabled="!!updatingCurrencyId"
          v-model="item.title"
        />
      </template>

      <template v-slot:[`item.public`]="{ item }">
        <v-switch
          :loading="updatingCurrencyId === item.id"
          dense
          hide-details
          :disabled="!!updatingCurrencyId"
          v-model="item.public"
        />
      </template>

      <template v-slot:[`item.rate`]="{ item }">
        <v-text-field
          type="number"
          :loading="updatingCurrencyId === item.id"
          dense
          hide-details
          :disabled="!!updatingCurrencyId || item.default"
          v-model.number="item.rate.value"
        />
      </template>

      <template v-slot:[`item.precision`]="{ item }">
        <v-text-field
          type="number"
          :loading="updatingCurrencyId === item.id"
          dense
          hide-details
          :disabled="!!updatingCurrencyId"
          v-model.number="item.precision"
        />
      </template>

      <template v-slot:[`item.rounding`]="{ item }">
        <v-select
          :loading="updatingCurrencyId === item.id"
          dense
          hide-details
          :disabled="!!updatingCurrencyId"
          v-model="item.rounding"
          item-text="text"
          item-value="value"
          :items="roundingItems"
        />
      </template>

      <template v-slot:[`item.actions`]="{ item }">
        <div class="d-flex justify-end">
          <v-btn
            :disabled="!editedCurrencies[item.id]"
            @click="saveCurrency(item)"
            icon
            ><v-icon :color="editedCurrencies[item.id] ? 'primary' : ''"
              >mdi-content-save</v-icon
            ></v-btn
          >
        </div>
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import { useStore } from "@/store";
import { computed, ref, watch } from "vue";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { onMounted } from "vue";
import {
  ChangeDefaultCurrencyRequest,
  CreateCurrencyRequest,
  CreateExchangeRateRequest,
  Rounding,
  UpdateCurrencyRequest,
  UpdateExchangeRateRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";

const store = useStore();

const currenciesHeaders = [
  { text: "Code", value: "code", width: 200, sortable: false },
  { text: "Title", value: "title", sortable: false },
  { text: "Rate", value: "rate", sortable: false },
  { text: "Precision", value: "precision", width: 100, sortable: false },
  { text: "Rounding", value: "rounding", width: 150, sortable: false },
  { text: "Public ", value: "public", width: 100, sortable: false },
  { text: "Actions ", value: "actions", width: 100, sortable: false },
];

const currenciesItems = ref([]);
const originalCurrenciesItems = ref([]);
const newDefaultCurrency = ref();
const fetchError = ref("");
const isCreateCurrencyOpen = ref(false);
const isChangeDefaultCurrencyLoading = ref(false);
const isCreateCurrencyLoading = ref(false);
const newCurrency = ref({ title: "", code: "" });
const updatingCurrencyId = ref("");

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

const rates = computed(() => store.getters["currencies/rates"]);
const currencies = computed(() =>
  store.getters["currencies/all"].filter((c) => !!c.id)
);
const defaultCurrency = computed(() => store.getters["currencies/default"]);
const isLoading = computed(() => store.getters["currencies/isLoading"]);
const roundingItems = computed(() =>
  Object.keys(Rounding)
    .filter((v) => !Number.isInteger(+v))
    .map((val) => ({
      text: val.split("_")[1],
      value: Rounding[val],
    }))
);
const editedCurrencies = computed(() =>
  currenciesItems.value.reduce((acc, item, index) => {
    acc[item.id] =
      JSON.stringify(item) !==
      JSON.stringify(originalCurrenciesItems.value[index]);
    return acc;
  }, {})
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

    newCurrency.value = { title: "", code: "" };
    isCreateCurrencyOpen.value = false;

    fetchCurrencies();
  } catch (err) {
    const message = err.response?.data?.message ?? err.message ?? err;
    store.commit("snackbar/showSnackbarError", { message: message });
  } finally {
    isCreateCurrencyLoading.value = false;
  }
};

const saveCurrency = async (item) => {
  try {
    updatingCurrencyId.value = item.id;

    await currenciesClient.value.updateCurrency(
      UpdateCurrencyRequest.fromJson({
        currency: {
          code: item.code,
          title: item.title,
          id: item.id,
          precision: item.precision,
          format: item.format,
          rounding: item.rounding,
          public: item.public,
          default: item.default,
        },
      })
    );

    if (!item.default) {
      if (item.rate.isExists) {
        await currenciesClient.value.updateExchangeRate(
          UpdateExchangeRateRequest.fromJson({
            rate: item.rate.value,
            to: { id: item.id },
            from: { id: 0 },
          })
        );
      } else {
        await currenciesClient.value.createExchangeRate(
          CreateExchangeRateRequest.fromJson({
            rate: item.rate.value,
            to: { id: item.id },
            from: { id: 0 },
          })
        );

        currenciesItems.value = currenciesItems.value.map((c) =>
          c.id === item.id ? { ...c, rate: { ...c.rate, isExists: true } } : c
        );
      }
    }

    originalCurrenciesItems.value = originalCurrenciesItems.value.map((c) =>
      c.id === item.id ? JSON.parse(JSON.stringify(item)) : c
    );
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    updatingCurrencyId.value = "";
  }
};

const changeDefaultCurrency = async () => {
  try {
    isChangeDefaultCurrencyLoading.value = true;

    await currenciesClient.value.changeDefaultCurrency(
      ChangeDefaultCurrencyRequest.fromJson({
        id: newDefaultCurrency.value,
      })
    );

    await fetchCurrencies();
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isChangeDefaultCurrencyLoading.value = false;
  }
};

watch([currencies, rates], () => {
  newDefaultCurrency.value = defaultCurrency.value.id;

  const items = currencies.value.map((currency) => {
    const rate = rates.value.find(
      (val) => val.to.id == currency.id && !val.from.id
    );

    return {
      ...currency,
      rate: {
        value: rate?.rate || 0,
        isExists: !!rate,
      },
    };
  });
  currenciesItems.value = JSON.parse(JSON.stringify(items));
  originalCurrenciesItems.value = JSON.parse(JSON.stringify(items));
});

watch(rates, async () => {
  if (
    defaultCurrency.value &&
    rates.value.length &&
    !rates.value.find(
      (rate) => rate.to.id == defaultCurrency.value.id && !rate.from.id
    )
  ) {
    await currenciesClient.value.createExchangeRate(
      CreateExchangeRateRequest.fromJson({
        rate: 1,
        to: { id: defaultCurrency.value.id },
        from: { id: 0 },
      })
    );

    fetchCurrencies();
  }
});
</script>

<script>
export default {
  name: "currencies-view",
};
</script>
