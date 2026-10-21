<template>
  <v-card outlined class="pool-editor pa-4" color="background">
    <div class="d-flex align-center mb-1">
      <span class="subtitle-1 font-weight-medium">{{ title }}</span>
      <v-spacer />
      <v-btn icon small :title="rawMode ? 'form' : 'edit as JSON'" @click="rawMode = !rawMode">
        <v-icon small>{{ rawMode ? "mdi-form-select" : "mdi-code-json" }}</v-icon>
      </v-btn>
    </div>

    <json-editor v-if="rawMode" :json="pool" @changeValue="emitPool($event)" />

    <template v-else>
      <!-- Network attributes -->
      <div class="section-label">Network</div>
      <v-row dense>
        <v-col cols="8">
          <v-text-field dense outlined hide-details="auto" label="Bridge" placeholder="vmbr0" :value="pool.bridge" :rules="[required]" @change="set('bridge', $event)" />
        </v-col>
        <v-col cols="4">
          <v-text-field dense outlined hide-details="auto" type="number" label="VLAN tag" placeholder="none" :value="pool.vlan_tag || ''" @change="set('vlan_tag', +$event || 0)" />
        </v-col>
        <v-col cols="8">
          <v-text-field dense outlined hide-details="auto" label="Gateway" placeholder="185.66.69.1" :value="pool.gateway" :rules="[ipRule]" @change="set('gateway', $event)" />
        </v-col>
        <v-col cols="4">
          <v-text-field dense outlined hide-details="auto" type="number" label="Prefix" placeholder="24" prefix="/" :value="pool.prefix" :rules="[required, prefixRule]" @change="set('prefix', +$event)" />
        </v-col>
        <v-col cols="12">
          <v-text-field dense outlined hide-details="auto" label="DNS servers" placeholder="1.1.1.1, 8.8.8.8" :value="(pool.dns || []).join(', ')" @change="set('dns', splitList($event))" />
        </v-col>
      </v-row>

      <!-- Address ranges -->
      <div class="d-flex align-center mt-4 mb-1">
        <span class="section-label mb-0">Address ranges</span>
        <v-spacer />
        <v-btn x-small text color="primary" @click="addAR">
          <v-icon x-small left>mdi-plus</v-icon>add range
        </v-btn>
      </div>

      <div v-if="!ars.length" class="text--secondary text-body-2 py-2">
        No ranges yet. A range is a first address and a size — like an OpenNebula AR.
      </div>

      <v-card v-for="(ar, i) in ars" :key="i" outlined class="ar-row mb-2" color="background-light">
        <v-row dense align="center" class="px-3 pt-2">
          <v-col cols="auto">
            <v-chip small label outlined class="ar-id">AR {{ ar.id }}</v-chip>
          </v-col>
          <v-col>
            <v-text-field dense outlined hide-details="auto" label="First IP" placeholder="185.66.69.141" :value="ar.ip" :rules="[required, ipRule]" @change="setAR(i, 'ip', $event)" />
          </v-col>
          <v-col cols="3">
            <v-text-field dense outlined hide-details="auto" type="number" min="1" label="Size" :value="ar.size" @change="setAR(i, 'size', Math.max(1, +$event || 1))" />
          </v-col>
          <v-col cols="auto">
            <v-btn icon small :color="expanded[i] ? 'primary' : ''" title="per-range overrides" @click="toggle(i)">
              <v-icon small>mdi-tune-variant</v-icon>
            </v-btn>
            <v-btn icon small title="remove range" @click="removeAR(i)">
              <v-icon small>mdi-close</v-icon>
            </v-btn>
          </v-col>
        </v-row>
        <div class="px-4 pb-2 text-caption text--secondary">
          <template v-if="lastIP(ar)">{{ ar.ip }} – {{ lastIP(ar) }} · {{ ar.size }} address{{ ar.size === 1 ? "" : "es" }}</template>
          <template v-else>enter the first address</template>
          <span v-if="hasOverrides(ar)"> · overrides: {{ overridesSummary(ar) }}</span>
        </div>
        <v-expand-transition>
          <div v-show="expanded[i]" class="px-3 pb-3">
            <div class="text-caption text--secondary mb-1">Overrides for this range (empty = inherit from network)</div>
            <v-row dense>
              <v-col cols="8">
                <v-text-field dense outlined hide-details label="Gateway" :value="ar.gateway" @change="setAR(i, 'gateway', $event)" />
              </v-col>
              <v-col cols="4">
                <v-text-field dense outlined hide-details type="number" label="Prefix" prefix="/" :value="ar.prefix || ''" @change="setAR(i, 'prefix', +$event || 0)" />
              </v-col>
              <v-col cols="8">
                <v-text-field dense outlined hide-details label="Bridge" :value="ar.bridge" @change="setAR(i, 'bridge', $event)" />
              </v-col>
              <v-col cols="4">
                <v-text-field dense outlined hide-details type="number" label="VLAN tag" :value="ar.vlan_tag || ''" @change="setAR(i, 'vlan_tag', +$event || 0)" />
              </v-col>
            </v-row>
          </div>
        </v-expand-transition>
      </v-card>

      <!-- Holds -->
      <div class="section-label mt-4">Holds</div>
      <v-combobox
        dense
        outlined
        multiple
        small-chips
        deletable-chips
        hide-details="auto"
        placeholder="type an IP and press Enter"
        hint="addresses inside the ranges that must never be assigned"
        persistent-hint
        :value="pool.holds || []"
        @change="set('holds', ($event || []).map((s) => String(s).trim()).filter(Boolean))"
      />

      <div class="d-flex align-center mt-3 text-caption text--secondary">
        <v-icon x-small class="mr-1">mdi-ip-network-outline</v-icon>
        {{ totalAddresses }} addresses in {{ ars.length }} range{{ ars.length === 1 ? "" : "s" }} · {{ (pool.holds || []).length }} on hold
      </div>
      <v-alert v-if="overlapError" dense text type="error" class="mt-2 mb-0">{{ overlapError }}</v-alert>
    </template>
  </v-card>
</template>

<script>
import JsonEditor from "@/components/JsonEditor.vue";

const ip2n = (ip) => {
  const p = String(ip || "").split(".").map(Number);
  if (p.length !== 4 || p.some((x) => Number.isNaN(x) || x < 0 || x > 255)) return null;
  return ((p[0] << 24) >>> 0) + (p[1] << 16) + (p[2] << 8) + p[3];
};
const n2ip = (n) => [n >>> 24, (n >>> 16) & 255, (n >>> 8) & 255, n & 255].join(".");

export default {
  name: "proxmox-ip-pool-editor",
  components: { JsonEditor },
  props: {
    value: { type: Object, default: () => ({}) },
    title: { type: String, default: "Network" },
  },
  data: () => ({ rawMode: false, expanded: {} }),
  computed: {
    pool() {
      return this.value || {};
    },
    ars() {
      return Array.isArray(this.pool.ars) ? this.pool.ars : [];
    },
    totalAddresses() {
      return this.ars.reduce((s, ar) => s + (+ar.size || 0), 0);
    },
    overlapError() {
      const spans = this.ars
        .map((ar) => {
          const lo = ip2n(ar.ip);
          return lo === null ? null : { id: ar.id, lo, hi: lo + (+ar.size || 1) - 1 };
        })
        .filter(Boolean);
      for (let i = 0; i < spans.length; i++) {
        for (let j = i + 1; j < spans.length; j++) {
          if (spans[i].lo <= spans[j].hi && spans[j].lo <= spans[i].hi) {
            return `AR ${spans[i].id} overlaps AR ${spans[j].id}: an address may belong to one range only`;
          }
        }
      }
      const outside = (this.pool.holds || []).find((h) => {
        const n = ip2n(h);
        return n === null || !spans.some((s) => n >= s.lo && n <= s.hi);
      });
      return outside ? `Hold ${outside} is not inside any range` : "";
    },
  },
  methods: {
    required(v) {
      return (v !== "" && v !== null && v !== undefined) || "Required";
    },
    ipRule(v) {
      return !v || ip2n(v) !== null || "Not an IPv4 address";
    },
    prefixRule(v) {
      return (+v >= 1 && +v <= 32) || "1..32";
    },
    splitList(s) {
      return String(s || "")
        .split(/[,\s]+/)
        .map((x) => x.trim())
        .filter(Boolean);
    },
    lastIP(ar) {
      const n = ip2n(ar.ip);
      if (n === null || !ar.size) return "";
      return n2ip(n + (+ar.size || 1) - 1);
    },
    hasOverrides(ar) {
      return !!(ar.gateway || ar.prefix || ar.bridge || ar.vlan_tag);
    },
    overridesSummary(ar) {
      return [ar.gateway && `gw ${ar.gateway}`, ar.prefix && `/${ar.prefix}`, ar.bridge, ar.vlan_tag && `vlan ${ar.vlan_tag}`]
        .filter(Boolean)
        .join(", ");
    },
    toggle(i) {
      this.$set(this.expanded, i, !this.expanded[i]);
    },
    emitPool(pool) {
      this.$emit("input", pool);
    },
    set(key, val) {
      const pool = { ...this.pool, [key]: val };
      if (key === "ars") {
        delete pool.ranges;
        delete pool.range;
      }
      if (val === "" || val === null || val === undefined || (Array.isArray(val) && !val.length)) {
        if (key !== "ars") delete pool[key];
      }
      this.emitPool(pool);
    },
    nextID() {
      const used = new Set(this.ars.map((a) => a.id));
      let id = 0;
      while (used.has(id)) id++;
      return id;
    },
    addAR() {
      this.set("ars", [...this.ars, { id: this.nextID(), ip: "", size: 1 }]);
    },
    removeAR(i) {
      this.set("ars", this.ars.filter((_, idx) => idx !== i));
    },
    setAR(i, key, val) {
      const ars = this.ars.map((a, idx) => {
        if (idx !== i) return a;
        const next = { ...a, [key]: val };
        if (val === "" || val === 0 || val === null) delete next[key];
        if (key === "size") next.size = val;
        return next;
      });
      this.set("ars", ars);
    },
  },
  created() {
    // migrate legacy `ranges` into ARs once, so the list has something to show
    if (!Array.isArray(this.pool.ars) && Array.isArray(this.pool.ranges) && this.pool.ranges.length) {
      const ars = [];
      let id = 0;
      for (const r of this.pool.ranges) {
        if (r.includes("-")) {
          const [a, b] = r.split("-").map((s) => s.trim());
          const na = ip2n(a);
          const nb = ip2n(b);
          if (na !== null && nb !== null && nb >= na) ars.push({ id: id++, ip: a, size: nb - na + 1 });
        } else if (r.includes("/")) {
          const [a, bits] = r.split("/");
          const na = ip2n(a);
          const size = 2 ** (32 - +bits);
          if (na !== null) ars.push({ id: id++, ip: n2ip(na + (+bits < 31 ? 1 : 0)), size: +bits < 31 ? size - 2 : size });
        } else if (ip2n(r) !== null) {
          ars.push({ id: id++, ip: r, size: 1 });
        }
      }
      const pool = { ...this.pool, ars };
      delete pool.ranges;
      this.emitPool(pool);
    }
  },
};
</script>

<style scoped>
.section-label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  opacity: 0.6;
  margin-bottom: 6px;
}
.ar-id {
  font-family: monospace;
}
.ar-row {
  border-radius: 8px;
}
</style>
