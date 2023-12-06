<template>
  <div>
    <v-menu :value="true" :close-on-content-click="false">
      <template v-slot:activator="{ on, attrs }">
        <v-icon class="group-icon" v-bind="attrs" v-on="on">mdi-filter</v-icon>
      </template>

      <v-list dense>
        <v-list-item dense v-for="item of filters[column]" :key="item">
          <v-checkbox
            dense
            v-model="selected[column]"
            :value="item"
            :label="item"
            @change="selected = Object.assign({}, selected)"
          />
        </v-list-item>
      </v-list>
    </v-menu>
    <v-row class="my-3" v-if="!isPlansLoading">
      <v-btn
        class="ml-3"
        :loading="isAddonsLoading && setToAllValue === true"
        :disabled="isAddonsLoading && setToAllValue !== true"
        @click="setEnabledToTab(true)"
        >Enable all</v-btn
      >
      <v-btn
        class="ml-3"
        :loading="isAddonsLoading && setToAllValue === false"
        :disabled="isAddonsLoading && setToAllValue !== false"
        @click="setEnabledToTab(false)"
        >Disable all</v-btn
      >
    </v-row>
    <nocloud-table
      table-name="dedicated-prices"
      sort-by="isBeenSell"
      sort-desc
      item-key="id"
      :show-expand="true"
      :show-select="false"
      :items="filteredPlans"
      :headers="headers"
      :loading="isPlansLoading"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.name`]="{ item }">
        <v-text-field dense style="width: 200px" v-model="item.name" />
      </template>

      <template v-slot:[`item.group`]="{ item }">
        <template
          v-if="newGroup.mode === 'edit' && newGroup.planId === item.id"
        >
          <v-text-field
            dense
            class="d-inline-block mr-1"
            style="width: 200px"
            v-model="newGroup.name"
          />
          <v-icon @click="editGroup(item.group)">mdi-content-save</v-icon>
          <v-icon @click="newGroup.mode = 'none'">mdi-close</v-icon>
        </template>

        <template
          v-if="newGroup.mode === 'create' && newGroup.planId === item.id"
        >
          <v-text-field
            dense
            class="d-inline-block mr-1"
            style="width: 200px"
            v-model="newGroup.name"
          />
          <v-icon @click="createGroup(item)">mdi-content-save</v-icon>
          <v-icon @click="newGroup.mode = 'none'">mdi-close</v-icon>
        </template>

        <template v-if="newGroup.mode === 'none'">
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

        <template v-else-if="newGroup.planId !== item.id">{{
          item.group
        }}</template>
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
          style="min-width: 200px"
          v-model="item.value"
          :suffix="defaultCurrency"
        />
      </template>

      <template v-slot:expanded-item="{ headers, item }">
        <template v-if="item.installation_fee">
          <td></td>
          <td></td>
          <td :colspan="headers.length - 6">Installation price</td>
          <td>
            {{ item.installation_fee.price.value }}
            {{ defaultCurrency }}
          </td>
          <td>
            <v-text-field
              dense
              style="width: 150px"
              v-model="item.installation_fee.value"
            />
          </td>
          <td></td>
          <td></td>
        </template>
      </template>

      <template v-slot:[`item.addons`]="{ item }">
        <v-dialog width="90vw">
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              icon
              v-bind="attrs"
              v-on="on"
              :disabled="isAddonsLoading"
              @click="fetchAddonsToOne(item)"
            >
              <v-icon> mdi-menu-open </v-icon>
            </v-btn>
          </template>

          <nocloud-table
            table-name="dedicated-addons-prices"
            class="pa-4"
            item-key="id"
            :show-select="false"
            :items="getAddons(item)"
            :headers="addonsHeaders"
            :loading="isAddonsLoading"
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
        </v-dialog>
      </template>
      <template v-slot:[`item.sell`]="{ item }">
        <v-switch
          :loading="isAddonsLoading && fetchAddonsItemId === item.id"
          :disabled="isAddonsLoading && fetchAddonsItemId !== item.id"
          v-model="item.sell"
          @change="fetchAddonsToOne(item)"
        />
      </template>
    </nocloud-table>
  </div>
</template>

<script>
import api from "@/api.js";
import nocloudTable from "@/components/table.vue";
import { getMarginedValue } from "@/functions";

export default {
  name: "dedicated-table",
  components: { nocloudTable },
  props: {
    fee: { type: Object, required: true },
    template: { type: Object, required: true },
    isPlansLoading: { type: Boolean, required: true },
    getPeriod: { type: Function, required: true },
  },
  data: () => ({
    plans: [],
    addons: {},
    headers: [
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
      { text: "Addons", value: "addons", sortable: false },
      {
        text: "Sell",
        value: "sell",
        sortable: false,
        class: "groupable",
        width: 100,
      },
    ],
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

    filters: { Sell: ["true", "false"], Payment: ["monthly", "yearly"] },
    selected: { Sell: ["true", "false"], Payment: ["monthly", "yearly"] },

    column: "",
    fetchError: "",
    isAddonsLoading: false,
    fetchAddonsItemId: "",
    setToAllValue: null,

    groups: [],
    newGroup: { mode: "none", name: "", planId: "" },
    usedFee: {},
  }),
  methods: {
    getAddons({ planCode, duration }) {
      return this.addons[planCode]?.filter((a) => a.duration === duration);
    },
    async fetchAddons({ planCode, sell }) {
      if (this.addons[planCode]) {
        this.addons[planCode].forEach(({ price }, i) => {
          if (price.value !== 0) return;
          this.addons[planCode][i].sell = sell;
        });
        return;
      }

      try {
        const {
          meta: { options },
        } = await api.post(`/sp/${this.sp.uuid}/invoke`, {
          method: "get_baremetal_options",
          params: { planCode },
        });
        const {
          bandwidth = [],
          memory = [],
          storage = [],
          vrack = [],
          ["system-storage"]: sys = [],
        } = options;
        const plans = [...bandwidth, ...memory, ...storage, ...vrack, ...sys];
        const value = this.setPlans({ plans }, planCode);
        this.setFee(value);
        this.$set(
          this.addons,
          planCode,
          value.map((addon) => {
            const resource = this.template.resources.find(
              (el) => addon.id === el.key
            );

            if (resource) {
              addon.value = resource.price;
              addon.sell = true;
            }
            if (+addon.price.value === 0) addon.sell = true;

            return addon;
          })
        );
      } catch (err) {
        console.error(err);
      }
    },
    async fetchAddonsToOne(tariff) {
      this.isAddonsLoading = true;
      this.fetchAddonsItemId = tariff.id;
      try {
        await this.fetchAddons(tariff);
      } finally {
        this.fetchAddonsItemId = "";
        this.isAddonsLoading = false;
      }
    },
    async changePlan(plan) {
      if (!this.plans.every(({ group }) => this.groups.includes(group))) {
        this.$store.commit("snackbar/showSnackbarError", {
          message: "You must select a group for the tariff!",
        });
        return "error";
      }

      plan.products = {};

      const sp = this.$store.getters["servicesProviders/all"];
      const { uuid } = sp.find((el) => el.type === "ovh");

      plan.resources = this.template.resources;

      const enabledPlans = this.plans.filter((el) => el.sell);
      const configurations = await Promise.all(
        enabledPlans.map((el) =>
          api.post(`/sp/${uuid}/invoke`, {
            method: "get_required_configuration",
            params: {
              planCode: el.planCode,
              duration: el.duration,
              pricingMode: el.duration === "P1M" ? "default" : "upfront12",
            },
          })
        )
      );

      for (let i = 0; i < configurations.length; i++) {
        const {
          meta: { requiredConfiguration },
        } = configurations[i];
        const el = enabledPlans[i];

        const addons = this.getAddons(el)
          ? this.getAddons(el)
              .filter((addon) => addon.sell)
              ?.map((el) => ({ id: el.planCode, title: el.name }))
          : this.template.products[el.id]?.meta.addons;

        const datacenter =
          requiredConfiguration.find((el) => el.label.includes("datacenter"))
            ?.allowedValues ?? [];

        const os =
          requiredConfiguration.find((el) => el.label.includes("os"))
            ?.allowedValues ?? [];

        plan.products[el.id] = {
          kind: "PREPAID",
          title: el.name,
          price: el.value,
          group: el.group,
          period: this.getPeriod(el.duration),
          sorter: Object.keys(plan.products).length,
          installation_fee: el.installation_fee.value,
          meta: { addons, datacenter, os },
        };
      }
      Object.values(this.addons).forEach((planCodeAddons) => {
        planCodeAddons.forEach((el) => {
          if (el.sell) {
            const resource = {
              key: el.id,
              kind: "PREPAID",
              title: el.name,
              price: el.value,
              period: this.getPeriod(el.duration),
              except: false,
              on: [],
            };
            const existedIndex = plan.resources.findIndex(
              (res) => res.key === resource.key
            );
            if (existedIndex !== -1) {
              plan.resources[existedIndex] = resource;
            } else {
              plan.resources.push(resource);
            }
          }
        });
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
                this.filters.Group = this.groups;
                this.selected.Group = this.groups;
              }

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
    getTarrifId({ duration, planCode }) {
      return `${duration} ${planCode}`;
    },
    getAddonId({ duration, planCode, tariff }) {
      return `${duration} ${tariff} ${planCode}`;
    },
    setPlans({ plans }, tariff = null) {
      const result = [];

      plans.forEach(({ prices, planCode, productName }) => {
        prices.forEach(({ pricingMode, price, duration }) => {
          const isMonthly = duration === "P1M" && pricingMode === "default";
          const isYearly = duration === "P1Y" && pricingMode === "upfront12";

          if (isMonthly || isYearly) {
            const newPrice = this.convertPrice(price.value);

            const id = tariff
              ? this.getAddonId({ planCode, duration, tariff })
              : this.getTarrifId({ planCode, duration });

            const installation = prices.find(
              (price) =>
                price.capacities.includes("installation") &&
                price.pricingMode === pricingMode
            );

            result.push({
              planCode,
              duration,
              installation_fee: {
                price: { value: +this.convertPrice(installation.price.value) },
                value: installation.price.value,
              },
              price: { value: newPrice },
              name: productName,
              apiName: productName,
              group: productName.split(/[\W0-9]/)[0],
              value: price.value,
              sell: false,
              id,
            });
          }
        });
      });

      return result;
    },
    changeFilters(plan, filters) {
      filters.forEach((text) => {
        if (!this.filters[text]) this.filters[text] = [];
        if (!this.selected[text]) this.selected[text] = [];

        const { value } = this.headers.find((el) => el.text === text);
        const filter = `${plan[value]}`;

        if (!this.filters[text].includes(filter)) {
          this.filters[text].push(filter);
        }
        if (!this.selected[text].includes(filter)) {
          this.selected[text].push(filter);
        }
      });
    },
    setFee(values) {
      this.filters.Margin = ["manual"];
      this.selected.Margin = ["manual"];
      this.usedFee = JSON.parse(JSON.stringify(this.fee));

      if (!values) {
        this.setFee(this.plans);
        Object.values(this.addons).forEach((el) => {
          this.setFee(el);
        });
      }
      values?.forEach((plan, i, arr) => {
        if (arr[i].installation_fee) {
          const price = arr[i].installation_fee.price.value;

          arr[i].installation_fee.value = getMarginedValue(this.fee, price);
        }
        arr[i].value = getMarginedValue(this.fee, plan.price.value);
        this.getMargin(arr[i]);
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
    getMargin({ value, price }, filter = true) {
      if (!this.usedFee.ranges) {
        if (filter) this.changeFilters({ margin: "none" }, ["Margin"]);
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

      if (filter) this.changeFilters({ margin }, ["Margin"]);
      return margin;
    },
    editGroup(group) {
      const i = this.groups.indexOf(group);

      this.groups.splice(i, 1, this.newGroup.name);
      this.plans.forEach((plan, index) => {
        if (plan.group !== group) return;
        this.plans[index].group = this.newGroup.name;
      });

      this.changeMode("none", { id: -1, group: "" });
    },
    createGroup(plan) {
      this.groups.push(this.newGroup.name);
      plan.group = this.newGroup.name;

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
      this.newGroup.mode = mode;
      this.newGroup.planId = id;
      this.newGroup.name = group;
    },
    applyFilter(values) {
      return values.filter((plan) => {
        const result = [];

        Object.entries(this.selected).forEach(([key, filters]) => {
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
    convertPrice(price) {
      return (price * this.plnRate).toFixed(2);
    },
    async setEnabledToTab(value) {
      this.plans = this.plans.map((p) => {
        const { planCode, duration } = p;
        const inFilter = !!this.filteredPlans.find(
          (fp) => fp.planCode === planCode && duration === fp.duration
        );
        if (inFilter) {
          p.sell = value;
        }

        return p;
      });
      this.setToAllValue = value;
      this.isAddonsLoading = true;
      try {
        await Promise.all(this.filteredPlans.map((p) => this.fetchAddons(p)));
        this.filteredPlans.forEach(({ planCode }) =>
          this.$set(
            this.addons,
            planCode,
            this.addons[planCode].map((addon) => {
              addon.sell = value || +addon.price.value===0;
              return addon;
            })
          )
        );
      } finally {
        this.setToAllValue = null;
        this.isAddonsLoading = false;
      }
    },
  },
  created() {
    this.$emit("changeLoading");

    api.servicesProviders
      .action({
        action: "get_baremetal_plans",
        uuid: this.sp.uuid,
      })
      .then(({ meta }) => {
        this.plans = this.setPlans(meta);
        this.groups = [];
        this.plans.forEach((plan, i) => {
          const product = this.template.products[plan.id];
          const title = (product?.title ?? plan.name).toUpperCase();
          const group = product?.group || title.split(/[\W0-9]/)[0];

          if (product) {
            this.plans[i].name = product.title;
            this.plans[i].value = product.price;
            this.plans[i].sell = true;
            this.plans[i].isBeenSell = true;
          } else {
            this.plans[i].name = title;
          }
          this.plans[i].group = group;

          if (!this.groups.includes(group)) this.groups.push(group);
        });

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
    sp() {
      return this.$store.getters["servicesProviders/all"].find(
        (sp) => sp.type === "ovh"
      );
    },
    filteredPlans() {
      return this.applyFilter(this.plans);
    },
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
        (r) => r.from === "PLN" && r.to === this.defaultCurrency
      )?.rate;
    },
  },
  watch: {
    plans() {
      this.$emit("changeFee", this.template.fee ?? {});
      setTimeout(() => {
        this.setFee(this.plans);
      });
    },
  },
};
</script>

<style>
.v-card .v-icon.group-icon {
  display: none;
  margin: 0 0 2px 4px;
  font-size: 18px;
  opacity: 1;
  cursor: pointer;
  color: #fff;
}

.v-data-table__expanded__content {
  background: var(--v-background-base);
}
</style>
