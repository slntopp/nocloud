<template>
  <div>
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

        <div v-else-if="tab === 'Tariffs'">
          <v-row class="my-3" v-if="!isPlansLoading">
            <v-btn
              :loading="isRefreshLoading"
              class="ml-3"
              @click="refreshApiPlans"
              >Fetch plans</v-btn
            >
            <v-btn :disabled="!newPlans" class="ml-3" @click="setRefreshedPlans"
              >Set api plans</v-btn
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
            sort-by="isBeenSell"
            sort-desc
            item-key="id"
            table-name="dedicated-prices"
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
                v-if="
                  newplanGroup.mode === 'edit' &&
                  newplanGroup.planId === item.id
                "
              >
                <v-text-field
                  dense
                  class="d-inline-block mr-1"
                  style="width: 200px"
                  v-model="newplanGroup.name"
                />
                <v-icon @click="editPlanGroup(item.group)"
                  >mdi-content-save</v-icon
                >
                <v-icon @click="newplanGroup.mode = 'none'">mdi-close</v-icon>
              </template>

              <template
                v-if="
                  newplanGroup.mode === 'create' &&
                  newplanGroup.planId === item.id
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
                <v-icon @click="changePlanMode('create', item)"
                  >mdi-plus</v-icon
                >
                <v-icon @click="changePlanMode('edit', item)"
                  >mdi-pencil</v-icon
                >
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
                v-if="
                  newcpuGroup.mode === 'edit' && newcpuGroup.planId === item.id
                "
              >
                <v-text-field
                  dense
                  class="d-inline-block mr-1"
                  style="width: 200px"
                  v-model="newcpuGroup.name"
                />
                <v-icon @click="editCpuGroup(item.cpu)"
                  >mdi-content-save</v-icon
                >
                <v-icon @click="newcpuGroup.mode = 'none'">mdi-close</v-icon>
              </template>

              <template
                v-if="
                  newcpuGroup.mode === 'create' &&
                  newcpuGroup.planId === item.id
                "
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
                <v-icon
                  v-if="planGroups.length > 1"
                  @click="deleteCpuGroup(item.cpu)"
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
              {{ value }} {{ defaultCurrency?.title }}
            </template>

            <template v-slot:[`item.price`]="{ item }">
              <v-text-field
                dense
                style="min-width: 200px"
                v-model="item.price"
                :suffix="defaultCurrency?.title"
              />
            </template>

            <template v-slot:expanded-item="{ headers, item }">
              <template v-if="item.installation_fee">
                <td></td>
                <td></td>
                <td :colspan="headers.length - 6">Installation price</td>
                <td>
                  {{ item.installation_fee.price.value }}
                  {{ defaultCurrency?.title }}
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
                  class="pa-4"
                  item-key="id"
                  :show-select="false"
                  :items="getPlanAddons(item)"
                  :headers="addonsHeaders"
                >
                  <template v-slot:[`item.duration`]="{ item }">
                    {{
                      getPayment(item.duration) || getBillingPeriod(item.period)
                    }}
                  </template>
                  <template v-slot:[`item.basePrice`]="{ value }">
                    {{ value }} {{ value ? defaultCurrency?.title : "" }}
                  </template>
                  <template v-slot:[`item.price`]="{ item }">
                    <span v-if="item.virtual">{{ item.price }}</span>
                    <v-text-field
                      v-else
                      dense
                      style="width: 200px"
                      :suffix="defaultCurrency?.title"
                      v-model="item.price"
                    />
                  </template>
                  <template v-slot:[`item.public`]="{ item: addon }">
                    <v-switch
                      :input-value="addon.public"
                      @change="changeAddonPublic(item, addon, $event)"
                    />
                  </template>
                </nocloud-table>
              </v-dialog>
            </template>
            <template v-slot:[`item.public`]="{ item }">
              <v-switch v-model="item.public" />
            </template>
          </nocloud-table>
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
import nocloudTable from "@/components/table.vue";
import { mapGetters } from "vuex";
import { getBillingPeriod, getMarginedValue } from "@/functions";
import api from "@/api";
import {
  Addon,
  ListAddonsRequest,
} from "nocloud-proto/proto/es/billing/addons/addons_pb";
import planAddonsTable from "@/components/planAddonsTable.vue";

export default {
  name: "dedicated-table",
  components: { nocloudTable, planAddonsTable },
  props: {
    fee: { type: Object, required: true },
    template: { type: Object, required: true },
    isPlansLoading: { type: Boolean, required: true },
    getPeriod: { type: Function, required: true },
  },
  data: () => ({
    plans: [],
    newPlans: null,
    headers: [
      { text: "Title", value: "title" },
      { text: "API title", value: "apiName" },
      { text: "Group", value: "group" },
      { text: "Cpu", value: "cpu" },
      {
        text: "Payment",
        value: "duration",
      },
      { text: "Incoming price", value: "basePrice" },
      { text: "Sale price", value: "price" },
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
      { text: "Incoming price", value: "basePrice" },
      { text: "Sale price", value: "price" },
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

    tabsIndex: 0,
    tabs: ["Tariffs", "Custom addons"],
    planAddons: [],
  }),
  methods: {
    getBillingPeriod,
    setRefreshedPlans() {
      this.plans = JSON.parse(JSON.stringify(this.newPlans));
      this.newPlans = null;
    },
    getProductDescription(products, code) {
      return products.find((p) => p.name === code)?.description;
    },
    async refreshApiPlans() {
      try {
        this.isRefreshLoading = true;

        const {
          meta: {
            catalog: { plans, addons, products },
          },
        } = await api.servicesProviders.action({
          action: "get_baremetal_plans",
          uuid: this.sp.uuid,
        });

        const newPlans = [];
        plans.map((plan) => {
          const { planCode, invoiceName } = plan;

          const allowedDurations = { default: "P1M", upfront12: "P1Y" };
          const allowedCapacities = ["installation", "renew"];
          const allowedAddonTypes = [
            "vrack",
            "storage",
            "system-storage",
            "bandwidth",
            "memory",
          ];
          const configurationsMap = {
            dedicated_datacenter: "datacenter",
            dedicated_os: "os",
          };

          const prices = plan.pricings.reduce((acc, pricing) => {
            const capacity = pricing.capacities[0];
            const { mode } = pricing;
            if (
              allowedDurations[mode] &&
              allowedCapacities.includes(capacity)
            ) {
              const tariff = this.getTariffId({
                planCode,
                duration: allowedDurations[mode],
              });
              if (!acc[tariff]) {
                acc[tariff] = {};
              }
              acc[tariff][capacity] = pricing.price / 10 ** 8;
            }

            return acc;
          }, {});
          Object.keys(prices).forEach((tariffId) => {
            const realPlan =
              this.plans.find((real) => real.id === tariffId) || {};
            const duration = tariffId.split(" ")[0];
            const mode = Object.keys(allowedDurations).find(
              (a) => allowedDurations[a] === duration
            );

            const tariffConfigurations = {};
            const tariffAddons = [];

            plan.configurations.forEach((configuration) => {
              const configurationKey = configurationsMap[configuration.name];
              if (configurationKey) {
                tariffConfigurations[configurationKey] = configuration.values;
              }
            });

            plan.addonFamilies.forEach((addonsTyped) => {
              if (allowedAddonTypes.includes(addonsTyped.name)) {
                addonsTyped.addons.forEach((addon) => {
                  const addonInfo = addons.find((a) => a.planCode === addon);
                  let basePrice = addonInfo?.pricings?.find(
                    (p) => p.mode === mode && p.capacities[0] === "renew"
                  )?.price;
                  if (!basePrice && basePrice !== 0) {
                    return;
                  }

                  const addonId = this.getAddonId(
                    { duration, planCode },
                    addon
                  );

                  const realAddon =
                    this.template.resources?.find(
                      ({ key }) => key === addonId
                    ) || {};

                  basePrice = this.convertPrice(basePrice / 10 ** 8);
                  const apiName = this.getProductDescription(
                    products,
                    addonInfo.product
                  );

                  tariffAddons.push({
                    duration,
                    planCode: addon,
                    title: realAddon.title || apiName,
                    price: realAddon.price || basePrice,
                    group: realAddon.group || apiName.split(/[\W0-9]/)[0],
                    id: addonId,
                    basePrice,
                    apiName,
                    public: realAddon.public,
                    meta: {},
                  });
                });
              }
            });

            const cpu = this.getProductDescription(products, planCode);

            const basePrice = this.convertPrice(prices[tariffId].renew);
            const installationPrice = this.convertPrice(
              prices[tariffId].installation
            );
            const apiName = `${invoiceName} (${planCode})`;

            newPlans.push({
              planCode,
              id: tariffId,
              basePrice,
              apiName,
              duration,
              price: realPlan.price || basePrice,
              title: realPlan.title || apiName,
              group: realPlan.group || apiName.split(/[\W0-9]/)[0],
              meta: {
                addons: tariffAddons,
                os: tariffConfigurations.os,
                datacenter: tariffConfigurations.datacenter,
              },
              installation_fee: {
                price: {
                  value: +installationPrice,
                },
                value: installationPrice,
              },
              public: realPlan.public,
              cpu: realPlan.cpu || cpu,
            });
          });
        });

        this.newPlans = newPlans;

        if (!this.plans?.length) {
          this.setRefreshedPlans();
        }
      } catch (err) {
        this.newPlans = null;
        this.$store.commit("snackbar/showSnackbarError", {
          message: err.response?.data?.message ?? err.message ?? err,
        });
      } finally {
        this.isRefreshLoading = false;
      }
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

      this.$emit("changeFee", this.fee);

      const addonsForCreate = [];
      const addonsForUpdate = [];
      let allAddons = [];

      this.plans.forEach((p) => {
        p.meta.addons?.forEach((addon) => {
          const data = Addon.fromJson({
            system: true,
            group: this.template.uuid,
            kind: "PREPAID",
            public: !!addon.public,
            meta: {
              key: addon.planCode,
              apiName: addon.apiName,
              id: addon.id,
              basePrice: addon.basePrice,
            },
            title: addon.title,
            periods: { [this.getPeriod(p.duration)]: addon.price },
          });
          if (addon.uuid) {
            data.uuid = addon.uuid;
            addonsForUpdate.push(data);
          } else {
            addonsForCreate.push(data);
          }
        });
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

      this.plans.forEach((p) => {
        const addons = (p.meta.addons || [])?.map(
          (el) => allAddons.find((addon) => addon.meta.id === el.id)?.uuid
        );

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
            baseInstallationFee: p.installation_fee?.price?.value,
            addons: undefined,
          },
          addons,
        };
      });

      plan.addons = this.planAddons;
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
          installation_fee: {
            ...p.installation_fee,
            value: getMarginedValue(this.fee, p.installation_fee.price.value),
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
              public: value && +a.basePrice === 0,
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
    getPlanAddons(item) {
      return item.meta?.addons || [];
    },
    changeAddonPublic(plan, addon, value) {
      if (!addon.virtual) {
        addon.public = value;
        plan.meta.addons = plan.meta.addons.map((a) =>
          a.id === addon.id ? addon : a
        );
      } else {
        if (value) {
          plan.meta.addons.push({ ...addon, public: true });
        } else {
          plan.meta.addons = plan.meta.addons.filter((a) => a.id !== addon.id);
        }
      }

      const planIndex = this.plans.findIndex((p) => p.id === plan.id);
      this.$set(this.plans, planIndex, plan);
    },
  },
  mounted() {
    this.refreshApiPlans();

    this.planAddons = [...this.template.addons];
  },
  async created() {
    this.$emit("changeFee", this.template.fee);

    const { addons } = await this.addonsClient.list(
      ListAddonsRequest.fromJson({ filters: { group: [this.template.uuid] } })
    );

    this.plans = Object.keys(this.template.products || {}).map((key) => {
      const product = this.template.products[key];

      const [duration, planCode] = key.split(" ");

      const tariffAddons = addons
        .filter((a) => {
          if (product.addons.includes(a.uuid)) {
            return true;
          }
          return false;
        })
        .map((data) => {
          const addon = data.toJson();

          return {
            basePrice: addon.meta.basePrice,
            id: addon.meta.id,
            group: "",
            duration,
            planCode: addon.meta.key,
            apiName: addon.meta.apiName,
            title: addon.title,
            price: addon.periods[this.getPeriod(duration)],
            public: addon.public,
            uuid: addon.uuid,
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
            value: product.meta.baseInstallationFee,
          },
          value: product.installationFee,
        },
        id: key,
        meta: {
          ...product.meta,
          addons: tariffAddons,
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
      if (this.defaultCurrency.title === "PLN") {
        return 1;
      }
      return this.rates.find(
        (r) =>
          r.from.title === "PLN" && r.to.title === this.defaultCurrency.title
      )?.rate;
    },
    addonsClient() {
      return this.$store.getters["addons/addonsClient"];
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
