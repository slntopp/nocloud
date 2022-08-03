<template>
  <div class="pa-4">
    <v-btn
      class="mr-2"
      color="background-light"
      :to="{ name: 'Transactions create' }"
    >
      Create
    </v-btn>

    <v-select
      label="Account"
      item-text="title"
      item-value="uuid"
      class="d-inline-block mr-2"
      v-model="accountId"
      :items="accounts"
    />
    <v-select
      label="Service"
      item-text="title"
      item-value="uuid"
      class="d-inline-block"
      v-model="serviceId"
      :items="services"
    />

    <v-progress-linear indeterminate class="pt-1" v-if="chartLoading" />
    <template v-else-if="series.length < 1">
      <v-subheader v-if="balance.values.length > 1">
        Balance:
      </v-subheader>
      <v-sparkline
        color="primary"
        height="25vh"
        line-width="1"
        label-size="4"
        :labels="balance.labels"
        :value="balance.values"
      />
    </template>
    <apexcharts
      v-else
      type="line"
      height="250"
      :options="chartOptions"
      :series="series"
    />

    <nocloud-table
      class="mt-4"
      sort-by="proc"
      :items="transactions"
      :headers="headers"
      :loading="isLoading"
      :sort-desc="true"
      :footer-error="fetchError"
      @input="selectTransaction"
    >
      <template v-slot:[`item.account`]="{ item }">
        {{ account(item.account) }}
      </template>

      <template v-slot:[`item.service`]="{ item, index }">
        <template v-if="item.service">
          <router-link
            :to="{ name: 'Service', params: { serviceId: item.service } }"
          >
            {{ service(item.service) }}
          </router-link>

          <v-icon
            class="ml-2"
            v-if="!visibleItems.includes(index)"
            @click="visibleItems.push(index)"
          >
            mdi-eye-outline
          </v-icon>
          <template v-else>
            ({{ hashTrim(item.service) }})
            <v-btn icon @click="addToClipboard(item.service, index)">
              <v-icon v-if="copyed === index"> mdi-check </v-icon>
              <v-icon v-else> mdi-content-copy </v-icon>
            </v-btn>
          </template>
        </template>
        <template v-else>-</template>
      </template>

      <template v-slot:[`item.total`]="{ item }">
        <balance :value="-item.total" />
      </template>
      <template v-slot:[`item.proc`]="{ item }">
        {{ date(item.proc) }}
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
import snackbar from '@/mixins/snackbar.js';
import nocloudTable from '@/components/table.vue';
import balance from '@/components/balance.vue';
import apexcharts from 'vue-apexcharts';

export default {
  name: 'transactions-view',
  components: { nocloudTable, balance, apexcharts },
  mixins: [snackbar],
  data: () => ({
    headers: [
      { text: 'Account ', value: 'account' },
      { text: 'Service ', value: 'service' },
      { text: 'Amount ', value: 'total' },
      { text: 'Date ', value: 'proc' }
    ],
    accountId: null,
    serviceId: null,
    visibleItems: [],
    selected: [],
    copyed: -1,
    fetchError: '',

    series: [],
    chartLoading: false,
    chartOptions: {
      chart: { height: 250, type: 'line' },
      dataLabels: { enabled: false },
      stroke: { curve: 'smooth' },
      xaxis: { type: 'datetime', categories: [] },
      tooltip: { x: { format: 'dd.MM.yy HH:mm' } },
      theme: { palette: 'palette10', mode: 'dark' },
      legend: { showForSingleSeries: true }
    }
  }),
  methods: {
    account(uuid) {
      return this.accounts.find((acc) =>
        acc.uuid === uuid
      )?.title;
    },
    service(uuid) {
      const service = this.$store.getters['services/all']
        .find((serv) => serv.uuid === uuid);

      return service?.title;
    },
    date(timestamp, bool) {
      const date = new Date(timestamp * 1000);
      const time = date.toUTCString().split(' ')[4];
      
      const day = date.getUTCDate();
      const month = date.getUTCMonth() + 1;
      const year = date.toUTCString().split(' ')[3];

      if (bool) return `${year}-${month}-${day}T${time}Z`;
      return `${day}.${month}.${year} ${time}`;
    },
    hashTrim(hash) {
      if (hash) return ` ${hash.slice(0, 12)}... `;
      else return ' XXXXXXXX... ';
    },
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((err) => {
            this.showSnackbarError({
              message: err
            });
          });
      } else {
        this.showSnackbarError({
          message: 'Clipboard is not supported!'
        });
      }
    },
    getTransactions() {
      const accounts = [];

      this.accounts.forEach((acc) => {
        if (acc.uuid) accounts.push(acc.uuid);
      });

      this.$store.dispatch('services/fetch')
      this.$store.dispatch('transactions/fetch', { accounts, service: this.serviceId })
        .then(() => {
          this.fetchError = '';
        })
        .catch((err) => {
          console.error(err);

          this.fetchError = 'Can\'t reach the server';
          if (err.response) {
            this.fetchError += `: [ERROR]: ${err.response.data.message}`;
          } else {
            this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
          }
        });
    },
    setTransactions(dates, labels, values) {
      const min = Math.min(...dates);
      let counter = 1;

      for (let i = 1; i < dates.length; i++) {
        const curr = Math.round(dates[i] / min);
        const spaces = (curr < 10) ? curr : 9;

        if (spaces < 2) continue;
        const newValues = values.splice(i + counter);
        const newLabels = labels.splice(i + counter);
        const diff = (newValues[0] - values.at(-1)) / spaces;

        for (let j = 0; j < spaces - 1; j++) {
          const prev = values[i + j + counter - 1];

          values[i + j + counter] = prev + diff;
          labels[i + j + counter] = ' ';
        }

        counter += spaces - 1;
        values = values.concat(newValues);
        labels = labels.concat(newLabels);
      }

      return [labels, values];
    },
    selectTransaction(value) {
      this.series = [];
      this.chartLoading = true;
      this.chartOptions.xaxis.categories = [];

      value.forEach(({ total, service, proc }) => {
        const name = service.slice(0, 8);
        const data = { data: [total], name, service };
        const i = this.series.findIndex(
          (item) => item.name === name
        );

        if (i !== -1) {
          this.series[i].data.push(total);
        } else {
          this.series.push(data);
        }

        this.chartOptions.xaxis.categories
          .push(this.date(proc, true));
      });
      setTimeout(() => { this.chartLoading = false }, 300);

      if (this.series.length < 1) {
        this.showSnackbar({
          message: 'Records not found',
          buttonColor: 'white',
          color: 'blue darken-3'
        });
      }
    },
    setListenerToLegend() {
      const legend = document.querySelectorAll('.apexcharts-legend-text');

      legend.forEach((el) => {
        el.addEventListener('click', (e) => {
          const { service } = this.series.find((item) =>
            item.name === e.target.innerText
          );

          this.$router.push({
            name: 'Service',
            params: { serviceId: service }
          });
        });
      });
    }
  },
  mounted() {
    const accounts = [];
    if (!this.$store.getters['transactions/all'].length) {
      this.getTransactions();
    }

    this.accountId = this.user.uuid || null;
    this.accounts.forEach((acc) => {
      if (acc.uuid) accounts.push(acc.uuid);
    });

    this.$store.commit("reloadBtn/setCallback", {
      type: "transactions/fetch",
      params: { accounts, service: this.serviceId }
    });
  },
  computed: {
    transactions() {
      const transactions = this.$store.getters['transactions/all'];

      if (!this.accountId && !this.serviceId) {
        return transactions;
      }

      return transactions
        .filter((item) => {
          const equalAccounts = item.account === this.accountId;
          const equalServices = item.service === this.serviceId;

          if (!this.accountId) return equalServices;
          else if (!this.serviceId) return equalAccounts;
          else return equalAccounts && equalServices;
        });
    },
    isLoading() {
      return this.$store.getters['transactions/isLoading'];
    },
    user() {
      return this.$store.getters['auth/userdata'];
    },
    accounts() {
      const accounts = this.$store.getters['accounts/all'];

      return [...accounts, { title: 'all', uuid: null }];
    },
    services() {
      const services = this.$store.getters['services/all'].map((el) => ({
        title: `${el.title} (${el.uuid.slice(0, 8)})`,
        uuid: el.uuid
      }));

      return [...services, { title: 'all', uuid: null }];
    },
    balance() {
      const dates = [];
      let labels = ['0 NCU'];
      let values = [0];
      let balance = 0;

      if (!this.accountId) {
        return { labels, values };
      }

      this.transactions?.forEach((el, i, arr) => {
        values.push(balance -= el.total);
        labels.push(`${balance} NCU`);
        dates.push(el.proc - arr[i - 1]?.proc ||
          arr[i + 1].proc - el.proc);
      });

      [labels, values] = this.setTransactions(dates, labels, values);

      const amount = values.length - 12;
      return {
        labels: (amount > 0) ? labels.slice(amount) : labels,
        values: (amount > 0) ? values.slice(amount) : values
      };
    }
  },
  watch: {
    transactions() {
      this.fetchError = '';
    },
    chartLoading() {
      setTimeout(this.setListenerToLegend);
    },
    user() {
      this.accountId = this.user.uuid;
    }
  }
}
</script>

<style>
.apexcharts-svg {
  background: none !important;
}
</style>
