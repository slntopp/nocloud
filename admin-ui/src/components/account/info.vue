<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-text-field v-model="title" label="name" style="width: 330px" />
    <v-card-title class="px-0">SSH keys:</v-card-title>

    <div class="pt-4">
      <v-menu
        bottom
        offset-y
        transition="slide-y-transition"
        v-model="isVisible"
        :close-on-content-click="false"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn class="mr-2" v-bind="attrs" v-on="on">
            Create
          </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row>
            <v-col>
              <v-text-field
                dense
                label="title"
                v-model="newKey.title"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-textarea
                dense
                label="key"
                v-model="newKey.value"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-btn @click="addKey">
                Send
              </v-btn>
            </v-col>
          </v-row>
        </v-card>
      </v-menu>

      <v-btn
        class="mr-8"
        :disabled="selected.length < 1"
        @click="deleteKeys"
      >
        Delete
      </v-btn>
    </div>

    <nocloud-table
      class="mt-4"
      item-key="value"
      v-model="selected"
      :items="keys"
      :headers="headers"
    />

    <v-btn
      class="mt-4"
      :loading="isEditLoading"
      @click="editAccount"
    >
      Submit
    </v-btn>

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
  </v-card>
</template>

<script>
import config from '@/config.js';
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";

export default {
  name: 'account-info',
  components: { nocloudTable },
  mixins: [snackbar],
  props: ['account'],
  data: () => ({
    newKey: { title: '', value: '' },
    headers: [
      { text: 'Title', value: 'title' },
      { text: 'Key', value: 'value' }
    ],
    generalRule: [v => !!v || 'Required field'],
    navTitles: config.navTitles ?? {},

    title: '',
    keys: [],
    selected: [],
    isVisible: false,
    isEditLoading: false
  }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    addKey() {
      this.keys.push(this.newKey);
      this.isVisible = false;
      this.newKey = { title: '', value: '' };
    },
    deleteKeys() {
      if (this.selected.length < 1) return;
      const arr = this.selected.map(
        (el) => el.value
      );

      this.keys = this.keys.filter((el) =>
        !arr.includes(el.value)
      );
      this.selected = [];
    },
    editAccount() {
      const newAccount = { ...this.account, title: this.title };
      newAccount.data.ssh_keys = this.keys;

      this.isEditLoading = true;
      api.accounts.update(this.account.uuid, newAccount)
        .then(() => {
          this.showSnackbarSuccess({
            message: 'Account edited successfully'
          });

          setTimeout(() => {
            this.$router.push({ name: 'Accounts' });
          }, 1500);
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isEditLoading = false;
        });
    }
  }
}
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
