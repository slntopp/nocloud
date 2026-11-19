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
          <th>Host</th>
          <th v-if="!namespace">Domain</th>
          <th>Type</th>
          <th>State</th>
          <th>Updated</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="item.asset + item.mrn">
          <td>
            <v-chip :color="ratingColor(item.rating)" variant="flat" size="x-small">
              {{ item.rating }}
            </v-chip>
          </td>
          <td>{{ item.title }}</td>
          <td>{{ item.asset_name }}</td>
          <td v-if="!namespace">{{ item.namespace }}</td>
          <td>{{ kindTitle(item.kind) }}</td>
          <td>{{ item.state }}</td>
          <td class="text-no-wrap">{{ shortDate(item.last_updated) }}</td>
        </tr>
        <tr v-if="!items.length">
          <td :colspan="namespace ? 6 : 7" class="text-center text-disabled">
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
import { ratingColor } from "./serviceAccount";

const props = defineProps({
  uuid: { type: String, required: true },
  rating: { type: String, default: "" },
  state: { type: String, default: "" },
  // Limits the list to one domain namespace; all of them when empty.
  namespace: { type: String, default: "" },
  pageSize: { type: Number, default: 15 },
  // Changing it reloads the current page.
  refreshKey: { type: Number, default: 0 },
});
const emit = defineEmits(["total"]);

const page = ref(1);
const loading = ref(false);
const error = ref("");
const items = ref([]);
const total = ref(0);

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
      action: "findings",
      params: {
        limit: props.pageSize,
        offset: (page.value - 1) * props.pageSize,
        ...(props.rating && { rating: props.rating }),
        ...(props.state && { state: props.state }),
        ...(props.namespace && { namespace: props.namespace }),
      },
    });
    const meta = response?.meta ?? {};
    items.value = meta.items ?? [];
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
  () => [props.rating, props.state, props.namespace],
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
</style>
