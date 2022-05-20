<template>
  <div class="pa-4">
    <v-btn
      class="mr-2"
      color="background-light"
      :to="{ name: 'Plans create' }"
    >
      Create
    </v-btn>
    <v-btn
      color="background-light"
      :loading="isDeleteLoading"
      @click="deleteSelectedPlans"
    >
      Delete
    </v-btn>

    <nocloud-table
      class="mt-4"
      :items="plans"
      :headers="headers"
      :value="selected"
      :loading="isLoading"
      :footer-error="fetchError"
      @input="(v) => (selected = v)"
    >
      <template v-slot:[`item.title`]="{ item }">
        <router-link
          :to="{ name: 'Plan', params: { planId: item.uuid } }"
        >
          {{ item.title }}
        </router-link>
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
import api from '@/api.js';
import snackbar from '@/mixins/snackbar.js';
import noCloudTable from '@/components/table.vue';

export default {
  name: 'plans-view',
  components: {
    'nocloud-table': noCloudTable
  },
  mixins: [snackbar],
  data: () => ({
    headers: [
      { text: 'Title ', value: 'title' },
      { text: 'UUID ', value: 'uuid' },
      { text: 'Public ', value: 'public' },
      { text: 'Type ', value: 'type' }
    ],
    isDeleteLoading: false,
    selected: [],
    copyed: -1,
    fetchError: ''
  }),
  methods: {
    deleteSelectedPlans() {
      if (this.selected.length > 0) {
        this.isDeleteLoading = true;

        const deletePromises = this.selected.map((el) =>
          api.plans.delete(el.uuid)
        );
        Promise.all(deletePromises)
          .then(() => {
            const ending = deletePromises.length === 1 ? '' : 's';

              this.$store.dispatch('plans/fetch');
              this.showSnackbar({
                message: `Plan${ending} deleted successfully.`,
              });
          })
          .catch((err) => {
            if (err.response.status >= 500 || err.response.status < 600) {
              this.showSnackbarError({
                message: `Plan Unavailable: ${
                  err?.response?.data?.message ?? 'Unknown'
                }.`,
                timeout: 0,
              });
            } else {
              this.showSnackbarError({
                message: `Error: ${err?.response?.data?.message ?? 'Unknown'}.`,
              });
            }
          })
          .finally(() => {
            this.isDeleteLoading = false;
          });
      }
    }
  },
  created() {
    this.$store.dispatch('plans/fetch')
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
    plans() {
      return this.$store.getters['plans/all'];
    },
    isLoading() {
      return this.$store.getters['plans/isLoading'];
    }
  },
  watch: {
    plans() {
      this.fetchError = '';
    }
  }
}
</script>
