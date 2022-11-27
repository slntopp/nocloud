<template>
  <v-row class="mt-4">
    <v-col cols="12">
      <nocloud-table
        :show-select="false"
        :items="plans"
        :headers="headers"
        :loading="isPlansLoading"
        :footer-error="fetchError"
      >
        <template v-slot:[`item.price`]="{ value }">
          {{ value }} {{ 'NCU' }}
        </template>
        <template v-slot:[`item.windows`]="{ value }">
          <template v-if="value">{{ value }} {{ 'NCU' }}</template>
          <template v-else>none</template>
        </template>
        <template v-slot:[`item.period`]="{ item }">
          {{ date(item.period) }}
        </template>
      </nocloud-table>
    </v-col>
    <v-col cols="12">
      <nocloud-table
        :show-select="false"
        :items="plan.resources"
        :headers="addonsHeaders"
        :loading="isPlansLoading"
        :footer-error="fetchError"
      >
        <template v-slot:[`item.price`]="{ value }">
          {{ value }} {{ 'NCU' }}
        </template>
        <template v-slot:[`item.period`]="{ item }">
          {{ date(item.period) }}
        </template>
      </nocloud-table>
    </v-col>
  </v-row>
</template>

<script>
import nocloudTable from '@/components/table.vue';

export default {
  components: { nocloudTable },
  props: { plan: { type: Object, required: true } },
  data: () => ({
    headers: [
      { text: 'OVH plan', value: 'title' },
      { text: 'Price', value: 'price' },
      { text: 'Windows', value: 'windows' },
      { text: 'Period', value: 'period' }
    ],
    addonsHeaders: [
      { text: 'Addon', value: 'key' },
      { text: 'Price', value: 'price' },
      { text: 'Period', value: 'period' }
    ],
    fetchError: ''
  }),
  methods: {
    date(timestamp) {
      if (timestamp / 3600 / 24 / 30 > 1) return '1 year';
      else return '1 month';
    }
  },
  computed: {
    isPlansLoading() {
      return this.$store.getters['plans/isLoading'];
    },
    plans() {
      const products = Object.values(this.plan.products);

      return products.map((el) => {
        const key = Object.keys(el.meta).find((el) => el.includes('windows'));

        return { ...el, windows: el.meta[key] };
      });
    }
  }
}
</script>
