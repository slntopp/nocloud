<template>
  <nocloud-table
    table-name="services"
    show-expand
    :value="value"
    @input="emit('input', $event)"
    :items="services"
    :headers="headers"
    :loading="isLoading"
    :expanded.sync="expanded"
    :footer-error="fetchError"
    :server-items-length="total"
    :server-side-page="options.page"
    @update:options="setOptions"
  >
    <template v-slot:[`item.hash`]="{ item, index }">
      <v-btn icon @click="addToClipboard(item.hash, index)">
        <v-icon v-if="copyed == index"> mdi-check </v-icon>
        <v-icon v-else> mdi-content-copy </v-icon>
      </v-btn>
      {{ hashTrim(item.hash) }}
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'Service', params: { serviceId: item.uuid } }">
        {{ getShortName(item.title, 45) }}
      </router-link>
    </template>

    <template v-slot:[`item.status`]="{ value }">
      <v-chip small :color="chipColor(value)">
        {{ value }}
      </v-chip>
    </template>
    <template v-slot:[`item.access`]="{ item }">
      <v-chip
        v-if="!isNamespacesLoading"
        :color="accessColor(item.access?.level)"
      >
        {{ getName(item.access?.namespace) }} ({{
          item.access?.level ?? "NONE"
        }})
      </v-chip>
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:expanded-item="{ headers, item }">
      <td :colspan="headers.length" style="padding: 0">
        <div v-for="(itemService, index) in services" :key="index">
          <v-expansion-panels
            inset
            multiple
            v-model="opened[index]"
            v-if="item.uuid == itemService.uuid"
          >
            <v-expansion-panel
              style="background: var(--v-background-light-base)"
              v-for="(group, i) in itemService.instancesGroups"
              :key="i"
              :disabled="!group.instances.length"
            >
              <v-expansion-panel-header>
                {{ group.title }} | Type: {{ group.type }} -
                {{ titleSP(group) }}
                <v-chip
                  small
                  class="instance-group-status"
                  :color="instanceCountColor(group)"
                >
                  {{ group.instances.length }}
                </v-chip>
              </v-expansion-panel-header>
              <v-expansion-panel-content
                style="background: var(--v-background-base)"
              >
                <service-instances-item
                  :instances="group.instances"
                  :spId="group.sp"
                  :type="group.type"
                />
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
        </div>
      </td>
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import serviceInstancesItem from "@/components/service_instances_item.vue";
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { debounce, getShortName } from "@/functions";
import api from "@/api";

const props = defineProps(["value", "refetch"]);
const { value, refetch } = toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();

const expanded = ref([]);
const headers = ref([
  { text: "Title", value: "title" },
  { text: "Status", value: "status" },
  { text: "UUID", value: "uuid", align: "start" },
  { text: "Hash", value: "hash" },
  { text: "Access", value: "access" },
]);
const opened = ref({});
const fetchError = ref("");
const copyed = ref(-1);
const options = ref({});

const namespaces = ref({});
const isNamespacesLoading = ref(false);

const isLoading = computed(() => store.getters["services/isLoading"]);
const services = computed(() => store.getters["services/all"]);
const total = computed(() => store.getters["services/total"]);
const servicesProviders = computed(
  () => store.getters["servicesProviders/all"]
);

const searchParam = computed(() => store.getters["appSearch/param"]);
const filter = computed(() => store.getters["appSearch/filter"]);
const requestOptions = computed(() => ({
  filters: {
    ...filter.value,
    search_param: filter.value.title || searchParam.value || undefined,
  },
  page: options.value.page,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
}));

const searchFields = computed(() => {
  return [
    {
      type: "input",
      title: "Title",
      key: "title",
    },
    {
      items: Object.values(stateMap),
      type: "select",
      item: {
        value: "id",
        title: "title",
      },
      title: "Status",
      key: "status",
    },
    {
      key: "access.level",
      items: Object.values(accessMap),
      item: {
        value: "id",
        title: "title",
      },
      type: "select",
      title: "Access",
    },
  ];
});

onMounted(() => {
  store.commit("appSearch/setFields", searchFields.value);
});

const setOptions = (newOptions) => {
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const fetchServices = async () => {
  try {
    await store.dispatch("services/fetch", requestOptions.value);
  } catch (err) {
    console.log(`err`, err);
    fetchError.value = "Can't reach the server";
    if (err.response) {
      fetchError.value += `: [ERROR]: ${err.response.data.message}`;
    } else {
      fetchError.value += `: [ERROR]: ${err.toJSON().message}`;
    }
  }
};

const fetchServicesDebounce = debounce(fetchServices);

const chipColor = (state) => {
  return stateMap[state]?.color ?? "blue-grey darken-2";
};
const accessColor = (level) => {
  return accessMap[level]?.color;
};

const titleSP = (group) => {
  const data = servicesProviders.value.find((el) => el.uuid == group?.sp);
  return data?.title || "not found";
};
const getName = (namespace) => {
  return namespaces.value[namespace]?.title ?? "";
};

const hashTrim = (hash) => {
  if (hash) return hash.slice(0, 8) + "...";
  else return "XXXXXXXX...";
};
const addToClipboard = async (text, index) => {
  if (navigator?.clipboard) {
    await navigator.clipboard.writeText(text);
    copyed.value = index;
  } else {
    alert("Clipboard is not supported!");
  }
};
const instanceCountColor = (group) => {
  if (group.instances.length) {
    return chipColor(group.status);
  }
  return chipColor("DEL");
};

watch(services, (value) => {
  fetchError.value = "";

  value.forEach(async ({ access: { namespace: uuid } }) => {
    if (!uuid) {
      return;
    }

    isNamespacesLoading.value = true;
    try {
      if (!namespaces.value[uuid]) {
        namespaces.value[uuid] = api.get("namespaces/" + uuid);
        namespaces.value[uuid] = await namespaces.value[uuid];
      }
    } catch {
      namespaces.value[uuid] = undefined;
    } finally {
      isNamespacesLoading.value = Object.values(namespaces.value).some(
        (acc) => acc instanceof Promise
      );
    }
  });
});

watch(filter, fetchServicesDebounce, { deep: true });
watch(options, fetchServicesDebounce);
watch(searchParam, fetchServicesDebounce);
watch(refetch, fetchServicesDebounce);
</script>

<script>
import search from "@/mixins/search.js";

const stateMap = {
  INIT: { color: "orange darken-2", id: 1, title: "INIT" },
  SUS: { color: "orange darken-2", title: "SUSPENDED", id: 7 },
  UP: { color: "green darken-2", id: 3, title: "UP" },
  DEL: { color: "gray darken-2", id: 5, title: "DEL" },
  UNSPECIFIED: { color: "red darken-2", title: "UNKNOWN", id: 0 },
  DOWN: { color: "orange darken-2", id: 4, title: "STOPPED" },
};

const accessMap = {
  ROOT: { color: "info", id: 4, title: "ROOT" },
  ADMIN: { color: "success", id: 3, title: "ADMIN" },
  MGMT: { color: "warning", id: 2, title: "MGMT" },
  READ: { color: "gray", id: 1, title: "READ" },
  NONE: { color: "error", id: 0, title: "NONE" },
};

export default {
  name: "services-table",
  mixins: [
    search({
      name: "services",
      defaultLayout: {
        title: "Default",
        filter: {
          status: Object.values(stateMap)
            .filter((s) => s.title !== "DEL")
            .map((s) => s.id),
        },
      },
    }),
  ],
};
</script>
