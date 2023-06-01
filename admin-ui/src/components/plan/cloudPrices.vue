<template>
  <div>
    <v-row>
      <v-col cols="3">
        <v-autocomplete
          :items="regions"
          label="Region"
          v-model="selectedRegion"
        />
      </v-col>
    </v-row>
    <nocloud-table
      item-key="uniqueId"
      table-name="cloudPrices"
      :show-select="false"
      :loading="isPlansLoading"
      :headers="pricesHeaders"
      :items="flavours"
      show-expand
      :expanded.sync="expanded"
    >
      <template v-slot:[`item.endPrice`]="{ item }">
        <v-text-field v-model.number="item.endPrice" type="number" />
      </template>
      <template v-slot:[`item.enabled`]="{ item }">
        <v-switch v-model.number="item.enabled" readonly />
      </template>
      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length" style="padding: 0">
          <v-card color="background-light">
            <v-card-title> Capabilities </v-card-title>
            <nocloud-table
              hide-default-footer
              :headers="capabilitiesHeaders"
              :items="item.capabilities"
              :show-select="false"
              no-hide-uuid
            />
            <v-card-title> Images </v-card-title>
            <nocloud-table
              item-key="id"
              :loading="!item?.images"
              :value="selectedImages[getKey(item)]"
              @input="selectedImages[getKey(item)] = $event"
              :headers="imagesHeaders"
              :items="item?.images"
            />
            <v-card-actions class="d-flex justify-end">
              <v-btn :loading="isSaveLoading" @click="saveProduct(item)"
                >Add</v-btn
              >
            </v-card-actions>
          </v-card>
        </td>
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import { onMounted, ref, defineProps, toRefs, watch, computed } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import NocloudTable from "@/components/table.vue";

const props = defineProps({
  fee: { type: Object, required: true },
  template: { type: Object, required: true },
  isPlansLoading: { type: Boolean, required: true },
  getPeriod: { type: Function, required: true },
  sp: { type: Object, required: true },
});

const { sp, template } = toRefs(props);

const store = useStore();

const expanded = ref([]);
const regions = ref([]);
const flavours = ref([]);
const selectedImages = ref({});
const prices = ref([]);
const selectedRegion = ref("");
const isPlansLoading = ref(false);
const isSaveLoading = ref(false);

const pricesHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Os type", value: "osType" },
  { text: "Disk", value: "disk" },
  { text: "In bound bandwidth", value: "inboundBandwidth" },
  { text: "Out bound bandwidth", value: "outboundBandwidth" },
  { text: "Period", value: "period" },
  { text: "Quota", value: "quota" },
  { text: "RAM", value: "ram" },
  { text: "VCPUS", value: "vcpus" },
  { text: "Type", value: "type" },
  { text: "Base price", value: "price" },
  { text: "End price", value: "endPrice" },
  { text: "Enabled", value: "enabled" },
]);

const capabilitiesHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Enabled", value: "enabled" },
]);

const imagesHeaders = ref([
  { text: "ID", value: "id" },
  { text: "Name", value: "name" },
  { text: "Size", value: "size" },
  { text: "Сreation date", value: "creationDate" },
  { text: "Status", value: "status" },
  { text: "Visibility", value: "visibility" },
]);

const projectId = computed(() => {
  return sp.value.vars?.projectId?.value?.default;
});

onMounted(async () => {
  Object.keys(template.value.products).forEach((key) => {
    selectedImages.value[key] = template.value.products[key].resources?.images;
  });
  await store.dispatch("servicesProviders/fetchById", sp.value.uuid);
  try {
    const { meta } = await api.servicesProviders.action({
      action: "regions",
      uuid: sp.value.uuid,
      params: {
        projectId: projectId.value,
        type: "ovh cloud",
      },
    });
    regions.value = meta.datacenters;
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Erorr during fetch regions",
    });
  }
});

const getKey = (item) => {
  return `${item.period} ${item.name}`;
};

watch(selectedRegion, async () => {
  isPlansLoading.value = true;
  try {
    const { meta } = await api.servicesProviders.action({
      action: "get_cloud_flavors",
      uuid: sp.value.uuid,
      params: {
        region: selectedRegion.value,
        projectId: projectId.value,
      },
    });
    prices.value = meta.codes;
    const newFlavours = [];
    meta.flavours.forEach((flavour) => {
      if (!flavour.available) {
        return;
      }

      Object.keys(flavour.planCodes || {}).forEach((key) => {
        newFlavours.push({
          ...flavour,
          period: key,
          price: prices.value[flavour.planCodes[key]],
          endPrice:
            template.value.products[`${key} ${flavour.name}`]?.price || 0,
          enabled: !!template.value.products[`${key} ${flavour.name}`],
          uniqueId: `${key}_${flavour.id}`,
        });
      });
    });

    flavours.value = newFlavours;
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Erorr during fetch plans",
    });
  } finally {
    isPlansLoading.value = false;
  }
});

watch(expanded, async (newValue) => {
  newValue.map(async (flavour) => {
    const flavourIndex = flavours.value.findIndex(
      (f) => flavour.uniqueId === f.uniqueId
    );
    if (!flavours.value[flavourIndex].images) {
      const data = await api.servicesProviders.action({
        uuid: sp.value.uuid,
        action: "get_cloud_images",
        params: {
          projectId: projectId.value,
          region: selectedRegion,
          osType: flavour.osType,
          flavorType: flavour.type,
        },
      });
      flavours.value[flavourIndex] = {
        ...flavours.value[flavourIndex],
        images: data.meta.images,
      };
      flavours.value = [...flavours.value];
    }
  });
});

const saveProduct = async (item) => {
  isSaveLoading.value = true;

  const products = { ...template.value.products };
  const itemName = `${item.period} ${item.name}`;
  products[itemName] = {
    title: itemName,
    kind: "PREPAID",
    price: item.endPrice,
    period: item.period === "hourly" ? 60 * 60 : 60 * 60 * 24 * 30,
    resources: {
      osType: item.osType,
      disk: item.disk,
      inboundBandwidth: item.inboundBandwidth,
      outboundBandwidth: item.outboundBandwidth,
      period: item.period,
      quota: item.quota,
      ram: item.ram,
      vcpus: item.vcpus,
      images: selectedImages.value[itemName],
    },
  };
  try {
    await api.plans.update(template.value.uuid, {
      ...template.value,
      products,
    });
    selectedImages.value[item.uniqueId] = [];
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Product added in plan successfully",
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during save plan",
    });
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<style scoped></style>
