<template>
  <v-card color="background-light" class="pa-5" elevation="0">
    <div class="d-flex align-center mb-2">
      <span class="text-h6">Virtual networks</span>
      <span class="text--secondary ml-3">address ranges · leases · holds (as OpenNebula vnets)</span>
      <v-spacer />
      <v-btn small text :loading="isLoading" @click="load">
        <v-icon small left>mdi-refresh</v-icon>refresh leases
      </v-btn>
    </div>
    <v-alert v-if="error" dense text type="error">{{ error }}</v-alert>

    <v-expansion-panels v-model="opened" multiple flat>
      <v-expansion-panel v-for="(net, kind) in networks" :key="kind" class="mb-2">
        <v-expansion-panel-header color="background">
          <div class="d-flex align-center" style="gap: 16px; flex-wrap: wrap">
            <strong class="text-capitalize">{{ kind }}</strong>
            <span class="text--secondary">
              {{ net.bridge }}<span v-if="net.vlan_tag"> · vlan {{ net.vlan_tag }}</span>
              · gw {{ net.gateway || "-" }} · /{{ net.prefix }}
              <span v-if="(net.dns || []).length"> · dns {{ net.dns.join(", ") }}</span>
            </span>
            <v-spacer />
            <v-chip x-small color="success" outlined>free {{ net.free }}</v-chip>
            <v-chip x-small color="info" outlined>used {{ net.used }}</v-chip>
            <v-chip x-small color="warning" outlined>hold {{ net.hold }}</v-chip>
            <v-chip v-if="conflicts(kind)" x-small color="error">conflict {{ conflicts(kind) }}</v-chip>
            <v-chip x-small outlined>total {{ net.total }}</v-chip>
          </div>
        </v-expansion-panel-header>

        <v-expansion-panel-content color="background">
          <!-- Address ranges -->
          <div class="d-flex align-center mt-2 mb-1">
            <span class="subtitle-2">Address ranges</span>
            <v-spacer />
            <v-btn x-small text @click="openAddAR(kind)">
              <v-icon small left>mdi-plus</v-icon>add range
            </v-btn>
          </div>
          <v-simple-table dense>
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
                <td>{{ ar.ip }}</td>
                <td>{{ ar.last }}</td>
                <td>{{ ar.size }}</td>
                <td>{{ ar.used }}</td>
                <td>{{ ar.hold }}</td>
                <td :class="ar.free === 0 ? 'error--text' : ''">{{ ar.free }}</td>
                <td class="text--secondary">
                  <span v-if="ar.gateway">gw {{ ar.gateway }} </span>
                  <span v-if="ar.prefix">/{{ ar.prefix }} </span>
                  <span v-if="ar.bridge">{{ ar.bridge }} </span>
                  <span v-if="ar.vlan_tag">vlan {{ ar.vlan_tag }}</span>
                </td>
                <td>
                  <v-tooltip bottom>
                    <template v-slot:activator="{ on }">
                      <span v-on="on">
                        <v-btn icon x-small :disabled="ar.used > 0 || isSaving" @click="removeAR(kind, ar.id)">
                          <v-icon small>mdi-delete</v-icon>
                        </v-btn>
                      </span>
                    </template>
                    <span>{{ ar.used > 0 ? "range has leased addresses" : "remove range" }}</span>
                  </v-tooltip>
                </td>
              </tr>
            </tbody>
          </v-simple-table>

          <!-- Leases -->
          <div class="d-flex align-center mt-4 mb-1" style="gap: 12px">
            <span class="subtitle-2">Leases</span>
            <v-btn-toggle v-model="stateFilter[kind]" dense mandatory>
              <v-btn x-small value="">all</v-btn>
              <v-btn x-small value="free">free</v-btn>
              <v-btn x-small value="used">used</v-btn>
              <v-btn x-small value="hold">hold</v-btn>
              <v-btn x-small value="conflict" :disabled="!conflicts(kind)">conflict{{ conflicts(kind) ? ` (${conflicts(kind)})` : "" }}</v-btn>
            </v-btn-toggle>
            <v-text-field
              v-model="search[kind]"
              dense
              hide-details
              clearable
              prepend-inner-icon="mdi-magnify"
              placeholder="ip / vm / instance"
              style="max-width: 260px"
            />
            <v-spacer />
            <v-text-field
              v-model="holdInput[kind]"
              dense
              hide-details
              placeholder="hold IP…"
              style="max-width: 160px"
              @keyup.enter="hold(kind, holdInput[kind])"
            />
            <v-btn x-small :disabled="!holdInput[kind]" :loading="isSaving" @click="hold(kind, holdInput[kind])">hold</v-btn>
          </div>
          <v-data-table
            dense
            :headers="leaseHeaders"
            :items="filteredLeases(kind)"
            :items-per-page="25"
            :footer-props="{ 'items-per-page-options': [25, 50, 100, -1] }"
            item-key="ip"
          >
            <template v-slot:[`item.state`]="{ item }">
              <v-chip x-small :color="stateColor(item.state)" outlined>{{ item.state }}</v-chip>
            </template>
            <template v-slot:[`item.vm`]="{ item }">
              <template v-if="item.owners">
                <div v-for="o in item.owners" :key="o.vmid" class="error--text">
                  {{ o.vmid }} · {{ o.vm || "-" }} @ {{ o.node }}
                </div>
              </template>
              <span v-else-if="item.vmid">{{ item.vmid }} · {{ item.vm || "-" }} @ {{ item.node }}</span>
              <span v-else class="text--disabled">-</span>
            </template>
            <template v-slot:[`item.instance`]="{ item }">
              <router-link v-if="item.instance" :to="{ name: 'Instance', params: { instanceId: item.instance } }">
                {{ item.instance.slice(0, 8) }}…
              </router-link>
              <span v-else-if="item.vmid" class="text--secondary">not managed</span>
              <span v-else class="text--disabled">-</span>
            </template>
            <template v-slot:[`item.actions`]="{ item }">
              <v-btn v-if="item.state === 'free'" x-small text :loading="isSaving" @click="hold(kind, item.ip)">hold</v-btn>
              <v-btn v-else-if="item.state === 'hold'" x-small text color="warning" :loading="isSaving" @click="release(kind, item.ip)">release</v-btn>
            </template>
          </v-data-table>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <!-- add AR dialog -->
    <v-dialog v-model="addDialog" max-width="520">
      <v-card color="background-light">
        <v-card-title>Add address range — {{ addKind }}</v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col cols="7"><v-text-field label="First IP" v-model="newAR.ip" placeholder="185.66.69.151" /></v-col>
            <v-col cols="5"><v-text-field label="Size" type="number" min="1" v-model.number="newAR.size" /></v-col>
          </v-row>
          <div class="text-caption mb-2">or a range / CIDR:</div>
          <v-text-field dense label="Range" v-model="newAR.range" placeholder="185.66.69.151-185.66.69.160 or 10.0.5.0/24" />
          <v-expansion-panels flat>
            <v-expansion-panel>
              <v-expansion-panel-header class="px-0">Overrides (optional)</v-expansion-panel-header>
              <v-expansion-panel-content>
                <v-row dense>
                  <v-col cols="6"><v-text-field dense label="Gateway" v-model="newAR.gateway" /></v-col>
                  <v-col cols="6"><v-text-field dense type="number" label="Prefix" v-model.number="newAR.prefix" /></v-col>
                  <v-col cols="6"><v-text-field dense label="Bridge" v-model="newAR.bridge" /></v-col>
                  <v-col cols="6"><v-text-field dense type="number" label="VLAN tag" v-model.number="newAR.vlan_tag" /></v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-card-text>
        <v-card-actions class="justify-end">
          <v-btn text @click="addDialog = false">Cancel</v-btn>
          <v-btn :loading="isSaving" :disabled="!newAR.range && !(newAR.ip && newAR.size > 0)" @click="addAR">Add</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script>
import api from "@/api";

const VAR_BY_KIND = { public: "public_ip_pool", private: "private_vnet_tmpl" };

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
    leaseHeaders: [
      { text: "IP", value: "ip", width: 150 },
      { text: "AR", value: "ar", width: 60 },
      { text: "State", value: "state", width: 90 },
      { text: "VM", value: "vm" },
      { text: "Instance", value: "instance", width: 140 },
      { text: "Source", value: "source", width: 90 },
      { text: "", value: "actions", sortable: false, width: 90 },
    ],
  }),
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
    stateColor(s) {
      return { free: "success", used: "info", hold: "warning", conflict: "error" }[s] || "";
    },

    // ---- persistence: edit the pool inside SP vars, exactly what the driver reads ----
    poolVar(kind) {
      const key = VAR_BY_KIND[kind];
      const value = { ...(this.template.vars?.[key]?.value || {}) };
      const pool = { ...(value.default || {}) };
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
