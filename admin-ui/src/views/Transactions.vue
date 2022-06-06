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
      class="d-inline-block ml-2"
      v-model="accountTitle"
      :items="accountsTitles"
    />

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

    <nocloud-table
      class="mt-4"
      sort-by="proc"
      :items="transactions"
      :headers="headers"
      :loading="isLoading"
      :sort-desc="true"
      :show-select="false"
      :footer-error="fetchError"
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
import noCloudTable from '@/components/table.vue';
import balance from '@/components/balance.vue';

export default {
  name: 'transactions-view',
  components: {
    'nocloud-table': noCloudTable,
    balance
  },
  mixins: [snackbar],
  data: () => ({
    headers: [
      { text: 'Account ', value: 'account' },
      { text: 'Service ', value: 'service' },
      { text: 'Amount ', value: 'total' },
      { text: 'Date ', value: 'proc' }
    ],
    accountTitle: '',
    visibleItems: [],
    copyed: -1,
    fetchError: ''
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
    date(timestamp) {
      const date = new Date(timestamp * 1000);
      const time = date.toUTCString().split(' ')[4];
      
      const day = date.getUTCDate();
      const month = date.getUTCMonth() + 1;
      const year = date.toUTCString().split(' ')[3];

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
    }
  },
  created() {
    const { title } = this.$store.getters['auth/userdata'];
    const accounts = this.accounts.map((acc) => acc.uuid);

    this.accountTitle = title;
    this.$store.dispatch('services/fetch')
    this.$store.dispatch('transactions/fetch', accounts)
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
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "transactions/fetch",
      params: this.accounts.map((acc) => acc.uuid)
    });
  },
  computed: {
    transactions() {
      const transactions = this.$store.getters['transactions/all'];
      const account = this.accounts.find((acc) =>
        acc.title === this.accountTitle
      );

      if (this.accountTitle === 'all') {
        return transactions;
      }

      return transactions.filter((item) =>
        item.account === account?.uuid
      );
    },
    isLoading() {
      return this.$store.getters['transactions/isLoading'];
    },
    accounts() {
      return this.$store.getters['accounts/all'];
    },
    accountsTitles() {
      return [...this.accounts.map((acc) => acc.title), 'all'];
    },
    balance() {
      const amount = this.transactions?.length - 11;
      const labels = [];
      const values = [];
      let balance = 0;

      if (this.accountTitle === 'all') {
        return { labels, values };
      }

      this.transactions?.forEach((el) => {
        values.push(balance -= el.total);
        labels.push(`${balance} NCU`);
      });
      values.unshift(0);
      labels.unshift('0 NCU');

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
    accounts() {
      const { title } = this.$store.getters['auth/userdata'];
      const accounts = this.accounts.map((acc) => acc.uuid);

      this.accountTitle = title;
      this.$store.dispatch('transactions/fetch', accounts)
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
    }
  }
}
</script>

<style>

</style>