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
    <v-tabs background-color="background-light" v-model="tab">
      <v-tab key="flavors"> Flavors </v-tab>
      <v-tab key="images"> Images </v-tab>
    </v-tabs>

    <v-tabs-items v-model="tab">
      <v-tab-item key="flavors">
        <nocloud-table
          item-key="uniqueId"
          table-name="cloudPrices"
          :show-select="false"
          :loading="isFlavoursLoading"
          :headers="pricesHeaders"
          :items="flavors"
          show-expand
          :expanded.sync="expanded"
        >
          <template v-slot:[`item.endPrice`]="{ item }">
            <v-text-field v-model.number="item.endPrice" type="number" />
          </template>
          <template v-slot:[`item.enabled`]="{ item }">
            <v-switch
              :input-value="item.enabled"
              @change="item.enabled = $event"
            />
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
              </v-card>
            </td>
          </template>
        </nocloud-table>
      </v-tab-item>
      <v-tab-item key="images">
        <nocloud-table
          item-key="id"
          :loading="isImagesLoading"
          :headers="imagesHeaders"
          :items="images"
        >
          <template v-slot:[`item.enabled`]="{ item }">
            <v-switch
              :input-value="item.enabled"
              @change="item.enabled = $event"
            />
          </template>
        </nocloud-table>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import {
  onMounted,
  ref,
  defineProps,
  toRefs,
  watch,
  computed,
  defineExpose,
} from "vue";
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

const { sp, template, fee } = toRefs(props);

const store = useStore();

const expanded = ref([]);
const tab = ref("prices");
const regions = ref([]);
const flavors = ref([]);
const prices = ref([]);
const images = ref([]);
const selectedRegion = ref("");
const isFlavoursLoading = ref(false);
const isImagesLoading = ref(false);

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
  { text: "OS type", value: "osType" },
  { text: "Visibility", value: "visibility" },
  { text: "Enabled", value: "enabled" },
]);

const projectId = computed(() => {
  return sp.value.vars?.projectId?.value?.default;
});

onMounted(async () => {
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

watch(selectedRegion, () => {
  fetchImages();
  fetchFlavours();
});

const fetchFlavours = async () => {
  isFlavoursLoading.value = true;
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

    flavors.value = newFlavours;
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Erorr during fetch flavors",
    });
  } finally {
    isFlavoursLoading.value = false;
  }
};

const fetchImages = async () => {
  try {
    isImagesLoading.value = true;
    const { meta } = await api.servicesProviders.action({
      uuid: sp.value.uuid,
      action: "get_cloud_images",
      params: {
        projectId: projectId.value,
        region: selectedRegion.value,
      },
    });
    images.value = meta.images.map((i) => {
      return {
        ...i,
        enabled: !!template.value.meta?.images?.[selectedRegion.value].find(
          (real) => real.id === i.id
        ),
      };
    });
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during fetch images",
    });
  } finally {
    isImagesLoading.value = false;
  }
};

const changePlan = (plan) => {
  plan.products = template.value.products;

  flavors.value.forEach((item) => {
    const itemName = `${item.period} ${item.name}`;

    if (item.enabled) {
      plan.products[itemName] = {
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
        },
      };
    }
  });

  plan.meta = template.value.meta || { images: {} };
  images.value.forEach((item) => {
    if (item.enabled) {
      if (!plan.meta.images?.[item.region]) {
        plan.meta.images[item.region] = [];
      }
      plan.meta.images[item.region].push(item);
    }
  });
};

const setFee = () => {
  flavors.value = flavors.value.map((i) => {
    if (!i.enabled) {
      i.endPrice = Math.round(i.price + (i.price / 100) * fee.value.default);
    }
    return i;
  });
};

defineExpose({ changePlan, setFee });
</script>

<style scoped></style>
