<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Plans' }">{{ navTitle('Plans') }}</router-link>
      /{{ planTitle }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="tabs"
    >
      <v-tab>Info</v-tab>
      <v-tab>Template</v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="tabs"
    >
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="planLoading" />
        <plans-create v-if="plan" :item="plan" />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="planLoading" />
        <plans-template v-if="plan" :template="plan" />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import config from '@/config.js';
import PlansCreate from '@/views/PlansCreate.vue';
import PlansTemplate from '@/components/plan/template.vue';

export default {
  name: 'plan-view',
  components: { PlansCreate, PlansTemplate },
  data: () => ({ tabs: 0, navTitles: config.navTitles ?? {} }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    }
  },
  computed: {
    plan() {
      return this.$store.getters['plans/one'];
    },
    planTitle() {
      return this?.plan.title ?? 'not found';
    },
    planLoading() {
      return this.$store.getters['plans/loading'];
    },
  },
  created() {
    const id = this.$route.params?.planId;

    this.$store.dispatch('plans/fetchItem', id)
      .then(() => {
        document.title = `${this.planTitle} | NoCloud`;
      });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: 'plans/fetchItem',
      params: this.$route.params?.planId,
    });
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
