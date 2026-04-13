<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5 justify-space-between">
      <div class="d-flex align-center">
        <v-btn
          v-for="layout in Object.keys(defaultLayouts)"
          class="mr-1"
          :key="layout"
          :disabled="defaultLayouts[layout].id === currentSearchLayout"
          @click="setInvoicesLayout(layout)"
        >
          <v-icon small class="mr-1">mdi-filter</v-icon>
          {{ layout }}</v-btn
        >
      </div>

      <invoices-actions
        :selected-invoices="selectedInvoices"
        @input="selectedInvoices = $event"
        @refresh="onRefresh"
      />
    </div>
    <invoices-table
      v-model="selectedInvoices"
      :loading="isLoading"
      :refetch="refetch"
      :items="invoices"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import InvoicesTable from "@/components/invoicesTable.vue";
import InvoicesActions from "@/components/invoicesActions.vue";

const selectedInvoices = ref([]);
const refetch = ref(false);

const store = useStore();

const isLoading = computed(() => store.getters["invoices/isLoading"]);
const invoices = computed(() => store.getters["invoices/all"]);

const defaultLayouts = computed(() => ({
  unpaid: {
    title: "Unpaid",
    filter: { status: [2] },
    fields: ["status"],
    id: "id-unpaid",
  },
  paid: {
    title: "Paid",
    filter: { status: [1] },
    fields: ["status"],
    id: "id-paid",
  },
  overdue: {
    id: "id-overdue",
    title: "Overdue",
    fields: ["deadline", "status"],
    filter: {
      status: [2],
      deadline: [
        new Date(Date.now() - 86000 * 365 * 1000).toISOString().split("T")[0],
        new Date(Date.now()).toISOString().split("T")[0],
      ],
    },
  },
}));

const currentSearchLayout = computed(
  () => store.getters["appSearch/currentLayout"],
);

const setDefaultLayouts = () => {
  const defaults = Object.values(defaultLayouts.value);
  const layouts = JSON.parse(
    JSON.stringify(store.getters["appSearch/layouts"]),
  );

  defaults.forEach((layout) => {
    const index = layouts.findIndex((l) => l.title === layout.title);
    if (index == -1) {
      layouts.push(layout);
    } else {
      layouts[index] = layout;
    }
  });

  store.commit("appSearch/setLayouts", layouts);
};

const setInvoicesLayout = (key) => {
  store.commit("appSearch/setCurrentLayout", defaultLayouts.value[key].id);
};

const onRefresh = () => {
  refetch.value = !refetch.value;
};

onMounted(() => {
  setTimeout(() => setDefaultLayouts(), 500);
});
</script>

<script>
export default { name: "invoices-view" };
</script>
<style scoped></style>
