<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Namespaces' }">{{
        navTitle("Namespaces")
      }}</router-link>
      / {{ 'NS_' + (namespaceTitle) }}
    </h1>
    <v-tabs class="rounded-t-lg" background-color="background-light" v-model="tabs">
      <v-tab>Info</v-tab>
    </v-tabs>
    <v-tabs-items class="rounded-b-lg" style="background: var(--v-background-light-base)" v-model="tabs">
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isFetchLoading" />
        <namespace-info :loading="isFetchLoading" v-if="namespace && namespaceTitle" :namespace="namespace"
          @input:title="namespace.title = $event" />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import api from '@/api.js'
import config from "@/config.js";
import namespaceInfo from '../components/namespace/info.vue';

export default {
  name: "account-view",
  data: () => ({ navTitles: config.navTitles ?? {}, isFetchLoading: false, tabs: 0, namespace: {} }),
  components: { namespaceInfo },
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    editNamespace() {
      this.isEditLoading = true

      api.namespaces.edit(this.namespace).finally(() => {
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
    },
    namespaceTitle() {
      if (!this.namespace || !this.namespace.title) {
        return '...'
      }

      return this.namespace.title
    }
  },
  async mounted() {
    if (!this.all || this.all.length === 0) {
      this.isFetchLoading = true
      await this.$store.dispatch("namespaces/fetch")
      this.isFetchLoading = false
    }

    this.namespace = this.all.find(n => n.uuid == this.namespaceId)
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
