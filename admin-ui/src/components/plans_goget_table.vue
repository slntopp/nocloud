<template>
  <v-row class="mt-4">
    <v-col :cols="12" :md="6">
      <nocloud-table
        :show-select="false"
        :items="products"
        :headers="headers"
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
      { text: 'Product', value: 'title' },
      { text: 'Price', value: 'price' },
      { text: 'Period', value: 'period' }
    ],
    fetchError: ''
  }),
  methods: {
    date(timestamp) {
      return `${timestamp / 3600 / 24 / 30} months`;
    }
  },
  computed: {
    isPlansLoading() {
      return this.$store.getters['plans/isLoading'];
    },
    products() {
      return Object.values(this.plan.products);
    }
  }
}
</script>
