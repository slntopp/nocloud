<template>
  <div>
    <v-row class="my-3" v-if="!isPlansLoading">
      <v-btn :loading="isRefreshLoading" class="ml-3" @click="refreshApiPlans"
        >Refresh plans</v-btn
      >
      <v-btn class="ml-3" @click="setEnabledToPlans(true)"
        >Enable all plans</v-btn
      >
      <v-btn class="ml-3" @click="setEnabledToPlans(false)"
        >Disable all plans</v-btn
      >
      <v-btn class="ml-3" @click="setEnabledToAddons(true)"
        >Enable all addons</v-btn
      >
      <v-btn class="ml-3" @click="setEnabledToAddons(false)"
        >Disable all addons</v-btn
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
    >
      <template v-slot:[`item.title`]="{ item }">
        <v-text-field dense style="width: 200px" v-model="item.title" />
      </template>

      <template v-slot:[`item.group`]="{ item }">
        <template
          v-if="newplanGroup.mode === 'edit' && newplanGroup.planId === item.id"
        >
          <v-text-field
            dense
            class="d-inline-block mr-1"
            style="width: 200px"
            v-model="newplanGroup.name"
          />
          <v-icon @click="editPlanGroup(item.group)">mdi-content-save</v-icon>
          <v-icon @click="newplanGroup.mode = 'none'">mdi-close</v-icon>
        </template>

        <template
          v-if="
            newplanGroup.mode === 'create' && newplanGroup.planId === item.id
          "
        >
          <v-text-field
            dense
            class="d-inline-block mr-1"
            style="width: 200px"
            v-model="newplanGroup.name"
          />
          <v-icon @click="createPlanGroup(item)">mdi-content-save</v-icon>
          <v-icon @click="newplanGroup.mode = 'none'">mdi-close</v-icon>
        </template>

        <template v-if="newplanGroup.mode === 'none'">
          <v-autocomplete
            dense
            class="d-inline-block"
            style="width: 200px"
            v-model="item.group"
            :items="planGroups"
          />
          <v-icon @click="changePlanMode('create', item)">mdi-plus</v-icon>
          <v-icon @click="changePlanMode('edit', item)">mdi-pencil</v-icon>
          <v-icon
            v-if="planGroups.length > 1"
            @click="deletePlanGroup(item.group)"
            >mdi-delete</v-icon
          >
        </template>

        <template v-else-if="newplanGroup.planId !== item.id">{{
          item.group
        }}</template>
      </template>

      <template v-slot:[`item.cpu`]="{ item }">
        <template
          v-if="newcpuGroup.mode === 'edit' && newcpuGroup.planId === item.id"
        >
          <v-text-field
            dense
            class="d-inline-block mr-1"
            style="width: 200px"
            v-model="newcpuGroup.name"
          />
          <v-icon @click="editCpuGroup(item.cpu)">mdi-content-save</v-icon>
          <v-icon @click="newcpuGroup.mode = 'none'">mdi-close</v-icon>
        </template>

        <template
          v-if="newcpuGroup.mode === 'create' && newcpuGroup.planId === item.id"
        >
          <v-text-field
            dense
            class="d-inline-block mr-1"
            style="width: 200px"
            v-model="newcpuGroup.name"
          />
          <v-icon @click="createCpuGroup(item)">mdi-content-save</v-icon>
          <v-icon @click="newcpuGroup.mode = 'none'">mdi-close</v-icon>
        </template>

        <template v-if="newcpuGroup.mode === 'none'">
          <v-autocomplete
            dense
            class="d-inline-block"
            style="width: 200px"
            v-model="item.cpu"
            :items="cpuGroups"
          />
          <v-icon @click="changeCpuMode('create', item)">mdi-plus</v-icon>
          <v-icon @click="changeCpuMode('edit', item)">mdi-pencil</v-icon>
          <v-icon v-if="planGroups.length > 1" @click="deleteCpuGroup(item.cpu)"
            >mdi-delete</v-icon
          >
        </template>

        <template v-else-if="newcpuGroup.planId !== item.id">{{
          item.cpu
        }}</template>
      </template>

      <template v-slot:[`item.duration`]="{ value }">
        {{ getPayment(value) }}
      </template>

      <template v-slot:[`item.basePrice`]="{ value }">
        {{ value }} {{ defaultCurrency }}
      </template>

      <template v-slot:[`item.price`]="{ item }">
        <v-text-field
          dense
          style="min-width: 200px"
          v-model="item.price"
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
            <v-btn icon v-bind="attrs" v-on="on">
              <v-icon> mdi-menu-open </v-icon>
            </v-btn>
          </template>

          <nocloud-table
            table-name="dedicated-addons-prices"
            class="pa-4"
            item-key="id"
            :show-select="false"
            :items="item.meta?.addons"
            :headers="addonsHeaders"
          >
            <template v-slot:[`item.duration`]="{ value }">
              {{ getPayment(value) }}
            </template>
            <template v-slot:[`item.basePrice`]="{ value }">
              {{ value }} {{ defaultCurrency }}
            </template>
            <template v-slot:[`item.price`]="{ item }">
              <v-text-field
                dense
                style="width: 200px"
                :suffix="defaultCurrency"
                v-model="item.price"
              />
            </template>
            <template v-slot:[`item.public`]="{ item }">
              <v-switch v-model="item.public" />
            </template>
          </nocloud-table>
        </v-dialog>
      </template>
      <template v-slot:[`item.public`]="{ item }">
        <v-switch v-model="item.public" />
      </template>
    </nocloud-table>
  </div>
</template>

<script>
import nocloudTable from "@/components/table.vue";
import { mapGetters } from "vuex";
import { getMarginedValue } from "@/functions";
import api from "@/api";

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
    headers: [
      { text: "Name", value: "title" },
      { text: "API name", value: "apiName" },
      { text: "Group", value: "group" },
      { text: "Cpu", value: "cpu" },
      {
        text: "Payment",
        value: "duration",
      },
      { text: "Income price", value: "price" },
      { text: "Sale price", value: "basePrice" },
      { text: "Addons", value: "addons" },
      {
        text: "Sell",
        value: "public",
        width: 100,
      },
    ],
    addonsHeaders: [
      { text: "Addon", value: "title" },
      {
        text: "Payment",
        value: "duration",
      },
      { text: "Income price", value: "price" },
      { text: "Sale price", value: "basePrice" },
      {
        text: "Sell",
        value: "public",
        width: 100,
      },
    ],

    isRefreshLoading: false,
    planGroups: [],
    newplanGroup: { mode: "none", name: "", planId: "" },

    cpuGroups: [],
    newcpuGroup: { mode: "none", name: "", planId: "" },

    usedFee: {},
  }),
  methods: {
    getApiProducts(plans, tariffPlanCode = null) {
      const result = [];

      plans.forEach(({ prices, planCode, productName }) => {
        prices.forEach(({ pricingMode, price, duration }) => {
          const isMonthly = duration === "P1M" && pricingMode === "default";
          const isYearly = duration === "P1Y" && pricingMode === "upfront12";
          if (!isMonthly && !isYearly) return;

          const newPrice = this.convertPrice(price.value);

          const id = tariffPlanCode
            ? this.getAddonId({ planCode, duration }, tariffPlanCode)
            : this.getTariffId({ planCode, duration });

          const installation = prices.find(
            (price) =>
              price.capacities.includes("installation") &&
              price.pricingMode === pricingMode
          );

          result.push({
            planCode,
            duration,
            installation_fee: {
              price: {
                value: +this.convertPrice(installation.price.value),
              },
              value: installation.price.value,
            },
            price: newPrice,
            title: productName,
            apiName: `${productName} (${planCode})`,
            group: productName.split(/[\W0-9]/)[0],
            basePrice: price.value,
            public: false,
            meta: {},
            id,
          });
        });
      });

      return result;
    },
    async refreshApiPlans() {
      try {
        this.isRefreshLoading = true;
        let plans = await this.fetchPlans();

        const allAddons = await Promise.allSettled(
          plans.map(async (p) => ({
            planId: p.id,
            data: await this.fetchAddonsToPlan(p),
          }))
        );

        const allConfigurations = await Promise.allSettled(
          plans.map(async (p) => ({
            planId: p.id,
            data: await this.fetchConfigurationToPlan(p),
          }))
        );

        plans.forEach((p) => {
          p.meta.addons = allAddons.find(
            (result) => result.value.planId === p.id
          ).value.data;
          const { os, datacenter } = allConfigurations.find(
            (result) => result.value.planId === p.id
          ).value.data;
          p.meta.os = os;
          p.meta.datacenter = datacenter;
        });

        //for all cpu needs os,datacenters,addons ^
        const allCpus = await Promise.allSettled(
          plans.map(async (p) => ({
            planId: p.id,
            data: await this.fetchCpuToPlan(p),
          }))
        );

        allCpus.forEach((result) => {
          if (result.status === "rejected") {
            return;
          }
          const { planId, data } = result.value;
          const planIndex = plans.findIndex((p) => p.id === planId);
          plans[planIndex].cpu = data;
        });

        this.plans = plans.map((plan) => {
          const realPlan = this.plans.find((real) => real.id === plan.id);
          if (!realPlan) {
            return plan;
          }

          const { os, datacenter, cpu } = plan.meta;

          const addons = plan.meta.addons.map((addon) => {
            const realAddon = realPlan.meta?.addons?.find(
              ({ id }) => id === addon.id
            );
            if (!realAddon) {
              return addon;
            }
            const { price, public: Public } = realAddon;
            return { ...addon, public: Public, price };
          });

          return {
            ...plan,
            ...realPlan,
            cpu: plan.cpu || realPlan.cpu,
            apiName: plan.apiName,
            basePrice: plan.basePrice,
            meta: {
              addons,
              os,
              datacenter,
              cpu,
            },
          };
        });
      } catch (err) {
        this.$store.commit("snackbar/showSnackbarError", {
          message: err.response?.data?.message ?? err.message ?? err,
        });
      } finally {
        this.isRefreshLoading = false;
      }
    },
    async fetchPlans() {
      const {
        meta: { plans },
      } = await api.servicesProviders.action({
        action: "get_baremetal_plans",
        uuid: this.sp.uuid,
      });

      return this.getApiProducts(plans);
    },
    async fetchAddonsToPlan({ planCode, duration }) {
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

      return this.getApiProducts(plans, planCode).filter(
        (p) => p.duration === duration
      );
    },
    async fetchConfigurationToPlan({ planCode, duration }) {
      const {
        meta: { requiredConfiguration },
      } = await api.post(`/sp/${this.sp.uuid}/invoke`, {
        method: "get_required_configuration",
        params: {
          planCode: planCode,
          duration: duration,
          pricingMode: duration === "P1M" ? "default" : "upfront12",
        },
      });

      const datacenter =
        requiredConfiguration.find((el) => el.label.includes("datacenter"))
          ?.allowedValues ?? [];

      const os =
        requiredConfiguration.find((el) => el.label.includes("os"))
          ?.allowedValues ?? [];

      return { os, datacenter };
    },
    async fetchCpuToPlan(plan) {
      const { planCode, duration } = plan;
      const addons = plan.meta.addons
        ?.filter(
          (a, index, arr) =>
            +a.basePrice === 0 &&
            index ===
              arr.findIndex((dublicate) =>
                dublicate.planCode.startsWith(a.planCode.split("-")?.[0])
              )
        )
        ?.map((a) => a.planCode);
      const configuration = {
        dedicated_datacenter: plan.meta.datacenter[0],
        dedicated_os: plan.meta.os[0],
      };

      const { meta } = await api.servicesProviders.action({
        uuid: this.sp.uuid,
        action: "checkout_baremetal",
        params: {
          configuration,
          addons,
          planCode,
          duration,
          pricingMode: duration === "P1M" ? "default" : "upfront12",
        },
      });

      return meta.order.details.find(
        (d) =>
          d.description.toLowerCase().includes("intel") ||
          d.description.toLowerCase().includes("amd")
      )?.description;
    },
    setPlanGroups() {
      const groups = [];

      this.plans.forEach((plan, i) => {
        const title = plan.title?.toUpperCase();
        const group = plan?.group || title.split(/[\W0-9]/)[0];

        this.plans[i].group = group;

        if (!groups.includes(group)) groups.push(group);
      });

      this.planGroups = groups;
    },
    setCpuGroups() {
      const groups = [];

      this.plans.forEach((plan) => {
        if (!groups.includes(plan.cpu)) groups.push(plan.cpu);
      });

      this.cpuGroups = groups;
    },
    async changePlan(plan) {
      plan.products = {};
      plan.resources = [];

      this.plans.forEach((p) => {
        p.meta.addons?.forEach((a) => {
          plan.resources.push({
            key: a.id,
            kind: "PREPAID",
            title: a.title,
            price: a.price,
            public: a.public,
            period: this.getPeriod(a.duration),
            except: false,
            on: [],
          });
        });

        p.meta.addons = (p.meta.addons || [])
          .filter((addon) => addon.public)
          ?.map((el) => ({ id: el.id, title: el.title }));

        plan.products[p.id] = {
          kind: "PREPAID",
          title: p.title,
          public: p.public,
          price: p.price,
          group: p.group,
          period: this.getPeriod(p.duration),
          sorter: Object.keys(plan.products).length,
          installation_fee: p.installation_fee?.value,
          meta: {
            ...p.meta,
            basePrice: p.basePrice,
            apiName: p.apiName,
            cpu: p.cpu,
          },
        };
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
    getTariffId({ duration, planCode }) {
      return `${duration} ${planCode}`;
    },
    getAddonId({ duration, planCode }, name) {
      return `${duration} ${planCode} ${name}`;
    },
    convertPrice(price) {
      return (price * this.plnRate).toFixed(2);
    },

    createPlanGroup(plan) {
      this.createGroup({ type: "plan", path: "group" }, plan);
    },
    deletePlanGroup(group) {
      this.deleteGroup({ type: "plan", path: "group" }, group);
    },
    changePlanMode(mode, { id, group }) {
      this.changeGroupMode({ type: "plan" }, mode, { id, group });
    },
    editPlanGroup(group) {
      this.editGroup({ type: "plan", path: "group" }, group);
    },

    createCpuGroup(plan) {
      this.createGroup({ type: "cpu", path: "cpu" }, plan);
    },
    deleteCpuGroup(group) {
      this.deleteGroup({ type: "cpu", path: "cpu" }, group);
    },
    changeCpuMode(mode, { id, group }) {
      this.changeGroupMode({ type: "cpu" }, mode, { id, group });
    },
    editCpuGroup(group) {
      this.editGroup({ type: "cpu", path: "cpu" }, group);
    },

    createGroup({ type, path }, plan) {
      const name = this[`new${type}Group`].name;
      this[`${type}Groups`].push(name);
      plan[path] = name;

      this.changeGroupMode({ type }, "none", { id: -1, group: "" });
    },
    deleteGroup({ type, path }, group) {
      this[`${type}Groups`] = this[`${type}Groups`].filter(
        (el) => el !== group
      );
      this.plans.forEach((plan, i) => {
        if (plan[path] !== group) return;
        this.plans[i][path] = this[`${type}Groups`][0];
      });
    },
    editGroup({ type, path }, group) {
      const name = this[`new${type}Group`].name;
      const i = this[`${type}Groups`].indexOf(group);

      this[`${type}Groups`].splice(i, 1, name);
      this.plans.forEach((plan, index) => {
        if (plan[path] !== group) return;
        this.plans[index][path] = name;
      });

      this.changeGroupMode({ type }, "none", { id: -1, group: "" });
    },
    changeGroupMode({ type }, mode, { id, group }) {
      this[`new${type}Group`].mode = mode;
      this[`new${type}Group`].planId = id;
      this[`new${type}Group`].name = group;
    },
    setFee() {
      this.plans = this.plans.map((p) => {
        return {
          ...p,
          meta: {
            ...p.meta,
            addons: p.meta.addons?.map((a) => ({
              ...a,
              price: getMarginedValue(this.fee, a.basePrice),
            })),
          },
          price: getMarginedValue(this.fee, p.basePrice),
        };
      });
    },
    setEnabledToPlans(value) {
      this.plans = this.plans.map((p) => {
        return {
          ...p,
          meta: {
            ...p.meta,
            addons: p.meta.addons?.map((a) => ({
              ...a,
              public: value && a.basePrice === 0,
            })),
          },
          public: value,
        };
      });
    },
    setEnabledToAddons(value) {
      this.plans = this.plans.map((p) => {
        return {
          ...p,
          meta: {
            ...p.meta,
            addons: p.meta.addons?.map((a) => ({
              ...a,
              public: value,
            })),
          },
        };
      });
    },
  },
  created() {
    this.plans = Object.keys(this.template.products || {}).map((key) => {
      const product = this.template.products[key];

      const [duration, planCode] = key.split(" ");

      const addons = (product.meta?.addons || []).map(({ id }) => {
        return {
          ...(this.template.resources.find(({ key }) => key === id) || {}),
          id,
          duration,
        };
      });

      return {
        ...product,
        duration,
        planCode,
        cpu: product.meta.cpu,
        apiName: product.meta.apiName,
        basePrice: product.meta.basePrice,
        installation_fee: {
          price: {
            value: +this.convertPrice(product.installationFee),
          },
          value: product.installationFee,
        },
        id: key,
        meta: {
          ...product.meta,
          addons,
        },
      };
    });
  },
  computed: {
    ...mapGetters("currencies", { rates: "rates", defaultCurrency: "default" }),
    sp() {
      return this.$store.getters["servicesProviders/all"].find(
        (sp) => sp.type === "ovh"
      );
    },
    filteredPlans() {
      return this.plans;
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
      this.setPlanGroups();
      this.setCpuGroups();
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
