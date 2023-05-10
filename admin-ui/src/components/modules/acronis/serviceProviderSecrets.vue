<template>
  <v-container>
    <v-row>
      <v-col>
        <v-card-title>Client id</v-card-title>
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.secrets.clientId" />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-card-title>Client secret</v-card-title>
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.secrets.clientSecret" />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-card-title>Host</v-card-title>
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.secrets.datacenterUrl" />
      </v-col>
    </v-row>
    <v-card-title>Offering items</v-card-title>
    <v-row>
      <nocloud-table
        style="width: 100%"
        table-name="acronisPrices"
        :no-hide-uuid="false"
        :show-select="false"
        :loading="isLoading"
        :headers="headers"
        :items="offeringItems"
      >
        <template v-slot:[`item.price`]="{ item }">
          <v-text-field
            type="number"
            v-model.number="item.price"
          ></v-text-field>
        </template>
      </nocloud-table>
    </v-row>
    <v-row justify="end">
      <v-btn :loading="isSaveLoading" @click="saveOffering">Save</v-btn>
    </v-row>
  </v-container>
</template>

<script setup>
import { onMounted, ref,defineProps } from "vue";
import api from "@/api";
import NocloudTable from "@/components/table.vue";

const props=defineProps(['template'])

const offeringItems = ref([]);
const isLoading = ref(false);
const isSaveLoading = ref(false);

const headers = ref([
  { text: "Name", value: "name" },
  { text: "Usage name", value: "usage_name" },
  { text: "Type", value: "type" },
  { text: "Application id", value: "application_id" },
  { text: "Edition", value: "edition" },
  { text: "Price", value: "price" },
]);

onMounted(async () => {
  isLoading.value = true;
  try {
    offeringItems.value = (
      await api.servicesProviders.action({
        action: "get_offering_items",
        uuid: props.template.uuid,
      })
    ).meta.offeringItems.map((of) => ({
      ...of,
      price: props.template.secrets.offeringItems[of.name],
    }));
  } finally {
    isLoading.value = false;
  }
});

const saveOffering = async () => {
  const offerings = offeringItems.value.filter((of) => !!of.price);
  if (offerings.length === 0) {
    return;
  }
  const sp = JSON.parse(JSON.stringify(props.template));
  sp.secrets.offeringItems = {};
  offerings.forEach((of) => {
    sp.secrets.offeringItems[of.name] = of.price;
  });
  isSaveLoading.value = true;
  try {
    await api.servicesProviders.update(props.template.uuid, sp);
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<style scoped></style>
