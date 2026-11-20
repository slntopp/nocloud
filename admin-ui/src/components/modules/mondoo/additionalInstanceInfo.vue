<template>
  <div class="mondoo-info">
    <v-row>
      <v-col v-if="isSsh" cols="12" md="6">
        <v-text-field
          readonly
          label="Scanned machine"
          :model-value="scan.asset_name || config.instance"
          hint="Scanned over SSH from the scan runner. Nothing runs on the machine itself."
          persistent-hint
        />
      </v-col>
      <v-col v-else cols="12" md="6">
        <v-combobox
          label="Domains"
          :model-value="domains"
          multiple
          chips
          closable-chips
          :readonly="imported"
          :hint="
            imported
              ? 'Imported integration: its domains are managed in Mondoo.'
              : 'The driver updates the integration on its next monitoring pass.'
          "
          persistent-hint
          :delimiters="[',', ' ']"
          @update:model-value="changeDomains"
        />
      </v-col>
      <v-col cols="6" md="2">
        <v-text-field readonly label="Hosts" :model-value="summary.assets ?? '-'" />
      </v-col>
      <v-col cols="6" md="2">
        <v-text-field
          readonly
          label="Open findings"
          :model-value="summary.findings ?? '-'"
        />
      </v-col>
      <v-col cols="12" md="2">
        <v-text-field readonly label="Last sync" :model-value="syncedAt" />
      </v-col>
    </v-row>

    <v-row class="align-center">
      <v-col cols="12" md="6" class="d-flex flex-wrap align-center ga-2">
        <v-chip
          v-for="rating in ratings"
          :key="rating.key"
          :color="rating.color"
          variant="flat"
          size="small"
          :class="{ 'mondoo-dim': !byRating[rating.key] }"
          @click="filterRating(rating.key)"
        >
          {{ rating.title }}: {{ byRating[rating.key] ?? 0 }}
        </v-chip>
      </v-col>
      <v-col cols="12" md="6" class="d-flex flex-wrap align-center ga-2 justify-md-end">
        <v-tooltip v-if="isSsh && scan.error" location="top" :text="scan.error">
          <template #activator="{ props: tip }">
            <v-chip v-bind="tip" size="small" color="warning" variant="outlined">
              last scan failed
            </v-chip>
          </template>
        </v-tooltip>
        <v-chip v-else-if="isSsh && scan.finished_at" size="small" color="success" variant="outlined">
          scanned {{ formatTs(scan.finished_at) }}
        </v-chip>
        <v-chip v-else-if="isSsh" size="small" color="primary" variant="outlined">
          <v-progress-circular indeterminate size="12" width="2" class="mr-2" />
          {{ scan.requested_at ? "first scan running" : "waiting for the first scan" }}
        </v-chip>
        <v-tooltip v-if="fix.status" location="top" :text="fix.error || fix.title">
          <template #activator="{ props: tip }">
            <v-chip v-bind="tip" size="small" variant="outlined" :color="fixColor">
              <v-progress-circular v-if="fixActive" indeterminate size="12" width="2" class="mr-2" />
              fix {{ fix.status }}
            </v-chip>
          </template>
        </v-tooltip>
        <v-chip v-if="imported" size="small" color="primary" variant="outlined">
          imported, read-only
        </v-chip>
        <v-chip v-if="trialEnds" size="small" :color="trialActive ? 'success' : undefined" variant="outlined">
          trial {{ trialActive ? "until" : "ended" }} {{ trialEnds }}
        </v-chip>
        <v-tooltip v-if="integrationMrn" location="top" :text="integrationMrn">
          <template #activator="{ props: tip }">
            <v-chip v-bind="tip" size="small" variant="outlined" @click="copy(integrationMrn)">
              integration mrn
              <v-icon end size="small">mdi-content-copy</v-icon>
            </v-chip>
          </template>
        </v-tooltip>
      </v-col>
    </v-row>

    <v-row v-if="namespaces.length">
      <v-col cols="12" class="d-flex flex-wrap ga-2">
        <v-chip
          v-for="ns in namespaces"
          :key="ns.name"
          size="small"
          variant="outlined"
          :color="ns.configured === false ? 'warning' : undefined"
        >
          {{ ns.name }}
          <template v-if="ns.grade">&nbsp;· grade {{ ns.grade }} ({{ ns.score }})</template>
          <template v-if="ns.configured === false">&nbsp;· discovered</template>
        </v-chip>
      </v-col>
    </v-row>

    <nocloud-expansion-panels title="Scan results" class="mb-4 mt-2">
      <div class="d-flex align-center flex-wrap ga-2 mb-2">
        <v-btn-toggle v-model="tab" mandatory density="compact" color="primary">
          <v-btn value="findings" size="small">Findings ({{ findings.total }})</v-btn>
          <v-btn value="assets" size="small">Assets ({{ assets.total }})</v-btn>
        </v-btn-toggle>

        <template v-if="tab === 'findings'">
          <v-select
            v-model="findingsFilter.rating"
            :items="ratingItems"
            item-title="title"
            item-value="key"
            label="Rating"
            density="compact"
            hide-details
            style="max-width: 160px"
          />
          <v-select
            v-model="findingsFilter.state"
            :items="stateItems"
            item-title="title"
            item-value="key"
            label="State"
            density="compact"
            hide-details
            style="max-width: 160px"
          />
          <v-select
            v-if="namespaces.length > 1"
            v-model="findingsFilter.namespace"
            :items="domainItems"
            label="Domain"
            density="compact"
            hide-details
            style="max-width: 220px"
          />
          <v-text-field
            v-model="searchText"
            label="Search"
            placeholder="title or CVE"
            density="compact"
            hide-details
            clearable
            prepend-inner-icon="mdi-magnify"
            style="max-width: 240px"
          />
        </template>
        <v-spacer />
        <v-btn size="small" variant="text" :loading="loading" @click="reload">
          <v-icon start>mdi-refresh</v-icon>refresh
        </v-btn>
      </div>

      <v-alert v-if="error" type="warning" variant="tonal" density="compact" class="mb-2">
        {{ error }}
      </v-alert>

      <mondoo-findings-table
        v-if="tab === 'findings'"
        :uuid="template.uuid"
        :rating="findingsFilter.rating"
        :state="findingsFilter.state"
        :namespace="findingsFilter.namespace"
        :search="findingsFilter.search"
        :page-size="pageSize"
        :refresh-key="refreshKey"
        @total="findings.total = $event"
        @fix="watchFix"
      />

      <template v-else>
      <v-table density="compact" class="mondoo-table">
        <thead>
          <tr>
            <th>Host</th>
            <th>Namespace</th>
            <th>Type</th>
            <th>Grade</th>
            <th>State</th>
            <th>Updated</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in assets.items" :key="item.mrn">
            <td>{{ item.name }}</td>
            <td>{{ item.namespace }}</td>
            <td>{{ item.platform || item.type }}</td>
            <td>{{ item.grade ? `${item.grade} (${item.score})` : "-" }}</td>
            <td>{{ item.state }}</td>
            <td class="text-no-wrap">{{ shortDate(item.updated_at) }}</td>
          </tr>
          <tr v-if="!assets.items.length">
            <td colspan="6" class="text-center text-disabled">
              {{ loading ? "Loading..." : "No assets" }}
            </td>
          </tr>
        </tbody>
      </v-table>

      <div class="d-flex align-center justify-end mt-2">
        <span class="text-caption mr-4">{{ pageLabel }}</span>
        <v-pagination
          v-if="pages > 1"
          v-model="page"
          :length="pages"
          :total-visible="5"
          density="compact"
        />
      </div>
      </template>
    </nocloud-expansion-panels>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, toRefs, watch } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import NocloudExpansionPanels from "@/components/ui/nocloudExpansionPanels.vue";
import MondooFindingsTable from "./findingsTable.vue";
import { ratings } from "./serviceAccount";

const props = defineProps(["template", "sp", "account", "addons"]);
const emit = defineEmits(["update"]);
const { template } = toRefs(props);

const store = useStore();
const pageSize = 15;

const tab = ref("findings");
const page = ref(1);
const loading = ref(false);
const error = ref("");
const findings = reactive({ total: 0 });
const assets = reactive({ items: [], total: 0 });
const refreshKey = ref(0);
const findingsFilter = reactive({ rating: "", state: "", namespace: "", search: "" });
// Typing must not fire a request per key.
const searchText = ref("");
let searchTimer;
watch(searchText, (value) => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => (findingsFilter.search = (value ?? "").trim()), 350);
});

const ratingItems = [{ key: "", title: "Any rating" }, ...ratings];
const stateItems = [
  { key: "", title: "Open" },
  { key: "CLOSED", title: "Passing" },
  { key: "ALL", title: "All" },
];

const data = computed(() => template.value?.data ?? {});
const config = computed(() => template.value?.config ?? {});
// The copy in instance data is refreshed by the driver's monitoring pass and
// can lag a sync; ask the driver directly and fall back to the copy.
const liveSummary = ref(null);
const summary = computed(() => liveSummary.value ?? data.value.summary ?? {});
const byRating = computed(() => summary.value.findings_by_rating ?? {});
const namespaces = computed(() => summary.value.namespaces ?? []);
const domains = computed(() => config.value.domains ?? []);
const domainItems = computed(() => [
  { title: "All domains", value: "" },
  ...namespaces.value.map(({ name }) => ({ title: name, value: name })),
]);
const imported = computed(() => !!data.value.imported);
const isSsh = computed(() => summary.value.kind === "ssh" || !!config.value.instance);
const scan = computed(() => summary.value.scan ?? {});
// The last fix run on the machine, and whether it is still under way.
const fix = computed(() => summary.value.remediation ?? {});
const fixActive = computed(() => ["requested", "running"].includes(fix.value.status));
const fixColor = computed(
  () => ({ done: "success", failed: "warning" })[fix.value.status] ?? "primary"
);
const integrationMrn = computed(() => data.value.integration_mrn);

const formatTs = (seconds) =>
  seconds ? new Date(seconds * 1000).toLocaleString() : "";
const syncedAt = computed(() => formatTs(+summary.value.synced_at) || "never");
const trialEnds = computed(() => formatTs(+config.value.trial_ends));
const trialActive = computed(() => +config.value.trial_ends * 1000 > Date.now());

const total = computed(() => assets.total);
const pages = computed(() => Math.ceil(total.value / pageSize));
const pageLabel = computed(() => {
  if (!total.value) return "";
  const from = (page.value - 1) * pageSize + 1;
  return `${from}-${Math.min(from + pageSize - 1, total.value)} of ${total.value}`;
});

const shortDate = (value) => {
  if (!value) return "-";
  const date = new Date(typeof value === "number" ? value * 1000 : value);
  return isNaN(date) ? "-" : date.toLocaleDateString();
};

async function invoke(action, params) {
  const response = await api.instances.action({
    uuid: template.value.uuid,
    action,
    params,
  });
  return response?.meta ?? {};
}

// The assets page; findings tables load themselves.
async function load() {
  if (!template.value?.uuid || tab.value !== "assets") return;
  loading.value = true;
  error.value = "";
  try {
    const meta = await invoke("assets", {
      limit: pageSize,
      offset: (page.value - 1) * pageSize,
    });
    assets.items = meta.items ?? [];
    assets.total = +meta.total || 0;
  } catch (err) {
    error.value =
      err?.response?.data?.message ?? err?.message ?? "Failed to load results";
  } finally {
    loading.value = false;
  }
}

async function loadCounts() {
  // Totals for the toggle, without waiting for a tab switch.
  try {
    const [a, live] = await Promise.all([
      invoke("assets", { limit: 1 }),
      invoke("summary"),
    ]);
    liveSummary.value = live;
    assets.total = +a.total || 0;
  } catch {
    // The tables' own loads report errors.
  }
}

function reload() {
  loadCounts();
  load();
  refreshKey.value++;
}

// After a fix is requested, follow it: its status first, then the scan that
// comes after it, and reload the findings once that scan is in.
let fixTimer;
function watchFix() {
  clearInterval(fixTimer);
  let ticks = 0;
  fixTimer = setInterval(async () => {
    await loadCounts();
    const rescanned = +scan.value.finished_at >= +fix.value.finished_at;
    if (++ticks > 72 || (!fixActive.value && fix.value.finished_at && rescanned)) {
      clearInterval(fixTimer);
      refreshKey.value++;
    }
  }, 5000);
}
onBeforeUnmount(() => clearInterval(fixTimer));

function filterRating(rating) {
  tab.value = "findings";
  findingsFilter.rating = findingsFilter.rating === rating ? "" : rating;
}

function changeDomains(values) {
  const next = [
    ...new Set(
      (values ?? [])
        .map((d) =>
          String(d).trim().toLowerCase().replace(/^https?:\/\//, "").replace(/\/.*$/, "")
        )
        .filter(Boolean)
    ),
  ];
  emit("update", { key: "config.domains", value: next });
}

function copy(text) {
  navigator.clipboard?.writeText(text);
  store.commit("snackbar/showSnackbar", { message: "Copied" });
}

watch(tab, () => {
  page.value = 1;
  load();
});
watch(page, load);

onMounted(reload);
</script>

<script>
export default {
  name: "instance-mondoo-info",
};
</script>

<style scoped>
.mondoo-dim {
  opacity: 0.45;
}
.mondoo-table {
  background: transparent;
}
</style>
