<template>
  <v-card color="background-light" class="pa-5" elevation="0">
    <div class="d-flex align-center mb-2">
      <span class="text-h6">Virtual networks</span>
      <v-spacer />
      <v-btn size="small" variant="text" :loading="isLoading" @click="load">
        <v-icon size="small" start>mdi-refresh</v-icon>refresh leases
      </v-btn>
    </div>
    <v-alert v-if="error" density="compact" variant="tonal" type="error">{{ error }}</v-alert>

    <v-expansion-panels v-model="opened" multiple flat>
      <v-expansion-panel v-for="(net, kind) in networks" :key="kind" class="mb-2">
        <v-expansion-panel-title color="background">
          <div class="d-flex align-center" style="gap: 16px; flex-wrap: wrap">
            <strong class="text-capitalize">{{ kind }}</strong>
            <span class="text-medium-emphasis">
              {{ net.bridge }}<span v-if="net.vlan_tag"> · vlan {{ net.vlan_tag }}</span>
              · gw {{ net.gateway || "-" }} · /{{ net.prefix }}
              <span v-if="(net.dns || []).length"> · dns {{ net.dns.join(", ") }}</span>
            </span>
            <v-spacer />
            <v-chip size="x-small" color="success" variant="outlined">free {{ net.free }}</v-chip>
            <v-chip size="x-small" color="info" variant="outlined">used {{ net.used }}</v-chip>
            <v-chip size="x-small" color="warning" variant="outlined">hold {{ net.hold }}</v-chip>
            <v-chip variant="flat" v-if="conflicts(kind)" size="x-small" color="error">conflict {{ conflicts(kind) }}</v-chip>
            <v-chip v-if="outside(kind)" size="x-small" color="error" variant="outlined">outside {{ outside(kind) }}</v-chip>
            <v-chip size="x-small" variant="outlined">total {{ net.total }}</v-chip>
          </div>
        </v-expansion-panel-title>

        <v-expansion-panel-text color="background">
          <!-- Address ranges -->
          <div class="d-flex align-center mt-2 mb-1">
            <span class="text-subtitle-2">Address ranges</span>
            <v-spacer />
            <v-btn size="x-small" variant="text" @click="openAddAR(kind)">
              <v-icon size="small" start>mdi-plus</v-icon>add range
            </v-btn>
          </div>
          <v-table density="compact">
            <thead>
              <tr>
                <th>AR</th>
                <th>First</th>
                <th>Last</th>
                <th>Size</th>
                <th>Used</th>
                <th>Hold</th>
                <th>Free</th>
                <th>Overrides</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="ar in net.ars" :key="ar.id">
                <td>{{ ar.id }}</td>
                <template v-if="isEditingAR(kind, ar.id)">
                  <td>
                    <v-text-field v-model="editAR.ip" density="compact" hide-details placeholder="First IP" />
                  </td>
                  <td class="text-medium-emphasis">{{ editLast }}</td>
                  <td>
                    <v-text-field v-model.number="editAR.size" density="compact" hide-details type="number" min="1" style="max-width: 80px" />
                  </td>
                  <td>{{ ar.used }}</td>
                  <td>{{ ar.hold }}</td>
                  <td>{{ ar.free }}</td>
                  <td>
                    <v-row density="compact">
                      <v-col cols="6">
                        <v-text-field v-model="editAR.gateway" density="compact" hide-details placeholder="gateway" />
                      </v-col>
                      <v-col cols="6">
                        <v-text-field v-model.number="editAR.prefix" density="compact" hide-details type="number" placeholder="prefix" />
                      </v-col>
                      <v-col cols="6">
                        <v-text-field v-model="editAR.bridge" density="compact" hide-details placeholder="bridge" />
                      </v-col>
                      <v-col cols="6">
                        <v-text-field v-model.number="editAR.vlan_tag" density="compact" hide-details type="number" placeholder="vlan" />
                      </v-col>
                    </v-row>
                  </td>
                  <td class="text-no-wrap">
                    <v-btn icon variant="text" size="x-small" color="success" :loading="isSaving" :disabled="!editAR.ip || !(editAR.size > 0)" @click="saveAR(kind, ar.id)">
                      <v-icon size="small">mdi-check</v-icon>
                    </v-btn>
                    <v-btn icon variant="text" size="x-small" @click="cancelEditAR">
                      <v-icon size="small">mdi-close</v-icon>
                    </v-btn>
                  </td>
                </template>
                <template v-else>
                  <td>{{ ar.ip }}</td>
                  <td>{{ ar.last }}</td>
                  <td>{{ ar.size }}</td>
                  <td>{{ ar.used }}</td>
                  <td>{{ ar.hold }}</td>
                  <td :class="ar.free === 0 ? 'text-error' : ''">{{ ar.free }}</td>
                  <td class="text-medium-emphasis">
                    <span v-if="ar.gateway">gw {{ ar.gateway }} </span>
                    <span v-if="ar.prefix">/{{ ar.prefix }} </span>
                    <span v-if="ar.bridge">{{ ar.bridge }} </span>
                    <span v-if="ar.vlan_tag">vlan {{ ar.vlan_tag }}</span>
                    <span v-if="!ar.gateway && !ar.prefix && !ar.bridge && !ar.vlan_tag">—</span>
                  </td>
                  <td class="text-no-wrap">
                    <v-btn icon variant="text" size="x-small" :disabled="isSaving" title="edit range" @click="startEditAR(kind, ar)">
                      <v-icon size="small">mdi-pencil</v-icon>
                    </v-btn>
                    <v-tooltip location="bottom">
                      <template v-slot:activator="{ props }">
                        <span v-bind="props">
                          <v-btn icon variant="text" size="x-small" :disabled="ar.used > 0 || isSaving" @click="removeAR(kind, ar.id)">
                            <v-icon size="small">mdi-delete</v-icon>
                          </v-btn>
                        </span>
                      </template>
                      <span>{{ ar.used > 0 ? "range has leased addresses" : "remove range" }}</span>
                    </v-tooltip>
                  </td>
                </template>
              </tr>
            </tbody>
          </v-table>

          <!-- Leases -->
          <div class="d-flex align-center mt-4 mb-1" style="gap: 12px">
            <span class="text-subtitle-2">Leases</span>
            <v-btn-toggle v-model="stateFilter[kind]" density="compact" mandatory>
              <v-btn size="x-small" value="">all</v-btn>
              <v-btn size="x-small" value="free">free</v-btn>
              <v-btn size="x-small" value="used">used</v-btn>
              <v-btn size="x-small" value="hold">hold</v-btn>
              <v-btn size="x-small" value="conflict" :disabled="!conflicts(kind)">conflict{{ conflicts(kind) ? ` (${conflicts(kind)})` : "" }}</v-btn>
              <v-btn size="x-small" value="outside" :disabled="!outside(kind)">outside{{ outside(kind) ? ` (${outside(kind)})` : "" }}</v-btn>
            </v-btn-toggle>
            <v-text-field
              v-model="search[kind]"
              density="compact"
              hide-details
              clearable
              prepend-inner-icon="mdi-magnify"
              placeholder="ip / vm / instance"
              style="max-width: 260px"
            />
            <v-spacer />
            <v-text-field
              v-model="holdInput[kind]"
              density="compact"
              hide-details
              placeholder="hold IP…"
              style="max-width: 160px"
              @keyup.enter="hold(kind, holdInput[kind])"
            />
            <v-btn size="x-small" :disabled="!holdInput[kind]" :loading="isSaving" @click="hold(kind, holdInput[kind])">hold</v-btn>
          </div>
          <v-data-table
            density="compact"
            :headers="leaseHeaders"
            :items="filteredLeases(kind)"
            :items-per-page="25"
            :items-per-page-options="[
              { value: 25, title: '25' },
              { value: 50, title: '50' },
              { value: 100, title: '100' },
              { value: -1, title: 'All' },
            ]"
            item-value="ip"
          >
            <template v-slot:[`item.ar`]="{ item }">
              <span v-if="item.ar === 0 || item.ar > 0">{{ item.ar }}</span>
              <span v-else class="text-disabled">—</span>
            </template>
            <template v-slot:[`item.state`]="{ item }">
              <v-select
                v-if="item.state === 'free' || item.state === 'hold'"
                :model-value="item.state"
                :items="[
                  { title: 'free', value: 'free' },
                  { title: 'hold', value: 'hold' },
                ]"
                density="compact"
                hide-details
                :disabled="isSaving"
                style="max-width: 110px"
                @update:model-value="setLeaseState(kind, item, $event)"
              />
              <v-chip v-else size="x-small" :color="stateColor(item.state)" variant="outlined">{{ item.state }}</v-chip>
            </template>
            <template v-slot:[`item.vm`]="{ item }">
              <template v-if="item.owners">
                <div v-for="o in item.owners" :key="o.vmid" class="text-error">
                  {{ o.vmid }} · {{ o.vm || "-" }} @ {{ o.node }}
                </div>
              </template>
              <span v-else-if="item.vmid">{{ item.vmid }} · {{ item.vm || "-" }} @ {{ item.node }}</span>
              <span v-else class="text-disabled">-</span>
            </template>
            <template v-slot:[`item.instance`]="{ item }">
              <router-link v-if="item.instance" :to="{ name: 'Instance', params: { instanceId: item.instance } }">
                {{ item.instance.slice(0, 8) }}…
              </router-link>
              <span v-else-if="item.vmid" class="text-medium-emphasis">not managed</span>
              <span v-else class="text-disabled">-</span>
            </template>
            <template v-slot:[`item.actions`]="{ item }">
              <v-btn v-if="item.state === 'free'" size="x-small" variant="text" :loading="isSaving" @click="hold(kind, item.ip)">hold</v-btn>
              <v-btn v-else-if="item.state === 'hold'" size="x-small" variant="text" color="warning" :loading="isSaving" @click="release(kind, item.ip)">release</v-btn>
            </template>
          </v-data-table>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>

    <!-- add AR dialog -->
    <v-dialog v-model="addDialog" max-width="520">
      <v-card color="background-light">
        <v-card-title>Add address range — {{ addKind }}</v-card-title>
        <v-card-text>
          <v-row density="compact">
            <v-col cols="7"><v-text-field label="First IP" v-model="newAR.ip" placeholder="185.66.69.151" /></v-col>
            <v-col cols="5"><v-text-field label="Size" type="number" min="1" v-model.number="newAR.size" /></v-col>
          </v-row>
          <div class="text-caption mb-2">or a range / CIDR:</div>
          <v-text-field density="compact" label="Range" v-model="newAR.range" placeholder="185.66.69.151-185.66.69.160 or 10.0.5.0/24" />
          <v-expansion-panels flat>
            <v-expansion-panel>
              <v-expansion-panel-title class="px-0">Overrides (optional)</v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-row density="compact">
                  <v-col cols="6"><v-text-field density="compact" label="Gateway" v-model="newAR.gateway" /></v-col>
                  <v-col cols="6"><v-text-field density="compact" type="number" label="Prefix" v-model.number="newAR.prefix" /></v-col>
                  <v-col cols="6"><v-text-field density="compact" label="Bridge" v-model="newAR.bridge" /></v-col>
                  <v-col cols="6"><v-text-field density="compact" type="number" label="VLAN tag" v-model.number="newAR.vlan_tag" /></v-col>
                </v-row>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-card-text>
        <v-card-actions class="justify-end">
          <v-btn variant="text" @click="addDialog = false">Cancel</v-btn>
          <v-btn :loading="isSaving" :disabled="!newAR.range && !(newAR.ip && newAR.size > 0)" @click="addAR">Add</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script>
import api from "@/api";

const VAR_BY_KIND = { public: "public_ip_pool", private: "private_vnet_tmpl" };

const ip2n = (ip) => {
  const p = String(ip || "").split(".").map(Number);
  if (p.length !== 4 || p.some((x) => Number.isNaN(x) || x < 0 || x > 255)) return null;
  return ((p[0] << 24) >>> 0) + (p[1] << 16) + (p[2] << 8) + p[3];
};
const n2ip = (n) => [n >>> 24, (n >>> 16) & 255, (n >>> 8) & 255, n & 255].join(".");

export default {
  name: "proxmox-networks",
  props: { template: { type: Object, required: true } },
  data: () => ({
    networks: {},
    opened: [0],
    isLoading: false,
    isSaving: false,
    error: "",
    search: { public: "", private: "" },
    stateFilter: { public: "", private: "" },
    holdInput: { public: "", private: "" },
    addDialog: false,
    addKind: "public",
    newAR: { ip: "", size: 1, range: "", gateway: "", prefix: null, bridge: "", vlan_tag: null },
    editKind: "",
    editId: null,
    editAR: { ip: "", size: 1, gateway: "", prefix: null, bridge: "", vlan_tag: null },
    leaseHeaders: [
      { title: "IP", key: "ip", width: 150 },
      { title: "AR", key: "ar", width: 60 },
      { title: "State", key: "state", width: 120 },
      { title: "VM", key: "vm" },
      { title: "Instance", key: "instance", width: 140 },
      { title: "Source", key: "source", width: 90 },
      { title: "", key: "actions", sortable: false, width: 90 },
    ],
  }),
  computed: {
    editLast() {
      const n = ip2n(this.editAR.ip);
      const size = +this.editAR.size || 0;
      if (n === null || size < 1) return "—";
      return n2ip(n + size - 1);
    },
  },
  methods: {
    async load() {
      this.isLoading = true;
      this.error = "";
      try {
        const { meta } = await api.servicesProviders.action({
          action: "get_leases",
          uuid: this.template.uuid,
          params: {},
        });
        this.networks = meta?.networks || {};
      } catch (e) {
        // fall back to the summary published in public_data (no lease list)
        const pd = this.template.publicData || {};
        this.networks = {};
        if (pd.public_ips?.total) this.networks.public = { ...pd.public_ips, leases: [] };
        if (pd.private_ips?.total) this.networks.private = { ...pd.private_ips, leases: [] };
        this.error = e.response?.data?.message || e.message || "Cannot load leases (driver unavailable?)";
      } finally {
        this.isLoading = false;
      }
    },
    filteredLeases(kind) {
      const q = (this.search[kind] || "").toLowerCase();
      const st = this.stateFilter[kind];
      return (this.networks[kind]?.leases || []).filter(
        (l) =>
          (!st || l.state === st) &&
          (!q || [l.ip, l.vm, l.instance, l.node, String(l.vmid)].some((x) => (x || "").toLowerCase().includes(q)))
      );
    },
    conflicts(kind) {
      return (this.networks[kind]?.leases || []).filter((l) => l.state === "conflict").length;
    },
    outside(kind) {
      return (this.networks[kind]?.leases || []).filter((l) => l.state === "outside").length;
    },
    stateColor(s) {
      return { free: "success", used: "info", hold: "warning", conflict: "error", outside: "error" }[s] || "";
    },

    // ---- persistence: edit the pool inside SP vars, exactly what the driver reads ----
    poolVar(kind) {
      const key = VAR_BY_KIND[kind];
      const value = { ...(this.template.vars?.[key]?.value || {}) };
      const pool = { ...(value.default || {}) };
      // migrate: the AR list is the single source of truth, a stale `ranges` must not linger
      if (!Array.isArray(pool.ars)) {
        pool.ars = (this.networks[kind]?.ars || []).map(({ id, ip, size, gateway, prefix, bridge, vlan_tag, dns }) => {
          const ar = { id, ip, size };
          if (gateway) ar.gateway = gateway;
          if (prefix) ar.prefix = prefix;
          if (bridge) ar.bridge = bridge;
          if (vlan_tag) ar.vlan_tag = vlan_tag;
          if (dns && dns.length) ar.dns = dns;
          return ar;
        });
      }
      delete pool.ranges;
      delete pool.range;
      return { key, value, pool };
    },
    async savePool(kind, pool) {
      const { key, value } = this.poolVar(kind);
      const body = {
        ...this.template,
        vars: { ...this.template.vars, [key]: { value: { ...value, default: pool } } },
      };
      this.isSaving = true;
      try {
        // Update does not re-validate on the core side: ask the driver first
        // (overlapping ranges, hold outside ranges, public/private clash, …)
        const test = await api.servicesProviders.testConfig(body);
        if (test && test.result === false) {
          throw new Error(test.error || "driver rejected the network configuration");
        }
        await api.servicesProviders.update(this.template.uuid, body);
        await this.$store.dispatch("servicesProviders/fetchById", this.template.uuid);
        await this.load();
        this.$store.commit("snackbar/showSnackbarSuccess", { message: "Network updated" });
      } catch (e) {
        this.$store.commit("snackbar/showSnackbarError", {
          message: e.response?.data?.message || e.message || "Error during network update",
        });
      } finally {
        this.isSaving = false;
      }
    },
    hold(kind, ip) {
      ip = String(ip || "").trim();
      if (!ip) return;
      const { pool } = this.poolVar(kind);
      const holds = new Set(pool.holds || []);
      holds.add(ip);
      this.holdInput[kind] = "";
      return this.savePool(kind, { ...pool, holds: [...holds] });
    },
    release(kind, ip) {
      const { pool } = this.poolVar(kind);
      return this.savePool(kind, { ...pool, holds: (pool.holds || []).filter((h) => h !== ip) });
    },
    setLeaseState(kind, item, state) {
      if (state === item.state) return;
      if (state === "hold") return this.hold(kind, item.ip);
      if (state === "free") return this.release(kind, item.ip);
    },
    isEditingAR(kind, id) {
      return this.editKind === kind && this.editId === id;
    },
    startEditAR(kind, ar) {
      this.editKind = kind;
      this.editId = ar.id;
      this.editAR = {
        ip: ar.ip || "",
        size: ar.size || 1,
        gateway: ar.gateway || "",
        prefix: ar.prefix || null,
        bridge: ar.bridge || "",
        vlan_tag: ar.vlan_tag || null,
      };
    },
    cancelEditAR() {
      this.editKind = "";
      this.editId = null;
    },
    saveAR(kind, id) {
      const { pool } = this.poolVar(kind);
      const ars = (pool.ars || []).map((a) => {
        if (a.id !== id) return a;
        const next = { id, ip: String(this.editAR.ip || "").trim(), size: +this.editAR.size || 1 };
        if (this.editAR.gateway) next.gateway = String(this.editAR.gateway).trim();
        if (this.editAR.prefix > 0) next.prefix = +this.editAR.prefix;
        if (this.editAR.bridge) next.bridge = String(this.editAR.bridge).trim();
        if (this.editAR.vlan_tag > 0) next.vlan_tag = +this.editAR.vlan_tag;
        return next;
      });
      this.cancelEditAR();
      return this.savePool(kind, { ...pool, ars });
    },
    openAddAR(kind) {
      this.addKind = kind;
      this.newAR = { ip: "", size: 1, range: "", gateway: "", prefix: null, bridge: "", vlan_tag: null };
      this.addDialog = true;
    },
    addAR() {
      const { pool } = this.poolVar(this.addKind);
      const ars = Array.isArray(pool.ars) ? [...pool.ars] : [];
      const used = new Set(ars.map((a) => a.id));
      let id = 0;
      while (used.has(id)) id++;
      const ar = { id };
      if (this.newAR.range) ar.range = this.newAR.range.trim();
      else {
        ar.ip = this.newAR.ip.trim();
        ar.size = +this.newAR.size || 1;
      }
      for (const k of ["gateway", "bridge"]) if (this.newAR[k]) ar[k] = this.newAR[k].trim();
      for (const k of ["prefix", "vlan_tag"]) if (this.newAR[k] > 0) ar[k] = +this.newAR[k];
      ars.push(ar);
      this.addDialog = false;
      return this.savePool(this.addKind, { ...pool, ars });
    },
    removeAR(kind, id) {
      const { pool } = this.poolVar(kind);
      return this.savePool(kind, { ...pool, ars: (pool.ars || []).filter((a) => a.id !== id) });
    },
  },
  mounted() {
    this.load();
  },
};
</script>
