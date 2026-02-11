<template>
  <v-card color="background-light" class="pa-4">
    <div class="d-flex align-center">
      <v-switch label="Individual" v-model="isIndividual" />
    </div>

    <nocloud-table
      table-name="addons-products-table"
      :loading="isPlansLoading"
      :headers="planHeaders"
      :items="items"
      show-expand
      item-key="id"
      :show-select="false"
      :expanded.sync="expanded"
      :footer-error="fetchError"
      @update:options="setOptions"
      :server-items-length="total"
      :server-side-page="page"
    >
      <template v-slot:[`item.enabled`]="{ item }">
        <v-skeleton-loader v-if="updatingId === item.id" type="text" />
        <v-switch
          v-else
          dense
          hide-details
          :disabled="!!updatingId"
          :input-value="item.enabled"
          @change="changePlanAddons(item, $event)"
        />
      </template>
      <template v-slot:expanded-item="{ headers, item: plan }">
        <td :colspan="headers.length" style="padding: 0">
          <nocloud-table
            :server-items-length="-1"
            hide-default-footer
            :show-select="false"
            :headers="productHeaders"
            :items="plan.children"
          >
            <template v-slot:[`item.enabled`]="{ item }">
              <v-skeleton-loader v-if="updatingId === item.id" type="text" />
              <v-switch
                v-else
                dense
                hide-details
                :disabled="!!updatingId || plan.enabled"
                :input-value="item.enabled"
                @change="changeProductAddons(item, $event)"
              />
            </template>
          </nocloud-table>
        </td>
      </template>
    </nocloud-table>
  </v-card>
</template>

<script setup>
import { useStore } from "@/store";
import { computed, onMounted, ref, toRefs, watch } from "vue";
import NocloudTable from "@/components/table.vue";
import api from "@/api";
import { debounce } from "@/functions";
import { PlanKind } from "nocloud-proto/proto/es/billing/billing_pb";
import { NoCloudStatus } from "nocloud-proto/proto/es/statuses/statuses_pb";
import useSearch from "@/hooks/useSearch";

const props = defineProps({
  addon: {},
});

useSearch({
  name: "addon-products-tab",
  defaultLayout: {
    title: "Default",
    filter: {
      public: [],
      status: [],
    },
  },
});

const { addon } = toRefs(props);

const store = useStore();

const page = ref(1);
const options = ref({});
const fetchError = ref("");

const statusMap = {
  DEL: { title: "DELETED", color: "blue-grey darken-2" },
  UNSPECIFIED: { title: "ACTIVE", color: "success" },
};

const expanded = ref([]);
const updatingId = ref("");
const isIndividual = ref(false);
const unique = ref({});

const planHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Enabled", value: "enabled" },
]);

const productHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Enabled", value: "enabled" },
]);

onMounted(() => {
  setSearchFields();
  fetchUnique();
});

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

const plans = computed(() => store.getters["plans/all"]);
const isPlansLoading = computed(() => store.getters["plans/loading"]);
const total = computed(() => store.getters["plans/total"]);

const searchParam = computed(() => store.getters["appSearch/param"]);
const filter = computed(() => store.getters["appSearch/filter"]);

const filters = computed(() => {
  const filters = Object.keys(filter.value).reduce((newFilter, key) => {
    if (filter.value[key] === undefined || filter.value[key] === null) {
      delete newFilter[key];
    } else {
      newFilter[key] = filter.value[key];
    }

    return newFilter;
  }, {});

  filters["meta.isIndividual"] = [isIndividual.value];

  if (searchParam.value || filters.title) {
    filters.search_param = searchParam.value || filters.title;
  }

  return filters;
});

const requestOptions = computed(() => ({
  filters: filters.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));

const items = computed(() => {
  return plans.value.map((plan) => {
    const children = Object.keys(plan.products || {}).map((key) => ({
      id: key,
      name: plan.products[key].title,
      plan: plan.uuid,
      enabled:
        plan.addons?.includes(addon.value.uuid) ||
        plan.products[key].addons?.includes(addon.value.uuid),
    }));
    return {
      id: plan.uuid,
      name: plan.title,
      enabled: plan.addons?.includes(addon.value.uuid),
      children,
    };
  });
});

const fetchPlans = async () => {
  fetchError.value = "";
  try {
    await store.dispatch("plans/fetch", requestOptions.value);
  } catch (e) {
    fetchError.value = e.message;
  }
};

const fetchUnique = async () => {
  const response = await store.getters["plans/plansClient"].plansUnique();

  unique.value = response.toJson().unique;
};

const fetchPlansDebounce = debounce(fetchPlans, 500);

const setOptions = (newOptions) => {
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const changePlanAddons = async (item, val) => {
  try {
    const plan = plans.value.find((p) => p.uuid === item.id);

    updatingId.value = plan.uuid;
    if (!plan.addons) {
      plan.addons = [];
    }

    if (val) {
      plan.addons.push(addon.value.uuid);
    } else {
      plan.addons = plan.addons.filter(
        (addonId) => addonId !== addon.value.uuid,
      );
    }
    await api.plans.update(plan.uuid, plan);

    item.enabled = !!val;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    updatingId.value = "";
  }
};

const changeProductAddons = async (item, val) => {
  try {
    const plan = plans.value.find((p) => p.uuid === item.plan);
    const product = plan.products[item.id];

    if (!product.addons) {
      product.addons = [];
    }

    updatingId.value = item.id;

    if (val) {
      product.addons.push(addon.value.uuid);
    } else {
      product.addons = plan.addons.filter(
        (addonId) => addonId !== addon.value.uuid,
      );
    }

    plan.products[item.id] = product;
    item.enabled = val;
    await api.plans.update(plan.uuid, plan);
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    updatingId.value = "";
  }
};

const setSearchFields = () => {
  store.commit("appSearch/setFields", searchFields.value);
};

watch(searchFields, setSearchFields, { deep: true });
watch(requestOptions, fetchPlansDebounce, { deep: true });

watch(
  filters,
  (newVal, oldVal) => {
    if (JSON.stringify(newVal) !== JSON.stringify(oldVal)) {
      page.value = 1;
    }
  },
  { deep: true },
);
</script>

<script>
export default { name: "addon-products" };
</script>

<style scoped></style>
