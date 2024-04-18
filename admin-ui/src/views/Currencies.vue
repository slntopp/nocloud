<template>
  <div class="pa-4">
    <div class="buttons__inline pb-2 mt-4">
      <v-dialog max-width="400">
        <template v-slot:activator="{ on, attrs }">
          <v-btn class="mr-2" color="background-light" v-bind="attrs" v-on="on">
            Add
          </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row dense>
            <v-col cols="12">
              <v-autocomplete
                dense
                label="Currency 1"
                v-model="currency.from"
                item-text="title"
                item-value="id"
                return-object
                :items="currenciesFrom"
              />
            </v-col>
            <v-col cols="12">
              <v-autocomplete
                dense
                label="Currency 2"
                v-model="currency.to"
                item-text="title"
                item-value="id"
                return-object
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
              <v-btn :loading="isCreateLoading" @click="addCurrency"
                >Save</v-btn
              >
            </v-col>
          </v-row>
        </v-card>
      </v-dialog>
      <confirm-dialog
        :disabled="selected.length < 1"
        @confirm="deleteSelectedCurrencies"
      >
        <v-btn
          class="mr-2"
          color="background-light"
          :disabled="selected.length < 1"
        >
          Delete
        </v-btn>
      </confirm-dialog>

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
      table-name="currencys"
      class="mt-4"
      item-key="id"
      v-model="selected"
      :items="currencies"
      :headers="headers"
      :loading="isLoading"
      :footer-error="fetchError"
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
      <template v-slot:[`item.actions`]="{ item }">
        <v-icon title="Save edit" @click="editCurrency(item)">
          mdi-content-save-edit-outline
        </v-icon>
      </template>
    </nocloud-table>
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
    selected: [],

    rules: {
      number: [
        (value) =>
          /^[-+]?[0-9]*[.,]?[0-9]+(?:[eE][-+]?[0-9]+)?$/.test(value) ||
          "Invalid!",
      ],
    },

    currency: { from: "", to: "", rate: "1" },
    isCreateLoading: false,
    fetchError: "",
  }),
  methods: {
    addCurrency() {
      if (!this.currency.from || !this.currency.to) return;
      if (typeof this.rules.number[0](this.currency.rate) === "string") return;

      const newCurrency = {
        rate: +this.currency.rate.replace(",", "."),
        from: this.currency.from,
        to: this.currency.to,
      };

      this.isCreateLoading = true;
      api
        .post("/billing/currencies/rates", newCurrency)
        .then(() => {
          this.$store.dispatch("currencies/fetch");
          this.currency = { from: "", to: "", rate: "1" };
        })
        .catch((err) => {
          const message = err.response?.data?.message ?? err.message ?? err;

          this.showSnackbarError({ message });
          console.error(err);
        })
        .finally(() => (this.isCreateLoading = false));
    },
    editCurrency(currency) {
      if (typeof this.rules.number[0](currency.rate) === "string") return;
      const newCurrency = {
        rate: +currency.rate.replace(",", "."),
        from: currency.from,
        to: currency.to,
      };

      this.$store.commit("currencies/setLoading", true);
      api
        .put("/billing/currencies/rates", newCurrency)
        .then(() => {
          this.showSnackbarSuccess({ message: "Done" });
        })
        .catch((err) => {
          const message = err.response?.data?.message ?? err.message ?? err;

          this.showSnackbarError({ message });
          console.error(err);
        })
        .finally(() => {
          this.$store.commit("currencies/setLoading", false);
        });
    },
    deleteSelectedCurrencies() {
      const promises = this.selected
        .filter(({ id }) => this.currencies.find((el) => el.id === id))
        .map(({ from, to }) =>
          api.delete(`/billing/currencies/rates/rate?from.id=${from.id}&to.id=${to.id}`)
        );

      Promise.all(promises)
        .then(() => {
          this.$store.dispatch("currencies/fetch");
        })
        .catch((err) => {
          const message = err.response?.data?.message ?? err.message ?? err;

          this.showSnackbarError({ message });
          console.error(err);
        });
    },
  },
  created() {
    this.$store.dispatch("currencies/fetch").catch((err) => {
      const message = err.response?.data?.message ?? err.message ?? err;

      this.showSnackbarError({ message });
      console.error(err);
    });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "currencies/fetch",
    });
  },
  computed: {
    isLoading() {
      return this.$store.getters["currencies/isLoading"];
    },
    currencies() {
      return this.$store.getters["currencies/rates"];
    },
    currenciesList() {
      return this.$store.getters["currencies/all"];
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    currenciesFrom() {
      const currencies = [];

      this.currencies.forEach(({ from, to }) => {
        if (to === this.currency.to) currencies.push(from);
      });

      return this.currenciesList.filter(
        (el) => el !== this.currency.to && !currencies.includes(el)
      );
    },
    currenciesTo() {
      const currencies = [];

      this.currencies.forEach(({ to, from }) => {
        if (from === this.currency.from) currencies.push(to);
      });

      return this.currenciesList.filter(
        (el) => el !== this.currency.from && !currencies.includes(el)
      );
    },
  },
};
</script>
