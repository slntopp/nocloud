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
          :value="tarrif.title"
          append-icon="mdi-pencil"
          @click:append="priceModelDialog = true"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Price instance total"
          :value="type === 'dedicated' ? +totalNewPrice?.toFixed(2) : getPrice"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Account price instance total"
          :value="accountTotalNewPrice"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Date (create)"
          :value="template.data.creation"
        />
      </v-col>
      <v-col>
        <v-text-field readonly label="Due to date/next payment" :value="date" />
      </v-col>
    </v-row>
    <nocloud-table
      table-name="ovh-billing"
      hide-default-footer
      sort-by="index"
      item-key="key"
      :show-select="false"
      :headers="pricesHeaders"
      :items="pricesItems"
    >
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field
          v-model="item.price"
          @change="onUpdatePrice(item, false)"
          type="number"
          append-icon="mdi-pencil"
        ></v-text-field>
      </template>
      <template v-slot:[`item.accountPrice`]="{ item }">
        <v-text-field
          v-model="item.accountPrice"
          @change="onUpdatePrice(item, true)"
          type="number"
          append-icon="mdi-pencil"
        ></v-text-field>
      </template>
      <template v-slot:[`item.basePrice`]="{ item }">
        <v-text-field
          :loading="isBasePricesLoading"
          readonly
          :value="convertedBasePrices[item.key]"
        ></v-text-field>
      </template>
      <template v-slot:body.append>
        <tr>
          <td>Total instance price</td>
          <td></td>
          <td>{{ isBasePricesLoading ? "Loading..." : totalBasePrice }}</td>
          <td>
            <div class="d-flex justify-space-between align-center">
              {{ totalNewPrice?.toFixed(2) }}
            </div>
          </td>
          <td></td>
          <td>{{ accountTotalNewPrice }}</td>
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
import {
  formatSecondsToDate,
  getBillingPeriod,
} from "@/functions";

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
  { text: "Account price", value: "accountPrice" },
  { text: "Billing period", value: "period" },
]);
const totalNewPrice = ref(0);
const isBasePricesLoading = ref(false);
const priceModelDialog = ref(false);
const accountRate = ref(0);

const namespace = computed(() =>
  store.getters["namespaces/all"]?.find(
    (n) => n.uuid == template.value.access.namespace
  )
);
const accountTotalNewPrice = computed(() =>
  toAccountPrice(totalNewPrice.value)
);
const account = computed(() => {
  if (!namespace.value) {
    return;
  }
  return store.getters["accounts/all"]?.find(
    (a) => a?.uuid == namespace.value.access.namespace
  );
});
const setTotalNewPrice = () => {
  totalNewPrice.value = +pricesItems.value
    .reduce((acc, i) => +i.price + acc, 0)
    .toFixed(2);
};

const onUpdatePrice = (item, isAccount) => {
  if (isAccount) {
    emit("update", {
      key: item.path,
      value: fromAccountPrice(item.accountPrice),
    });
    pricesItems.value = pricesItems.value.map((p) => {
      if (p.path === item.path) {
        p.price = fromAccountPrice(item.accountPrice);
      }
      return p;
    });
  } else {
    emit("update", { key: item.path, value: item.price });
    pricesItems.value = pricesItems.value.map((p) => {
      if (p.path === item.path) {
        p.accountPrice = toAccountPrice(item.price);
      }
      return p;
    });
  }
  setTotalNewPrice();
};

const convertedBasePrices = computed(() => {
  if (!rate.value) {
    return basePrices.value;
  }
  const converted = {};
  Object.keys(basePrices.value).forEach((key) => {
    converted[key] = basePrices.value[key] * rate.value;
  });

  return converted;
});

const totalBasePrice = computed(() => {
  return Object.keys(convertedBasePrices.value)
    .reduce((acc, key) => acc + +convertedBasePrices.value[key], 0)
    .toFixed(2);
});

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
  prices["tarrif"] = meta.codes[tarrif.value?.meta?.priceCode];
  return prices;
};
const getPriceFromProduct = (product) => {
  return product.prices
    .find(
      (p) =>
        duration.value === p.duration &&
        template.value.config.pricingMode === p.pricingMode
    )
    ?.price?.value.toFixed(2);
};

const getAddonKey = (key) => {
  let keys = [];
  if (template.value.config.type === "dedicated") {
    keys = [duration.value, planCode.value, key];
  } else {
    keys = [duration.value, key];
  }
  return keys.join(" ");
};

const date = computed(() => {
  if (type.value === "cloud") {
    return formatSecondsToDate(
      +template.value?.data?.last_monitoring + +tarrif.value.period
    );
  }

  return template.value.data.expiration;
});

const initPrices = () => {
  pricesItems.value.push({
    title: "tarrif",
    key: "tarrif",
    ind: 0,
    path: `billingPlan.products.${[duration.value, planCode.value].join(
      " "
    )}.price`,
    price: tarrif.value?.price,
    period: tarrif.value?.period,
  });

  addons.value.forEach((key, ind) => {
    const addonIndex = template.value.billingPlan.resources.findIndex(
      (p) => p.key === getAddonKey(key)
    );

    pricesItems.value.push({
      price: template.value.billingPlan.resources[addonIndex]?.price || 0,
      path: `billingPlan.resources.${addonIndex}.price`,
      title: key,
      key: key,
      index: ind + 1,
      period: template.value.billingPlan.resources[addonIndex]?.period,
    });
  });

  pricesItems.value = pricesItems.value.map((i) => {
    i.period = getBillingPeriod(i.period);
    i.accountPrice = i.price * accountRate.value;

    return i;
  });
  setTotalNewPrice();
};

const planCode = computed(() => template.value.config.planCode);
const duration = computed(() => template.value.config.duration);
const addons = computed(() => template.value.config.addons || []);
const type = computed(() => template.value.config.type);
const tarrif = computed(() => {
  let key = "";
  if (!duration.value) {
    key = template.value.product;
  } else {
    key = [duration.value, planCode.value].join(" ");
  }
  return template.value.billingPlan.products[key];
});
const getPrice = computed(() => {
  const prices = [];
  prices.push(tarrif.value?.price);
  addons.value.forEach((name) => {
    prices.push(
      template.value.billingPlan.resources.find(
        (p) =>
          p.key === [duration.value, planCode, name].join(" ") ||
          p.key === [duration.value, name].join(" ")
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
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch base prices",
    });
  } finally {
    isBasePricesLoading.value = false;
  }
};

const defaultCurrency = computed(() => {
  return store.getters["currencies/default"];
});

const toAccountPrice = (price) => {
  return (price / accountRate.value).toFixed(2);
};
const fromAccountPrice = (price) => {
  return (price * accountRate.value).toFixed(2);
};

onMounted(() => {
  initPrices();
  getBasePrices();
  api
    .get(
      `/billing/currencies/rates/${account.value.currency}/${defaultCurrency.value}`
    )
    .then((res) => {
      accountRate.value = res.rate;
      pricesItems.value = pricesItems.value.map((i) => {
        i.accountPrice = toAccountPrice(i.price);
        return i;
      });
    });
});
</script>

<style scoped></style>
