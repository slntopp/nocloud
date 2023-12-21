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
          table-name="vps-tarrifs"
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

          <template v-slot:[`item.addons`]="{ item }">
            <v-dialog width="90vw">
              <template v-slot:activator="{ on, attrs }">
                <v-btn icon v-bind="attrs" v-on="on">
                  <v-icon> mdi-menu-open </v-icon>
                </v-btn>
              </template>
              <nocloud-table
                table-name="vps-external-addons"
                :items="getExternalAddons(item)"
                :headers="externalAddonsHeaders"
                :show-select="false"
              >
                <template v-slot:[`item.sell`]="{ item: addon }">
                  <v-switch
                    :input-value="addon.sell"
                    @change="changeExternalAddonSell(item, addon, $event)"
                  />
                </template>

                <template v-slot:[`item.period`]="{ value }">
                  {{getBillingPeriod(value)}}
                </template>
              </nocloud-table>
            </v-dialog>
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
            {{ value }} {{ defaultCurrency }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field
              dense
              style="width: 200px"
              :suffix="defaultCurrency"
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

        <nocloud-table
          table-name="vps-addons"
          v-else-if="tab === 'Addons'"
          v-model="selectedAddons"
          :items="addons"
          item-key="id"
          :headers="addonsHeaders"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:top>
            <v-toolbar flat color="background">
              <v-toolbar-title>Actions</v-toolbar-title>
              <v-divider inset vertical class="mx-4" />
              <v-spacer />

              <v-btn class="mr-2" color="background-light" @click="addAddon">
                Create
              </v-btn>
              <v-btn
                @click="removeAddons"
                color="background-light"
                :disabled="selectedAddons.length < 1"
                >Delete</v-btn
              >
            </v-toolbar>
          </template>
          <template v-slot:[`item.duration`]="{ item }">
            <date-field
              v-if="item.virtual"
              :period="item.fullDate"
              @changeDate="item.period = getTimestamp($event.value)"
            />
            <template v-else>
              {{ getPayment(item.duration) }}
            </template>
          </template>
          <template v-slot:[`item.price.value`]="{ item }">
            {{ item.price?.value }}
            {{ item.price?.value ? defaultCurrency : "" }}
          </template>
          <template v-slot:[`item.key`]="{ item }">
            <v-text-field
              dense
              :disabled="!item.virtual"
              style="width: 200px"
              v-model="item.key"
            />
          </template>
          <template v-slot:[`item.name`]="{ item }">
            <v-text-field dense style="width: 200px" v-model="item.name" />
          </template>
          <template v-slot:[`item.type`]="{ item }">
            {{ item.virtual ? "External" : "Internal" }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field
              dense
              style="width: 200px"
              :suffix="defaultCurrency"
              v-model="item.value"
            />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model="item.public" />
          </template>
          <template v-slot:[`item.min`]="{ item }">
            <v-text-field dense type="number" v-model="item.min" />
          </template>
          <template v-slot:[`item.max`]="{ item }">
            <v-text-field dense type="number" v-model="item.max" />
          </template>
          <template v-slot:[`item.kind`]="{ item }">
            <v-radio-group
              :disabled="!item.virtual"
              row
              mandatory
              v-model="item.kind"
            >
              <v-radio
                v-for="kind of ['PREPAID', 'POSTPAID']"
                :key="kind"
                :value="kind"
                :label="kind.toLowerCase()"
              />
            </v-radio-group>
          </template>
          <template v-slot:[`item.autoEnable`]="{ item }">
            <v-switch v-model="item.autoEnable" />
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
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import api from "@/api.js";
import nocloudTable from "@/components/table.vue";
import {getBillingPeriod, getFullDate, getMarginedValue, getTimestamp} from "@/functions";
import DateField from "@/components/date.vue";

export default {
  name: "vps-table",
  components: { DateField, nocloudTable },
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
    tabs: ["Tariffs", "Addons", "OS"],
    images: [],

    plans: [],
    headers: [
      { text: "", value: "data-table-expand" },
      { text: "Name", value: "name" },
      { text: "API name", value: "apiName" },
      { text: "Group", value: "group" },
      {
        text: "Payment",
        value: "duration",
      },
      { text: "Income price", value: "price.value" },
      { text: "Sale price", value: "value" },
      { text: "Addons", value: "addons" },
      {
        text: "Sell",
        value: "sell",
        width: 100,
      },
    ],

    selectedAddons: [],
    addons: [],
    addonsHeaders: [
      { text: "Key", value: "key" },
      { text: "Addon", value: "name" },
      { text: "Api name", value: "apiName" },
      {
        text: "Payment",
        value: "duration",
      },
      { text: "Income price", value: "price.value" },
      { text: "Sale price", value: "value" },
      { text: "Kind", value: "kind" },
      { text: "Type", value: "type" },
      { text: "Min count", value: "min" },
      { text: "Max count", value: "max" },
      { text: "Auto enable", value: "autoEnable" },
      {
        text: "Sell",
        value: "sell",
        width: 100,
      },
    ],

    externalAddonsHeaders: [
      { text: "Name", value: "name" },
      {
        text: "Payment",
        value: "period",
      },
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
  }),
  methods: {
    getBillingPeriod,
    getTimestamp,
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
    changePlan(plan) {
      this.plans.forEach((el) => {
        const [, , cpu, ram, disk] = el.planCode.split("-");
        const meta = {
          addons: el.addons,
          datacenter: el.datacenter,
          os: el.os.filter((item) => this.images.includes(item)),
          hidedOs: el.os.filter((item) => !this.images.includes(item)),
        };

        if (el.windows) meta.windows = el.windows.value;
        plan.products[el.id] = {
          kind: "PREPAID",
          title: el.name,
          price: el.value,
          public: el.public,
          group: el.group,
          period: this.getPeriod(el.duration),
          resources: { cpu: +cpu, ram: ram * 1024, disk: disk * 1024 },
          meta: {
            ...meta,
            basePrice: el.price.value,
            apiName: el.apiName,
          },
          sorter: Object.keys(plan.products).length,
          installation_fee: el.installation_fee,
        };
      });

      this.addons.forEach((el) => {
        plan.resources.push({
          key: el.virtual ? el.key : el.id,
          public: el.public,
          kind: el.kind,
          min: el.min || undefined,
          max: el.max || undefined,
          title: el.name,
          price: el.value,
          virtual: el.virtual,
          period: el.virtual ? el.period : this.getPeriod(el.duration),
          except: false,
          meta: {
            basePrice: el.price?.value,
            autoEnable: el.autoEnable,
            apiName: el.apiName,
          },
          on: [],
        });
      });
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
                kind: "PREPAID",
                autoEnable: false,
                min: undefined,
                max: undefined,
                price: { value: newPrice },
                duration,
                apiName: productName,
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
      this.addCustomAddonsToPlan();
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
    addAddon() {
      this.addons.push({
        key: "",
        kind: "POSTPAID",
        value: 0,
        period: 0,
        virtual: true,
        public: true,
        max: undefined,
        min: undefined,
        id: Math.random().toString(16).slice(2),
      });
    },
    removeAddons() {
      if (this.selectedAddons.some((a) => !a.virtual)) {
        return this.$store.commit("snackbar/showSnackbarError", {
          message: "You cant delete internal plans",
        });
      }
      this.addons = this.addons.filter(
        (a) => !this.selectedAddons.find((sa) => sa.key === a.key)
      );
    },
    addCustomAddonsToPlan() {
      this.addons.push(
        ...this.template.resources
          .filter((r) => r.virtual)
          .map((r) => ({
            ...r,
            fullDate: getFullDate(r.period),
            autoEnable: r.meta.autoEnable,
            value: r.price,
            id: r.key,
            name: r.title,
          }))
      );
    },
    getExternalAddons(item) {
      return this.addons
        .filter((a) => a.virtual && a.public)
        .map((a) => ({ ...a, sell: item.addons.includes(a.id) }));
    },
    changeExternalAddonSell(item, addon, value) {
      if (value) {
        item.addons.push(addon.id);
      } else {
        item.addons = item.addons.filter((a) => a !== addon.id);
      }
    },
  },
  created() {
    const newImages = [];
    this.plans = Object.keys(this.template.products || {}).map((key) => {
      const [duration, planCode] = key.split(" ");
      const product = this.template.products[key];

      const { meta } = product;

      const enabledOs = meta.os || [];
      newImages.push(...enabledOs);
      const os = enabledOs.concat(...(meta.hidedOs || []));

      const { apiName, addons, datacenter, basePrice } = meta;

      return {
        ...product,
        duration,
        planCode,
        price: { value: basePrice },
        value: product.price,
        datacenter,
        addons,
        installation_fee: product.installationFee,
        os,
        name: product.title,
        apiName,
        id: key,
      };
    });

    this.addons = this.template.resources
      .filter((r) => !r.virtual)
      .map((r) => {
        const { key } = r;
        const [duration, planCode] = key.split(" ");

        return {
          ...r,
          duration,
          planCode,
          price: { value: r.meta.basePrice },
          value: r.price,
          name: r.title,
          id: key,
        };
      });

    this.addCustomAddonsToPlan();
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
      if (this.defaultCurrency === "PLN") {
        return 1;
      }
      return this.rates.find(
        (r) => r.to === this.defaultCurrency && r.from === "PLN"
      )?.rate;
    },
    allImages() {
      const imagesSet = new Set();
      this.plans?.forEach((p) => {
        p.os?.forEach((os) => imagesSet.add(os));
      });

      const imagesArr = [...imagesSet.values()];
      imagesArr.sort();
      return imagesArr;
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
