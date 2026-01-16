<template>
  <v-card>
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
import { computed, ref, toRefs, watch } from "vue";
import NocloudTable from "@/components/table.vue";
import api from "@/api";
import { debounce } from "@/functions";

const props = defineProps({
  addon: {},
});
const { addon } = toRefs(props);

const store = useStore();

const page = ref(1);
const options = ref({});
const fetchError = ref("");

const expanded = ref([]);
const updatingId = ref("");

const planHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Enabled", value: "enabled" },
]);
const productHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Enabled", value: "enabled" },
]);

const plans = computed(() => store.getters["plans/all"]);
const isPlansLoading = computed(() => store.getters["plans/loading"]);
const total = computed(() => store.getters["plans/total"]);

const searchParam = computed(() => store.getters["appSearch/param"]);
const filters = computed(() => {
  var filters = { "meta.isIndividual": [false] };
  if (searchParam.value) {
    filters[searchParam] = searchParam.value;
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
        (addonId) => addonId !== addon.value.uuid
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
        (addonId) => addonId !== addon.value.uuid
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

watch(requestOptions, fetchPlansDebounce, { deep: true });
</script>

<script>
export default { name: "addon-products" };
</script>

<style scoped></style>
