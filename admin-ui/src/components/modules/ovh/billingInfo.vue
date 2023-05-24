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
          :value="template.config.planCode"
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
          v-model.number="prices[item.key]"
          @change="setTotalNewPrice"
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
          <td>{{ totalBasePrice || "Loading..." }}</td>
          <td>
            <div class="d-flex justify-space-between align-center">
              <span>{{ totalNewPrice.toFixed(2) }} </span>
              <v-btn
                class="mx-5"
                :disabled="isBasePricesLoading"
                :loading="isPlanChangeLoading"
                @click="saveNewPrices"
                >Save prices</v-btn
              >
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
import { getTodayFullDate } from "@/functions";

const props = defineProps(["template", "plans"]);
const emit = defineEmits(["refresh"]);

const store = useStore();

const { template, plans } = toRefs(props);
const pricesItems = ref([]);
const prices = ref({});
const basePrices = ref({});
const rate = ref(0);
const pricesHeaders = ref([
  { text: "Name", value: "title" },
  { text: "Base price", value: "basePrice" },
  { text: "Price", value: "price" },
]);
const isPlanChangeLoading = ref(false);
const totalNewPrice = ref(0);
const totalBasePrice = ref(0);
const isBasePricesLoading = ref(false);
const priceModelDialog = ref(false);

const saveNewPrices = async () => {
  const instance = JSON.parse(JSON.stringify(template.value));
  const planCodeLocal = "IND_" + instance.title + "_" + getTodayFullDate();
  const plan = {
    title: planCodeLocal,
    public: false,
    kind: instance.billingPlan.kind,
    type: instance.billingPlan.type,
    resources: [],
  };
  const product = { ...tarrif.value, price: prices.value.tarrif };
  plan.products = {
    [duration.value + " " + template.value.config.planCode]: product,
  };
  addons.value.forEach((key) => {
    plan.resources.push({
      ...template.value.billingPlan.resources.find(
        (p) => p.key === [duration.value, key].join(" ")
      ),
      price: prices.value[key],
    });
  });

  isPlanChangeLoading.value = true;
  try {
    const data = await api.plans.create(plan);
    await api.servicesProviders.bindPlan(template.value.sp, [data.uuid]);
    const tempService = JSON.parse(JSON.stringify(service.value));
    const igIndex = tempService.instancesGroups.findIndex((ig) =>
      ig.instances.find((i) => i.uuid === template.value.uuid)
    );
    const instanceIndex = tempService.instancesGroups[
      igIndex
    ].instances.findIndex((i) => i.uuid === template.value.uuid);

    tempService.instancesGroups[igIndex].instances[instanceIndex].billingPlan =
      data;
    await api.services._update(tempService);
    isPlanChangeLoading.value = false;
    emit("refresh");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save prices",
    });
  } finally {
    isPlanChangeLoading.value = false;
  }
};

const setTotalNewPrice = () => {
  totalNewPrice.value = Object.keys(prices.value).reduce(
    (acc, key) => acc + +prices.value[key],
    0
  );
};

const getBasePrices = () => {
  isBasePricesLoading.value = true;
  api
    .get(`/billing/currencies/rates/PLN/${defaultCurrency.value}`)
    .then((res) => {
      rate.value = res.rate;
    })
    .catch(() =>
      api.get(`/billing/currencies/rates/${defaultCurrency.value}/PLN`)
    )
    .then((res) => {
      if (res) rate.value = 1 / res.rate;
    })
    .catch((err) => console.error(err));
  api
    .post(`/sp/${template.value.sp}/invoke`, { method: "get_plans" })
    .then(({ meta }) => {
      const planCodeCurr = meta.plans.find(
        (p) => planCode.value === p.planCode
      );
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
      isBasePricesLoading.value = false;
    })
    .catch((e) =>
      store.commit("snackbar/showSnackbarError", {
        message: e.response?.data?.message || "Error during fetch base prices",
      })
    );
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
  });
  prices.value["tarrif"] = tarrif.value.price;

  addons.value.forEach((key, ind) => {
    prices.value[key] = template.value.billingPlan.resources.find(
      (p) => p.key === [duration.value, key].join(" ")
    ).price;
    pricesItems.value.push({
      title: key,
      key: key,
      index: ind + 1,
    });
  });
  setTotalNewPrice();
};

const planCode = computed(() => template.value.config.planCode);
const duration = computed(() => template.value.config.duration);
const addons = computed(() => template.value.config.addons);
const tarrif = computed(
  () =>
    template.value.billingPlan.products[
      [duration.value, planCode.value].join(" ")
    ]
);
const getPrice = computed(() => {
  const prices = [];
  prices.push(tarrif.value.price);
  addons.value.forEach((name) => {
    prices.push(
      template.value.billingPlan.resources.find(
        (p) => p.key === [duration.value, name].join(" ")
      ).price
    );
  });
  return prices.reduce((acc, val) => acc + val, 0);
});
const defaultCurrency = computed(() => {
  return store.getters["currencies/default"];
});

const service = computed(() =>
  store.getters["services/all"].find((s) => s.uuid === template.value.service)
);

onMounted(() => {
  initPrices();
  getBasePrices();
});
</script>

<style scoped></style>
