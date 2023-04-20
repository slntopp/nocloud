<template>
  <div class="pa-4">
    <div class="d-flex">
      <h1 class="page__title">Showcases</h1>
    </div>

    <div v-for="provider of sp" :key="provider.uuid">
      <v-card-title>{{ provider.title }}:</v-card-title>
      <showcase
        :template="provider"
        :is-display="true"
        @update:showcase="updateShowcase"
      />
    </div>

    <v-btn class="mt-4" :isLoading="isLoading" @click="tryToSend">Save</v-btn>

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
import snackbar from "@/mixins/snackbar.js";
import showcase from '@/components/ServicesProvider/showcase.vue';

export default {
  name: 'showcases-view',
  mixins: [snackbar],
  components: { showcase },
  data: () => ({ updated: [], isLoading: false }),
  methods: {
    updateShowcase(provider) {
      this.$store.commit('servicesProviders/updateService', provider);

      if (this.updated.find(({ uuid }) => uuid === provider.uuid)) return;
      this.updated.push(provider);
    },
    tryToSend() {
      const promises = this.updated.map((provider) =>
        api.servicesProviders.update(provider.uuid, provider)
      );

      this.isLoading = true;
      Promise.all(promises).then(() => {
        this.showSnackbarSuccess({
          message: "Showcases changed successfully"
        });
      })
      .catch((err) => {
        this.showSnackbarError({ message: err });
        console.error(err);
      })
      .finally(() => {
        this.isLoading = false;
      });
    }
  },
  created() {
    if (this.sp.length > 0) return;
    this.$store.dispatch("servicesProviders/fetch")
      .catch((err) => {
        this.showSnackbarError({ message: err });
        console.error(err);
      });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "servicesProviders/fetch",
    });
  },
  computed: {
    sp() {
      return this.$store.getters['servicesProviders/all'];
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
