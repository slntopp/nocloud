<template>
  <nocloud-table
    table-name="addons-products"
    :loading="isPlansLoading"
    :headers="planHeaders"
    :items="items"
    show-expand
    item-key="id"
    :show-select="false"
    :expanded.sync="expanded"

  >
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
            <v-switch v-model="item.enabled"></v-switch>
          </template>
        </nocloud-table>
      </td>
    </template>
  </nocloud-table>
</template>

<script setup>
import { useStore } from "@/store";
import { computed, onMounted, ref } from "vue";
import NocloudTable from "@/components/table.vue";

const store = useStore();

const planHeaders = ref([{ text: "Name", value: "name" }]);
const productHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Enabled", value: "enabled" },
]);
const expanded = ref([]);

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
    }));
    return { id: plan.uuid, name: plan.title, children };
  });
});
</script>

<script>
export default { name: "addon-products" };
</script>

<style scoped></style>
