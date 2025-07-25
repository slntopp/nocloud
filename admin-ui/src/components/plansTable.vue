<template>
  <nocloud-table
    :table-name="tableName"
    class="mt-4"
    :items="plans"
    :headers="headers"
    :value="value"
    @input="emit('input', $event)"
    :loading="isLoading"
    :footer-error="fetchError"
    @update:options="setOptions"
    :server-items-length="total"
    :server-side-page="page"
  >
    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'Plan', params: { planId: item.uuid } }">
        {{ getShortName(item.title, 45) }}
      </router-link>
    </template>

    <template v-slot:[`item.meta.auto_start`]="{ item }">
      <v-skeleton-loader v-if="updatedPlanUuid === item.uuid" type="text" />
      <v-switch
        v-else
        dense
        hide-details
        :readonly="!!updatedPlanUuid"
        :input-value="item.meta?.auto_start"
        :disabled="isDeleted(item)"
        @change="
          updatePlan(item, {
            key: 'meta',
            value: { ...(item.meta || {}), auto_start: $event },
          })
        "
      />
    </template>

    <template v-slot:[`item.properties.autoRenew`]="{ item }">
      <v-skeleton-loader v-if="updatedPlanUuid === item.uuid" type="text" />
      <v-switch
        v-else
        dense
        hide-details
        :readonly="!!updatedPlanUuid"
        :input-value="item.properties?.autoRenew"
        :disabled="isDeleted(item)"
        @change="
          updatePlan(item, {
            key: 'properties.autoRenew',
            value: $event,
          })
        "
      />
    </template>

    <template v-slot:[`item.public`]="{ item }">
      <v-skeleton-loader v-if="updatedPlanUuid === item.uuid" type="text" />
      <v-switch
        v-else
        dense
        hide-details
        :readonly="!!updatedPlanUuid"
        :input-value="item.public"
        :disabled="isDeleted(item)"
        @change="
          updatePlan(item, {
            key: 'public',
            value: $event,
          })
        "
      />
    </template>
    <template v-slot:[`item.kind`]="{ value }">
      {{ value?.toLowerCase && value?.toLowerCase() }}
    </template>
    <template v-slot:[`item.instanceCount`]="{ item }">
      <v-progress-circular
        v-if="isInstanceCountLoading"
        size="20"
        indeterminate
      />
      <template v-else>
        {{ instanceCountMap[item.uuid] }}
      </template>
    </template>
    <template v-slot:[`item.status`]="{ item }">
      <v-chip small :color="getStatus(item).color">
        {{ getStatus(item).title }}
      </v-chip>
    </template>
    <template v-slot:footer.prepend>
      <div class="d-flex align-center mt-2">
        <v-select
          style="max-width: 80px"
          :items="fileTypes"
          label="File type"
          v-model="selectedFileType"
          class="d-inline-block mx-1"
        />
        <download-template-button
          @click:xlsx="downloadXlsx"
          class="mx-1"
          small
          title="Download"
          :disabled="!value.length || isPlansUploadLoading"
          name="selected_copy"
          :type="selectedFileType"
          :template="value"
        />
        <v-file-input
          class="file-input mx-1 mt-4"
          :label="`upload ${selectedFileType} price models...`"
          :accept="`.${selectedFileType}`"
          @change="onJsonInputChange"
        />
        <confirm-dialog
          @confirm="uploadPlans"
          :disabled="!uploadedPlans.length"
          :text="uploadedPlansText"
        >
          <v-btn
            :disabled="!uploadedPlans.length"
            :loading="isPlansUploadLoading"
            small
            >Upload</v-btn
          >
        </confirm-dialog>
      </div>
    </template>
  </nocloud-table>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";
import {
  downloadPlanXlsx,
  readJSONFile,
  readYAMLFile,
  getShortName,
  debounce,
} from "@/functions";
import { NoCloudStatus } from "nocloud-proto/proto/es/statuses/statuses_pb";
import { Plan, PlanKind } from "nocloud-proto/proto/es/billing/billing_pb";
import useSearch from "@/hooks/useSearch";

const statusMap = {
  DEL: { title: "DELETED", color: "blue-grey darken-2" },
  UNSPECIFIED: { title: "ACTIVE", color: "success" },
};

const props = defineProps({
  value: {},
  tableName: { type: String, default: "plans-table" },
  showSelect: { type: Boolean, default: false },
  customParams: { type: Object, default: () => ({}) },
  customHeaders: { type: Array, default: () => [] },
  refetch: { type: Boolean, default: false },
  noSearch: { type: Boolean, default: false },
  plans: { type: Array, default: () => [] },
  total: { type: Number, default: 0 },
  isLoading: { type: Boolean, default: false },
});
const {
  value,
  refetch,
  customParams,
  noSearch,
  isLoading,
  plans,
  total,
  customHeaders,
} = toRefs(props);

const emit = defineEmits(["input", "update:options", "fetch:plans"]);

const store = useStore();
useSearch({
  name: props.tableName,
  noSearch: props.noSearch,
  defaultLayout: {
    title: "Default",
    filter: {
      public: [true],
      status: [NoCloudStatus.UNSPECIFIED],
    },
  },
});

const fetchError = ref("");
const options = ref({});
const page = ref(1);

const unique = ref({});

const isInstanceCountLoading = ref(true);
const instanceCountMap = ref({});

const updatedPlanUuid = ref("");

const isPlansUploadLoading = ref(false);
const uploadedPlans = ref([]);
const selectedFileType = ref("JSON");
const fileTypes = ref(["JSON", "YAML", "XLSX"]);

onMounted(() => {
  setSearchFields();

  fetchUnique();
});

const headers = computed(() => {
  if (!customHeaders.value.length) {
    return [
      { text: "Title ", value: "title" },
      { text: "UUID ", value: "uuid" },
      { text: "Kind ", value: "kind" },
      { text: "Type ", value: "type" },
      { text: "Status ", value: "status" },
      { text: "Public ", value: "public" },
      { text: "Auto start ", value: "meta.auto_start" },
      { text: "Automatic debit", value: "properties.autoRenew" },
      { text: "Linked instances count ", value: "instanceCount" },
    ];
  }
  return customHeaders.value;
});

const filter = computed(() => store.getters["appSearch/filter"]);
const param = computed(() => store.getters["appSearch/param"]);

const filters = computed(() => {
  const filters = Object.keys(filter.value).reduce((newFilter, key) => {
    if (filter.value[key] === undefined || filter.value[key] === null) {
      delete newFilter[key];
    } else {
      newFilter[key] = filter.value[key];
    }

    return newFilter;
  }, {});

  for (const key of Object.keys(customParams.value.filters || {})) {
    filters[key] = customParams.value.filters[key];
  }

  if (param.value || filters.title) {
    filters.search_param = param.value || filters.title;
  }

  return filters;
});

const requestOptions = computed(() => ({
  ...customParams.value,
  filters: filters.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));

const searchFields = computed(() => [
  {
    title: "Title",
    key: "title",
    type: "input",
  },
  {
    items: Object.keys(PlanKind)
      .filter((value) => !Number.isInteger(+value))
      .map((key) => ({
        text: key,
        value: PlanKind[key],
      })),
    title: "Kind",
    key: "kind",
    type: "select",
  },
  {
    items: unique.value.type?.map((type) => ({ text: type, value: type })),
    title: "Type",
    key: "type",
    type: "select",
  },
  {
    items: Object.keys(statusMap).map((key) => ({
      text: statusMap[key].title,
      value: NoCloudStatus[key],
    })),
    title: "Status",
    key: "status",
    type: "select",
  },
  {
    title: "Public",
    key: "public",
    type: "select",
    items: [
      { text: "Enabled", value: true },
      { text: "Disabled", value: false },
    ],
  },
]);

const uploadedPlansText = computed(
  () =>
    "Uploaded plans:<br/>" +
    uploadedPlans.value.map((p) => p.title).join("<br/>")
);
const isJson = computed(() => selectedFileType.value === "JSON");

const fetchPlans = async () => {
  fetchError.value = "";
  try {
    await emit("fetch:plans", requestOptions.value);
  } catch (e) {
    fetchError.value = e.message;
  }
};

const fetchPlansDebounced = debounce(fetchPlans, 300);

const fetchInstancesCount = async () => {
  isInstanceCountLoading.value = true;
  try {
    instanceCountMap.value = (
      await store.getters["plans/plansClient"].listPlansInstances({
        uuids: plans.value.map((p) => p.uuid),
      })
    ).plans;
  } finally {
    isInstanceCountLoading.value = false;
  }
};

const fetchUnique = async () => {
  const response = await store.getters["plans/plansClient"].plansUnique();

  unique.value = response.toJson().unique;
};

const setOptions = (newOptions) => {
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const updatePlan = async (item, { key, value }) => {
  try {
    updatedPlanUuid.value = item.uuid;
    const newPlan = { ...item };

    const subkeys = key.split(".");

    if (subkeys.length === 1) {
      newPlan[subkeys[0]] = value;
    } else {
      let data = newPlan;

      subkeys.forEach((subkey, index) => {
        if (index === subkeys.length - 1) {
          data[subkey] = value;
        } else {
          if (!data[subkey] || typeof data[subkey] !== "object") {
            data[subkey] = {};
          }
          data = data[subkey];
        }
      });
    }

    await store.getters["plans/plansClient"].updatePlan(
      Plan.fromJson(newPlan, { ignoreUnknownFields: true })
    );
    item[key] = value;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during update plan",
    });
  } finally {
    updatedPlanUuid.value = "";
  }
};

const getStatus = (item) => {
  return statusMap[item.status] || statusMap.UNSPECIFIED;
};

const isDeleted = (plan) => {
  return plan.status === "DEL";
};

const setSearchFields = () => {
  if (noSearch.value) {
    return;
  }
  store.commit("appSearch/setFields", searchFields.value);
};

const downloadXlsx = () => {
  return downloadPlanXlsx(value.value);
};

const onJsonInputChange = async (file) => {
  uploadedPlans.value = [];
  try {
    const data = isJson.value
      ? await readJSONFile(file)
      : await readYAMLFile(file);
    if (Array.isArray(data)) {
      uploadedPlans.value.push(...data);
    } else {
      uploadedPlans.value.push(data);
    }
  } catch (err) {
    uploadedPlans.value = [];
    store.commit("snackbar/showSnackbarError", {
      message: err,
    });
  }
};

const uploadPlans = async () => {
  isPlansUploadLoading.value = true;
  try {
    await Promise.all(
      uploadedPlans.value.map((p) =>
        store.getters["plans/plansClient"].createPlan({
          ...p,
          uuid: undefined,
        })
      )
    );
    fetchPlansDebounced();
  } catch (err) {
    store.commit("snackbar/showSnackbarError", {
      message: err,
    });
  } finally {
    uploadedPlans.value = [];
    isPlansUploadLoading.value = false;
  }
};

watch(searchFields, setSearchFields, { deep: true });

watch(
  filters,
  (newVal, oldVal) => {
    if (JSON.stringify(newVal) !== JSON.stringify(oldVal)) {
      page.value = 1;
      fetchPlansDebounced();
    }
  },
  { deep: true }
);

watch(
  customParams,
  (newVal, oldVal) => {
    if (JSON.stringify(newVal) !== JSON.stringify(oldVal)) {
      fetchPlansDebounced();
    }
  },
  { deep: true }
);

watch([options, refetch], fetchPlansDebounced);

watch(plans, () => {
  if (plans.value.length) {
    fetchInstancesCount();
  }
});
</script>
