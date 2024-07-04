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
            {{ value }} {{ defaultCurrency?.title }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field
              dense
              style="width: 200px"
              :suffix="defaultCurrency?.title"
              v-model="item.value"
            />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model="item.public" />
          </template>
          <template v-slot:expanded-item="{ headers, item }">
            <template v-if="item.windows">
              <td></td>
              <td :colspan="headers.length - 4">{{ item.windows.name }}</td>
              <td>
                {{ item.windows.price.value }}
                {{ defaultCurrency?.title }}
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
            {{ value }} {{ defaultCurrency?.title }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field
              dense
              style="width: 200px"
              :suffix="defaultCurrency?.title"
              v-model="item.value"
            />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model="item.public" />
          </template>
        </nocloud-table>

        <div class="os-tab__card" v-else-if="tab === 'OS'">
          <template v-if="allImages.length">
            <v-card
              outlined
              class="pt-4 pl-4 d-flex"
              style="gap: 10px"
              color="background"
              v-for="item of allImages"
              :key="item"
            >
              <v-chip
                close
                :color="images.includes(item) ? 'info' : 'error'"
                :close-icon="
                  images.includes(item) ? 'mdi-close-circle' : 'mdi-plus-circle'
                "
                @click:close="changeImage(item)"
              >
                <span>
                  {{ item }}
                </span>
              </v-chip>
            </v-card>
          </template>
          <template v-else>
            <v-card
              outlined
              class="pt-4 pl-4 d-flex"
              style="gap: 10px"
              color="background"
            >
              <v-card-title>No os in selected tariffs</v-card-title>
            </v-card>
          </template>
        </div>

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
    images: [],

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
        value: "sell",
        width: 100,
      },
    ],

    addons: [],
    addonsHeaders: [
      { text: "Addon", value: "name" },
      { text: "Margin", value: "margin" },
      {
        text: "Payment",
        value: "duration",
      },
      { text: "Incoming price", value: "price.value" },
      { text: "Sale price", value: "value" },
      {
        text: "Sell",
        value: "sell",
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
    isRefreshLoading: false,

    planAddons: [],
  }),
  methods: {
    changeImage(value) {
      const i = this.images.indexOf(value);

      if (i !== -1) this.images.splice(i, 1);
      else this.images.push(value);
    },
    testConfig() {
      if (!this.plans.every(({ group }) => this.groups.includes(group))) {
        return "You must select a group for the tariff!";
      }
    },
    async changePlan(plan) {
      plan.resources = [];
      plan.products = {};

      const addonsForCreate = [];
      const addonsForUpdate = [];
      let allAddons = [];

      this.addons.forEach((addon) => {
        const data = Addon.fromJson({
          system: true,
          title: addon.name,
          group: this.template.uuid,
          periods: { [this.getPeriod(addon.duration)]: addon.value },
          public: addon.public,
          kind: "PREPAID",
          meta: {
            basePrice: addon.price.value,
            key: addon.id,
          },
        });
        if (addon.uuid) {
          data.uuid = addon.uuid;
          addonsForUpdate.push(data);
        } else {
          addonsForCreate.push(data);
        }
      });

      if (addonsForCreate.length) {
        const createdAddons = await this.addonsClient.createBulk({
          addons: addonsForCreate,
        });

        allAddons.push(...createdAddons.toJson().addons);
      }

      if (addonsForUpdate.length) {
        const updatedAddons = await this.addonsClient.updateBulk({
          addons: addonsForUpdate,
        });
        allAddons.push(...updatedAddons.toJson().addons);
      }

      this.plans.forEach((el) => {
        const [, , cpu, ram, disk] = el.planCode.split("-");
        const meta = {
          datacenter: el.datacenter,
          os: el.os.filter((item) => this.images.includes(item)),
          hidedOs: el.os.filter((item) => !this.images.includes(item)),
        };
        const addons = el.addons
          .map(
            (key) =>
              allAddons.find(
                (addon) => [el.duration, key].join(" ") === addon.meta.key
              )?.uuid
          )
          .filter((a) => !!a);

        if (el.windows) meta.windows = el.windows.value;
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
    changePlans({ plans, windows, catalog }) {
      const result = [];

      plans.forEach(({ prices, planCode, productName }) => {
        prices.forEach(({ pricingMode, price, duration }) => {
          const isMonthly = duration === "P1M" && pricingMode === "default";
          const isYearly = duration === "P1Y" && pricingMode === "upfront12";

          if (isMonthly || isYearly) {
            const id = `${duration} ${planCode}`;
            const realProduct = this.plans.find((p) => p.id === id) || {};

            const code = planCode.split("-").slice(1).join("-");
            const option = windows.find((el) => el.planCode.includes(code));
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
            const plan = { windows: null, addons, os, datacenter };

            if (option) {
              const {
                price: { value },
              } = option.prices.find(
                (el) =>
                  el.duration === duration && el.pricingMode === pricingMode
              );
              const newPrice = this.convertPrice(value);

              plan.windows = {
                value,
                price: { value: newPrice },
                name: option.productName,
                code: option.planCode,
              };
            }

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
                value: realAddon.value || price.value,
                public: !!realAddon.public,
                id,
              });
            }
          });
        });
      });

      return result;
    },
    setFee() {
      const windows = [];

      this.usedFee = JSON.parse(JSON.stringify(this.fee));
      this.plans.forEach((el) => {
        if (el.windows) windows.push(el.windows);
      });

      [this.plans, this.addons, windows].forEach((el) => {
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
          this.images = [];
          if (status) {
            this.allImages.forEach((img) => {
              this.images.push(img);
            });
          }
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
      return (price * this.plnRate).toFixed(2);
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

        if (!this.plans?.length || !this.addons?.length) {
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
    const newImages = [];
    const { addons = [] } = (
      await this.addonsClient.list(
        ListAddonsRequest.fromJson({ filters: { group: [this.template.uuid] } })
      )
    ).toJson();

    this.addons = addons.map((addon) => {
      const { key, basePrice } = addon.meta;
      const [duration, planCode] = key.split(" ");
      return {
        ...addon,
        duration,
        planCode,
        price: { value: basePrice },
        value: Object.values(addon.periods)[0],
        name: addon.title,
        id: key,
        uuid: addon.uuid,
      };
    });

    this.plans = Object.keys(this.template.products || {}).map((key) => {
      const [duration, planCode] = key.split(" ");
      const product = this.template.products[key];

      const { meta } = product;

      const enabledOs = meta.os || [];
      newImages.push(...enabledOs);
      const os = enabledOs.concat(...(meta.hidedOs || []));

      const { apiName, datacenter, basePrice } = meta;

      return {
        ...product,
        duration,
        planCode,
        price: { value: basePrice },
        value: product.price,
        datacenter,
        addons: product.addons?.map(
          (uuid) =>
            this.addons
              .find((addon) => addon.uuid === uuid)
              ?.meta.key.split(" ")[1]
        ),
        installation_fee: product.installationFee,
        os,
        name: product.title,
        apiName,
        id: key,
      };
    });

    this.images = [...new Set(newImages)];
  },
  mounted() {
    this.$emit("changeFee", this.template.fee);

    this.refreshPlans();
  },
  computed: {
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    rates() {
      return this.$store.getters["currencies/rates"];
    },
    plnRate() {
      if (this.defaultCurrency?.title === "PLN") {
        return 1;
      }
      return this.rates.find(
        (r) =>
          r.to?.title === this.defaultCurrency?.title && r.from?.title === "PLN"
      )?.rate;
    },
    allImages() {
      const imagesSet = new Set();
      this.plans.forEach((p) => {
        p.os.forEach((os) => imagesSet.add(os));
      });

      const imagesArr = [...imagesSet.values()];
      imagesArr.sort();
      return imagesArr;
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
