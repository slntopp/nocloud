<template>
  <div class="pa-4">
    <v-dialog max-width="400">
      <template v-slot:activator="{ on, attrs }">
        <v-btn class="mr-2" color="background-light" v-bind="attrs" v-on="on">
          Add
        </v-btn>
      </template>
      <v-card class="pa-4">
        <v-row dense>
          <v-col cols="12">
            <v-select
              dense
              label="Currency 1"
              v-model="currency.from"
              :items="currenciesFrom"
            />
          </v-col>
          <v-col cols="12">
            <v-select
              dense
              label="Currency 2"
              v-model="currency.to"
              :items="currenciesTo"
            />
          </v-col>
          <v-col cols="12">
            <v-text-field
              dense
              type="number"
              label="Rate"
              v-model="currency.rate"
              :rules="rules.number"
            />
          </v-col>
          <v-col>
            <v-btn :loading="isCreateLoading" @click="addCurrency">Save</v-btn>
          </v-col>
        </v-row>
      </v-card>
    </v-dialog>

    <confirm-dialog @confirm="deleteSelectedCurrencies">
      <v-btn class="mr-2" color="background-light" :disabled="selected.length < 1">
        Delete
      </v-btn>
    </confirm-dialog>

    <v-text-field
      dense
      readonly
      label="Default currency"
      class="d-inline-block"
      style="width: 200px"
      :value="defaultCurrency"
    />

    <nocloud-table
      class="mt-4"
      item-key="id"
      v-model="selected"
      :items="currencies"
      :headers="headers"
      :loading="isLoading"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.rate`]="{ item }">
        <v-text-field
          dense
          type="number"
          style="width: 200px"
          v-model="item.rate"
          :rules="rules.number"
        />
      </template>
      <template v-slot:[`item.actions`]="{ item }">
        <v-icon title="Save edit" @click="editCurrency(item)">
          mdi-content-save-edit-outline
        </v-icon>
      </template>
    </nocloud-table>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";

export default {
  name: "currencies-view",
  components: { nocloudTable, confirmDialog },
  mixins: [snackbar],
  data: () => ({
    headers: [
      { text: "Currency 1 ", value: "from" },
      { text: "Currency 2 ", value: "to" },
      { text: "Rate ", value: "rate" },
      { text: "Actions", value: "actions", sortable: false },
    ],
    currenciesList: ['NCU', 'USD', 'EUR', 'BYN', 'PLN'],
    currencies: [],
    selected: [],

    rules: {
      number: [(value) => /^[-+]?[0-9]*[.,]?[0-9]+(?:[eE][-+]?[0-9]+)?$/.test(value) || 'Invalid!']
    },

    currency: { from: "", to: "", rate: "1" },
    isLoading: false,
    isCreateLoading: false,
    fetchError: "",
  }),
  methods: {
    addCurrency() {
      if (this.currency.from === "" || this.currency.to === "") return;
      if (typeof this.rules.number[0](this.currency.rate) === 'string') return;

      const newCurrency = {
        rate: +this.currency.rate.replace(',', '.'),
        from: this.currenciesList.indexOf(this.currency.from),
        to: this.currenciesList.indexOf(this.currency.to)
      };

      this.isCreateLoading = true;
      api.post('/billing/currencies/rates', newCurrency)
        .then(() => {
          this.currencies.push({
            ...this.currency,
            id: `${this.currency.from} ${this.currency.to}`
          });
          this.currency = { from: "", to: "", rate: "1" };
        })
        .catch((err) => {
          const message = err.response?.data?.message ?? err.message ?? err;

          this.showSnackbarError({ message });
          console.error(err);
        })
        .finally(() => this.isCreateLoading = false);
    },
    editCurrency(currency) {
      if (typeof this.rules.number[0](currency.rate) === 'string') return;
      const newCurrency = {
        rate: +currency.rate.replace(',', '.'),
        from: this.currenciesList.indexOf(currency.from),
        to: this.currenciesList.indexOf(currency.to)
      };

      this.isLoading = true;
      api.put('/billing/currencies/rates', newCurrency)
        .then(() => {
          this.showSnackbarSuccess({ message: 'Done' });
        })
        .catch((err) => {
          const message = err.response?.data?.message ?? err.message ?? err;

          this.showSnackbarError({ message });
          console.error(err);
        })
        .finally(() => this.isLoading = false);
    },
    deleteSelectedCurrencies() {
      this.isLoading = true;
      const promises = this.selected.filter(({ id }) =>
        this.currencies.find((el) => el.id === id)
      ).map(({ from, to }) =>
        api.delete(`/billing/currencies/rates/${from}/${to}`)
      );

      Promise.all(promises).then(() => {
        this.currencies = this.currencies.filter(({ id }) =>
          !this.selected.find((el) => el.id === id)
        );
        this.selected = [];
      })
      .catch((err) => {
        const message = err.response?.data?.message ?? err.message ?? err;

        this.showSnackbarError({ message });
        console.error(err);
      })
      .finally(() => this.isLoading = false);
    }
  },
  created() {
    this.isLoading = true;
    api.get('/billing/currencies')
      .then((res) => {
        this.currenciesList = res.currencies;

        return api.get('/billing/currencies/rates');
      })
      .then((res) => {
        this.currencies = res.rates.map((el) => ({
          ...el, id: `${el.from} ${el.to}`
        }));
      })
      .catch((err) => {
        const message = err.response?.data?.message ?? err.message ?? err;

        this.showSnackbarError({ message });
        console.error(err);
      })
      .finally(() => this.isLoading = false);
  },
  computed: {
    currenciesFrom() {
      const currencies = [];

      this.currencies.forEach(({ from, to }) => {
        if (to === this.currency.to) currencies.push(from);
      });

      return this.currenciesList.filter((el) =>
        el !== this.currency.to && !currencies.includes(el)
      );
    },
    currenciesTo() {
      const currencies = [];

      this.currencies.forEach(({ to, from }) => {
        if (from === this.currency.from) currencies.push(to);
      });

      return this.currenciesList.filter((el) =>
        el !== this.currency.from && !currencies.includes(el)
      );
    },
    defaultCurrency() {
      const currency = this.currencies.find((el) =>
        el.rate === 1 && [el.from, el.to].includes('NCU')
      );

      if (!currency) return '';
      return (currency.from === 'NCU') ? currency.to : currency.from;
    }
  }
}
</script>
