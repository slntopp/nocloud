<template>
  <div class="finding">
    <div v-if="loading" class="text-disabled pa-2">Loading...</div>
    <v-alert v-else-if="error" type="warning" variant="tonal" density="compact">
      {{ error }}
    </v-alert>
    <v-row v-else>
      <v-col cols="12" md="7">
        <div v-if="facts.length" class="d-flex flex-wrap ga-2 mb-3">
          <v-chip v-for="fact in facts" :key="fact.label" size="small" variant="outlined" :href="fact.href" target="_blank">
            {{ fact.label }}: {{ fact.value }}
          </v-chip>
        </div>

        <template v-if="details.description || details.message">
          <div class="finding__title">What it means</div>
          <!-- eslint-disable-next-line vue/no-v-html -- renderMarkdown escapes its input -->
          <div class="finding__text" v-html="renderMarkdown(details.description || details.message)" />
        </template>

        <template v-if="details.remediation">
          <div class="finding__title">How to fix it</div>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <div class="finding__text" v-html="renderMarkdown(details.remediation)" />
        </template>
        <div v-else class="text-disabled mb-3">
          Mondoo provides no remediation steps for this finding.
        </div>

        <template v-if="details.audit">
          <div class="finding__title finding__title--toggle" @click="showAudit = !showAudit">
            How to verify
            <v-icon size="small">{{ showAudit ? "mdi-chevron-up" : "mdi-chevron-down" }}</v-icon>
          </div>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <div v-if="showAudit" class="finding__text" v-html="renderMarkdown(details.audit)" />
        </template>

        <template v-if="details.references?.length">
          <div class="finding__title">References</div>
          <ul class="finding__text">
            <li v-for="ref in details.references" :key="ref.url">
              <a :href="ref.url" target="_blank" rel="noopener noreferrer">{{ ref.title || ref.url }}</a>
            </li>
          </ul>
        </template>
      </v-col>

      <v-col cols="12" md="5">
        <div class="finding__title">Affected hosts ({{ details.assets?.length ?? 0 }})</div>
        <v-table density="compact" class="finding__assets">
          <thead>
            <tr>
              <th>Host</th>
              <th>Domain</th>
              <th v-if="state">State</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="asset in details.assets" :key="asset.asset">
              <td>
                {{ asset.asset_name }}
                <div v-for="pkg in asset.packages" :key="pkg.name" class="text-caption text-medium-emphasis">
                  {{ pkg.name }} {{ pkg.installed }}
                  <template v-if="pkg.available">→ {{ pkg.available }}</template>
                </div>
              </td>
              <td>{{ asset.namespace }}</td>
              <td v-if="state">{{ asset.state }}</td>
              <td class="text-no-wrap">{{ shortDate(asset.last_updated) }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import api from "@/api";
import { renderMarkdown } from "./markdown";

const props = defineProps({
  uuid: { type: String, required: true },
  mrn: { type: String, required: true },
  rating: { type: String, default: "" },
  state: { type: String, default: "" },
  namespace: { type: String, default: "" },
});

const loading = ref(true);
const error = ref("");
const details = ref({});
const showAudit = ref(false);

const facts = computed(() => {
  const d = details.value;
  const out = [];
  if (d.id) {
    out.push({
      label: "ID",
      value: d.id,
      href: /^CVE-/.test(d.id) ? `https://nvd.nist.gov/vuln/detail/${d.id}` : undefined,
    });
  }
  if (d.cvss !== undefined) out.push({ label: "CVSS", value: d.cvss });
  if (d.epss !== undefined) out.push({ label: "EPSS", value: `${(d.epss * 100).toFixed(1)}%` });
  if (d.finding_type) out.push({ label: "Type", value: d.finding_type.toLowerCase() });
  return out;
});

const shortDate = (value) => {
  const date = new Date(value);
  return !value || isNaN(date) ? "-" : date.toLocaleDateString();
};

onMounted(async () => {
  try {
    const response = await api.instances.action({
      uuid: props.uuid,
      action: "finding",
      params: {
        mrn: props.mrn,
        ...(props.rating && { rating: props.rating }),
        ...(props.state && { state: props.state }),
        ...(props.namespace && { namespace: props.namespace }),
      },
    });
    details.value = response?.meta ?? {};
  } catch (err) {
    error.value =
      err?.response?.data?.message ?? err?.message ?? "Failed to load the finding";
  } finally {
    loading.value = false;
  }
});
</script>

<script>
export default {
  name: "mondoo-finding-details",
};
</script>

<style scoped>
.finding {
  padding: 12px 4px 16px;
  white-space: normal;
}
.finding__title {
  font-weight: 600;
  margin: 4px 0 6px;
}
.finding__title--toggle {
  cursor: pointer;
  user-select: none;
}
.finding__text {
  margin-bottom: 14px;
  line-height: 1.5;
}
.finding__text :deep(p),
.finding__text :deep(ul),
.finding__text :deep(ol) {
  margin: 0 0 8px;
}
.finding__text :deep(ul),
.finding__text :deep(ol),
ul.finding__text {
  padding-left: 22px;
}
.finding__text :deep(h4) {
  margin: 10px 0 4px;
}
.finding__text :deep(code) {
  padding: 1px 4px;
  border-radius: 4px;
  background: rgba(128, 128, 128, 0.2);
  font-size: 0.9em;
}
.finding__text :deep(pre) {
  margin: 0 0 8px;
  padding: 8px 10px;
  border-radius: 6px;
  background: rgba(128, 128, 128, 0.16);
  overflow-x: auto;
}
.finding__text :deep(pre code) {
  padding: 0;
  background: none;
}
.finding__text :deep(a),
ul.finding__text a {
  color: rgb(var(--v-theme-primary));
}
.finding__assets {
  background: transparent;
}
</style>
