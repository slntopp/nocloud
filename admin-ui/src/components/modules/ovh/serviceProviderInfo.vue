<template>
  <div>
    <slot></slot>

    <v-card-title class="px-0 mb-3">Projects:</v-card-title>
    <v-row class="flex-column">
      <v-col>
        <nocloud-table
          table-name="ovh-service-provider-info"
          :items="projects"
          :headers="headers"
          :show-select="false"
        >
          <template v-slot:[`item.balance`]="{ value }">
            <balance :value="value.toFixed()" />
          </template>
        </nocloud-table>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import nocloudTable from "@/components/table.vue";
import balance from "@/components/balance.vue";

export default {
  name: "service-provider-ovh",
  components: { nocloudTable, balance },
  props: { template: { type: Object, required: true } },
  data: () => ({
    headers: [
      { text: "Title", value: "desc" },
      { text: "UUID", value: "uuid" },
      { text: "Balance", value: "balance" },
    ],
  }),
  computed: {
    projects() {
      const wilds = this.template.state?.meta?.wilds;

      if (!wilds) return [];
      return Object.entries(wilds).map(([key, value]) => ({
        ...value,
        uuid: key,
      }));
    },
  },
};
</script>
