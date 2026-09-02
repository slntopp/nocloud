<template>
  <div>
    <v-alert v-if="meta.error" type="error" dense>{{ meta.error }}</v-alert>
    <v-row class="mb-2">
      <v-col cols="12" md="3">
        <v-text-field readonly label="PVE version" :value="publicData.pve_version || meta.pve_version" />
      </v-col>
      <v-col cols="12" md="3">
        <v-text-field readonly label="Cluster" :value="clusterLabel" />
      </v-col>
      <v-col cols="12" md="3">
        <v-text-field readonly label="Nodes" :value="nodes.length" />
      </v-col>
      <v-col cols="12" md="3">
        <v-text-field readonly label="Templates" :value="templates.length" />
      </v-col>
    </v-row>
    <slot></slot>

    <!-- Nodes -->
    <v-card-title class="px-0 mb-3">Nodes:</v-card-title>
    <v-row class="mb-7">
      <v-col v-for="node in nodes" :key="node.name" cols="12">
        <v-row>
          <v-col class="order-2 order-lg-1" cols="12" lg="6">
            <div class="title_progress">
              <span>Memory, MB</span>
              <div>
                <span>{{ node.used_ram }}</span> / <span>{{ node.total_ram }}</span>
              </div>
            </div>
            <v-progress-linear :value="percent(node.used_ram, node.total_ram)" color="green" height="20">
              <template v-slot:default="{ value }"><strong>{{ value }}%</strong></template>
            </v-progress-linear>
            <div class="title_progress mt-3">
              <span>CPU load</span>
              <div>
                <span>{{ (+node.used_cpu || 0).toFixed(1) }}</span> / <span>{{ node.total_cpu }}</span>
              </div>
            </div>
            <v-progress-linear :value="percent(node.used_cpu, node.total_cpu)" color="green" height="20">
              <template v-slot:default="{ value }"><strong>{{ value }}%</strong></template>
            </v-progress-linear>
          </v-col>
          <v-col cols="12" lg="6" class="order-1 order-lg-2">
            <p>name: {{ node.name }}</p>
            <p>
              state:
              <v-chip x-small :color="node.state === 'ONLINE' ? 'success' : 'error'">{{ node.state }}</v-chip>
            </p>
            <p>uptime: {{ uptime(node.uptime) }}</p>
            <p>bridges: {{ (node.bridges || []).join(", ") || "-" }}</p>
          </v-col>
        </v-row>
      </v-col>
    </v-row>

    <!-- Networking -->
    <v-card-title class="px-0 mb-3">Networking (static pools):</v-card-title>
    <v-row class="mb-7">
      <v-col cols="12" lg="6" v-for="(pool, kind) in ipPools" :key="kind">
        <div class="title_progress">
          <span>{{ kind }} IPs</span>
          <div>
            <span>{{ pool.used }}</span> / <span>{{ pool.total }}</span>
          </div>
        </div>
        <v-progress-linear :value="percent(pool.used, pool.total)" :color="poolColor(pool)" height="20">
          <template v-slot:default="{ value }"><strong>{{ value }}%</strong></template>
        </v-progress-linear>
        <div class="text-caption mt-1">free: {{ pool.free }}</div>
      </v-col>
    </v-row>

    <!-- Storages -->
    <v-card-title class="px-0 mb-3">Storages:</v-card-title>
    <v-row class="mb-7">
      <v-col v-for="(st, key) in publicData.storages || {}" :key="key" cols="12">
        <v-row>
          <v-col class="order-2 order-lg-1" cols="12" lg="6">
            <div class="title_progress">
              <span>{{ st.drive_type || "not mapped in sched_ds" }}</span>
              <div>
                <span>{{ (st.used / 1024).toFixed(2) }}</span> /
                <span>{{ (st.total / 1024).toFixed(2) }} GiB</span>
              </div>
            </div>
            <v-progress-linear :value="percent(st.used, st.total)" color="green" height="20">
              <template v-slot:default="{ value }"><strong>{{ value }}%</strong></template>
            </v-progress-linear>
          </v-col>
          <v-col class="order-1 order-lg-2" cols="12" lg="6">
            <p>name: {{ st.name }} <span v-if="!st.shared">({{ st.node }})</span></p>
            <p>type: {{ st.type }}{{ st.shared ? ", shared" : "" }}</p>
            <p>content: {{ st.content }}</p>
          </v-col>
        </v-row>
      </v-col>
    </v-row>

    <!-- Templates -->
    <v-card-title class="px-0 mb-3">Templates:</v-card-title>
    <v-simple-table class="mb-7">
      <thead>
        <tr>
          <th>VMID</th>
          <th>Name</th>
          <th>Node</th>
          <th>OS</th>
          <th>Disk, GB</th>
          <th>cloud-init</th>
          <th>agent</th>
          <th>Public</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="t in templates" :key="t.id">
          <td>{{ t.id }}</td>
          <td>{{ t.name }}</td>
          <td>{{ t.node }}</td>
          <td>{{ t.windows ? "windows" : t.ostype }}</td>
          <td>{{ ((t.min_size || 0) / 1024).toFixed(0) }}</td>
          <td><v-icon small :color="t.ci ? 'success' : 'error'">{{ t.ci ? "mdi-check" : "mdi-close" }}</v-icon></td>
          <td><v-icon small :color="t.agent ? 'success' : 'warning'">{{ t.agent ? "mdi-check" : "mdi-close" }}</v-icon></td>
          <td><v-icon small :color="t.is_public ? 'success' : 'grey'">{{ t.is_public ? "mdi-eye" : "mdi-eye-off" }}</v-icon></td>
          <td class="warning--text">{{ t.warning }}</td>
        </tr>
      </tbody>
    </v-simple-table>

    <!-- Orphans -->
    <v-card-title class="px-0 mb-3">
      Orphan VMs (tag nocloud, no instance): {{ orphans.length }}
    </v-card-title>
    <v-simple-table v-if="orphans.length">
      <thead>
        <tr>
          <th>VMID</th>
          <th>Name</th>
          <th>Node</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in orphans" :key="o.vmid">
          <td>{{ o.vmid }}</td>
          <td>{{ o.name }}</td>
          <td>{{ o.node }}</td>
          <td>{{ o.status }}</td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script>
export default {
  name: "service-provider-proxmox",
  props: { template: { type: Object, required: true } },
  computed: {
    publicData() {
      return this.template.publicData || {};
    },
    meta() {
      return this.template.state?.meta || {};
    },
    clusterLabel() {
      const c = this.publicData.cluster;
      if (!c?.is_cluster) return "single node";
      return c.quorate ? "quorate" : "NO QUORUM";
    },
    nodes() {
      const n = this.publicData.nodes || {};
      return Object.values(n).sort((a, b) => (a.name > b.name ? 1 : -1));
    },
    templates() {
      const t = this.publicData.templates || {};
      return Object.keys(t)
        .map((id) => ({ id: +id, ...t[id] }))
        .sort((a, b) => a.id - b.id);
    },
    orphans() {
      return this.publicData.orphans || [];
    },
    ipPools() {
      const out = {};
      if (this.publicData.public_ips) out.public = this.publicData.public_ips;
      if (this.publicData.private_ips?.total) out.private = this.publicData.private_ips;
      return out;
    },
  },
  methods: {
    percent(used, total) {
      if (!total) return 0;
      return Math.min(100, Math.round(((+used || 0) / +total) * 100));
    },
    poolColor(pool) {
      const p = this.percent(pool.used, pool.total);
      if (p >= 95) return "red";
      if (p > 80) return "orange";
      return "green";
    },
    uptime(sec) {
      if (!sec) return "-";
      const d = Math.floor(sec / 86400);
      const h = Math.floor((sec % 86400) / 3600);
      return `${d}d ${h}h`;
    },
  },
};
</script>

<style scoped>
.title_progress {
  display: flex;
  justify-content: space-between;
}
</style>
