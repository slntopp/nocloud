<template>
  <div class="pt-4">
    <v-list-subheader class="px-0">Free trial</v-list-subheader>
    <v-row class="align-center">
      <v-col cols="12" md="2">
        <v-switch
          label="Offer a trial"
          color="primary"
          density="compact"
          hide-details
          :model-value="enabled"
          @update:model-value="toggle"
        />
      </v-col>
      <template v-if="enabled">
        <v-col cols="6" md="2">
          <v-text-field
            density="compact"
            type="number"
            min="1"
            label="Length"
            suffix="days"
            hide-details
            :model-value="toDays(trial.period)"
            @change="set('period', fromDays($event.target.value))"
          />
        </v-col>
        <v-col cols="6" md="2">
          <v-text-field
            density="compact"
            type="number"
            min="0"
            label="Invoice before end"
            suffix="days"
            hide-details
            :model-value="toDays(trial.invoice_before ?? defaultInvoiceBefore)"
            @change="set('invoice_before', fromDays($event.target.value))"
          />
        </v-col>
        <v-col cols="6" md="2">
          <v-select
            density="compact"
            label="One trial per"
            hide-details
            :items="scopes"
            item-title="title"
            item-value="value"
            :model-value="scopeKind"
            @update:model-value="setScopeKind"
          />
        </v-col>
        <v-col cols="6" md="2" v-if="scopeKind === 'config'">
          <v-text-field
            density="compact"
            label="Config key"
            hide-details
            :model-value="scopeKey"
            @change="set('scope', `config:${$event.target.value}`)"
          />
        </v-col>
        <v-col cols="6" md="2">
          <v-text-field
            density="compact"
            label="Group"
            placeholder="plan type"
            persistent-placeholder
            hide-details
            :model-value="trial.group"
            @change="set('group', $event.target.value || undefined)"
          />
        </v-col>
      </template>
    </v-row>
    <div v-if="enabled" class="text-caption text-medium-emphasis pt-1">
      The first instance per {{ scopeKind === "config" ? `value of config.${scopeKey || "<key>"}` : "account" }}
      starts without a start invoice. Its first renewal invoice is issued
      {{ toDays(trial.invoice_before ?? defaultInvoiceBefore) }} day(s) before the trial ends.
      Products sharing a group share the trial.
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";

// Edits product.meta.trial, which billing reads when an instance is created:
// { period, scope: "account" | "config:<key>", group, invoice_before } (seconds).
const props = defineProps({ meta: { type: Object, default: () => ({}) } });
const emit = defineEmits(["update:meta"]);

const day = 86400;
const defaultInvoiceBefore = 2 * day;
const scopes = [
  { value: "account", title: "account" },
  { value: "config", title: "config value" },
];

const trial = computed(() => props.meta?.trial ?? {});
const enabled = computed(() => !!props.meta?.trial);
const scopeKind = computed(() =>
  String(trial.value.scope ?? "").startsWith("config:") ? "config" : "account"
);
const scopeKey = computed(() =>
  scopeKind.value === "config" ? trial.value.scope.slice("config:".length) : ""
);

const toDays = (seconds) => Math.round(((+seconds || 0) / day) * 100) / 100;
const fromDays = (days) => Math.max(0, Math.round((+days || 0) * day));

function update(next) {
  const meta = { ...props.meta };
  if (next) meta.trial = next;
  else delete meta.trial;
  emit("update:meta", meta);
}

function toggle(value) {
  update(value ? { period: 7 * day, scope: "account" } : null);
}

function set(key, value) {
  const next = { ...trial.value, [key]: value };
  if (value === undefined) delete next[key];
  update(next);
}

function setScopeKind(kind) {
  set("scope", kind === "config" ? "config:domain" : "account");
}
</script>
