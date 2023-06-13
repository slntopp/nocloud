<template>
  <div>
    <v-row>
      <v-col cols="3">
        <v-autocomplete
          :items="regions"
          label="Region"
          v-model="selectedRegion"
          :loading="isRegionsLoading"
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
          sort-by="enabled"
          sort-desc
          item-key="uniqueId"
          table-name="cloudFlavors"
          :show-select="false"
          :loading="isFlavoursLoading"
          :headers="pricesHeaders"
          :items="flavors[selectedRegion]"
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
          :items="images[selectedRegion]"
          sort-by="enabled"
          table-name="cloudImages"
          sort-desc
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
const flavors = ref({});
const prices = ref({});
const images = ref({});
const selectedRegion = ref("");
const isFlavoursLoading = ref(false);
const isImagesLoading = ref(false);
const isRegionsLoading = ref(false);

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
  { text: "Enabled", value: "enabled" },
]);

const projectId = computed(() => {
  return sp.value.vars?.projectId?.value?.default;
});

onMounted(async () => {
  await store.dispatch("servicesProviders/fetchById", sp.value.uuid);
  isRegionsLoading.value = true;
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
  } finally {
    isRegionsLoading.value = false;
  }
});

watch(selectedRegion, () => {
  fetchImages();
  fetchFlavours();
});

const fetchFlavours = async () => {
  if (
    flavors.value[selectedRegion.value] &&
    prices.value[selectedRegion.value]
  ) {
    return;
  }

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
    prices.value[selectedRegion.value] = meta.codes;

    const newFlavours = [];
    meta.flavours.forEach((flavour) => {
      if (!flavour.available) {
        return;
      }
      Object.keys(flavour.planCodes || {}).forEach((key) => {
        const period = key === "monthly" ? "P1M" : "P1H";
        const planCode = `${period} ${flavour.name}-${selectedRegion.value}`;
        newFlavours.push({
          ...flavour,
          period,
          name: planCode,
          price: prices.value[selectedRegion.value]?.[flavour.planCodes[key]],
          endPrice: template.value.products[planCode]?.price || 0,
          enabled: !!template.value.products[planCode],
          uniqueId: `${period} ${flavour.id}`,
          meta: { region: selectedRegion.value },
        });
      });
    });

    flavors.value[selectedRegion.value] = newFlavours;
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Erorr during fetch flavors",
    });
  } finally {
    isFlavoursLoading.value = false;
  }
};

const fetchImages = async () => {
  if (images.value[selectedRegion.value]) {
    return;
  }

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
    images.value[selectedRegion.value] = meta.images.map((i) => {
      return {
        ...i,
        enabled: !!Object.keys(template.value.products).find(
          (k) =>
            template.value.products[k].meta?.region === selectedRegion.value &&
            template.value.products[k].meta?.os?.find((o) => i.id === o.id)
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

  Object.keys(flavors.value).forEach((regionKey) => {
    const regionFlavors = flavors.value[regionKey].filter((f) => f.enabled);

    const regionImages = images.value[regionKey]
      .filter((item) => item.enabled)
      .map((item) => ({ name: item.name, id: item.id }));

    regionFlavors.forEach((item) => {
      plan.products[item.name] = {
        title: item.name,
        kind: "PREPAID",
        price: item.endPrice,
        period: item.period === "P1H" ? 60 * 60 : 60 * 60 * 24 * 30,
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
        meta: {
          ...item.meta,
          datacenter: [regionKey],
          os: regionImages,
        },
      };
    });
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
