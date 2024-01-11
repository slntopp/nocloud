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
  { text: "Key", value: "name" },
  { text: "Title", value: "usage_name" },
  { text: "Edition", value: "edition" },
  { text: "Required", value: "mandatory" },
  { text: "Price", value: "price" },
  { text: "Plan price", value: "planPrice" },
]);

onMounted(async () => {
  isLoading.value = true;
  try {
    await store.dispatch("servicesProviders/fetch", { anonymously: false });
    const spUuid = sps.value.find((sp) => sp.type === props.template.type).uuid;
    await store.dispatch("servicesProviders/fetchById", spUuid);
    sp.value = sps.value.find((sp) => sp.uuid === spUuid);
    Object.keys(sp.value.secrets.offeringItems).forEach((key) => {
      availableResources.value.push(sp.value.secrets.offeringItems[key]);
    });
    Object.keys(props.template.products).forEach((key) => {
      const resource = availableResources.value.find((r) => r.name === key);
      if (resource) {
        resource.planPrice = props.template.products[key].price;
      }
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch acronis prices",
    });
  } finally {
    isLoading.value = false;
  }
});

const save = async () => {
  const products = {};
  availableResources.value.forEach((res) => {
    if (res.planPrice) {
      products[res.name] = {
        title: res.usage_name,
        price: res.planPrice,
        kind: "PREPAID",
        period: 60 * 60 * 24 * 30,
        resources: {},
        meta: {
          edition: res.edition,
          application_id: res.application_id,
          infra_id: res.infra_id,
          mandatory: res.mandatory,
          measurement_unit: res.measurement_unit,
        },
      };
    }
  });
  isSaveLoading.value = true;
  try {
    await api.plans.update(props.template.uuid, {
      ...props.template,
      products,
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save prices",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const sps = computed(() => store.getters["servicesProviders/all"]);
</script>

<style scoped></style>
