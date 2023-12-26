<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-expansion-panels v-if="!isLoading">
      <v-expansion-panel>
        <v-expansion-panel-header color="background">
          Margin rules:
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background">
          <plan-opensrs
            :fee="fee"
            :isEdit="true"
            @changeFee="changeFee"
            @onValid="(data) => (isValid = data)"
          />
          <confirm-dialog
            text="This will apply the rules markup parameters to all prices"
            @confirm="setFee"
          >
            <v-btn class="mt-4" color="secondary">Set rules</v-btn>
          </confirm-dialog>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <v-tabs
      class="rounded-t-lg"
      v-model="tabsIndex"
      background-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab">{{ tab }}</v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabsIndex"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item v-for="tab in tabs" :key="tab">
        <v-progress-linear v-if="isLoading" indeterminate class="pt-1" />

        <div v-if="tab === 'Tariffs'">
          <div class="mt-4" v-if="!isLoading">
            <v-btn class="mx-1" @click="setSellToAllTariffs(true)"
              >Enable all</v-btn
            >
            <v-btn class="mx-1" @click="setSellToAllTariffs(false)"
              >Disable all</v-btn
            >
          </div>
          <nocloud-table
            :headers="headers"
            table-name="keyweb-prices"
            class="pa-4"
            item-key="text"
            :show-select="false"
            :items="tariffs"
          >
            <template v-slot:[`item.price`]="{ item }">
              <v-text-field v-model.number="item.price" type="number" />
            </template>
            <template v-slot:[`item.name`]="{ item }">
              <v-text-field v-model="item.name" />
            </template>
            <template v-slot:[`item.sell`]="{ item }">
              <v-switch v-model.number="item.sell" />
            </template>

            <template v-slot:[`item.os`]="{ item }">
              <v-btn icon @click="openOs(item)">
                <v-icon> mdi-menu-open </v-icon>
              </v-btn>
            </template>
            <template v-slot:[`item.addons`]="{ item }">
              <v-btn icon @click="openAddons(item)">
                <v-icon> mdi-menu-open </v-icon>
              </v-btn>
            </template>
          </nocloud-table>
        </div>

        <plans-resources-table
          v-show="tab === 'Addons'"
          :default-virtual="true"
          @change:resource="changeCustomAddons($event)"
          :resources="customAddonsResources"
          type="ovh dedicated"
        />
      </v-tab-item>
    </v-tabs-items>

    <v-dialog max-width="70vw" v-model="isAddonsOpen">
      <v-card color="background-light">
        <v-card-title
          >Addons
          <v-btn class="ml-3 mr-2" small @click="setSellToAllAddons(true)"
            >Enable all</v-btn
          >
          <v-btn small @click="setSellToAllAddons(false)">Disable all</v-btn>
        </v-card-title>
        <nocloud-table
          :show-select="false"
          :items="selectedTariff?.addons"
          :headers="addonsHeaders"
        >
          <template v-slot:[`item.name`]="{ item }">
            <v-text-field :disabled="item.virtual" v-model="item.name" />
          </template>
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field
              :disabled="item.virtual"
              v-model.number="item.price"
              type="number"
            />
          </template>
          <template v-slot:[`item.sell`]="{ item: addon }">
            <v-switch v-model="addon.sell" />
          </template>
        </nocloud-table>
      </v-card>
    </v-dialog>
    <v-dialog max-width="70vw" v-model="isOsOpen">
      <v-card color="background-light">
        <v-card-title
          >Os
          <v-btn class="ml-3 mr-2" small @click="setSellToAllOs(true)"
            >Enable all</v-btn
          >
          <v-btn small @click="setSellToAllOs(false)">Disable all</v-btn>
        </v-card-title>
        <nocloud-table
          :show-select="false"
          :items="selectedTariff?.os"
          :headers="osHeaders"
        >
          <template v-slot:[`item.name`]="{ item }">
            <v-text-field v-model="item.name" />
          </template>
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field v-model.number="item.price" type="number" />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model.number="item.sell" />
          </template>
        </nocloud-table>
      </v-card>
    </v-dialog>

    <v-card-actions class="d-flex justify-end">
      <v-btn @click="save" :loading="isSaveLoading">save</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { ref, defineProps, onMounted, computed } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import planOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { getBillingPeriod, getMarginedValue, getTimestamp } from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import plansResourcesTable from "@/components/plans_resources_table.vue";

const props = defineProps(["template"]);

const store = useStore();
const { convertFrom } = useCurrency();

const isLoading = ref(false);
const isSaveLoading = ref(false);
const isAddonsOpen = ref(false);
const isOsOpen = ref(false);
const isValid = ref(false);
const selectedTariff = ref({});
const tariffs = ref([]);
const fee = ref({});
const tabsIndex = ref(0);
const tabs = ["Tariffs", "Addons"];
const customAddonsResources = ref([]);

const headers = ref([
  { text: "Key", value: "key" },
  { text: "Name", value: "name", width: "250" },
  { text: "Duration", value: "duration" },
  { text: "Os", value: "os", width: "50" },
  { text: "Addons", value: "addons", width: "50" },
  { text: "Base price", value: "basePrice", width: "75" },
  { text: "Price", value: "price", width: "200" },
  { text: "Sell", value: "sell", width: "100" },
]);

const addonsHeaders = ref([
  { text: "Base name", value: "baseName" },
  { text: "Name", value: "name", width: "250" },
  { text: "Duration", value: "duration" },
  { text: "Base price", value: "basePrice", width: "75" },
  { text: "Price", value: "price", width: "200" },
  { text: "Sell", value: "sell", width: "100" },
]);

const osHeaders = ref([
  { text: "Base name", value: "baseName" },
  { text: "Name", value: "name", width: "250" },
  { text: "Duration", value: "duration" },
  { text: "Base price", value: "basePrice", width: "75" },
  { text: "Price", value: "price", width: "200" },
  { text: "Sell", value: "sell", width: "100" },
]);

onMounted(async () => {
  isLoading.value = true;
  try {
    customAddonsResources.value = [
      ...props.template.resources.filter((p) => p.virtual),
    ];

    await store.dispatch("servicesProviders/fetch", { anonymously: false });
    const spUuid = sps.value.find((sp) =>
      sp.meta?.plans?.includes(props.template.uuid)
    ).uuid;

    const { meta } = await api.servicesProviders.action({
      action: "list_products",
      uuid: spUuid,
    });
    const products = [];
    meta.products.forEach((p) => {
      const data = {
        name: [p.integration, p.name].join(" "),
        id: p.id,
      };

      const keys = {
        monthly: `${p.integration}_${p.name}-monthly`,
        yearly: `${p.integration}_${p.name}-yearly`,
      };

      const realProducts = {
        monthly: props.template?.products[keys.monthly],
        yearly: props.template?.products[keys.yearly],
      };

      const addonsAnnotations = {
        backup: "Backups Files Limit|Backup",
        os: "VM Template|OS",
      };
      const addonsValues = {
        backup: [],
        os: [],
      };

      Object.keys(addonsAnnotations).forEach((key) => {
        const metaKey = addonsAnnotations[key];
        p.configOptions[metaKey]?.configurableSubOptions.forEach((c) => {
          const basePriceYearly = convertFrom(
            +p.pricing.configOptions[metaKey][c.optionname].annually,
            "EUR"
          );
          const basePriceMonthly = convertFrom(
            +p.pricing.configOptions[metaKey][c.optionname].monthly,
            "EUR"
          );

          const data = {
            name: c.optionname,
            sell: false,
            type: metaKey,
            baseName: c.optionname,
          };
          addonsValues[key].push({
            ...data,
            price: basePriceYearly,
            basePrice: basePriceYearly,
            duration: "yearly",
          });
          addonsValues[key].push({
            ...data,
            price: basePriceMonthly,
            basePrice: basePriceMonthly,
            duration: "monthly",
          });
        });
      });

      const getTariffAddons = (subKey, duration) => {
        return addonsValues[subKey]
          .filter((a) => a.duration === duration)
          .map((a) => {
            a.key = a.baseName + "$" + keys[duration];
            const realAddon = props.template.resources.find(
              (r) => r.key === a.key
            );
            if (realAddon) {
              a.price = realAddon?.price;
              a.name = realAddon?.title;
              a.sell = true;
            }

            return a;
          });
      };

      const getProduct = (duration) => {
        const basePrice = convertFrom(
          p.pricing.productrenew[
            duration === "yearly" ? "annually" : "monthly"
          ],
          "EUR"
        );

        return {
          ...data,
          name: realProducts[duration]?.title || data.name,
          key: keys[duration],
          duration: duration,
          os: getTariffAddons("os", duration),
          addons: getTariffAddons("backup", duration).concat(
            customAddonsResources.value
              .filter((a) => a.public)
              .map((a) => ({
                baseName: a.title,
                basePrice: null,
                duration: getBillingPeriod(a.period),
                name: a.title,
                price: a.price,
                virtual: true,
                key: a.key,
                sell: !!realProducts[duration]?.meta.addons?.includes(a.key),
                type: "Custom",
              }))
          ),
          price: realProducts[duration]?.price || basePrice,
          basePrice,
          sell: !!realProducts[duration],
        };
      };

      products.push(getProduct("monthly"));
      products.push(getProduct("yearly"));
    });

    tariffs.value = products;
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch keyweb prices",
    });
  } finally {
    isLoading.value = false;
  }
});

const sps = computed(() => store.getters["servicesProviders/all"]);

const openOs = (item) => {
  isOsOpen.value = true;
  selectedTariff.value = item;
};
const openAddons = (item) => {
  isAddonsOpen.value = true;
  selectedTariff.value = item;
};

const setSellToAllOs = (value) => {
  selectedTariff.value.os = selectedTariff.value.os.map((os) => ({
    ...os,
    sell: value,
  }));
};

const setSellToAllAddons = (value) => {
  selectedTariff.value.addons = selectedTariff.value.addons.map((os) => ({
    ...os,
    sell: value,
  }));
};

const setSellToAllTariffs = (value) => {
  tariffs.value = tariffs.value.map((t) => {
    t.sell = value;
    t.os = t.os.map((os) => ({ ...os, sell: value }));
    t.addons = t.addons.map((a) => ({ ...a, sell: value }));
    return t;
  });
};

const save = async () => {
  const products = {};
  const resources = [];

  resources.push(...customAddonsResources.value);

  const enabledTariffs = tariffs.value.filter((t) => t.sell);

  enabledTariffs.forEach((t) => {
    const period = t.duration === "monthly" ? 3600 * 24 * 30 : 3600 * 24 * 365;
    const kind = "PREPAID";

    const enabledAddons = t.addons.filter((a) => a.sell);
    const enabledOs = t.os.filter((os) => os.sell);
    enabledOs.concat(enabledAddons).forEach((r) => {
      if (r.virtual) {
        return;
      }
      resources.push({
        key: r.key,
        kind,
        price: r.price,
        title: r.name,
        meta: {
          type: r.type,
        },
        period,
      });
    });

    products[t.key] = {
      kind,
      period,
      price: t.price,
      title: t.name,
      meta: {
        addons: enabledAddons.map((a) => a.key),
        os: enabledOs.map((a) => a.key),
        keywebId: t.id,
      },
    };
  });

  isSaveLoading.value = true;
  try {
    await api.plans.update(props.template.uuid, {
      ...props.template,
      products,
      resources,
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save prices",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const changeFee = (value) => {
  fee.value = JSON.parse(JSON.stringify(value));
};

const changeCustomAddons = ({ key, value, id }) => {
  if (key === "resources") {
    customAddonsResources.value = value;
    return;
  }

  customAddonsResources.value = customAddonsResources.value.map((c) => {
    if (c.id === id) {
      if (key === "date") {
        c.period = getTimestamp(value);
      } else {
        c[key] = value;
      }
    }
    return c;
  });
};

const setFee = () => {
  tariffs.value.forEach((t) => {
    t.price = getMarginedValue(fee.value, t.basePrice);
    t.os = t.os.map((os) => ({
      ...os,
      price: getMarginedValue(fee.value, os.basePrice),
    }));
    t.addons = t.addons.map((a) => ({
      ...a,
      price: getMarginedValue(fee.value, a.basePrice),
    }));
  });
};
</script>

<style scoped></style>
