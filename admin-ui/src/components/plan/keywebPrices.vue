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

    <div>
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
        class="pa-4"
        :loading="isLoading"
        item-key="key"
        :show-select="false"
        :expanded.sync="expanded"
        show-expand
        :items="tariffs"
      >
        <template v-slot:expanded-item="{ headers, item }">
          <td :colspan="headers.length">
            <rich-editor v-model="item.description" />
          </td>
        </template>

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
import { ref, defineProps, onMounted, computed, toRefs } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import planOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { getMarginedValue } from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import RichEditor from "@/components/ui/richEditor.vue";
import {
  Addon,
  ListAddonsRequest,
} from "nocloud-proto/proto/es/billing/addons/addons_pb";

const props = defineProps(["template"]);
const { template } = toRefs(props);

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
const expanded = ref([]);

const headers = ref([
  { text: "Key", value: "key" },
  { text: "Title", value: "name", width: "250" },
  { text: "Duration", value: "duration" },
  { text: "Os", value: "os", width: "50" },
  { text: "Addons", value: "addons", width: "50" },
  { text: "Incoming price", value: "basePrice", width: "75" },
  { text: "Sale price", value: "price", width: "200" },
  { text: "Sell", value: "sell", width: "100" },
]);

const addonsHeaders = ref([
  { text: "Base title", value: "baseName" },
  { text: "Title", value: "name", width: "250" },
  { text: "Duration", value: "duration" },
  { text: "Incoming price", value: "basePrice", width: "75" },
  { text: "Sale price", value: "price", width: "200" },
  { text: "Sell", value: "sell", width: "100" },
]);

const osHeaders = ref([
  { text: "Base title", value: "baseName" },
  { text: "Title", value: "name", width: "250" },
  { text: "Duration", value: "duration" },
  { text: "Incoming price", value: "basePrice", width: "75" },
  { text: "Sale price", value: "price", width: "200" },
  { text: "Sell", value: "sell", width: "100" },
]);

onMounted(async () => {
  isLoading.value = true;
  try {
    const addonsPromise = addonsClient.value.list(
      ListAddonsRequest.fromJson({
        filters: { group: template.value.uuid },
      })
    );
    await store.dispatch("servicesProviders/fetch", { anonymously: false });
    const spUuid = sps.value.find((sp) =>
      sp.meta?.plans?.includes(props.template.uuid)
    ).uuid;

    const { meta } = await api.servicesProviders.action({
      action: "list_products",
      uuid: spUuid,
    });

    const { addons = [] } = (await addonsPromise).toJson();
    console.log(addons);
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
          .filter((addon) => addon.duration === duration)
          .map((addon) => {
            addon.key = addon.baseName + "$" + keys[duration];
            const realAddon = addons.find(
              (realAddon) => realAddon.meta?.key === addon.key
            );

            if (realAddon) {
              addon.price = Object.values(realAddon.periods)[0];
              addon.name = realAddon.title;
              addon.sell = realAddon.public;
              addon.uuid = realAddon.uuid;
            }

            return addon;
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
          addons: getTariffAddons("backup", duration),
          price: realProducts[duration]?.price || basePrice,
          basePrice,
          description: realProducts[duration]?.meta?.description,
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
const addonsClient = computed(() => store.getters["addons/addonsClient"]);

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

  isSaveLoading.value = true;
  try {
    await Promise.all(
      tariffs.value.map(async (t) => {
        const period =
          t.duration === "monthly" ? 3600 * 24 * 30 : 3600 * 24 * 365;
        const kind = "PREPAID";

        const addons = await Promise.all(
          t.os.concat(t.addons).map(async (addon) => {
            const data = Addon.fromJson({
              group: template.value.uuid,
              title: addon.name,
              public: addon.sell,
              system: true,
              meta: {
                type: addon.type,
                key: addon.key,
              },
              periods: { [period]: addon.price },
            });
            if (addon.uuid) {
              data.uuid = addon.uuid;
              return addonsClient.value.update(data);
            }

            return addonsClient.value.create(data);
          })
        );

        products[t.key] = {
          kind,
          period,
          price: t.price,
          title: t.name,
          public: t.sell,
          addons: addons.map((a) => a.uuid),
          meta: {
            keywebId: t.id,
            description: t.description,
          },
        };
      })
    );

    await api.plans.update(props.template.uuid, {
      ...props.template,
      products,
      resources,
    });
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Price model edited successfully",
    });
  } catch (e) {
    console.log(e);
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
