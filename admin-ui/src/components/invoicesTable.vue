<template>
  <nocloud-table
    table-name="invoices-table"
    class="mt-4"
    sort-by="proc"
    sort-desc
    :show-select="false"
    :items="items"
    :headers="headers"
    :loading="loading"
  >
    <template v-slot:[`item.account`]="{ value }">
      <router-link :to="{ name: 'Account', params: { accountId: value } }">
        {{ account(value) }}
      </router-link>
    </template>

    <template v-slot:[`item.total`]="{ item }">
      <balance abs :currency="item.currency" :value="-item.total" />
    </template>
    <template v-slot:[`item.proc`]="{ item }">
      {{ formatSecondsToDate(item.proc, true) }}
    </template>
    <template v-slot:[`item.exec`]="{ item }">
      {{ formatSecondsToDate(item.exec, true) }}
    </template>
    <template v-slot:[`item.created`]="{ item }">
      {{ formatSecondsToDate(item.created, true) }}
    </template>
    <template v-slot:[`item.status`]="{ item }">
      <v-chip>{{ item.status }}</v-chip>
    </template>
    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon :to="{ name: 'Invoice page', params: { uuid: item.uuid } }">
        <v-icon>mdi-login</v-icon>
      </v-btn>
    </template>
  </nocloud-table>
</template>

<script>
import nocloudTable from "@/components/table.vue";
import balance from "@/components/balance.vue";
import { formatSecondsToDate } from "../functions";

export default {
  name: "invoices-table",
  components: { nocloudTable, balance },
  props: {
    items: { type: Array, required: true },
    loading: { type: Boolean, default: false },
  },
  data: () => ({
    headers: [
      { text: "UUID ", value: "uuid" },
      { text: "Account ", value: "account" },
      { text: "Amount ", value: "total" },
      { text: "Payment date ", value: "proc" },
      { text: "Executed date ", value: "exec" },
      { text: "Created date ", value: "created" },
      { text: "Status ", value: "status" },
      { text: "Actions ", value: "actions" },
    ],
  }),
  methods: {
    formatSecondsToDate,
    account(uuid) {
      return this.accounts.find((acc) => acc.uuid === uuid)?.title;
    },
  },
  computed: {
    accounts() {
      return this.$store.getters["accounts/all"];
    },
    services() {
      return this.$store.getters["services/all"];
    },
  },
};
</script>
