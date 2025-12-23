<template>
  <div>
    <v-row
      class="my-4"
      v-if="!isPlansLoading"
      align="center"
      justify="space-between"
    >
      <div>
        <v-btn class="ml-3" @click="refreshPlans" :loading="isRefreshLoading"
          >Fetch plans</v-btn
        >
        <v-btn class="ml-3" :disabled="!newPlans" @click="setRefreshedPlans"
          >Set api plans</v-btn
        >
        <v-btn class="ml-3" @click="setSellToTab(true)">Enable all</v-btn>
        <v-btn class="ml-3" @click="setSellToTab(false)">Disable all</v-btn>
      </div>
      <div class="mr-10">
        <v-text-field
          v-model="searchParam"
          append-icon="mdi-magnify"
          label="Search"
          single-line
          hide-details
        ></v-text-field>
      </div>
    </v-row>
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
        <v-progress-linear v-if="isPlansLoading" indeterminate class="pt-1" />

        <nocloud-table
          item-key="id"
          v-else-if="tab === 'Tariffs'"
          :show-expand="true"
          :show-select="false"
          :items="filtredPlans"
          :headers="headers"
          :expanded.sync="expanded"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.name`]="{ item }">
            <v-text-field dense style="width: 200px" v-model="item.name" />
          </template>

          <template v-slot:[`item.resources`]="{ item }">
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <span color="primary" dark v-bind="attrs" v-on="on">
                  CPU:{{ item.resources.cpu || "N/A" }}
                </span>
              </template>
              <pre style="color: black">{{ item.resources }}</pre>
            </v-tooltip>
          </template>
          <template v-slot:[`item.group`]="{ item }">
            <template v-if="mode === 'edit' && planId === item.id">
              <v-text-field
                dense
                class="d-inline-block mr-1"
                style="width: 200px"
                v-model="newGroupName"
              />
              <v-icon @click="editGroup(item.group)">mdi-content-save</v-icon>
              <v-icon @click="mode = 'none'">mdi-close</v-icon>
            </template>

            <template v-if="mode === 'create' && planId === item.id">
              <v-text-field
                dense
                class="d-inline-block mr-1"
                style="width: 200px"
                v-model="newGroupName"
              />
              <v-icon @click="createGroup(item)">mdi-content-save</v-icon>
              <v-icon @click="mode = 'none'">mdi-close</v-icon>
            </template>

            <template v-if="mode === 'none'">
              <v-select
                dense
                class="d-inline-block"
                style="width: 200px"
                v-model="item.group"
                :items="groups"
              />
              <v-icon @click="changeMode('create', item)">mdi-plus</v-icon>
              <v-icon @click="changeMode('edit', item)">mdi-pencil</v-icon>
              <v-icon v-if="groups.length > 1" @click="deleteGroup(item.group)"
                >mdi-delete</v-icon
              >
            </template>

            <template v-else-if="planId !== item.id">{{ item.group }}</template>
          </template>
          <template v-slot:[`item.duration`]="{ value }">
            {{ getPayment(value) }}
          </template>
          <template v-slot:[`item.price.value`]="{ value }">
            {{ value }} {{ defaultCurrency?.code }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field
              dense
              style="width: 200px"
              :suffix="defaultCurrency?.code"
              v-model="item.value"
            />
          </template>
          <template v-slot:[`item.public`]="{ item }">
            <v-switch v-model="item.public" />
          </template>
        </nocloud-table>

        <nocloud-table
          v-else-if="tab === 'Addons'"
          :show-select="false"
          :items="filtredAddons"
          :headers="addonsHeaders"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.duration`]="{ value }">
            {{ getPayment(value) }}
          </template>
          <template v-slot:[`item.price.value`]="{ value }">
            {{ value }} {{ defaultCurrency?.code }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field
              dense
              style="width: 200px"
              :suffix="defaultCurrency?.code"
              v-model="item.value"
            />
          </template>
          <template v-slot:[`item.public`]="{ item }">
            <v-switch v-model="item.public" />
          </template>
        </nocloud-table>

        <nocloud-table
          v-else-if="tab === 'OS'"
          :show-select="false"
          :items="filtredImages"
          :headers="imagesHeaders"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.tariff`]="{ value }">
            {{ value || "Any" }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field
              v-if="item"
              dense
              style="width: 200px"
              :suffix="defaultCurrency?.code"
              v-model="item.value"
            />
          </template>
          <template v-slot:[`item.public`]="{ item }">
            <v-switch v-model="item.public" />
          </template>
        </nocloud-table>

        <div v-else>
          <plan-addons-table
            @change:addons="planAddons = $event"
            :addons="template.addons"
          />
        </div>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from "vue";
import { useStore } from "@/store";
import api from "@/api.js";
import nocloudTable from "@/components/table.vue";
import { getMarginedValue } from "@/functions";
import {
  Addon,
  ListAddonsRequest,
} from "nocloud-proto/proto/es/billing/addons/addons_pb";
import planAddonsTable from "@/components/planAddonsTable.vue";
import useCurrency from "@/hooks/useCurrency";
import { replaceNullWithUndefined } from "../../functions";

const props = defineProps({
  fee: { type: Object, required: true },
  template: { type: Object, required: true },
  isPlansLoading: { type: Boolean, required: true },
  getPeriod: { type: Function, required: true },
  sp: { type: Object, required: true },
});

const emit = defineEmits(["changeFee", "change:addons"]);

const store = useStore();
const { convertFrom } = useCurrency();

const searchParam = ref("");

const groups = ref([]);
const expanded = ref([]);
const tabs = ref(["Tariffs", "Addons", "OS", "Custom addons"]);

const plans = ref([]);
const headers = ref([
  { text: "", value: "data-table-expand" },
  { text: "Title", value: "name" },
  { text: "API title", value: "apiName" },
  { text: "Group", value: "group" },
  { text: "Resources", value: "resources", width: 50 },
  {
    text: "Payment",
    value: "duration",
  },
  { text: "Incoming price", value: "price.value" },
  { text: "Sale price", value: "value" },
  {
    text: "Sell",
    value: "public",
    width: 100,
  },
]);

const addons = ref([]);
const addonsHeaders = ref([
  { text: "Addon", value: "name" },
  {
    text: "Payment",
    value: "duration",
  },
  { text: "Incoming price", value: "price.value" },
  { text: "Sale price", value: "value" },
  {
    text: "Sell",
    value: "public",
    width: 100,
  },
]);

const images = ref([]);
const imagesHeaders = ref([
  { text: "OS", value: "name" },
  { text: "Tariff", value: "tariff" },
  { text: "Incoming price", value: "price.value" },
  { text: "Sale price", value: "value" },
  { text: "Payment", value: "duration" },
  {
    text: "Sell",
    value: "public",
    width: 100,
  },
]);

const fetchError = ref("");
const newGroupName = ref("");
const mode = ref("none");

const planId = ref(-1);
const tabsIndex = ref(0);
const usedFee = ref({});

const newPlans = ref(null);
const newAddons = ref(null);
const newImages = ref(null);
const isRefreshLoading = ref(false);

const planAddons = ref([]);

const defaultCurrency = computed(() => store.getters["currencies/default"]);
const addonsClient = computed(() => store.getters["addons/addonsClient"]);

const filtredPlans = computed(() => {
  if (!searchParam.value) return plans.value;

  return plans.value.filter(
    (plan) =>
      plan.name.toLowerCase().includes(searchParam.value.toLowerCase()) ||
      plan.apiName.toLowerCase().includes(searchParam.value.toLowerCase()) ||
      plan.group.toLowerCase().includes(searchParam.value.toLowerCase())
  );
});

const filtredAddons = computed(() => {
  if (!searchParam.value) return addons.value;

  return addons.value.filter((addon) =>
    addon.name.toLowerCase().includes(searchParam.value.toLowerCase())
  );
});

const filtredImages = computed(() => {
  if (!searchParam.value) return images.value;

  return images.value.filter((image) =>
    image.name.toLowerCase().includes(searchParam.value.toLowerCase())
  );
});

const testConfig = () => {
  if (!plans.value.every(({ group }) => groups.value.includes(group))) {
    return "You must select a group for the tariff!";
  }
};

const changePlan = async (plan) => {
  plan.resources = [];
  plan.products = {};

  let allAddons = [];
  let addonsArray = [];

  addons.value.forEach((addon) => {
    const addonkey = addon.id.split(" ")[1];
    const existedAddon = addonsArray.find(
      (existed) => existed.meta.key === addonkey
    );

    let data;
    if (existedAddon) {
      existedAddon.periods[props.getPeriod(addon.duration)] = addon.value;
      existedAddon.meta.basePrices[props.getPeriod(addon.duration)] =
        addon.price.value;
      data = existedAddon;
      return;
    } else {
      data = {
        system: true,
        title: addon.name,
        group: props.template.uuid,
        periods: { [props.getPeriod(addon.duration)]: addon.value },
        public: !!addon.public,
        kind: "PREPAID",
        meta: {
          basePrices: {
            [props.getPeriod(addon.duration)]: addon.price.value,
          },
          key: addonkey,
          type: "addon",
        },
      };
    }

    if (addon.uuid) {
      data.uuid = addon.uuid;
      addonsArray.push({ ...data, type: "update" });
    } else {
      addonsArray.push({ ...data, type: "create" });
    }
  });

  images.value.forEach((addon) => {
    const existedAddon = addonsArray.find(
      (existed) =>
        existed.meta.type === "os" &&
        existed.meta.key === addon.id &&
        existed.meta.tariff === addon.tariff
    );

    let data;
    if (existedAddon) {
      existedAddon.periods[props.getPeriod(addon.duration)] = addon.value;
      existedAddon.meta.basePrices[props.getPeriod(addon.duration)] =
        addon.price.value;
      data = existedAddon;
      return;
    } else {
      data = {
        system: true,
        title: addon.name,
        group: props.template.uuid,
        periods: { [props.getPeriod(addon.duration)]: addon.value },
        public: !!addon.public,
        kind: "PREPAID",
        meta: {
          basePrices: {
            [props.getPeriod(addon.duration)]: addon.price.value,
          },
          key: addon.id,
          type: "os",
        },
      };
    }

    if (addon.tariff) {
      data.meta.tariff = addon.tariff;
    }

    if (addon.uuid) {
      data.uuid = addon.uuid;
      addonsArray.push({ ...data, type: "update" });
    } else {
      addonsArray.push({ ...data, type: "create" });
    }
  });

  const addonsForCreate = addonsArray
    .filter((a) => a.type === "create")
    .map((a) => {
      delete a.type;
      return replaceNullWithUndefined(a);
    });

  const addonsForUpdate = addonsArray
    .filter((a) => a.type === "update")
    .map((a) => {
      delete a.type;
      return replaceNullWithUndefined(a);
    });

  if (addonsForCreate.length) {
    const createdAddons = await addonsClient.value.createBulk({
      addons: addonsForCreate.map((addon) => Addon.fromJson(addon)),
    });

    allAddons.push(...createdAddons.toJson().addons);
  }

  if (addonsForUpdate.length) {
    const updatedAddons = await addonsClient.value.updateBulk({
      addons: addonsForUpdate.map((addon) => Addon.fromJson(addon)),
    });
    allAddons.push(...updatedAddons.toJson().addons);
  }

  plans.value.forEach((el) => {
    const meta = {
      datacenter: el.datacenter,
    };

    const addonsForProduct = el.addons
      .map((key) => allAddons.find((addon) => key === addon.meta.key)?.uuid)
      .filter((a) => !!a)
      .concat(
        el.os
          .map(
            (key) =>
              allAddons.find(
                (addon) =>
                  key === addon.meta.key &&
                  (!addon.meta.tariff ||
                    addon.meta.tariff ===
                      `option-windows-${el.planCode.replace("vps-", "")}`)
              )?.uuid
          )
          .filter((a) => !!a)
      );

    console.log(
      el.addons
        .map((key) => allAddons.find((addon) => key === addon.meta.key))
        .filter((a) => !!a)
    );

    plan.products[el.id] = {
      kind: "PREPAID",
      title: el.name,
      price: el.value,
      public: el.public,
      group: el.group,
      addons: addonsForProduct,
      period: props.getPeriod(el.duration),
      resources: el.resources,
      meta: {
        ...meta,
        basePrice: el.price.value,
        apiName: el.apiName,
      },
      sorter: Object.keys(plan.products).length,
      installation_fee: el.installation_fee,
    };
  });

  plan.addons = planAddons.value;
};

const changePlans = ({ plans: plansData, catalog }) => {
  const result = [];

  plansData.forEach(({ prices, planCode, productName }) => {
    if (!catalog.products.find((p) => planCode === p.name)) {
      return;
    }

    prices.forEach(({ pricingMode, price, duration }) => {
      const isMonthly = duration === "P1M" && pricingMode === "default";
      const isYearly = duration === "P1Y" && pricingMode === "upfront12";

      if (isMonthly || isYearly) {
        const id = `${duration} ${planCode}`;
        const realProduct = plans.value.find((p) => p.id === id) || {};

        const code = planCode;
        const newPrice = convertPrice(price.value);

        const { configurations, addonFamilies } = catalog.plans.find(
          ({ planCode }) => planCode === code
        );

        const product = catalog.products.find((p) => planCode === p.name);
        const technical = product?.blobs?.technical || {};

        let cpu, ram, drive_size, drive_type;
        if (technical.cpu) {
          cpu = technical.cpu.cores || 0;
        }

        if (technical.memory) {
          ram = (technical.memory.size || 0) * 1024;
        }

        if (technical.storage) {
          drive_size = (technical.storage.disks || [])[0].capacity * 1024;
          drive_type = (technical.storage.disks || [])[0].technology;
        }

        const os = configurations.find((c) => c.name === "vps_os")?.values;
        const datacenter = configurations.find(
          (c) => c.name === "vps_datacenter"
        )?.values;

        const addonsData = addonFamilies.reduce(
          (res, { addons }) => [...res, ...addons],
          []
        );
        const plan = { addons: addonsData, os, datacenter };

        const installation = prices.find(
          (price) =>
            price.capacities.includes("installation") &&
            price.pricingMode === pricingMode
        );

        result.push({
          ...plan,
          planCode,
          duration,
          installation_fee:
            realProduct.installation_fee || installation.price.value,
          price: { value: newPrice },
          name: realProduct.name || productName,
          apiName: productName,
          resources: { cpu, ram, drive_size, drive_type },
          group:
            realProduct.group ||
            productName.replace(/VPS[\W0-9]/, "").split(/[\W0-9]/)[0],
          value: realProduct.value || newPrice,
          public: !!realProduct.public,
          id,
        });
      }
    });
  });
  result.sort((a, b) => {
    const isCpuEqual = b.resources.cpu === a.resources.cpu;
    const isRamEqual = b.resources.ram === a.resources.ram;

    if (isCpuEqual && isRamEqual)
      return a.resources.drive_size - b.resources.drive_size;
    if (isCpuEqual) return a.resources.ram - b.resources.ram;
    return a.resources.cpu - b.resources.cpu;
  });

  return result;
};

const changeAddons = ({ backup, disk, snapshot }) => {
  const result = [];

  [backup, disk, snapshot].forEach((el) => {
    el.forEach(({ prices, planCode, productName }) => {
      prices.forEach(({ pricingMode, price, duration }) => {
        const isMonthly = duration === "P1M" && pricingMode === "default";
        const isYearly = duration === "P1Y" && pricingMode === "upfront12";

        if (isMonthly || isYearly) {
          const id = `${duration} ${planCode}`;
          const realAddon = addons.value.find((a) => a.id === id) || {};

          const newPrice = convertPrice(price.value);

          result.push({
            price: { value: newPrice },
            duration,
            name: productName,
            value: realAddon.value || newPrice,
            public: !!realAddon.public,
            id,
          });
        }
      });
    });
  });

  return result;
};

const changeImages = ({ windows }) => {
  const newImagesArray = [];

  newPlans.value.forEach((plan) =>
    plan.os.forEach((key) => {
      let tariff, price, basePrice;
      const periods = ["P1M", "P1Y"];

      periods.forEach((duration) => {
        let product;
        if (key.toLowerCase().includes("windows")) {
          product = windows.find(
            (w) =>
              w.planCode ===
              `option-windows-${plan.planCode.replace("vps-", "")}`
          );

          if (!product || !product?.prices) return;

          const data = product.prices.find(
            (price) =>
              price.duration === duration &&
              ["upfront12", "default"].includes(price.pricingMode)
          );
          if (!data) return;
          const { price: priceObject } = data;
          tariff = product.planCode;
          const newPrice = convertPrice(priceObject.value);
          price = newPrice;
          basePrice = newPrice;
        }

        if (
          newImagesArray.find(
            (os) =>
              os.id === key &&
              os.duration === duration &&
              (!os.tariff || os.tariff === product?.planCode)
          )
        ) {
          return;
        }

        const realAddon = images.value.find(
          (a) => a.id === key && tariff === a.tariff && a.duration === duration
        );

        newImagesArray.push({
          name: key,
          price: { value: basePrice || 0 },
          value: realAddon?.value || price || 0,
          public: realAddon ? realAddon.public : true,
          tariff,
          id: key,
          duration,
        });
      });
    })
  );

  return newImagesArray;
};

const setFee = () => {
  usedFee.value = JSON.parse(JSON.stringify(props.fee));

  [plans.value, addons.value, images.value].forEach((el) => {
    el.forEach((plan, i, arr) => {
      arr[i].value = getMarginedValue(props.fee, plan.price.value);
    });
  });
};

const getPayment = (duration) => {
  switch (duration) {
    case "P1M":
      return "monthly";
    case "P1Y":
      return "yearly";
  }
};

const editGroup = (group) => {
  const i = groups.value.indexOf(group);

  groups.value.splice(i, 1, newGroupName.value);
  plans.value.forEach((plan, index) => {
    if (plan.group !== group) return;
    plans.value[index].group = newGroupName.value;
  });

  changeMode("none", { id: -1, group: "" });
};

const createGroup = (plan) => {
  groups.value.push(newGroupName.value);
  plan.group = newGroupName.value;

  changeMode("none", { id: -1, group: "" });
};

const deleteGroup = (group) => {
  groups.value = groups.value.filter((el) => el !== group);
  plans.value.forEach((plan, i) => {
    if (plan.group !== group) return;
    plans.value[i].group = groups.value[0];
  });
};

const changeMode = (modeValue, { id, group }) => {
  mode.value = modeValue;
  planId.value = id;
  newGroupName.value = group;
};

const setSellToTab = (status) => {
  switch (tabs.value[tabsIndex.value]) {
    case "Addons": {
      setSellToValue(addons.value, status);
      break;
    }
    case "OS": {
      setSellToValue(images.value, status);
      break;
    }
    case "Tariffs": {
      setSellToValue(plans.value, status);
      break;
    }
  }
};

const setSellToValue = (value, status) => {
  value = value.map((p) => {
    p.public = status;
    return p;
  });
};

const convertPrice = (price) => {
  return convertFrom(price, { code: "PLN" });
};

const refreshPlans = async () => {
  try {
    isRefreshLoading.value = true;
    const { meta } = await api.servicesProviders.action({
      action: "get_plans",
      uuid: props.sp.uuid,
    });

    newPlans.value = changePlans(meta);
    newAddons.value = changeAddons(meta);
    newImages.value = changeImages(meta);

    if (
      !plans.value?.length ||
      !addons.value?.length ||
      !images.value?.length
    ) {
      setRefreshedPlans();
    }
  } catch (err) {
    newPlans.value = null;
    newAddons.value = null;
    store.commit("snackbar/showSnackbarError", {
      message: err.response?.data?.message ?? err.message ?? err,
    });
  } finally {
    isRefreshLoading.value = false;
  }
};

const setRefreshedPlans = () => {
  addons.value = JSON.parse(JSON.stringify(newAddons.value));
  plans.value = JSON.parse(JSON.stringify(newPlans.value));
  images.value = JSON.parse(JSON.stringify(newImages.value));
  newPlans.value = null;
  newAddons.value = null;

  setGroups();
};

const setGroups = () => {
  groups.value = [];
  plans.value.forEach((plan) => {
    const group = plan?.group || plan?.name?.split(/[\W0-9]/)[0];
    if (!groups.value.includes(group)) groups.value.push(group);
  });
};

const initializeData = async () => {
  const periodsDurationMap = { 31536000: "P1Y", 2592000: "P1M" };

  const { addons: addonsData = [] } = (
    await addonsClient.value.list(
      ListAddonsRequest.fromJson({ filters: { group: [props.template.uuid] } })
    )
  ).toJson();

  const newAddonsArray = [];
  addonsData
    .filter((addon) => addon.meta?.type != "os")
    .forEach((addon) =>
      newAddonsArray.push(
        ...Object.keys(addon.periods).map((period) => {
          const duration = periodsDurationMap[period];
          const { key: planCode, basePrices } = addon.meta;
          return {
            ...addon,
            duration,
            planCode,
            price: { value: basePrices[period] },
            value: addon.periods[period],
            name: addon.title,
            id: [duration, planCode].join(" "),
            uuid: addon.uuid,
          };
        })
      )
    );
  addons.value = newAddonsArray;

  plans.value = Object.keys(props.template.products || {}).map((key) => {
    const [duration, planCode] = key.split(" ");
    const product = props.template.products[key];

    const { meta } = product;

    const { apiName, datacenter, basePrice } = meta;
    console.log(
      product.addons
        ?.map((uuid) => addonsData.find((addon) => addon.uuid === uuid))
        .filter((addon) => addon?.meta.type !== "os")
        .map((addon) => addon?.meta.key)
    );

    return {
      ...product,
      duration,
      planCode,
      price: { value: basePrice },
      value: product.price,
      datacenter,
      addons: product.addons
        ?.map((uuid) => addonsData.find((addon) => addon.uuid === uuid))
        .filter((addon) => addon?.meta.type !== "os")
        .map((addon) => addon?.meta.key),
      installation_fee: product.installationFee,
      os: product.addons
        ?.map((uuid) => addonsData.find((addon) => addon.uuid === uuid))
        .filter((addon) => addon?.meta.type === "os")
        .map((addon) => addon?.meta.key),
      name: product.title,
      apiName,
      id: key,
    };
  });

  const newImagesArray = [];
  addonsData
    .filter((addon) => addon.meta?.type == "os")
    .forEach((addon) =>
      newImagesArray.push(
        ...Object.keys(addon.periods).map((period) => {
          const duration = periodsDurationMap[period];
          const { key, basePrices, tariff } = addon.meta;
          return {
            ...addon,
            duration,
            price: { value: basePrices[period] },
            value: addon.periods[period],
            name: addon.title,
            tariff,
            id: key,
            uuid: addon.uuid,
          };
        })
      )
    );

  images.value = newImagesArray;
};

watch(
  () => plans.value,
  () => {
    setGroups();
  },
  { deep: true }
);

watch(tabsIndex, () => {
  searchParam.value = "";
});

onMounted(() => {
  emit("changeFee", props.template.fee);

  refreshPlans();

  planAddons.value = [...(props.template.addons || [])];

  initializeData();
});

defineExpose({
  testConfig,
  changePlan,
  setFee,
});
</script>

<style>
.v-card .v-icon.group-icon {
  display: none;
  margin: 0 0 1px 2px;
  font-size: 18px;
  opacity: 1;
  cursor: pointer;
  color: #fff;
}

.v-data-table__expanded__content {
  background: var(--v-background-base);
}

.os-tab__card {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  background: var(--v-background-base);
  padding-bottom: 16px;
}

.os-tab__card .v-chip {
  color: #fff !important;
}

.os-tab__card .v-chip .v-icon {
  color: #fff;
}

@media (max-width: 1200px) {
  .os-tab__card {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 1000px) {
  .os-tab__card {
    grid-template-columns: 1fr;
  }
}
</style>
