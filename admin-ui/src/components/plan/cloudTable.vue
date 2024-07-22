<template>
  <div>
    <v-row align="center">
      <v-col cols="3">
        <v-autocomplete
          :items="regions"
          label="Region"
          v-model="selectedRegion"
          :loading="isRegionsLoading"
        />
      </v-col>
      <template v-if="selectedRegion">
        <v-btn class="mr-1" @click="setEnabledToTab(true)">Enable all</v-btn>
        <v-btn class="ml-1" @click="setEnabledToTab(false)">Disable all</v-btn>
      </template>
    </v-row>
    <v-tabs background-color="background-light" v-model="tab">
      <v-tab v-for="tabKey in tabItems" :key="tabKey">
        {{ tabKey[0].toUpperCase() + tabKey.slice(1) }}</v-tab
      >
    </v-tabs>

    <v-tabs-items v-model="tab">
      <v-tab-item key="flavors">
        <nocloud-table
          sort-by="enabled"
          sort-desc
          item-key="uniqueId"
          :show-select="false"
          :loading="isFlavoursLoading"
          :headers="pricesHeaders"
          :items="flavors[selectedRegion]"
          show-expand
          :expanded.sync="expanded"
        >
          <template v-slot:[`item.name`]="{ item }">
            <v-text-field style="min-width: 200px" v-model="item.name" />
          </template>
          <template v-slot:[`item.endPrice`]="{ item }">
            <v-text-field
              style="width: 200px"
              :suffix="defaultCurrency?.title"
              v-model.number="item.endPrice"
              type="number"
            />
          </template>
          <template v-slot:[`item.price`]="{ value }">
            {{ value }} {{ defaultCurrency?.title }}
          </template>
          <template v-slot:[`item.gpu.model`]="{ item }">
            <template v-if="item.gpu.model !== ''">
              {{ item.gpu.model }} (x{{ item.gpu.number }})
            </template>
            <template v-else>-</template>
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
                <v-card-title> Capabilities</v-card-title>
                <nocloud-table
                  hide-default-footer
                  :headers="capabilitiesHeaders"
                  :items="item.capabilities"
                  :show-select="false"
                  no-hide-uuid
                  table-name="cloud-capabilities"
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
          :show-select="false"
          table-name="cloud-images"
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
      <v-tab-item key="addons">
        <plan-addons-table
          @change:addons="planAddons = $event"
          :addons="template.addons"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import {
  computed,
  defineExpose,
  defineProps,
  onMounted,
  ref,
  toRefs,
  watch,
} from "vue";
import api from "@/api";
import { useStore } from "@/store";
import NocloudTable from "@/components/table.vue";
import { getMarginedValue } from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import planAddonsTable from "@/components/planAddonsTable.vue";

const props = defineProps({
  fee: { type: Object, required: true },
  template: { type: Object, required: true },
  isPlansLoading: { type: Boolean, required: true },
  getPeriod: { type: Function, required: true },
  sp: { type: Object, required: true },
});

const { sp, template, fee } = toRefs(props);

const store = useStore();
const { convertFrom, defaultCurrency } = useCurrency();

const expanded = ref([]);
const tab = ref("prices");
const tabItems = ref(["flavors", "images", "addons"]);
const regions = ref([]);
const flavors = ref({});
const prices = ref({});
const images = ref({});
const selectedRegion = ref("");
const isFlavoursLoading = ref(false);
const isImagesLoading = ref(false);
const isRegionsLoading = ref(false);
const planAddons = ref([]);

const pricesHeaders = ref([
  { text: "Title", value: "name" },
  { text: "API Name", value: "apiName" },
  { text: "Os type", value: "osType" },
  { text: "Disk", value: "disk" },
  { text: "In bound bandwidth", value: "inboundBandwidth" },
  { text: "Out bound bandwidth", value: "outboundBandwidth" },
  { text: "Period", value: "period" },
  { text: "Quota", value: "quota" },
  { text: "RAM", value: "ram" },
  { text: "VCPUS", value: "vcpus" },
  { text: "GPU", value: "gpu.model" },
  { text: "Type", value: "type" },
  { text: "Incoming price", value: "price" },
  { text: "Sale price", value: "endPrice" },
  { text: "Enabled", value: "enabled" },
]);

const capabilitiesHeaders = ref([
  { text: "Title", value: "name" },
  { text: "Enabled", value: "enabled" },
]);

const imagesHeaders = ref([
  { text: "ID", value: "id" },
  { text: "Title", value: "name" },
  { text: "Type", value: "type" },
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
      Object.entries(flavour.planCodes || {}).forEach(([key, value]) => {
        const { gpu = { model: "", number: 0 } } = meta.technical[value] ?? {};
        const period = key === "monthly" ? "P1M" : "P1H";
        const planCode = `${period} ${flavour.id}`;
        const price = convertFrom(
          prices.value[selectedRegion.value]?.[flavour.planCodes[key]],
          "PLN"
        );

        newFlavours.push({
          ...flavour,
          gpu,
          name: template.value.products[planCode]?.title || flavour.name,
          apiName: flavour.name,
          period,
          key: planCode,
          priceCode: flavour.planCodes[key],
          price,
          endPrice: template.value.products[planCode]?.price || price,
          enabled: !!template.value.products[planCode],
          uniqueId: `${period} ${flavour.id}`,
          meta: { region: selectedRegion.value },
        });
      });
    });

    flavors.value[selectedRegion.value] = newFlavours;
  } catch (error) {
    console.log(error);
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
    const regionImages = {};
    images.value[regionKey]
      .filter((item) => item.enabled)
      .forEach((item) => {
        regionImages[item.type] = regionImages[item.type]
          ? [
              ...regionImages[item.type],
              {
                name: item.name,
                id: item.id,
              },
            ]
          : [{ name: item.name, id: item.id }];
      });

    if (!regionFlavors.length) {
      Object.keys(plan.products).forEach((key) => {
        if (plan.products[key]?.meta?.region === regionKey) {
          plan.products[key] = undefined;
        }
      });
    }
    regionFlavors.forEach((item) => {
      plan.products[item.key] = {
        title: item.name,
        kind: "PREPAID",
        price: item.endPrice,
        period: item.period === "P1H" ? 60 * 60 : 60 * 60 * 24 * 30,
        resources: {
          osType: item.osType,
          drive_size: item.disk && +item.disk ? +item.disk * 1024 : 0,
          drive_type: "SSD",
          inboundBandwidth: item.inboundBandwidth,
          outboundBandwidth: item.outboundBandwidth,
          period: item.period,
          quota: item.quota,
          ram: item.ram,
          cpu: item.vcpus,
          gpu_name: item.gpu.model,
          gpu_count: item.gpu.number,
        },
        public: true,
        meta: {
          priceCode: item.priceCode,
          ...item.meta,
          datacenter: [regionKey],
          os: regionImages[item.osType],
        },
      };
    });
  });

  plan.addons = planAddons.value;
};
const setEnabledToValues = (value, status) => {
  value = value[selectedRegion.value].map((i) => {
    i.enabled = status;
    return i;
  });
};
const setEnabledToTab = (status = false) => {
  switch (tabItems.value[tab.value]) {
    case "images": {
      setEnabledToValues(images.value, status);
      break;
    }
    case "flavors": {
      setEnabledToValues(flavors.value, status);
      break;
    }
  }
};

const setFee = () => {
  Object.keys(flavors.value).forEach((key) => {
    flavors.value[key] = flavors.value[key].map((i) => {
      i.endPrice = getMarginedValue(fee.value, i.price);
      return i;
    });
  });
};

defineExpose({ changePlan, setFee });
</script>

<style scoped></style>
