<template>
  <div>
    <v-tabs background-color="background-light" v-model="tab">
      <v-tab v-for="tabKey in tabItems" :key="tabKey">
        {{ tabKey[0].toUpperCase() + tabKey.slice(1) }}</v-tab
      >
    </v-tabs>

    <v-tabs-items v-model="tab">
      <v-tab-item key="tariffs">
        <nocloud-table
          sort-by="enabled"
          sort-desc
          item-key="id"
          table-name="vps-tariffs"
          :show-select="false"
          :loading="isTariffsLoading"
          :headers="tariffsHeaders"
          :items="filteredTariffs"
          show-expand
          :filters-items="filterItems"
          :filters-values="filtersValues"
          @input:filter="filtersValues[$event.key] = $event.value"
          :expanded.sync="expandedTarrifs"
        >
          <template v-slot:[`item.endPrice`]="{ item }">
            <v-text-field v-model.number="item.endPrice" type="number" />
          </template>
          <template v-slot:[`item.duration`]="{ value }">
            <span>{{ readablePeriodMap[value] }}</span>
          </template>
          <template v-slot:[`item.enabled`]="{ item }">
            <v-switch
              :input-value="item.enabled"
              @change="item.enabled = $event"
            />
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
                :value="item.group"
                :items="groups"
                @change="changeGroup(item, $event)"
              />
              <v-icon @click="changeMode('create', item)">mdi-plus</v-icon>
              <v-icon @click="changeMode('edit', item)">mdi-pencil</v-icon>
              <v-icon v-if="groups.length > 1" @click="deleteGroup(item.group)"
                >mdi-delete</v-icon
              >
            </template>

            <template v-else-if="planId !== item.id">{{ item.group }}</template>
          </template>

          <template v-slot:expanded-item="{ headers, item }">
            <template v-if="item.windows">
              <td></td>
              <td :colspan="headers.length - 4">{{ item.windows.name }}</td>
              <td>
                {{ item.windows.price.value }}
                {{ defaultCurrency }}
              </td>
              <td>
                <v-text-field
                  dense
                  style="width: 150px"
                  v-model="item.windows.value"
                />
              </td>
              <td></td>
            </template>
            <template v-else>
              <td></td>
              <td :colspan="headers.length - 1">{{ $t("Windows is none") }}</td>
            </template>
          </template>
        </nocloud-table>
      </v-tab-item>
      <v-tab-item key="addons">
        <nocloud-table
          sort-by="enabled"
          sort-desc
          item-key="id"
          table-name="vps-addons"
          :show-select="false"
          :loading="isTariffsLoading"
          :headers="addonsHeaders"
          :items="addons"
        >
          <template v-slot:[`item.duration`]="{ value }">
            <span>{{ readablePeriodMap[value] }}</span>
          </template>
          <template v-slot:[`item.name`]="{ item }">
            <v-text-field style="min-width: 200px" v-model="item.name" />
          </template>
          <template v-slot:[`item.endPrice`]="{ item }">
            <v-text-field v-model.number="item.endPrice" type="number" />
          </template>
          <template v-slot:[`item.enabled`]="{ item }">
            <v-switch
              :input-value="item.enabled"
              @change="item.enabled = $event"
            />
          </template>
        </nocloud-table>
      </v-tab-item>
      <v-tab-item key="images">
        <nocloud-table
          sort-by="enabled"
          sort-desc
          item-key="id"
          table-name="vps-images"
          :show-select="false"
          :loading="isTariffsLoading"
          :headers="imagesHeaders"
          :items="images"
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
import { onMounted, toRefs, ref, computed } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import NocloudTable from "@/components/table.vue";
import usePlnRate from "@/hooks/usePlnRate";

const props = defineProps({
  fee: { type: Object, required: true },
  template: { type: Object, required: true },
  isPlansLoading: { type: Boolean, required: true },
  getPeriod: { type: Function, required: true },
  sp: { type: Object, required: true },
});
const { sp } = toRefs(props);

const rate = usePlnRate();
const store = useStore();

//constants
const tabItems = ["tariffs", "addons", "images"];
const allowedDuration = ["P1M", "P1Y"];
const allowedPricingModel = ["default", "upfront12"];
const readablePeriodMap = { P1M: "monthly", P1Y: "yearly" };

const tab = ref(0);
const expandedTarrifs = ref([]);
const isTariffsLoading = ref(false);
const fetchError = ref("");
const filtersValues = ref({ group: [], duration: [], enabled: [] });

const groups = ref([]);
const newGroupName = ref("");
const mode = ref("none");
const planId = ref("");

const fullTariffs = ref({});
const tariffs = ref([]);
const tariffsHeaders = ref([
  { text: "Name", value: "name" },
  { text: "API name", value: "apiName" },
  { text: "Group", value: "group", customFilter: true },
  {
    text: "Payment",
    value: "duration",
    customFilter: true,
  },
  { text: "Income price", value: "price" },
  { text: "Sale price", value: "endPrice" },
  {
    text: "Enabled",
    value: "enabled",
  },
]);

const filterItems = computed(() => {
  return {
    group: groups.value,
    duration: ["P1M", "P1Y"],
  };
});

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const filteredTariffs = computed(() => {
  return tariffs.value.filter((t) => {
    return Object.keys(filtersValues.value).every(
      (key) =>
        !filtersValues.value[key]?.length ||
        filtersValues.value[key].includes(t[key])
    );
  });
});

const addons = ref([]);
const addonsHeaders = ref([
  { text: "Addon", value: "name" },
  {
    text: "Payment",
    value: "duration",
  },
  { text: "Income price", value: "price" },
  { text: "Sale price", value: "endPrice" },
  {
    text: "Enabled",
    value: "enabled",
  },
]);

const images = ref([]);
const imagesHeaders = ref([
  {
    text: "Name",
    value: "title",
  },
  {
    text: "Enabled",
    value: "enabled",
  },
]);

onMounted(() => {
  isTariffsLoading.value = true;
  api.servicesProviders
    .action({ action: "get_plans", uuid: sp.value.uuid })
    .then(({ meta }) => {
      setAddons(meta);
      setTariffs(meta);
      setImages(meta);
      fullTariffs.value = meta.catalog.plans;
      fetchError.value = "";
    })
    .catch((err) => {
      fetchError.value = err.response?.data?.message ?? err.message ?? err;
      console.error(err);
    })
    .finally(() => {
      isTariffsLoading.value = false;
    });
});

const setAddons = (meta) => {
  const addonsKeys = ["backup", "disk", "snapshot"];
  const newAddons = [];

  addonsKeys.forEach((key) => {
    meta[key].forEach(({ prices, planCode, productName: name }) => {
      const filterPrices = prices.filter(
        ({ duration, pricingMode }) =>
          allowedDuration.includes(duration) &&
          allowedPricingModel.includes(pricingMode)
      );

      filterPrices.forEach(({ price: { value }, duration }) => {
        const price = (value * rate.value).toFixed(2);

        newAddons.push({
          price: price,
          endPrice: price,
          name,
          planCode,
          id: [planCode, duration].join(" "),
          duration,
        });
      });
    });
  });

  addons.value = newAddons;
};

const setTariffs = (meta) => {
  const newTarrifs = [];
  meta.plans.forEach((p) => {
    const prices = p.prices.filter(
      ({ duration, pricingMode }) =>
        allowedDuration.includes(duration) &&
        allowedPricingModel.includes(pricingMode)
    );

    prices.forEach((i) => {
      const code = p.planCode.split("-").slice(1).join("-");
      const option = meta.windows.find((el) => el.planCode.includes(code));

      let windows = null;
      if (option) {
        const {
          price: { value },
        } = option.prices.find(
          (el) => el.duration === i.duration && el.pricingMode === i.pricingMode
        );
        const newPrice = (value * rate.value).toFixed(2);

        windows = {
          value,
          price: { value: newPrice },
          name: option.productName,
          code: option.planCode,
        };
      }

      const price = (i.price.value * rate.value).toFixed(2);
      newTarrifs.push({
        apiName: p.productName,
        name: p.productName,
        planCode: p.planCode,
        endPrice: price,
        id: [i.duration, p.planCode].join(" "),
        price,
        windows,
        duration: i.duration,
        enabled: false,
        group: p.productName.replace(/VPS[\W0-9]/, "").split(/[\W0-9]/)[0],
      });
    });
  });

  //sort by cpu and ram
  newTarrifs.sort((a, b) => {
    const resA = a.planCode.split("-");
    const resB = b.planCode.split("-");

    const isCpuEqual = resB.at(-3) === resA.at(-3);
    const isRamEqual = resB.at(-2) === resA.at(-2);

    if (isCpuEqual && isRamEqual) return resA.at(-1) - resB.at(-1);
    if (isCpuEqual) return resA.at(-2) - resB.at(-2);
    return resA.at(-3) - resB.at(-3);
  });

  const newGroups = new Map();
  newTarrifs.forEach((t) => newGroups.set(t.group, t.group));

  groups.value = [...newGroups.keys()];

  tariffs.value = newTarrifs;
};

const setImages = (meta) => {
  const newImages = new Map();
  meta.catalog.plans.forEach((p) => {
    p.configurations[1].values.forEach((os) => {
      newImages.set(os, { id: os, title: os, enabled: false });
    });
  });

  images.value = [...newImages.values()];
};

const editGroup = (group) => {
  const i = groups.value.indexOf(group);
  groups.value.splice(i, 1, newGroupName.value);
  tariffs.value.forEach((tariff) => {
    if (tariff.group !== group) return;

    changeGroup(tariff, newGroupName.value);
  });

  changeMode("none", { id: -1, group: "" });
};
const createGroup = (tariff) => {
  groups.value.push(newGroupName.value);
  changeGroup(tariff, newGroupName.value);
  changeMode("none", { id: -1, group: "" });
};
const deleteGroup = (group) => {
  groups.value = groups.value.filter((el) => el !== group);
  tariffs.value.forEach((tariff) => {
    if (tariff.group !== group) return;
    changeGroup(tariff, groups.value[0]);
  });
};
const changeMode = (newMode, { id, group }) => {
  mode.value = newMode;
  planId.value = id;
  newGroupName.value = group;
};

const changeGroup = (item, newGroup) => {
  item.name = item.apiName.replace(item.group, newGroup);
  item.group = newGroup;
};
</script>
