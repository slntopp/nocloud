<template>
  <v-card outlined class="pa-3" color="background">
    <div class="d-flex align-center mb-2">
      <span class="subtitle-1">{{ title }}</span>
      <v-spacer />
      <v-switch
        dense
        hide-details
        class="mt-0"
        label="JSON"
        :input-value="rawMode"
        @change="rawMode = !!$event"
      />
    </div>

    <json-editor v-if="rawMode" :json="pool" @changeValue="emitPool($event)" />

    <template v-else>
      <!-- Network (guest) attributes: what ONE puts on the Virtual Network -->
      <v-row dense>
        <v-col cols="12" md="3">
          <v-text-field dense label="Bridge" placeholder="vmbr0" :value="pool.bridge" :rules="[required]" @change="set('bridge', $event)" />
        </v-col>
        <v-col cols="12" md="2">
          <v-text-field dense type="number" label="VLAN tag" placeholder="0" :value="pool.vlan_tag || ''" @change="set('vlan_tag', +$event || 0)" />
        </v-col>
        <v-col cols="12" md="3">
          <v-text-field dense label="Gateway" placeholder="185.66.69.1" :value="pool.gateway" :rules="[ipRule]" @change="set('gateway', $event)" />
        </v-col>
        <v-col cols="12" md="2">
          <v-text-field dense type="number" label="Prefix" placeholder="24" :value="pool.prefix" :rules="[required, prefixRule]" @change="set('prefix', +$event)" />
        </v-col>
        <v-col cols="12" md="2">
          <v-text-field dense label="DNS" placeholder="1.1.1.1, 8.8.8.8" :value="(pool.dns || []).join(', ')" @change="set('dns', splitList($event))" />
        </v-col>
      </v-row>

      <!-- Address Ranges -->
      <div class="d-flex align-center mt-2">
        <span class="subtitle-2">Address ranges</span>
        <v-spacer />
        <v-btn x-small text @click="addAR"><v-icon small left>mdi-plus</v-icon>add range</v-btn>
      </div>
      <v-simple-table dense>
        <thead>
          <tr>
            <th style="width: 50px">AR</th>
            <th>First IP</th>
            <th style="width: 110px">Size</th>
            <th>Last IP</th>
            <th>Overrides (gateway / prefix / bridge / vlan)</th>
            <th style="width: 40px"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(ar, i) in ars" :key="i">
            <td>{{ ar.id }}</td>
            <td>
              <v-text-field dense hide-details :value="ar.ip" :rules="[ipRule]" placeholder="185.66.69.141" @change="setAR(i, 'ip', $event)" />
            </td>
            <td>
              <v-text-field dense hide-details type="number" min="1" :value="ar.size" @change="setAR(i, 'size', Math.max(1, +$event || 1))" />
            </td>
            <td class="text--secondary">{{ lastIP(ar) }}</td>
            <td>
              <div class="d-flex" style="gap: 6px">
                <v-text-field dense hide-details style="max-width: 130px" placeholder="gateway" :value="ar.gateway" @change="setAR(i, 'gateway', $event)" />
                <v-text-field dense hide-details style="max-width: 60px" type="number" placeholder="/" :value="ar.prefix || ''" @change="setAR(i, 'prefix', +$event || 0)" />
                <v-text-field dense hide-details style="max-width: 90px" placeholder="bridge" :value="ar.bridge" @change="setAR(i, 'bridge', $event)" />
                <v-text-field dense hide-details style="max-width: 60px" type="number" placeholder="vlan" :value="ar.vlan_tag || ''" @change="setAR(i, 'vlan_tag', +$event || 0)" />
              </div>
            </td>
            <td>
              <v-btn icon x-small @click="removeAR(i)"><v-icon small>mdi-delete</v-icon></v-btn>
            </td>
          </tr>
          <tr v-if="!ars.length">
            <td colspan="6" class="text--secondary">No address ranges — add at least one (first IP + size), like ONE's AR IP/SIZE.</td>
          </tr>
        </tbody>
      </v-simple-table>

      <!-- Holds -->
      <v-combobox
        class="mt-3"
        dense
        multiple
        small-chips
        deletable-chips
        label="Holds (addresses that are never assigned)"
        hint="type an IP and press Enter"
        persistent-hint
        :value="pool.holds || []"
        @change="set('holds', ($event || []).map((s) => String(s).trim()).filter(Boolean))"
      />

      <div class="text-caption mt-2">
        {{ totalAddresses }} addresses in {{ ars.length }} range(s), {{ (pool.holds || []).length }} on hold
      </div>
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
  data: () => ({ rawMode: false }),
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
    emitPool(pool) {
      this.$emit("input", pool);
    },
    set(key, val) {
      const pool = { ...this.pool, [key]: val };
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
    // migrate legacy `ranges` into ARs once, so the table has something to show
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
