<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="Price model"
          :value="template.billingPlan.title"
        >
          <template v-slot:append>
            <v-icon
              v-if="isPriceModelCanBeChange"
              @click="priceModelDialog = true"
              >mdi-pencil</v-icon
            >
            <v-icon
              @click="
                $router.push({
                  name: 'Plan',
                  params: { planId: template.billingPlan.uuid },
                })
              "
              >mdi-login</v-icon
            >
          </template>
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Product name"
          :value="tarrif.title"
          :append-icon="isPriceModelCanBeChange ? 'mdi-pencil' : undefined"
          @click:append="
            isPriceModelCanBeChange ? (priceModelDialog = true) : undefined
          "
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Date (create)"
          :value="formatSecondsToDate(template.created, true)"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="dueDate"
        />
      </v-col>
    </v-row>

    <instances-panels title="Prices">
      <nocloud-table
        hide-default-footer
        sort-by="index"
        item-key="key"
        :show-select="false"
        :headers="pricesHeaders"
        :items="pricesItems"
      >
        <template v-slot:[`item.prices`]="{ item }">
          <div class="d-flex">
            <v-text-field
              class="mr-2"
              v-model="item.price"
              @change="onUpdatePrice(item, false)"
              :suffix="defaultCurrency"
              type="number"
              append-icon="mdi-pencil"
            ></v-text-field>
            <v-text-field
              class="ml-2"
              style="color: var(--v-primary-base)"
              v-model="item.accountPrice"
              @change="onUpdatePrice(item, true)"
              :suffix="accountCurrency"
              type="number"
              append-icon="mdi-pencil"
            ></v-text-field>
          </div>
        </template>
        <template v-slot:[`item.basePrice`]="{ item }">
          <v-skeleton-loader type="text" v-if="isBasePricesLoading" />
          <span v-else> {{ convertedBasePrices[item.key] }} PLN </span>
        </template>
        <template v-slot:[`item.title`]="{ item }">
          <span v-html="item.title" />
          <v-chip v-if="item.isAddon" small class="ml-1">Addon</v-chip>
        </template>
        <template v-slot:body.append>
          <tr>
            <td></td>
            <td></td>
            <td>
              {{ getBillingPeriod(tarrif.period) }}
            </td>
            <td>
              {{
                isBasePricesLoading
                  ? "Loading..."
                  : [totalBasePrice, "PLN"].join(" ")
              }}
            </td>
            <td>
              <div class="d-flex justify-end">
                <v-chip outlined color="primary" class="mr-4">
                  {{ [totalNewPrice?.toFixed(2), defaultCurrency].join(" ") }}
                  /
                  {{ [accountTotalNewPrice, accountCurrency].join(" ") }}
                </v-chip>
              </div>
            </td>
          </tr>
        </template>
      </nocloud-table>
    </instances-panels>

    <edit-price-model
      v-if="isPriceModelCanBeChange"
      @refresh="emit('refresh')"
      :template="template"
      :plans="plans"
      :account-currency="accountCurrency"
      :account-rate="accountRate"
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
  watch,
} from "vue";
import NocloudTable from "@/components/table.vue";
import api from "@/api";
import { useStore } from "@/store";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";
import usePlnRate from "@/hooks/usePlnRate";
import { formatSecondsToDate, getBillingPeriod } from "@/functions";
import useInstancePrices from "@/hooks/useInstancePrices";
import InstancesPanels from "@/components/ui/nocloudExpansionPanels.vue";

const props = defineProps(["template", "plans", "account"]);
const emit = defineEmits(["refresh", "update"]);

const { template, plans, account } = toRefs(props);

const store = useStore();
const rate = usePlnRate();
const { toAccountPrice, fromAccountPrice, accountCurrency, accountRate } =
  useInstancePrices(template.value, account.value);

const pricesItems = ref([]);
const basePrices = ref({});
const pricesHeaders = ref([
  { text: "Name", value: "title" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Base price", value: "basePrice" },
  { text: "Price", value: "prices" },
]);
const totalNewPrice = ref(0);
const isBasePricesLoading = ref(false);
const priceModelDialog = ref(false);

onMounted(() => {
  initPrices();
  getBasePrices();
});

const accountTotalNewPrice = computed(() =>
  toAccountPrice(totalNewPrice.value)
);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

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

const dueDate = computed(() => {
  return formatSecondsToDate(+template.value?.data?.next_payment_date, true);
});

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
  return { ...(template.value.billingPlan.products[key] || {}), key };
});

const service = computed(() =>
  store.getters["services/all"].find((s) => s.uuid === template.value.service)
);

const isPriceModelCanBeChange = computed(() => ["cloud"].includes(type.value));

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
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch base prices",
    });
  } finally {
    isBasePricesLoading.value = false;
  }
};
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

const getVpsPrices = async () => {
  const { meta } = await api.servicesProviders.action({
    uuid: template.value.sp,
    action: "get_plans",
  });

  const prices = {};

  const planCodeCurr = meta.plans.find((p) => planCode.value === p.planCode);
  prices[planCode.value] = getPriceFromProduct(planCodeCurr);
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
  const prices = {};

  prices[planCode.value] = tarrif.value.meta?.basePrice;

  template.value.config.addons.forEach((a) => {
    prices[a] = template.value.billingPlan.resources.find(
      (r) => r.key === getAddonKey(a)
    ).meta.basePrice;
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
  prices[planCode.value] = meta.codes[tarrif.value?.meta?.priceCode];
  return prices;
};

const initPrices = () => {
  const productKey = [duration.value, planCode.value].join(" ");
  pricesItems.value.push({
    title: template.value.billingPlan.products[productKey]?.title,
    key: planCode.value,
    ind: 0,
    path: `billingPlan.products.${productKey}.price`,
    kind: tarrif.value.kind,
    price: tarrif.value?.price,
    period: tarrif.value?.period,
  });

  addons.value.forEach((key, ind) => {
    const addonIndex = template.value.billingPlan.resources.findIndex(
      (p) => p.key === getAddonKey(key)
    );

    const addon = template.value.billingPlan.resources[addonIndex];

    if (!addon) {
      return;
    }

    pricesItems.value.push({
      price: addon.price || 0,
      isAddon: true,
      path: `billingPlan.resources.${addonIndex}.price`,
      title: template.value.billingPlan.resources[addonIndex]?.title || key,
      kind: addon.kind,
      key: key,
      index: ind + 1,
      period: addon.period,
    });
  });

  pricesItems.value = pricesItems.value.map((i) => {
    i.period = getBillingPeriod(i.period);

    return i;
  });

  setAccountsPrices();
  setTotalNewPrice();
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

const setAccountsPrices = () => {
  pricesItems.value = pricesItems.value.map((i) => {
    i.accountPrice = toAccountPrice(i.price);

    return i;
  });
};

watch(accountRate, () => {
  setAccountsPrices();
});
</script>

<style scoped></style>
