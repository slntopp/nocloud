<template>
  <div>
    <v-alert v-if="error" type="warning" variant="tonal" density="compact" class="mb-2">
      {{ error }}
    </v-alert>

    <v-table density="compact" class="mondoo-table">
      <thead>
        <tr>
          <th>Rating</th>
          <th>Finding</th>
          <th>Type</th>
          <th>Affected hosts</th>
          <th>Updated</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <template v-for="item in items" :key="item.mrn">
          <tr class="mondoo-row" @click="toggle(item.mrn)">
            <td>
              <v-chip :color="ratingColor(item.rating)" variant="flat" size="x-small">
                {{ item.rating }}
              </v-chip>
            </td>
            <td>
              {{ item.title }}
              <span v-if="item.id" class="text-caption text-medium-emphasis ml-1">{{ item.id }}</span>
            </td>
            <td>{{ kindTitle(item.kind) }}</td>
            <td>{{ item.assets }}</td>
            <td class="text-no-wrap">{{ shortDate(item.last_updated) }}</td>
            <td class="text-right">
              <v-icon size="small">
                {{ open === item.mrn ? "mdi-chevron-up" : "mdi-chevron-down" }}
              </v-icon>
            </td>
          </tr>
          <tr v-if="open === item.mrn" class="mondoo-details">
            <td colspan="6">
              <mondoo-finding-details
                :uuid="uuid"
                :mrn="item.mrn"
                :rating="rating"
                :state="state"
                :namespace="namespace"
                @fix="emit('fix')"
              />
            </td>
          </tr>
        </template>
        <tr v-if="!items.length">
          <td colspan="6" class="text-center text-disabled">
            {{ loading ? "Loading..." : "No findings" }}
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
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import api from "@/api";
import MondooFindingDetails from "./findingDetails.vue";
import { ratingColor } from "./serviceAccount";

const props = defineProps({
  uuid: { type: String, required: true },
  rating: { type: String, default: "" },
  state: { type: String, default: "" },
  // Limits the list to one domain namespace; all of them when empty.
  namespace: { type: String, default: "" },
  // Free text over finding titles and CVE/advisory ids.
  search: { type: String, default: "" },
  pageSize: { type: Number, default: 15 },
  // Changing it reloads the current page.
  refreshKey: { type: Number, default: 0 },
});
const emit = defineEmits(["total", "fix"]);

const page = ref(1);
const loading = ref(false);
const error = ref("");
const items = ref([]);
const total = ref(0);
// MRN of the finding whose details are shown.
const open = ref("");
const toggle = (mrn) => (open.value = open.value === mrn ? "" : mrn);

const pages = computed(() => Math.ceil(total.value / props.pageSize));
const pageLabel = computed(() => {
  if (!total.value) return "";
  const from = (page.value - 1) * props.pageSize + 1;
  return `${from}-${Math.min(from + props.pageSize - 1, total.value)} of ${total.value}`;
});

const kindTitle = (kind) => String(kind ?? "").replace(/Finding$/, "");
const shortDate = (value) => {
  if (!value) return "-";
  const date = new Date(typeof value === "number" ? value * 1000 : value);
  return isNaN(date) ? "-" : date.toLocaleDateString();
};

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const response = await api.instances.action({
      uuid: props.uuid,
      action: "finding_groups",
      params: {
        limit: props.pageSize,
        offset: (page.value - 1) * props.pageSize,
        ...(props.rating && { rating: props.rating }),
        ...(props.state && { state: props.state }),
        ...(props.namespace && { namespace: props.namespace }),
        ...(props.search && { search: props.search }),
      },
    });
    const meta = response?.meta ?? {};
    items.value = meta.items ?? [];
    if (!items.value.some(({ mrn }) => mrn === open.value)) open.value = "";
    total.value = +meta.total || 0;
    emit("total", total.value);
  } catch (err) {
    error.value =
      err?.response?.data?.message ?? err?.message ?? "Failed to load findings";
  } finally {
    loading.value = false;
  }
}

watch(page, load);
watch(() => props.refreshKey, load);
watch(
  () => [props.rating, props.state, props.namespace, props.search],
  () => {
    if (page.value !== 1) page.value = 1;
    else load();
  }
);

onMounted(load);
</script>

<script>
export default {
  name: "mondoo-findings-table",
};
</script>

<style scoped>
.mondoo-table {
  background: transparent;
}
.mondoo-row {
  cursor: pointer;
}
.mondoo-details > td {
  background: rgba(128, 128, 128, 0.06);
}
</style>
