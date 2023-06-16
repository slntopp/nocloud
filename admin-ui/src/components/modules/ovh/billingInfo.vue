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

const getOvhPrices = async () => {
  const { meta } = await api.servicesProviders.action({
    uuid: template.value.sp,
    action: "get_plans",
  });

  return meta;
};

const getDedicatedPrice = async () => {
  const { meta } = await api.servicesProviders.action({
    uuid: template.value.sp,
    action: "get_baremetal_plans",
  });

  return meta;
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
    if (addonIndex === -1) {
      return;
    }

    pricesItems.value.push({
      price: template.value.billingPlan.resources[addonIndex].price,
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
      meta = await getOvhPrices();
    } else if (type.value === "dedicated") {
      meta = await getDedicatedPrice();
    }

    const planCodeCurr = meta.plans.find((p) => planCode.value === p.planCode);
    basePrices.value["tarrif"] = getPriceFromProduct(planCodeCurr);
    addons.value.forEach((addon) => {
      Object.keys(meta).forEach((metaKey) => {
        const product =
          meta[metaKey].find &&
          meta[metaKey].find((p) => p?.planCode === addon);
        if (product) {
          basePrices.value[addon] = getPriceFromProduct(product);
        }
      });
    });
    totalBasePrice.value = Object.keys(basePrices.value)
      .reduce((acc, key) => acc + +basePrices.value[key], 0)
      .toFixed(2);
  } catch (e) {
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
