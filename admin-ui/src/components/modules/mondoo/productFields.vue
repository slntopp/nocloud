<template>
  <div class="pt-4">
    <v-list-subheader class="px-0">Mondoo product</v-list-subheader>
    <v-row class="align-center">
      <v-col cols="12" md="4">
        <v-select
          density="compact"
          label="What it scans"
          hide-details
          :items="productKinds"
          item-title="title"
          item-value="value"
          :model-value="kind"
          @update:model-value="set('kind', $event)"
        >
          <template v-slot:item="{ props: itemProps, item }">
            <v-list-item v-bind="itemProps" :subtitle="item.raw.subtitle" />
          </template>
        </v-select>
      </v-col>

      <template v-if="kind === 'ssh'">
        <v-col cols="6" md="2">
          <v-text-field
            density="compact"
            label="Scan every"
            placeholder="24h"
            persistent-placeholder
            hint="At least 1h"
            persistent-hint
            :model-value="meta.schedule"
            @change="set('schedule', $event.target.value.trim() || undefined)"
          />
        </v-col>
        <v-col cols="6" md="4">
          <v-switch
            density="compact"
            color="primary"
            hide-details
            label="Customers may run fixes"
            :model-value="meta.remediation === 'customer'"
            @update:model-value="set('remediation', $event ? 'customer' : undefined)"
          />
        </v-col>
      </template>
    </v-row>

    <template v-if="kind !== 'ssh'">
      <div class="text-caption text-medium-emphasis pt-2 pb-3">
        Scan options. "Mondoo default" leaves the option out, so Mondoo's own
        default applies.
      </div>
      <v-row>
        <v-col v-for="option in scanOptions" :key="option.key" cols="6" md="3">
          <v-select
            density="compact"
            hide-details
            :label="option.title"
            :items="tristate"
            :model-value="toTristate(meta[option.key])"
            @update:model-value="set(option.key, fromTristate($event))"
          />
        </v-col>
      </v-row>
    </template>

    <div class="text-caption text-medium-emphasis pt-2">
      {{
        kind === "ssh"
          ? "One instance scans one server of the customer, over SSH from the scan runner."
          : "One instance scans one domain. Each further domain is another instance, ordered and paid on its own."
      }}
    </div>
  </div>
</template>

<script setup>
import { computed, toRefs } from "vue";
import { productKinds, productKind } from "./productKinds";

const props = defineProps({ meta: { type: Object, default: () => ({}) } });
const emit = defineEmits(["update:meta"]);
const { meta } = toRefs(props);

const kind = computed(() => productKind({ meta: meta.value }));

const scanOptions = [
  { key: "subdomains", title: "Subdomain discovery" },
  { key: "port_scan", title: "Port scan" },
  { key: "web_fingerprint", title: "Web fingerprinting" },
  { key: "vulnerability_scanning", title: "Vulnerability scanning" },
  { key: "subdomain_takeover", title: "Subdomain takeover" },
  { key: "cloud_discovery", title: "Cloud discovery" },
  { key: "webscan", title: "Web scan" },
];
const tristate = [
  { title: "Mondoo default", value: "default" },
  { title: "On", value: "on" },
  { title: "Off", value: "off" },
];
const toTristate = (v) => (v === true ? "on" : v === false ? "off" : "default");
const fromTristate = (v) => (v === "on" ? true : v === "off" ? false : undefined);

// Options of the other kind do not apply and are dropped on a kind change.
const sshOnly = ["remediation"];
const domainOnly = scanOptions.map(({ key }) => key);

function set(key, value) {
  const next = { ...meta.value };
  if (value === undefined) delete next[key];
  else next[key] = value;
  if (key === "kind") {
    for (const k of value === "ssh" ? domainOnly : sshOnly) delete next[k];
    if (value !== "ssh") delete next.schedule;
  }
  emit("update:meta", next);
}
</script>
