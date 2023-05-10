<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <nocloud-table
      :loading="isLoading"
      :headers="headers"
      table-text="cpanel-prices"
      class="pa-4"
      item-key="text"
      :show-select="false"
      :items="availableResources"
    >
      <template v-slot:[`item.planPrice`]="{ item }">
        <v-text-field
          v-model.number="item.planPrice"
          type="number"
        ></v-text-field>
      </template>
    </nocloud-table>
    <v-card-actions class="d-flex justify-end">
      <v-btn @click="save" :loading="isSaveLoading">save</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { ref, defineProps, onMounted, computed } from "vue";
import { useStore } from "@/store";
import api from "@/api";

const props = defineProps(["template"]);

const store = useStore();

const availableResources = ref([]);
const sp = ref(null);
const isLoading = ref(false);
const isSaveLoading = ref(false);

const headers = ref([
  { text: "name", value: "key" },
  { text: "sp price", value: "spPrice" },
  { text: "plan price", value: "planPrice" },
]);

onMounted(async () => {
  isLoading.value = true;
  try {
    await store.dispatch("servicesProviders/fetch", false);
    const spUuid = sps.value.find((sp) => sp.type === props.template.type).uuid;
    await store.dispatch("servicesProviders/fetchById", spUuid);
    sp.value = sps.value.find((sp) => sp.uuid === spUuid);
    Object.keys(sp.value.secrets.offeringItems).forEach((key) => {
      availableResources.value.push({
        key,
        spPrice: sp.value.secrets.offeringItems[key],
      });
    });
    props.template.resources.forEach((r) => {
      const indexRealItem = availableResources.value.findIndex(
        (res) => res.key === r.key
      );
      availableResources.value[indexRealItem].planPrice = r.price;
    });
  } finally {
    isLoading.value = false;
  }
});

const save = async () => {
  const resources = [];
  availableResources.value.forEach((res) => {
    if (res.planPrice) {
      resources.push({
        key: res.key,
        price: res.planPrice,
        kind: "PREPAID",
        period: 1000 * 60 * 60 * 24 * 30,
        except: false,
      });
    }
  });
  isSaveLoading.value = true;
  try {
    await api.plans.update(props.template.uuid, {
      ...props.template,
      resources,
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const sps = computed(() => store.getters["servicesProviders/all"]);
</script>

<style scoped></style>
