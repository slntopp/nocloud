<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Namespaces' }">{{
        navTitle("Namespaces")
      }}</router-link>
      / {{ 'NS_' + (namespaceTitle) }}
    </h1>
    <v-card :loading="isFetchLoading" elevation="0" color="background-light" class="pa-4">
      <v-card-title>Access</v-card-title>
      <v-text-field readonly :value="namespace.access?.level" label="level" style="width: 330px" />
      <v-text-field readonly :value="namespace.access?.role" label="role" style="width: 330px" />
      <v-text-field readonly :value="namespace.access?.namespace" label="namespace" style="width: 330px" />
      <v-text-field class="mt-5" v-model="namespace.title" label="title" style="width: 330px" />
      <div class="pt-4">
        <v-btn class="mt-4 mr-2" :loading="isEditLoading" @click="editNamespace">
          Submit
        </v-btn>
      </div>
    </v-card>
  </div>
</template>

<script>
import api from '@/api.js'
import config from "@/config.js";

export default {
  name: "account-view",
  data: () => ({ navTitles: config.navTitles ?? {}, namespace: {}, namespaceTitle: '...', isFetchLoading: false, isEditLoading: false }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    editNamespace() {
      this.isEditLoading = true

      api.namespaces.edit(this.editableNamespace).then(() => {
        this.namespace = this.editableNamespace
      }).finally(() => {
        this.isEditLoading = false
      })
    }
  },
  computed: {
    all() {
      return this.$store.getters['namespaces/all']
    },
    namespaceId() {
      return this.$route.params.namespaceId
    }
  },
  async mounted() {
    if (!this.all || this.all.length === 0) {
      this.isFetchLoading = true
      await this.$store.dispatch("namespaces/fetch")
      this.isFetchLoading = false
    }

    this.namespace = this.all.find(n => n.uuid == this.namespaceId)
    this.namespaceTitle = this.namespace.title
  },
};
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
