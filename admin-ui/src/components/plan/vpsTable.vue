<template>
  <div>
    <v-row class="my-4" v-if="!isPlansLoading" align="center">
      <v-btn class="ml-3" @click="refreshPlans" :loading="isRefreshLoading"
        >Fetch plans</v-btn
      >
      <v-btn class="ml-3" :disabled="!newPlans" @click="setRefreshedPlans"
        >Set api plans</v-btn
      >
      <v-btn class="ml-3" @click="setSellToTab(true)">Enable all</v-btn>
      <v-btn class="ml-3" @click="setSellToTab(false)">Disable all</v-btn>
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
          :items="plans"
          :headers="headers"
          :expanded.sync="expanded"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.name`]="{ item }">
            <v-text-field dense style="width: 200px" v-model="item.name" />
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
          :items="addons"
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
          :items="images"
          :headers="imagesHeaders"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.tariff`]="{ value }">
            {{ value || "Any" }}
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

<script>
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

export default {
  name: "vps-table",
  components: { nocloudTable, planAddonsTable },
  props: {
    fee: { type: Object, required: true },
    template: { type: Object, required: true },
    isPlansLoading: { type: Boolean, required: true },
    getPeriod: { type: Function, required: true },
    sp: { type: Object, required: true },
  },
  data: () => ({
    groups: [],
    expanded: [],
    tabs: ["Tariffs", "Addons", "OS", "Custom addons"],

    plans: [],
    headers: [
      { text: "", value: "data-table-expand" },
      { text: "Title", value: "name" },
      { text: "API title", value: "apiName" },
      { text: "Group", value: "group" },
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
    ],

    addons: [],
    addonsHeaders: [
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
    ],

    images: [],
    imagesHeaders: [
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
    ],

    column: "",
    fetchError: "",
    newGroupName: "",
    mode: "none",

    planId: -1,
    tabsIndex: 0,
    usedFee: {},

    newPlans: null,
    newAddons: null,
    newImages: null,
    isRefreshLoading: false,

    planAddons: [],
  }),
  setup() {
    const { convertFrom } = useCurrency();

    return { convertFrom };
  },
  methods: {
    testConfig() {
      if (!this.plans.every(({ group }) => this.groups.includes(group))) {
        return "You must select a group for the tariff!";
      }
    },
    async changePlan(plan) {
      plan.resources = [];
      plan.products = {};

      let allAddons = [];
      let addons = [];

      this.addons.forEach((addon) => {
        const addonkey = addon.id.split(" ")[1];
        const existedAddon = addons.find(
          (existed) => existed.meta.key === addonkey
        );

        let data;
        if (existedAddon) {
          existedAddon.periods[this.getPeriod(addon.duration)] = addon.value;
          existedAddon.meta.basePrices[this.getPeriod(addon.duration)] =
            addon.price.value;
          data = existedAddon;
          return;
        } else {
          data = {
            system: true,
            title: addon.name,
            group: this.template.uuid,
            periods: { [this.getPeriod(addon.duration)]: addon.value },
            public: !!addon.public,
            kind: "PREPAID",
            meta: {
              basePrices: {
                [this.getPeriod(addon.duration)]: addon.price.value,
              },
              key: addonkey,
              type: "addon",
            },
          };
        }

        if (addon.uuid) {
          data.uuid = addon.uuid;
          addons.push({ ...data, type: "update" });
        } else {
          addons.push({ ...data, type: "create" });
        }
      });

      this.images.forEach((addon) => {
        const existedAddon = addons.find(
          (existed) =>
            existed.meta.type === "os" &&
            existed.meta.key === addon.id &&
            existed.meta.tariff === addon.tariff
        );

        let data;
        if (existedAddon) {
          existedAddon.periods[this.getPeriod(addon.duration)] = addon.value;
          existedAddon.meta.basePrices[this.getPeriod(addon.duration)] =
            addon.price.value;
          data = existedAddon;
          return;
        } else {
          data = {
            system: true,
            title: addon.name,
            group: this.template.uuid,
            periods: { [this.getPeriod(addon.duration)]: addon.value },
            public: !!addon.public,
            kind: "PREPAID",
            meta: {
              basePrices: {
                [this.getPeriod(addon.duration)]: addon.price.value,
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
          addons.push({ ...data, type: "update" });
        } else {
          addons.push({ ...data, type: "create" });
        }
      });

      const addonsForCreate = addons
        .filter((a) => a.type === "create")
        .map((a) => {
          delete a.type;
          return replaceNullWithUndefined(a);
        });

      const addonsForUpdate = addons
        .filter((a) => a.type === "update")
        .map((a) => {
          delete a.type;
          return replaceNullWithUndefined(a);
        });

      if (addonsForCreate.length) {
        const createdAddons = await this.addonsClient.createBulk({
          addons: addonsForCreate.map((addon) => Addon.fromJson(addon)),
        });

        allAddons.push(...createdAddons.toJson().addons);
      }

      if (addonsForUpdate.length) {
        const updatedAddons = await this.addonsClient.updateBulk({
          addons: addonsForUpdate.map((addon) => Addon.fromJson(addon)),
        });
        allAddons.push(...updatedAddons.toJson().addons);
      }

      this.plans.forEach((el) => {
        const [, , cpu, ram, disk] = el.planCode.split("-");
        const meta = {
          datacenter: el.datacenter,
        };

        const addons = el.addons
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

        plan.products[el.id] = {
          kind: "PREPAID",
          title: el.name,
          price: el.value,
          public: el.public,
          group: el.group,
          addons,
          period: this.getPeriod(el.duration),
          resources: { cpu: +cpu, ram: ram * 1024, drive_size: disk * 1024 },
          meta: {
            ...meta,
            basePrice: el.price.value,
            apiName: el.apiName,
          },
          sorter: Object.keys(plan.products).length,
          installation_fee: el.installation_fee,
        };
      });

      plan.addons = this.planAddons;
    },
    changePlans({ plans, catalog }) {
      const result = [];
      console.log(plans, catalog);

      plans.forEach(({ prices, planCode, productName }) => {
        prices.forEach(({ pricingMode, price, duration }) => {
          const isMonthly = duration === "P1M" && pricingMode === "default";
          const isYearly = duration === "P1Y" && pricingMode === "upfront12";

          if (isMonthly || isYearly) {
            const id = `${duration} ${planCode}`;
            const realProduct = this.plans.find((p) => p.id === id) || {};

            const code = planCode.split("-").slice(1).join("-");
            const newPrice = this.convertPrice(price.value);

            const { configurations, addonFamilies } = catalog.plans.find(
              ({ planCode }) => planCode.includes(code)
            );
            const os = configurations.find((c) => c.name === "vps_os")?.values;
            const datacenter = configurations.find(
              (c) => c.name === "vps_datacenter"
            )?.values;

            const addons = addonFamilies.reduce(
              (res, { addons }) => [...res, ...addons],
              []
            );
            const plan = { addons, os, datacenter };

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
        const resA = a.planCode.split("-");
        const resB = b.planCode.split("-");

        const isCpuEqual = resB.at(-3) === resA.at(-3);
        const isRamEqual = resB.at(-2) === resA.at(-2);

        if (isCpuEqual && isRamEqual) return resA.at(-1) - resB.at(-1);
        if (isCpuEqual) return resA.at(-2) - resB.at(-2);
        return resA.at(-3) - resB.at(-3);
      });

      return result;
    },
    changeAddons({ backup, disk, snapshot }) {
      const result = [];

      [backup, disk, snapshot].forEach((el) => {
        el.forEach(({ prices, planCode, productName }) => {
          prices.forEach(({ pricingMode, price, duration }) => {
            const isMonthly = duration === "P1M" && pricingMode === "default";
            const isYearly = duration === "P1Y" && pricingMode === "upfront12";

            if (isMonthly || isYearly) {
              const id = `${duration} ${planCode}`;
              const realAddon = this.addons.find((a) => a.id === id) || {};

              const newPrice = this.convertPrice(price.value);

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
    },
    changeImages({ windows }) {
      const newImages = [];

      this.newPlans.forEach((plan) =>
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
              const newPrice = this.convertPrice(priceObject.value);
              price = newPrice;
              basePrice = newPrice;
            }

            if (
              newImages.find(
                (os) =>
                  os.id === key &&
                  os.duration === duration &&
                  (!os.tariff || os.tariff === product?.planCode)
              )
            ) {
              return;
            }

            const realAddon = this.images.find(
              (a) =>
                a.id === key && tariff === a.tariff && a.duration === duration
            );

            newImages.push({
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

      return newImages;
    },
    setFee() {
      this.usedFee = JSON.parse(JSON.stringify(this.fee));

      [this.plans, this.addons, this.images].forEach((el) => {
        el.forEach((plan, i, arr) => {
          arr[i].value = getMarginedValue(this.fee, plan.price.value);
        });
      });
    },
    getPayment(duration) {
      switch (duration) {
        case "P1M":
          return "monthly";
        case "P1Y":
          return "yearly";
      }
    },
    editGroup(group) {
      const i = this.groups.indexOf(group);

      this.groups.splice(i, 1, this.newGroupName);
      this.plans.forEach((plan, index) => {
        if (plan.group !== group) return;
        this.plans[index].group = this.newGroupName;
      });

      this.changeMode("none", { id: -1, group: "" });
    },
    createGroup(plan) {
      this.groups.push(this.newGroupName);
      plan.group = this.newGroupName;

      this.changeMode("none", { id: -1, group: "" });
    },
    deleteGroup(group) {
      this.groups = this.groups.filter((el) => el !== group);
      this.plans.forEach((plan, i) => {
        if (plan.group !== group) return;
        this.plans[i].group = this.groups[0];
      });
    },
    changeMode(mode, { id, group }) {
      this.mode = mode;
      this.planId = id;
      this.newGroupName = group;
    },
    setSellToTab(status) {
      switch (this.tabs[this.tabsIndex]) {
        case "Addons": {
          this.setSellToValue(this.addons, status);
          break;
        }
        case "OS": {
          this.setSellToValue(this.images, status);
          break;
        }
        case "Tariffs": {
          this.setSellToValue(this.plans, status);
          break;
        }
      }
    },
    setSellToValue(value, status) {
      value = value.map((p) => {
        p.public = status;
        return p;
      });
    },
    convertPrice(price) {
      return this.convertFrom(price, { code: "PLN" });
    },
    async refreshPlans() {
      try {
        this.isRefreshLoading = true;
        const { meta } = await api.servicesProviders.action({
          action: "get_plans",
          uuid: this.sp.uuid,
        });

        this.newPlans = this.changePlans(meta);
        this.newAddons = this.changeAddons(meta);
        this.newImages = this.changeImages(meta);

        if (
          !this.plans?.length ||
          !this.addons?.length ||
          !this.images?.length
        ) {
          this.setRefreshedPlans();
        }
      } catch (err) {
        this.newPlans = null;
        this.newAddons = null;
        this.$store.commit("snackbar/showSnackbarError", {
          message: err.response?.data?.message ?? err.message ?? err,
        });
      } finally {
        this.isRefreshLoading = false;
      }
    },
    setRefreshedPlans() {
      this.addons = JSON.parse(JSON.stringify(this.newAddons));
      this.plans = JSON.parse(JSON.stringify(this.newPlans));
      this.images = JSON.parse(JSON.stringify(this.newImages));
      this.newPlans = null;
      this.newAddons = null;

      this.setGroups();
    },
    setGroups() {
      this.groups = [];
      this.plans.forEach((plan) => {
        const group = plan?.group || plan?.name?.split(/[\W0-9]/)[0];
        if (!this.groups.includes(group)) this.groups.push(group);
      });
    },
  },
  async created() {
    const periodsDurationMap = { 31536000: "P1Y", 2592000: "P1M" };

    const { addons = [] } = (
      await this.addonsClient.list(
        ListAddonsRequest.fromJson({ filters: { group: [this.template.uuid] } })
      )
    ).toJson();

    const newAddons = [];
    addons
      .filter((addon) => addon.meta?.type != "os")
      .forEach((addon) =>
        newAddons.push(
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
    this.addons = newAddons;

    this.plans = Object.keys(this.template.products || {}).map((key) => {
      const [duration, planCode] = key.split(" ");
      const product = this.template.products[key];

      const { meta } = product;

      const { apiName, datacenter, basePrice } = meta;

      return {
        ...product,
        duration,
        planCode,
        price: { value: basePrice },
        value: product.price,
        datacenter,
        addons: product.addons
          ?.map((uuid) => addons.find((addon) => addon.uuid === uuid))
          .filter((addon) => addon?.meta.type !== "os")
          .map((addon) => addon?.meta.key),
        installation_fee: product.installationFee,
        os: product.addons
          ?.map((uuid) => addons.find((addon) => addon.uuid === uuid))
          .filter((addon) => addon?.meta.type === "os")
          .map((addon) => addon?.meta.key),
        name: product.title,
        apiName,
        id: key,
      };
    });

    const newImages = [];
    addons
      .filter((addon) => addon.meta?.type == "os")
      .forEach((addon) =>
        newImages.push(
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

    this.images = newImages;
  },
  mounted() {
    this.$emit("changeFee", this.template.fee);

    this.refreshPlans();

    this.planAddons = [...this.template.addons];
  },
  computed: {
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    addonsClient() {
      return this.$store.getters["addons/addonsClient"];
    },
  },
  watch: {
    plans() {
      this.setGroups();
    },
  },
};
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
