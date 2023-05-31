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

    <nocloud-table
      item-key="id"
      :show-select="false"
      :items="filteredPlans"
      :headers="headers"
      :loading="isPlansLoading"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.name`]="{ item }">
        <v-text-field dense style="width: 200px" v-model="item.name" />
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
      <template v-slot:[`item.addons`]="{ item }">
        <v-dialog width="90vw">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" @click="fetchAddons(item)">
              mdi-menu-open
            </v-icon>
          </template>

          <nocloud-table
            table-name="dedicated"
            class="pa-4"
            item-key="id"
            :show-select="false"
            :items="addons[item.planCode]"
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
              <v-text-field dense style="width: 150px" v-model="item.value" />
            </template>
            <template v-slot:[`item.sell`]="{ item }">
              <v-switch v-model="item.sell" />
            </template>
          </nocloud-table>
        </v-dialog>
      </template>
      <template v-slot:[`item.sell`]="{ item }">
        <v-switch v-model="item.sell" @change="fetchAddons(item)" />
      </template>
    </nocloud-table>
  </div>
</template>

<script>
import api from "@/api.js";
import nocloudTable from "@/components/table.vue";

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
      { text: "Tariff", value: "name" },
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
    rate: 1,
    isAddonsLoading: false,
  }),
  methods: {
    fetchAddons({ planCode, sell }) {
      if (this.addons[planCode]) {
        this.addons[planCode].forEach(({ price }, i) => {
          if (price.value !== 0) return;
          this.addons[planCode][i].sell = sell;
        });
        return;
      }

      const sp = this.$store.getters["servicesProviders/all"];
      const { uuid } = sp.find((el) => el.type === "ovh");

      this.isAddonsLoading = true;
      api
        .post(`/sp/${uuid}/invoke`, {
          method: "get_baremetal_options",
          params: { planCode },
        })
        .then(({ meta: { options } }) => {
          const {
            bandwidth = [],
            memory = [],
            storage = [],
            vrack = [],
            ["system-storage"]: sys = [],
          } = options;
          const plans = [...bandwidth, ...memory, ...storage, ...vrack, ...sys];
          const value = this.setPlans({ plans }).map((addon) => {
            const resource = this.template.resources.find(
              (el) => addon.id === el.key
            );

            if (resource) {
              addon.value = resource.price;
              addon.sell = true;
            }
            if (addon.price.value === 0 && sell) addon.sell = true;

            return addon;
          });

          this.$set(this.addons, planCode, value);
          this.setFee(this.addons[planCode]);
        })
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          this.isAddonsLoading = false;
        });
    },
    async changePlan(plan) {
      const sp = this.$store.getters["servicesProviders/all"];
      const { uuid } = sp.find((el) => el.type === "ovh");

      for await (const el of this.plans) {
        if (el.sell) {
          const {
            meta: { requiredConfiguration },
          } = await api.post(`/sp/${uuid}/invoke`, {
            method: "get_required_configuration",
            params: {
              planCode: el.planCode,
              duration: el.duration,
              pricingMode: el.duration === "P1M" ? "default" : "upfront12",
            },
          });

          const addons = this.addons[el.planCode]?.map((el) => el.planCode);

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
            period: this.getPeriod(el.duration),
            sorter: Object.keys(plan.products).length,
            meta: { addons, datacenter, os },
          };
        }
      }

      Object.values(this.addons).forEach((addon) => {
        addon.forEach((el) => {
          if (el.sell && !plan.resources.find((res) => res.key === el.id)) {
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
    setPlans({ plans }) {
      const result = [];

      plans.forEach(({ prices, planCode, productName }) => {
        prices.forEach(({ pricingMode, price, duration }) => {
          const isMonthly = duration === "P1M" && pricingMode === "default";
          const isYearly = duration === "P1Y" && pricingMode === "upfront12";

          if (isMonthly || isYearly) {
            const newPrice = parseFloat((price.value * this.rate).toFixed(2));

            result.push({
              planCode,
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

      if (!values) {
        this.setFee(this.plans);
        Object.values(this.addons).forEach((el) => {
          this.setFee(el);
        });
      }
      values?.forEach((plan, i, arr) => {
        const n = Math.pow(10, this.fee.precision ?? 0);
        let percent = (this.fee?.default ?? 0) / 100 + 1;
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
            break;
          default:
            round = "round";
        }
        if (this.fee.round === "NONE" || !this.fee.round) round = "round";
        else if (typeof this.fee.round === "string") {
          round = this.fee.round.toLowerCase();
        }

        for (let range of this.fee?.ranges ?? []) {
          if (plan.price.value <= range.from) continue;
          if (plan.price.value > range.to) continue;
          percent = range.factor / 100 + 1;
        }
        arr[i].value = Math[round](plan.price.value * percent * n) / n;

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
      if (!this.fee.ranges) {
        if (filter) this.changeFilters({ margin: "none" }, ["Margin"]);
        return "none";
      }

      const range = this.fee.ranges.find(
        ({ from, to }) => from <= price.value && to >= price.value
      );
      const n = Math.pow(10, this.fee.precision);
      let percent = range?.factor / 100 + 1;
      let margin;
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
      if (this.fee.round === "NONE") round = "round";
      else if (typeof this.fee.round === "string") {
        round = this.fee.round.toLowerCase();
      }

      if (value === Math[round](price.value * percent * n) / n) {
        margin = "ranged";
      } else {
        percent = this.fee.default / 100 + 1;
      }

      switch (value) {
        case price.value:
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
  },
  created() {
    this.$emit("changeLoading");
    api
      .get(`/billing/currencies/rates/PLN/${this.defaultCurrency}`)
      .then((res) => {
        this.rate = res.rate;
      })
      .catch(() =>
        api.get(`/billing/currencies/rates/${this.defaultCurrency}/PLN`)
      )
      .then((res) => {
        if (res) this.rate = 1 / res.rate;
      })
      .catch((err) => console.error(err));

    this.$store
      .dispatch("servicesProviders/fetch")
      .then(({ pool }) => {
        const sp = pool.find(({ type }) => type === "ovh");

        return api.post(`/sp/${sp.uuid}/invoke`, {
          method: "get_baremetal_plans",
        });
      })
      .then(({ meta }) => {
        this.plans = this.setPlans(meta);

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
      return this.applyFilter(this.plans);
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
  },
  watch: {
    plans() {
      this.$emit("changeFee", this.template.fee ?? {});
      setTimeout(() => {
        this.setFee(this.plans);

        this.plans.forEach((plan, i) => {
          const product = this.template.products[plan.id];

          if (product) {
            this.plans[i].name = product.title;
            this.plans[i].value = product.price;
            this.plans[i].sell = true;
          }
        });

        if (Object.keys(this.template.products).length === this.plans.length) {
          this.filters.Sell = ["true"];
          this.selected.Sell = ["true"];
        }
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
