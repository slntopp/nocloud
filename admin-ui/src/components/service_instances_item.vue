<template>
  <nocloud-table
    table-name="services-instances-item"
    :show-select="false"
    :items="instances"
    :headers="headers"
  >
    <template v-slot:[`item.title`]="{ item }">
      <router-link
        :to="{ name: 'Instance', params: { instanceId: item.uuid } }"
      >
        {{ item.title }}
      </router-link>
    </template>

    <template v-slot:[`item.resources.cpu`]="{ value }">
      {{ value }} {{ value > 1 ? "cores" : "core" }}
    </template>

    <template v-slot:[`item.resources.ram`]="{ value }">
      {{ driveSize(value) }} GB
    </template>

    <template v-slot:[`item.resources.drive_size`]="{ value }">
      {{ driveSize(value) }} GB
    </template>

    <template v-slot:[`item.config.template_id`]="{ value }">
      {{ getOSName(value, spId) }}
    </template>

    <template v-slot:[`item.config.planCode`]="{ item }">
      {{ getTariff(item) }}
    </template>

    <template v-slot:[`item.resources.period`]="{ value }">
      {{ value }} {{ value > 1 ? "months" : "month" }}
      <!--  {{ (item.type === 'goget') ? 'years' : 'months' }} -->
    </template>

    <template v-slot:[`item.state.meta.networking`]="{ item }">
      <template v-if="!item.state?.meta.networking?.public">-</template>
      <v-menu
        bottom
        open-on-hover
        v-else
        nudge-top="20"
        nudge-left="15"
        transition="slide-y-transition"
      >
        <template v-slot:activator="{ on, attrs }">
          <span v-bind="attrs" v-on="on">
            {{ item.state.meta.networking.public[0] }}
          </span>
        </template>

        <v-list dense>
          <v-list-item
            v-for="net of item.state.meta.networking.public"
            :key="net"
          >
            <v-list-item-title>{{ net }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </template>
  </nocloud-table>
</template>

<script>
import nocloudTable from "@/components/table.vue";

export default {
  name: "serviceInstancesItem",
  components: { nocloudTable },
  props: {
    instances: { type: Array, default: () => [] },
    spId: { type: String, default: () => "" },
    type: { type: String, default: () => "" },
  },
  methods: {
    driveSize(data) {
      return +(data / 1024).toFixed(2);
    },
    getOSName(id, sp) {
      if (!id || !sp) return;
      return this.sp.find(({ uuid }) => uuid === sp).publicData.templates[id]
        .name;
    },
    getTariff(item) {
      const {
        billingPlan,
        config: { planCode, duration },
      } = item;
      const key = `${duration} ${planCode}`;

      return billingPlan.products[key].title;
    },
  },
  computed: {
    headers() {
      const headers = [{ text: "Title", value: "title" }];

      switch (this.type) {
        case "ione":
          headers.push(
            { text: "CPU", value: "resources.cpu" },
            { text: "RAM", value: "resources.ram" },
            { text: "Disk", value: "resources.drive_size" },
            { text: "OS", value: "config.template_id" },
            { text: "IP", value: "state.meta.networking" }
          );
          break;
        case "ovh":
          headers.push(
            { text: "Tariff", value: "config.planCode" },
            { text: "IP", value: "state.meta.networking" },
            { text: "Creation", value: "data.creation" },
            { text: "Expiration", value: "data.expiration" }
          );
          break;
        case "goget":
        case "opensrs":
          headers.push(
            { text: "Domain", value: "resources.domain" },
            { text: "Period", value: "resources.period" }
          );

          if (this.type === "opensrs") break;
          else
            headers.push(
              { text: "DCV", value: "resources.dcv" },
              { text: "Email", value: "resources.approver_email" }
            );
      }

      return headers;
    },
    sp() {
      return this.$store.getters["servicesProviders/all"];
    },
  },
};
</script>

<style scoped>
.v-card__text .v-chip {
  width: 48%;
  display: flex;
  justify-content: center;
  align-items: center;
}
.v-card__text .hash {
  width: 100%;
}
</style>
