<template>
  <div>
    <v-card-title class="px-0 mb-3">Nodes</v-card-title>
    <v-row class="mb-7">
      <v-col
        v-for="(host, idx) in hosts"
        :key="idx"
        cols="12"
      >
        <div class="d-flex justify-space-between">
          <span>{{ host.name }}</span>
          <span>{{ host.status }}</span>
        </div>
        <v-progress-linear
          class="mt-2"
          :value="ramPercent(host)"
          color="green"
          height="16"
        >
          RAM {{ ramPercent(host) }}%
        </v-progress-linear>
      </v-col>
    </v-row>
    <v-card-title class="px-0 mb-3">Templates</v-card-title>
    <v-simple-table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Name</th>
          <th>Node</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="t in templates" :key="t.id">
          <td>{{ t.id }}</td>
          <td>{{ t.name }}</td>
          <td>{{ t.node }}</td>
        </tr>
      </tbody>
    </v-simple-table>
    <slot></slot>
  </div>
</template>

<script>
export default {
  name: "service-provider-proxmox",
  props: { template: { type: Object, required: true } },
  computed: {
    hosts() {
      return this.template.state?.meta?.hosts || [];
    },
    templates() {
      const t = this.template.publicData?.templates;
      return Array.isArray(t) ? t : Object.values(t || {});
    },
  },
  methods: {
    ramPercent(host) {
      if (!host.max_ram) return 0;
      return Math.ceil((host.ram_usage / host.max_ram) * 100);
    },
  },
};
</script>
