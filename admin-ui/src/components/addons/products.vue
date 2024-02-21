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
      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length" style="padding: 0">
          <nocloud-table
            :server-items-length="-1"
            hide-default-footer
            :show-select="false"
            :headers="productHeaders"
            :items="item.children"
          >
            <template v-slot:[`item.enabled`]="{ item }">
              <v-skeleton-loader v-if="updatingId === item.id" type="text" />
              <v-switch
                v-else
                dense
                hide-details
                :disabled="!!updatingId"
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
import { computed, onMounted, ref, toRefs } from "vue";
import NocloudTable from "@/components/table.vue";
import api from "@/api";

const props = defineProps({
  addon: {},
});
const { addon } = toRefs(props);

const store = useStore();

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

onMounted(() => {
  store.dispatch("plans/fetch");
});

const plans = computed(() => store.getters["plans/all"]);
const isPlansLoading = computed(() => store.getters["plans/isLoading"]);

const items = computed(() => {
  return plans.value.map((plan) => {
    const children = Object.keys(plan.products).map((key) => ({
      id: key,
      name: plan.products[key].title,
      plan: plan.uuid,
      enabled:
        plan.addons.includes(addon.value.uuid) ||
        plan.products[key].addons?.includes(addon.value.uuid),
    }));
    return {
      id: plan.uuid,
      name: plan.title,
      enabled: plan.addons.includes(addon.value.uuid),
      children,
    };
  });
});

const changePlanAddons = async (item, val) => {
  try {
    const plan = plans.value.find((p) => p.uuid === item.id);

    updatingId.value = plan.uuid;
    if (val) {
      plan.addons.push(addon.value.uuid);
    } else {
      plan.addons = plan.addons.filter(
        (addonId) => addonId !== addon.value.uuid
      );
    }
    await api.plans.update(plan.uuid, plan);
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

    updatingId.value = item.id;

    if (val) {
      product.addons.push(addon.value.uuid);
    } else {
      product.addons = plan.addons.filter(
        (addonId) => addonId !== addon.value.uuid
      );
    }

    plan.products[item.id] = product;
    await api.plans.update(plan.uuid, plan);
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    setTimeout(() => (updatingId.value = ""), 440000);
  }
};
</script>

<script>
export default { name: "addon-products" };
</script>

<style scoped></style>
