<template>
  <div>
    <v-menu :value="true" :close-on-content-click="false">
      <template v-slot:activator="{ on, attrs }">
        <v-icon class="group-icon" v-bind="attrs" v-on="on">mdi-filter</v-icon>
      </template>

      <v-list dense v-if="tabsIndex < 2">
        <v-list-item
          dense
          v-for="item of filters[tabsIndex][column]"
          :key="item"
        >
          <v-checkbox
            dense
            v-model="selected[tabsIndex][column]"
            :value="item"
            :label="item"
            @change="
              selected[tabsIndex] = Object.assign({}, selected[tabsIndex])
            "
          />
        </v-list-item>
      </v-list>
    </v-menu>
    <v-row class="mt-4" v-if="!isPlansLoading" align="center">
      <v-col cols="2">
        <v-btn @click="setEnableToAll(true)">Enable all</v-btn>
      </v-col>
      <v-col cols="2">
        <v-btn @click="setEnableToAll(false)">Disable all</v-btn>
      </v-col>
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
          :items="filteredPlans"
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
                @change="item.name = getName(item)"
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
            <v-text-field dense style="width: 150px" v-model="item.value" />
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
          :items="filteredAddons"
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
            <v-text-field dense style="width: 150px" v-model="item.value" />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model="item.sell" />
          </template>
        </nocloud-table>

        <div class="os-tab__card" v-else-if="tab === 'OS'">
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
              outlined
              :color="(images.includes(item)) ? 'info' : 'error'"
              :close-icon="(images.includes(item) ? 'mdi-close-circle' : 'mdi-plus-circle')"
              @click:close="changeImage(item)"
            >
              {{ item }}
            </v-chip>
          </v-card>
        </div>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import api from "@/api.js";
import nocloudTable from "@/components/table.vue";
import currencyRate from "@/mixins/currencyRate";
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
    allImages: [],
    images: [],

    plans: [],
    headers: [
      { text: "", value: "data-table-expand" },
      { text: "Name", value: "name" },
      { text: "API name", value: "apiName" },
      { text: "Group", value: "group", sortable: false, class: "groupable" },
      { text: "Margin", value: "margin", sortable: false, class: "groupable" },
      {
        text: "Payment",
        value: "duration",
        sortable: false,
        class: "groupable",
      },
      { text: "Income price", value: "price.value" },
      { text: "Sale price", value: "value" },
      {
        text: "Sell",
        value: "sell",
        sortable: false,
        class: "groupable",
        width: 100,
      },
    ],

    addons: [],
    addonsHeaders: [
      { text: "Addon", value: "name" },
      { text: "Margin", value: "margin", sortable: false, class: "groupable" },
      {
        text: "Payment",
        value: "duration",
        sortable: false,
        class: "groupable",
      },
      { text: "Income price", value: "price.value" },
      { text: "Sale price", value: "value" },
      {
        text: "Sell",
        value: "sell",
        sortable: false,
        class: "groupable",
        width: 100,
      },
    ],

    filters: {
      0: { Sell: ["true", "false"], Payment: ["monthly", "yearly"] },
      1: { Sell: ["true", "false"], Payment: ["monthly", "yearly"] },
    },
    selected: {
      0: { Sell: ["true", "false"], Payment: ["monthly", "yearly"] },
      1: { Sell: ["true", "false"], Payment: ["monthly", "yearly"] },
    },

    column: "",
    fetchError: "",
    newGroupName: "",
    mode: "none",

    planId: -1,
    tabsIndex: 0,
  }),
  mixins: [currencyRate],
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
            period: this.getPeriod(el.duration),
            resources: { cpu: +cpu, ram: ram * 1024, disk: disk * 1024 },
            sorter: Object.keys(plan.products).length,
            meta,
          };
        }
      });

      this.addons.forEach((el) => {
        if (el.sell) {
          plan.resources.push({
            key: el.id,
            kind: "PREPAID",
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

        headers.forEach(({ firstChild, children }) => {
          if (!children[1]?.className.includes("group-icon")) {
            const element = document.querySelector(".group-icon");
            const icon = element.cloneNode(true);

            firstChild.after(icon);
            icon.style = "display: inline-flex";

            icon.addEventListener("click", () => {
              const menu = document.querySelector(".v-menu__content");
              const { x, y } = icon.getBoundingClientRect();

              if (menu.className.includes("menuable__content__active")) return;

              this.column = firstChild.textContent.trim();
              if (this.column === "Group") {
                this.filters[this.tabsIndex].Group = this.groups;
                this.selected[this.tabsIndex].Group = this.groups;
              }

              this.filters[this.tabsIndex] = Object.assign(
                {},
                this.filters[this.tabsIndex]
              );
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
              }, 100);
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
            const newPrice = parseFloat((price.value * this.rate).toFixed(2));

            const { configurations, addonFamilies } = catalog.plans.find(
              ({ planCode }) => planCode.includes(code)
            );
            const os = configurations[1].values;
            const datacenter = configurations[0].values;
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
              const newPrice = parseFloat((value * this.rate).toFixed(2));

              plan.windows = {
                value,
                price: { value: newPrice },
                name: option.productName,
                code: option.planCode,
              };
            }

            result.push({
              ...plan,
              planCode,
              price: { value: newPrice },
              duration,
              name: productName,
              apiName: productName,
              group: productName.replace(/VPS[\W0-9]/, '').split(/[\W0-9]/)[0],
              value: price.value,
              sell: false,
              id: `${duration} ${planCode}`,
            });
          }
        });
      });
      this.allImages = result[1].os;
      this.allImages.sort();
      
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
              const newPrice = parseFloat((price.value * this.rate).toFixed(2));

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
    changeFilters(plan, filters) {
      filters.forEach((text) => {
        if (!this.filters[this.tabsIndex][text]) {
          this.filters[this.tabsIndex][text] = [];
        }
        if (!this.selected[this.tabsIndex][text]) {
          this.selected[this.tabsIndex][text] = [];
        }

        const { value } = this.headers.find((el) => el.text === text);
        const filter = `${plan[value]}`;

        if (!this.filters[this.tabsIndex][text].includes(filter)) {
          this.filters[this.tabsIndex][text].push(filter);
        }
        if (!this.selected[this.tabsIndex][text].includes(filter)) {
          this.selected[this.tabsIndex][text].push(filter);
        }
      });
    },
    setFee() {
      const windows = [];

      this.filters["0"].Margin = ["manual"];
      this.selected["0"].Margin = ["manual"];
      this.filters["1"].Margin = ["manual"];
      this.selected["1"].Margin = ["manual"];

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
    getName({ name, group }) {
      const newGroup = `${group[0].toUpperCase()}${group.slice(1)}`;
      const slicedName = name.replace(/VPS[\W0-9]/, '');
      const sep = /[\W0-9]/.exec(slicedName)[0];
      const newName = slicedName.split(sep).splice(1).join(sep);

      if (!name.startsWith('VPS')) return `${newGroup}${sep}${newName}`;
      else return `VPS ${newGroup} ${name.split(" ").at(-1)}`;
    },
    getMargin({ value, price }, filter = true) {
      if (!this.fee.ranges) {
        if (filter) this.changeFilters({ margin: "none" }, ["Margin"]);
        return "none";
      }
      const range = this.fee.ranges?.find(
        ({ from, to }) => from <= price.value && to >= price.value
      );
      const n = Math.pow(10, this.fee.precision ?? 0);
      let percent = range?.factor / 100 + 1;
      let round;

      switch (this.fee.round) {
        case 1:
          round = "floor";
          break;
        case 2:
          round = "round";
          break;
        case 3:
          round = "ceil";
      }
      if (!this.fee.round || this.fee.round === "NONE") round = "round";
      else if (typeof this.fee.round === "string") {
        round = this.fee.round.toLowerCase();
      }

      if (value === Math[round](price.value * percent * n) / n) {
        if (filter) this.changeFilters({ margin: "ranged" }, ["Margin"]);
        return "ranged";
      } else percent = (this.fee.default ?? 0) / 100 + 1;

      switch (value) {
        case price.value:
          if (filter) this.changeFilters({ margin: "none" }, ["Margin"]);
          return "none";
        case Math[round](price.value * percent * n) / n:
          if (filter) this.changeFilters({ margin: "fixed" }, ["Margin"]);
          return "fixed";
        default:
          if (filter) this.changeFilters({ margin: "manual" }, ["Margin"]);
          return "manual";
      }
    },
    editGroup(group) {
      const i = this.groups.indexOf(group);

      this.groups.splice(i, 1, this.newGroupName);
      this.plans.forEach((plan, index) => {
        if (plan.group !== group) return;
        this.plans[index].group = this.newGroupName;
        this.plans[index].name = this.getName(plan);
      });

      this.changeMode("none", { id: -1, group: "" });
    },
    createGroup(plan) {
      this.groups.push(this.newGroupName);
      plan.group = this.newGroupName;
      plan.name = this.getName(plan);

      this.changeMode("none", { id: -1, group: "" });
    },
    deleteGroup(group) {
      this.groups = this.groups.filter((el) => el !== group);
      this.plans.forEach((plan, i) => {
        if (plan.group !== group) return;
        this.plans[i].group = this.groups[0];
        this.plans[i].name = this.getName(plan);
      });
    },
    changeMode(mode, { id, group }) {
      this.mode = mode;
      this.planId = id;
      this.newGroupName = group;
    },
    applyFilter(values, i) {
      return values.filter((plan) => {
        const result = [];

        Object.entries(this.selected[i]).forEach(([key, filters]) => {
          const { value } = this.headers.find(({ text }) => text === key);
          let filter = `${plan[value]}`;

          switch (key) {
            case "Payment":
              filter = this.getPayment(plan[value]);
              break;
            case "Margin":
              filter = this.getMargin(plan, false);
          }

          if (filters.includes(filter)) result.push(true);
          else result.push(false);
        });

        return result.every((el) => el);
      });
    },
    setEnableToAll(status){
      this.plans=this.plans.map((p)=>{
        p.sell=status
        return p
      })
      this.addons.forEach((p,ind)=>{
        this.$set(this.addons,ind,{...p,sell:status})
      })
    }
  },
  created() {
    this.$emit("changeLoading");

    this.fetchRate();

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
    filteredPlans() {
      return this.applyFilter(this.plans, 0);
    },
    filteredAddons() {
      return this.applyFilter(this.addons, 1);
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
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
          addon.sell = false;
        });

        this.groups = [];
        this.plans.forEach((plan, i) => {
          const product = this.template.products[plan.id];
          const winKey = Object.keys(product?.meta || {}).find((el) =>
            el.includes("windows")
          );
          const title = (product?.title ?? plan.name).replace(/VPS[\W0-9]/, '');
          const group = title.split(/[\W0-9]/)[0];

          if (product) {
            this.plans[i].name = product.title;
            this.plans[i].value = product.price;
            this.plans[i].os = product.meta.os;
            this.plans[i].group = group;
            this.plans[i].sell = true;

            if (winKey) this.plans[i].windows.value = product.meta[winKey];
          }
          if (!this.groups.includes(group)) this.groups.push(group);
        });

        const sellingPlans = this.plans.filter(({ sell }) => sell);

        if (sellingPlans.length < 1) this.images = this.plans[1].os;
        else this.images = (sellingPlans[1] ?? sellingPlans[0]).os
          .filter((image) => this.allImages.includes(image));

        if (this.template.resources.length === this.addons.length) {
          this.filters["1"].Sell = ["true"];
          this.selected["1"].Sell = ["true"];
        }
        if (Object.keys(this.template.products).length === this.plans.length) {
          this.filters["0"].Sell = ["true"];
          this.selected["0"].Sell = ["true"];
        }
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
