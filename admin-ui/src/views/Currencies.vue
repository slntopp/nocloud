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
          <v-col>
            <v-btn @click="addCurrency">Save</v-btn>
          </v-col>
        </v-row>
      </v-card>
    </v-dialog>

    <confirm-dialog @confirm="deleteSelectedCurrencies">
      <v-btn color="background-light" :disabled="selected.length < 1">
        Delete
      </v-btn>
    </confirm-dialog>

    <nocloud-table
      class="mt-4"
      item-key="id"
      :items="currencies"
      :headers="headers"
      :loading="isLoading"
      :footer-error="fetchError"
      v-model="selected"
    />

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
    ],
    currenciesList: ['NCU', 'USD', 'EUR', 'BYN', 'PLN'],
    currencies: [],
    selected: [],
    currency: { from: "", to: "" },
    isLoading: false,
    fetchError: "",
  }),
  methods: {
    addCurrency() {
      this.currencies.push({
        from: this.currency.from,
        to: this.currency.to,
        rate: Math.round(Math.random() * 100),
        id: `${this.currency.from} ${this.currency.to}`
      });

      this.currency = { from: "", to: "" };
    },
    deleteSelectedCurrencies() {
      this.currencies = this.currencies.filter(({ id }) =>
        !this.selected.find((el) => el.id === id)
      );
      this.selected = [];
    }
  },
  created() {
    api.get('/billing/currencies')
      .then((res) => console.log(res))
      .catch((err) => console.error(err));
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
    }
  }
}
</script>
