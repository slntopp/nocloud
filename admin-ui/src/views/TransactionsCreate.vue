<template>
  <div class="pa-4">
    <h1 class="page__title">Create transaction</h1>
    <v-form v-model="isValid" ref="form">
      <v-row>
        <v-col lg="6" cols="12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Account</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-select
                label="Account"
                v-model="transaction.account"
                :items="accounts.map((acc) => acc.title)"
                :rules="generalRule"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Service</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-select
                label="Service"
                v-model="transaction.service"
                :items="services.map((service) => service.title)"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Amount</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                type="number"
                label="Amount"
                v-model="transaction.total"
                :rules="generalRule"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Date</v-subheader>
            </v-col>
            <v-col
              cols="4"
              v-for="type of [date, time]"
              :key="type.title"
            >
              <v-menu
                offset-y
                min-width="auto"
                transition="scale-transition"
                v-model="type.visible"
                :ref="`menu${type.title}`"
                :close-on-content-click="false"
                :return-value.sync="type.value"
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-text-field
                    readonly
                    v-model="type.value"
                    v-bind="attrs"
                    v-on="on"
                    :label="type.title"
                  />
                </template>
                <v-date-picker
                  no-title
                  scrollable
                  v-model="type.value"
                  v-if="date.visible"
                >
                  <v-spacer />
                  <v-btn
                    text
                    color="primary"
                    @click="type.visible = false"
                  >
                    Cancel
                  </v-btn>
                  <v-btn
                    text
                    color="primary"
                    @click="$refs.menuDate[0].save(type.value)"
                  >
                    OK
                  </v-btn>
                </v-date-picker>
                <v-time-picker
                  use-seconds
                  format="24hr"
                  v-if="time.visible"
                  v-model="type.value"
                  @click:second="$refs.menuTime[0].save(type.value)"
                />
              </v-menu>
            </v-col>
          </v-row>

          <v-row>
            <v-col cols="3">
              <v-subheader>Meta</v-subheader>
            </v-col>
            <v-col cols="9">
              <json-editor
                :json="transaction.meta"
                @changeValue="(data) => transaction.meta = data"
              />
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      
      <v-row>
        <v-col>
          <v-btn
            class="mr-2"
            color="background-light"
            :loading="isLoading"
            @click="tryToSend"
          >
            Create
          </v-btn>
        </v-col>
      </v-row>
    </v-form>

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
import api from '@/api.js';
import snackbar from '@/mixins/snackbar.js';
import JsonEditor from '@/components/JsonEditor.vue';

export default {
  components: { JsonEditor },
  name: 'transactionsCreate-view',
  mixins: [snackbar],
  data: () => ({
    transaction: {
      account: '',
      service: '',
      total: '',
      exec: 0,
      meta: {}
    },
    date: {
      title: 'Date',
      value: '',
      visible: false
    },
    time: {
      title: 'Time',
      value: '',
      visible: false
    },
    generalRule: [v => !!v || 'This field is required!'],

    isValid: false,
    isLoading: false,
  }),
  methods: {
    tryToSend() {
      if (!this.isValid) {
        this.$refs.form.validate();

        this.showSnackbarError({
          message: 'Validation failed!',
        });
        return;
      }

      this.isLoading = true;
      this.refreshData();

      api.transactions.create(this.transaction)
        .then(() => {
          this.showSnackbarSuccess({
            message: 'Transaction created successfully'
          });

          setTimeout(() => {
            this.$router.push({ name: 'Transactions' });
          }, 1500);
        })
        .catch((err) => {
          this.showSnackbarError({
              message: err,
          });
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    refreshData() {
      this.transaction.account = this.accounts.find((acc) =>
        acc.title === this.transaction.account
      ).uuid;

      this.transaction.service = this.services.find((service) =>
        service.title === this.transaction.service
      )?.uuid || '';

      this.transaction.exec = this.exec;
      this.transaction.total *= 1;
    }
  },
  created() {
    const date = new Date();
    const day = date.getDate();
    const month = date.getMonth() + 1;
    const year = date.getFullYear();
    const time = date.toString().split(' ')[4];

    this.date.value = `${year}-${month}-${day}`;
    this.time.value = `${time}`;
    
    this.$store.dispatch('services/fetch')
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
  computed: {
    accounts() {
      return this.$store.getters['accounts/all'];
    },
    services() {
      return this.$store.getters['services/all'];
    },
    exec() {
      return new Date(`${
        this.date.value
      }T${
        this.time.value
      }`).getTime() / 1000;
    }
  }
}
</script>

<style scoped>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>