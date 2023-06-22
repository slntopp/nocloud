<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="price model"
          append-icon="mdi-pencil"
          @click:append="priceModelDialog = true"
          :value="template.billingPlan.title"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Tarif (product plan)"
          :value="
            template.billingPlan.products[duration + ' ' + planCode]?.title
          "
          append-icon="mdi-pencil"
          @click:append="priceModelDialog = true"
        />
      </v-col>
      <v-col>
        <v-text-field readonly label="Price instance total" :value="getPrice" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Date (create)"
          :value="template.data.creation"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="template.data.expiration"
        />
      </v-col>
      ></v-row
    >
    <nocloud-table
      hide-default-footer
      sort-by="index"
      item-key="key"
      :show-select="false"
      :headers="pricesHeaders"
      :items="pricesItems"
    >
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field
          v-model.number="item.price"
          @change="onUpdatePrice(item)"
          type="number"
          append-icon="mdi-pencil"
        ></v-text-field>
      </template>
      <template v-slot:[`item.basePrice`]="{ item }">
        <v-text-field
          :loading="isBasePricesLoading"
          readonly
          :value="basePrices[item.key]"
        ></v-text-field>
      </template>
      <template v-slot:body.append>
        <tr>
          <td>Total instance price</td>
          <td>{{ isBasePricesLoading ? "Loading..." : totalBasePrice }}</td>
          <td>
            <div class="d-flex justify-space-between align-center">
              {{ totalNewPrice?.toFixed(2) }}
            </div>
          </td>
        </tr>
      </template>
    </nocloud-table>
    <edit-price-model
      @refresh="emit('refresh')"
      :template="template"
      :plans="plans"
      :service="service"
      v-model="priceModelDialog"
    />
  </div>
</template>

<script setup>
import {
  defineProps,
  ref,
  defineEmits,
  toRefs,
  computed,
  onMounted,
} from "vue";
import NocloudTable from "@/components/table.vue";
import api from "@/api";
import { useStore } from "@/store";
import EditPriceModel from "@/components/modules/ovh/editPriceModel.vue";
import useRate from "@/hooks/useRate";

const props = defineProps(["template", "plans"]);
const emit = defineEmits(["refresh", "update"]);

const store = useStore();
const rate = useRate();

const { template, plans } = toRefs(props);
const pricesItems = ref([]);
const basePrices = ref({});
const pricesHeaders = ref([
  { text: "Name", value: "title" },
  { text: "Base price", value: "basePrice" },
  { text: "Price", value: "price" },
]);
const totalNewPrice = ref(0);
const totalBasePrice = ref(0);
const isBasePricesLoading = ref(false);
const priceModelDialog = ref(false);

const setTotalNewPrice = () => {
  totalNewPrice.value = pricesItems.value.reduce((acc, i) => i.price + acc, 0);
};

const onUpdatePrice = (item) => {
  emit("update", { key: item.path, value: item.price });
  setTotalNewPrice();
};

const getVpsPrices = async () => {
  const { meta } = await api.servicesProviders.action({
    uuid: template.value.sp,
    action: "get_plans",
  });

  const prices = {};

  const planCodeCurr = meta.plans.find((p) => planCode.value === p.planCode);
  prices["tarrif"] = getPriceFromProduct(planCodeCurr);
  addons.value.forEach((addon) => {
    Object.keys(meta).forEach((metaKey) => {
      const product =
        meta[metaKey].find && meta[metaKey].find((p) => p?.planCode === addon);
      if (product) {
        prices[addon] = getPriceFromProduct(product);
      }
    });
  });

  return prices;
};

const getDedicatedPrice = async () => {
  const { meta } = await api.servicesProviders.action({
    uuid: template.value.sp,
    action: "get_baremetal_plans",
  });
  const prices = {};

  const planCodeCurr = meta.plans.find((p) => planCode.value === p.planCode);
  prices["tarrif"] = getPriceFromProduct(planCodeCurr);
  const addonsPrice = await api.servicesProviders.action({
    action: "get_baremetal_options",
    uuid: template.value.sp,
    params: { planCode: planCode.value },
  });
  const addonTypes = { softraid: "storage", ram: "memory" };
  addons.value.forEach((addon) => {
    const addonType = Object.keys(addonTypes).find((t) => addon.includes(t));
    const addonName = tarrif.value.meta.addons.find(
      (a) => a.id === addon
    ).title;
    pricesItems.value = pricesItems.value.map((p) => {
      if (p.title === addon) {
        p.title = addonName;
      }
      return p;
    });

    prices[addon] =
      addonsPrice.meta.options[addonTypes[addonType]]
        .find((m) => m.planCode === addon)
        ?.prices.find(
          (p) =>
            p.duration === duration.value &&
            p.pricingModel === template.value.config.pricingModel
        ).price.value || 0;
  });

  return prices;
};
const getCloudPrices = async () => {
  const fullSp = await api.servicesProviders.get(template.value.sp);
  const prices = {};
  const { meta } = await api.servicesProviders.action({
    action: "get_cloud_flavors",
    uuid: template.value.sp,
    params: {
      region: template.value.config.configuration.cloud_datacenter,
      projectId: fullSp.vars?.projectId?.value?.default,
    },
  });

  prices["tarrif"] = meta.codes[tarrif.value?.title] * rate.value;

  return prices;
};
const getPriceFromProduct = (product) => {
  return (
    product.prices.find(
      (p) =>
        duration.value === p.duration &&
        template.value.config.pricingMode === p.pricingMode
    )?.price?.value * rate.value
  ).toFixed(2);
};

const initPrices = () => {
  pricesItems.value.push({
    title: "tarrif",
    key: "tarrif",
    ind: 0,
    path: `billingPlan.products.${[duration.value, planCode.value].join(
      " "
    )}.price`,
    price: tarrif.value?.price,
  });

  addons.value.forEach((key, ind) => {
    const addonIndex = template.value.billingPlan.resources.findIndex(
      (p) => p.key === [duration.value, key].join(" ")
    );

    pricesItems.value.push({
      price: template.value.billingPlan.resources[addonIndex]?.price || 0,
      path: `billingPlan.resources.${addonIndex}.price`,
      title: key,
      key: key,
      index: ind + 1,
    });
  });
  setTotalNewPrice();
};

const planCode = computed(() => template.value.config.planCode);
const duration = computed(() => template.value.config.duration);
const addons = computed(() => template.value.config.addons || []);
const type = computed(() => template.value.config.type);
const tarrif = computed(
  () =>
    template.value.billingPlan.products[
      [duration.value, planCode.value].join(" ")
    ]
);
const getPrice = computed(() => {
  const prices = [];
  prices.push(tarrif.value?.price);
  addons.value.forEach((name) => {
    prices.push(
      template.value.billingPlan.resources.find(
        (p) => p.key === [duration.value, name].join(" ")
      )?.price || 0
    );
  });
  return prices.reduce((acc, val) => acc + val, 0);
});

const service = computed(() =>
  store.getters["services/all"].find((s) => s.uuid === template.value.service)
);

const getBasePrices = async () => {
  isBasePricesLoading.value = true;
  try {
    let meta = null;
    if (type.value === "vps") {
      meta = await getVpsPrices();
    } else if (type.value === "dedicated") {
      meta = await getDedicatedPrice();
    } else if (type.value === "cloud") {
      meta = await getCloudPrices();
    }

    basePrices.value = meta;
    totalBasePrice.value = Object.keys(basePrices.value)
      .reduce((acc, key) => acc + +basePrices.value[key], 0)
      .toFixed(2);
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch base prices",
    });
  } finally {
    isBasePricesLoading.value = false;
  }
};

onMounted(() => {
  initPrices();
  getBasePrices();
});
</script>

<style scoped></style>
