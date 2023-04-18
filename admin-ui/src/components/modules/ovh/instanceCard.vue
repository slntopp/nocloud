<template>
  <div>
    <v-row align="center">
      <v-col>
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          label="price"
          :value="getPrice"
        />
      </v-col>
    </v-row>
    <h3 v-if="dense">Data:</h3>
    <v-card-title v-else class="px-0">Data:</v-card-title>
    <v-row align="center">
      <v-col v-for="key in dataKeys" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key"
          :value="template.data[key]"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="login"
          style="display: inline-block; width: 200px"
          :value="template.state.meta.login"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="password"
          style="display: inline-block; width: 200px"
          :type="isVisible ? 'text' : 'password'"
          :value="template.state.meta.password"
          :append-icon="isVisible ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append="isVisible = !isVisible"
        />
      </v-col>
    </v-row>

    <h3 v-if="dense">Resources:</h3>
    <v-card-title v-else class="px-0">Resources:</v-card-title>
    <v-row align="center">
      <v-col v-for="(item, key) in resources" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key"
          :value="item"
        />
      </v-col>
      <v-col v-for="k in configKeys" :key="k.key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[k.key] ?? k.key"
          :value="getConfigValue(k.path)"
        />
      </v-col>
    </v-row>
    <h3 v-if="dense">Prices:</h3>
    <v-card-title v-else class="px-0">Prices:</v-card-title>
    <nocloud-table
      sort-by="index"
      item-key="key"
      :show-select="false"
      :headers="pricesHeaders"
      :items="pricesItems"
    >
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field
          v-model.number="prices[item.key]"
          @change="setTotalNewPrice"
          type="number"
        ></v-text-field>
      </template>
      <template v-slot:[`item.basePrice`]="{ item }">
        <v-text-field
          :loading="isBasePricesLoading"
          readonly
          :value="basePrices[item.key]"
        ></v-text-field>
      </template>
      <template v-slot:[`footer`]>
        <div class="d-flex justify-end ml-10 my-1">
          <span>Base total {{ totalBasePrice || "loading..."}} Total {{ totalNewPrice }}</span>
        </div>
      </template>
    </nocloud-table>
    <div class="d-flex justify-end align-center">
      <v-btn :loading="isPlanChangeLoading" @click="saveNewPrices"
        >Change</v-btn
      >
    </div>
  </div>
</template>

<script>
import NocloudTable from "@/components/table.vue";
import api from "@/api";

export default {
  name: "instance-card",
  components: { NocloudTable },
  props: {
    template: { type: Object, required: true },
    dense: { type: Boolean },
  },
  data: () => ({
    isVisible: false,
    dictionary: {
      cpu: "CPU",
      ram: "RAM",
      os: "OS",
      vpsId: "id",
    },
    configKeys: [
      { key: "datacenter", path: null },
      { key: "os", path: null },
      { key: "type", path: "type" },
    ],
    dataKeys: ["vpsId", "creation", "expiration"],
    pricesItems: [],
    prices: {},
    basePrices: {},
    rate: 0,
    pricesHeaders: [
      { text: "Name", value: "title" },
      { text: "Base price", value: "basePrice" },
      { text: "Price", value: "price" },
    ],
    isPlanChangeLoading: false,
    totalNewPrice: 0,
    totalBasePrice: 0,
    isBasePricesLoading: false,
  }),
  mounted() {
    this.initPrices();
    this.initConfigsKeys();
    this.getBasePrices();
  },
  methods: {
    initPrices() {
      this.pricesItems.push({
        title: "tarrif",
        key: "tarrif",
        ind: 0,
      });
      this.prices["tarrif"] = this.tarrif.price;

      this.addons.forEach((key, ind) => {
        this.prices[key] = this.template.billingPlan.resources.find(
          (p) => p.key === [this.duration, key].join(" ")
        ).price;
        this.pricesItems.push({
          title: key,
          key: key,
          index: ind + 1,
        });
      });
      this.setTotalNewPrice();
    },
    saveNewPrices() {
      const instance = JSON.parse(JSON.stringify(this.template));
      const planCode =
        "IND_" + instance.title + "_" + new Date().toISOString().slice(0, 10);
      const plan = {
        title: planCode,
        public: false,
        kind: instance.billingPlan.kind,
        type: instance.billingPlan.type,
        resources: [],
      };
      const product = { ...this.tarrif, price: this.prices.tarrif };
      plan.products = {
        [this.duration + " " + this.template.config.planCode]: product,
      };
      this.addons.forEach((key) => {
        plan.resources.push({
          ...this.template.billingPlan.resources.find(
            (p) => p.key === [this.duration, key].join(" ")
          ),
          price: this.prices[key],
        });
      });

      this.isPlanChangeLoading = true;
      api.plans.create(plan).then((data) => {
        api.servicesProviders.bindPlan(this.template.sp, data.uuid).then(() => {
          const tempService = JSON.parse(JSON.stringify(this.service));
          const igIndex = tempService.instancesGroups.findIndex((ig) =>
            ig.instances.find((i) => i.uuid === this.template.uuid)
          );
          const instanceIndex = tempService.instancesGroups[
            igIndex
          ].instances.findIndex((i) => i.uuid === this.template.uuid);

          tempService.instancesGroups[igIndex].instances[
            instanceIndex
          ].billingPlan = data;
          api.services._update(tempService).then(() => {
            this.isPlanChangeLoading = false;
            this.$emit("refresh");
          });
        });
      });
    },
    initConfigsKeys() {
      this.configKeys.forEach((k) => {
        if (!k.path) {
          k.path = this.getKeyFromConfiguration(k.key);
        }
      });
    },
    getKeyFromConfiguration(name) {
      for (const key of Object.keys(this.template.config.configuration))
        if (key.includes(name)) {
          return key;
        }
    },
    getConfigValue(path) {
      return (
        this.template.config[path] || this.template.config.configuration[path]
      );
    },
    getBasePrices() {
      this.isBasePricesLoading = true;
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
      api
        .post(`/sp/${this.template.sp}/invoke`, { method: "get_plans" })
        .then(({ meta }) => {
          const planCode = meta.plans.find((p) => this.planCode === p.planCode);
          this.basePrices["tarrif"] = this.getPriceFromProduct(planCode);

          this.addons.forEach((addon) => {
            Object.keys(meta).forEach((metaKey) => {
              const product =
                meta[metaKey].find &&
                meta[metaKey].find((p) => p?.planCode === addon);
              if (product) {
                this.basePrices[addon] = this.getPriceFromProduct(product);
              }
            });
          });

          this.totalBasePrice = Object.keys(this.basePrices).reduce(
            (acc, key) => acc + +this.basePrices[key],
            0
          ).toFixed(2);
          this.isBasePricesLoading = false;
        });
    },
    getPriceFromProduct(product) {
      return (
        product.prices.find(
          (p) =>
            this.duration === p.duration &&
            this.template.config.pricingMode === p.pricingMode
        )?.price?.value * this.rate
      ).toFixed(2);
    },
    setTotalNewPrice() {
      this.totalNewPrice = Object.keys(this.prices).reduce(
        (acc, key) => acc + +this.prices[key],
        0
      );
    },
  },
  computed: {
    service() {
      return this.$store.getters["services/all"].find(
        (s) => s.uuid === this.template.service
      );
    },
    resources() {
      return this.tarrif.resources;
    },
    planCode() {
      return this.template.config.planCode;
    },
    duration() {
      return this.template.config.duration;
    },
    addons() {
      return this.template.config.addons;
    },
    tarrif() {
      return this.template.billingPlan.products[
        [this.duration, this.planCode].join(" ")
      ];
    },
    getPrice() {
      const prices = [];
      prices.push(this.tarrif.price);
      this.addons.forEach((name) => {
        prices.push(
          this.template.billingPlan.resources.find(
            (p) => p.key === [this.duration, name].join(" ")
          ).price
        );
      });
      return prices.reduce((acc, val) => acc + val, 0);
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
  },
};
</script>
