<template>
  <div>
    <v-row class="my-4" v-if="!isPlansLoading" align="center">
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
          <template v-slot:[`item.margin`]="{ item }">
            {{ getMargin(item, false) }}
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
            <v-switch v-model="item.sell" />
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
          :show-select="false"
          :items="addons"
          :headers="addonsHeaders"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.margin`]="{ item }">
            {{ getMargin(item, false) }}
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
            <v-switch v-model="item.sell" />
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
import { getMarginedValue } from "@/functions";

export default {
  name: "vps-table",
  components: { nocloudTable },
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
      { text: "Margin", value: "margin" },
      {
        text: "Payment",
        value: "duration",
      },
      { text: "Income price", value: "price.value" },
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
      { text: "Income price", value: "price.value" },
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
    changePlan(plan) {
      this.plans.forEach((el) => {
        if (el.sell) {
          const [, , cpu, ram, disk] = el.planCode.split("-");
          const meta = {
            addons: el.addons,
            datacenter: el.datacenter,
            os: el.os.filter((item) => this.images.includes(item)),
          };

          if (el.windows) meta.windows = el.windows.value;
          plan.products[el.id] = {
            kind: "PREPAID",
            title: el.name,
            price: el.value,
            group: el.group,
            period: this.getPeriod(el.duration),
            resources: { cpu: +cpu, ram: ram * 1024, disk: disk * 1024 },
            sorter: Object.keys(plan.products).length,
            installation_fee: el.installation_fee,
            meta,
          };
        }
      });

      this.addons.forEach((el) => {
        if (el.sell) {
          plan.resources.push({
            key: el.id,
            kind: "PREPAID",
            title: el.name,
            price: el.value,
            period: this.getPeriod(el.duration),
            except: false,
            on: [],
          });
        }
      });
    },
    changeIcon() {
      setTimeout(() => {
        const headers = document.querySelectorAll(".groupable");

        headers.forEach(({ firstChild, childNodes }) => {
          if (!childNodes[1]?.className?.includes("group-icon")) {
            const element = document.querySelector(".group-icon");
            const icon = element.cloneNode(true);

            firstChild.after(icon);
            icon.style = "display: inline-flex";

            icon.addEventListener("click", () => {
              const menu = document.querySelector(".v-menu__content");
              const { x, y } = icon.getBoundingClientRect();

              if (menu.className.includes("menuable__content__active")) return;

              this.column = firstChild.textContent.trim();
              element.dispatchEvent(new Event("click"));

              setTimeout(() => {
                const width = document.documentElement.offsetWidth;
                const menuWidth = menu.offsetWidth;
                let marginLeft = 20;

                if (width < menuWidth + x)
                  marginLeft = width - (menuWidth + x) - 35;
                const marginTop = marginLeft < 20 ? 20 : 0;

                menu.style.left = `${x + marginLeft + window.scrollX}px`;
                menu.style.top = `${y + marginTop + window.scrollY}px`;
              }, 0);
            });
          }
        });
      }, 100);
    },
    changePlans({ plans, windows, catalog }) {
      const result = [];

      plans.forEach(({ prices, planCode, productName }) => {
        prices.forEach(({ pricingMode, price, duration }) => {
          const isMonthly = duration === "P1M" && pricingMode === "default";
          const isYearly = duration === "P1Y" && pricingMode === "upfront12";

          if (isMonthly || isYearly) {
            const code = planCode.split("-").slice(1).join("-");
            const option = windows.find((el) => el.planCode.includes(code));
            const newPrice = this.convertPrice(price.value);

            const { configurations, addonFamilies } = catalog.plans.find(
              ({ planCode }) => planCode.includes(code)
            )
            const os = configurations.find(c=>c.name==='vps_os')?.values;
            const datacenter = configurations.find(c=>c.name==='vps_datacenter')?.values;
           
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
              installation_fee: installation.price.value,
              price: { value: newPrice },
              name: productName,
              apiName: productName,
              group: productName.replace(/VPS[\W0-9]/, "").split(/[\W0-9]/)[0],
              value: price.value,
              sell: false,
              id: `${duration} ${planCode}`,
            });
          }
        });
      });
      this.plans = result;
      this.plans.sort((a, b) => {
        const resA = a.planCode.split("-");
        const resB = b.planCode.split("-");

        const isCpuEqual = resB.at(-3) === resA.at(-3);
        const isRamEqual = resB.at(-2) === resA.at(-2);

        if (isCpuEqual && isRamEqual) return resA.at(-1) - resB.at(-1);
        if (isCpuEqual) return resA.at(-2) - resB.at(-2);
        return resA.at(-3) - resB.at(-3);
      });
    },
    changeAddons({ backup, disk, snapshot }) {
      const result = [];

      [backup, disk, snapshot].forEach((el) => {
        el.forEach(({ prices, planCode, productName }) => {
          prices.forEach(({ pricingMode, price, duration }) => {
            const isMonthly = duration === "P1M" && pricingMode === "default";
            const isYearly = duration === "P1Y" && pricingMode === "upfront12";

            if (isMonthly || isYearly) {
              const newPrice = this.convertPrice(price.value);

              result.push({
                price: { value: newPrice },
                duration,
                name: productName,
                value: price.value,
                sell: false,
                id: `${duration} ${planCode}`,
              });
            }
          });
        });
      });

      this.addons = result;
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

          this.getMargin(arr[i]);
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
    getMargin({ value, price }) {
      if (!this.usedFee.ranges) {
        return "none";
      }

      const range = this.usedFee.ranges.find(
        ({ from, to }) => from <= price.value && to >= price.value
      );
      const n = Math.pow(10, this.usedFee.precision);
      let percent = range?.factor / 100 + 1;
      let margin;
      let round;

      switch (this.usedFee.round) {
        case 1:
          round = "floor";
          break;
        case 2:
          round = "round";
          break;
        case 3:
          round = "ceil";
      }
      if (this.usedFee.round === "NONE") round = "round";
      else if (typeof this.usedFee.round === "string") {
        round = this.usedFee.round.toLowerCase();
      }

      // value = Math[round](value * n) / n;

      if (value === Math[round](price.value * percent * n) / n) {
        margin = "ranged";
      } else if (this.usedFee.default <= 0) {
        margin = "none";
      } else {
        percent = this.usedFee.default / 100 + 1;
      }

      switch (value) {
        case Math[round](price.value * n) / n:
          margin = "none";
          break;
        case Math[round](price.value * percent * n) / n:
          if (!margin) margin = "fixed";
          break;
        default:
          margin = "manual";
      }

      return margin;
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
        p.sell = status;
        return p;
      });
    },
    convertPrice(price) {
      return (price * this.plnRate).toFixed(2);
    },
  },
  created() {
    this.$emit("changeLoading");

    api.servicesProviders
      .action({ action: "get_plans", uuid: this.sp.uuid })
      .then(({ meta }) => {
        this.changePlans(meta);
        this.changeAddons(meta);

        this.fetchError = "";
        this.changeIcon();
      })
      .catch((err) => {
        this.fetchError = err.response?.data?.message ?? err.message ?? err;
        console.error(err);
      })
      .finally(() => {
        this.$emit("changeLoading");
      });
  },
  mounted() {
    const icon = document.querySelector(".group-icon");

    icon.dispatchEvent(new Event("click"));
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
      this.plans
        .filter((p) => !!p.sell)
        .map((p) => {
          p.os.forEach((os) => imagesSet.add(os));
        });

      const imagesArr = [...imagesSet.values()];
      imagesArr.sort();
      return imagesArr;
    },
  },
  watch: {
    tabsIndex(value) {
      if (value > 1) return;
      this.changeIcon();

      const items = [this.plans, this.addons];

      items[value].forEach((el) => this.getMargin(el));
      this.$emit("changeFee", Object.assign({}, this.fee));
    },
    addons() {
      this.$emit("changeFee", this.template.fee ?? {});
      setTimeout(() => {
        this.setFee();

        this.template.resources.forEach(({ key, price }) => {
          const addon = this.addons.find((el) => el.id === key);

          addon.value = price;
          addon.sell = true;
        });

        this.groups = [];
        const imagesSet = new Set();
        this.plans.forEach((plan, i) => {
          const product = this.template.products[plan.id];
          const winKey = Object.keys(product?.meta || {}).find((el) =>
            el.includes("windows")
          );
          const title = (product?.title ?? plan.name).replace(/VPS[\W0-9]/, "");
          const group = product?.group || title.split(/[\W0-9]/)[0];

          if (product) {
            this.plans[i].name = product.title;
            this.plans[i].value = product.price;
            product.meta.os?.forEach((os) => imagesSet.add(os));

            this.plans[i].group = group;
            this.plans[i].sell = true;

            if (winKey) this.plans[i].windows.value = product.meta[winKey];
          }
          if (!this.groups.includes(group)) this.groups.push(group);
        });

        this.images = [...imagesSet.values()];
      });
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
