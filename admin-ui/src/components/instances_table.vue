<template>
  <nocloud-table
    table-name="instances"
    class="mt-4"
    :value="value"
    :items="instances"
    :headers="instancesHeaders"
    :loading="isLoading"
    :custom-sort="sortInstances"
    :footer-error="fetchError"
    @input="(value) => $emit('input', value)"
    :default-filtres="defaultFiltres"
    :show-select="showSelect"
  >
    <template v-slot:[`item.id`]="{ index }">
      {{ index + 1 }}
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <div class="d-flex justify-space-between">
        <router-link
          :target="openInNewTab ? '_blank' : null"
          :to="{ name: 'Instance', params: { instanceId: item.uuid } }"
        >
          {{ item.title }}
        </router-link>
        <login-in-account-icon
          :uuid="getAccount(item).uuid"
          :instanceId="item.uuid"
          :type="item.type"
        />
      </div>
    </template>

    <template v-slot:[`item.access`]="{ item }">
      <router-link
        v-if="!isAccountsLoading"
        :to="{ name: 'Account', params: { accountId: getAccount(item)?.uuid } }"
      >
        {{ getValue("access", item) }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:[`item.email`]="{ item }">
      {{ getValue("email", item) }}
    </template>

    <template v-slot:[`item.state`]="{ item }">
      <instance-state small :template="item" />
    </template>

    <template v-slot:[`item.product`]="{ item }">
      <router-link
        :to="{
          name: 'Plan',
          params: {
            planId: item.billingPlan?.uuid,
            search: getSearchParamForTariff(item),
          },
        }"
      >
        {{ getValue("product", item) }}
      </router-link>
    </template>

    <template v-slot:[`item.price`]="{ item }">
      {{ getValue("price", item) }}
    </template>

    <template v-slot:[`item.accountPrice`]="{ item }">
      {{ getValue("accountPrice", item) }}
    </template>

    <template v-slot:[`item.period`]="{ item }">
      {{ getValue("period", item) }}
    </template>

    <template v-slot:[`item.date`]="{ item }">
      {{ formatSecondsToDate(getValue("date", item), true) || "Unknown" }}
    </template>

    <template v-slot:[`item.dueDate`]="{ item }">
      {{
        typeof getExpirationDate(item) === "number"
          ? formatSecondsToDate(getExpirationDate(item))
          : getExpirationDate(item)
      }}
    </template>

    <template v-slot:[`item.service`]="{ item, value }">
      <router-link :to="{ name: 'Service', params: { serviceId: value } }">
        {{ getValue("service", item) }}
      </router-link>
    </template>

    <template v-slot:[`item.sp`]="{ item, value }">
      <router-link :to="{ name: 'ServicesProvider', params: { uuid: value } }">
        {{ getValue("sp", item) }}
      </router-link>
    </template>

    <template v-slot:[`item.billingPlan.title`]="{ item, value }">
      <router-link
        :to="{ name: 'Plan', params: { planId: item.billingPlan.uuid } }"
      >
        {{ value }}
      </router-link>
    </template>

    <template v-slot:[`item.resources.cpu`]="{ item }">
      {{ getValue("resources.cpu", item) }}
      {{ getValue("resources.cpu", item) || 0 > 1 ? "cores" : "core" }}
    </template>

    <template v-slot:[`item.resources.ram`]="{ item }">
      {{ getValue("resources.ram", item) }} GB
    </template>

    <template v-slot:[`item.resources.drive_size`]="{ item }">
      {{ getValue("resources.drive_size", item) }} GB
    </template>

    <template v-slot:[`item.config.template_id`]="{ item }">
      {{ getValue("config.template_id", item) }}
    </template>

    <template v-slot:[`item.state.meta.networking`]="{ item }">
      <template v-if="!item.state?.meta.networking?.public">-</template>
      <instance-ip-menu v-else :item="item" ui="span" />
    </template>

    <template v-slot:[`item.config.regular_payment`]="{ item }">
      <div class="d-flex justify-center align-center regular_payment">
        <v-switch
          dense
          hide-details
          :disabled="isChangeRegularPaymentLoading"
          :input-value="item.config.regular_payment"
          @change="changeRegularPayment(item, $event)"
        />
      </div>
    </template>
  </nocloud-table>
</template>

<script>
import nocloudTable from "@/components/table.vue";
import instanceIpMenu from "./ui/instanceIpMenu.vue";
import {
  compareSearchValue,
  formatSecondsToDate,
  getBillingPeriod,
  getDeepObjectValue,
  getInstancePrice,
  getState,
  isInstancePayg,
} from "@/functions";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import searchMixin from "@/mixins/search";
import InstanceState from "@/components/ui/instanceState.vue";
import { mapGetters } from "vuex";
import api from "@/api";

export default {
  name: "instances-table",
  components: {
    InstanceState,
    LoginInAccountIcon,
    nocloudTable,
    instanceIpMenu,
  },
  mixins: [
    searchMixin({
      name: "instances-table",
      defaultLayout: {
        title: "Default",
        filter: {
          state: [
            "INIT",
            "RUNNING",
            "STOPPED",
            "PENDING",
            "OPERATION",
            "SUSPENDED",
            "UNKNOWN",
            "ERROR",
          ],
        },
      },
    }),
  ],
  props: {
    value: { type: Array, required: false },
    headers: { type: Array, default: null },
    selected: { type: Object, default: null },
    showSelect: { type: Boolean, default: true },
    openInNewTab: { type: Boolean, default: false },
    items: { type: Array, default: () => [] },
    noSearch: { type: Boolean, default: false },
  },
  data: () => ({
    fetchError: "",
    defaultFiltres: [
      "id",
      "title",
      "service",
      "access.namespace",
      "access",
      "dueDate",
      "state",
      "product",
      "sp",
      "type",
      "price",
      "period",
      "date",
    ],
    accounts: {},
    instancesTypes: [],
    isChangeRegularPaymentLoading: false,
    isAccountsLoading: false,
  }),
  mounted() {
    const types = require.context(
      "@/components/modules/",
      true,
      /instanceCreate\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(
        /\.\/([A-Za-z0-9-_,\s]*)\/instanceCreate\.vue/i
      );
      if (matched && matched.length > 1) {
        const type = matched[1];
        this.instancesTypes.push(type);
      }
    });
  },
  methods: {
    formatSecondsToDate,
    sortInstances(items, sortBy, sortDesc) {
      return items.sort((a, b) => {
        for (let i = 0; i < sortBy.length; i++) {
          const key = sortBy[i];
          if (sortDesc[i]) [a, b] = [b, a];

          let valueA = a;
          let valueB = b;

          if (this.headersGetters[key]) {
            valueA = this.getValue(key, valueA);
            valueB = this.getValue(key, valueB);
          } else {
            key.split(".").forEach((subkey) => {
              valueA = valueA?.[subkey];
              valueB = valueB?.[subkey];
            });
          }

          if (typeof valueA === "string" && typeof valueB === "string") {
            return valueA.toLowerCase().localeCompare(valueB.toLowerCase());
          } else if (typeof valueA === "number" && typeof valueB === "number") {
            valueA = valueA || 0;
            valueB = valueB || 0;
            return valueA - valueB;
          } else {
            return valueA > valueB;
          }
        }
      });
    },
    getAccount({ access }) {
      const {
        access: { namespace },
      } = this.namespaces?.find(({ uuid }) => uuid === access.namespace) ?? {
        access: {},
      };

      return this.accounts?.[namespace] ?? {};
    },
    getEmail(inst) {
      const account = this.getAccount(inst);

      return account?.data?.email ?? "-";
    },
    getPrice(inst) {
      return getInstancePrice(inst);
    },
    getNcuPrice(inst) {
      const price = (this.getPrice(inst) || 0).toFixed(2);
      if (!price) {
        return "";
      }
      return price + " " + this.defaultCurrency;
    },
    getAccountPrice(inst) {
      const price = this.getPrice(inst);
      if (!price) {
        return "";
      }

      const accountCurrency =
        this.getAccount(inst)?.currency || this.defaultCurrency;
      return (
        (price * this.getRate(accountCurrency)).toFixed(2) +
        " " +
        accountCurrency
      );
    },
    getRate(currency) {
      if (this.defaultCurrency === currency) {
        return 1;
      }
      return this.rates?.find(
        (r) => r.to === currency && r.from === this.defaultCurrency
      )?.rate;
    },
    getPeriod(inst) {
      if (isInstancePayg(inst)) {
        return "PayG";
      } else if (inst.resources.period && inst.type !== "ovh") {
        const text = inst.resources.period > 1 ? "months" : "month";
        return `${inst.resources.period} ${text}`;
      }
      const period = getBillingPeriod(
        Object.values(inst.billingPlan.products || {})[0]?.period || 0
      );

      return period || "Unknown";
    },
    getCreationDate(inst) {
      return +inst.created;
    },
    getExpirationDate(inst) {
      if (isInstancePayg(inst)) return "PayG";
      if (this.getPeriod(inst) === "One time") return "One time";
      return (
        inst.data.expiry?.expiredate || inst.data.next_payment_date || "Unknown"
      );
    },
    getService({ service }) {
      return (
        this.services?.find(({ uuid }) => service === uuid)?.title ?? service
      );
    },
    getServiceProvider({ sp }) {
      return this.sp?.find(({ uuid }) => uuid === sp)?.title;
    },
    getOSName(id, sp) {
      if (!id) return;
      return this.sp?.find(({ uuid }) => uuid === sp)?.publicData.templates[id]
        ?.name;
    },
    getTariff(item) {
      const {
        billingPlan,
        config: { planCode, duration },
      } = item;
      let key;
      if (item.type === "ovh") {
        key = `${duration} ${planCode}`;
      } else {
        key = item.product;
      }

      return billingPlan.products[key]?.title;
    },
    getValue(key, item) {
      return this.headersGetters[key](item);
    },
    getSearchParamForTariff(item) {
      return {
        searchParam: {
          value: this.getValue("product", item),
          title: this.getValue("product", item),
        },
      };
    },
    getSearchKeyItems(key) {
      return [...new Set(this.items.map((i) => this.getValue(key, i)))];
    },
    async changeRegularPayment(instance, value) {
      this.isChangeRegularPaymentLoading = true;
      try {
        const tempService = JSON.parse(
          JSON.stringify(
            this.services?.find((s) => s.uuid === instance.service)
          )
        );
        const igIndex = tempService.instancesGroups.findIndex((ig) =>
          ig.instances?.find((i) => i.uuid === instance.uuid)
        );
        const instanceIndex = tempService.instancesGroups[
          igIndex
        ].instances.findIndex((i) => i.uuid === instance.uuid);

        instance.config.regular_payment = value;

        tempService.instancesGroups[igIndex].instances[instanceIndex] =
          instance;
        await api.services._update(tempService);
      } finally {
        this.isChangeRegularPaymentLoading = false;
      }
    },
    fetchAccounts() {
      this.items.forEach(async ({ access }) => {
        const {
          access: { namespace: uuid },
        } = this.namespaces?.find(({ uuid }) => uuid === access.namespace) ?? {
          access: {},
        };
        if (!uuid) {
          return;
        }

        this.isAccountsLoading = true;
        try {
          if (!this.accounts[uuid]) {
            this.accounts[uuid] = api.accounts.get(uuid);
            this.accounts[uuid] = await this.accounts[uuid];
          }
        } catch {
          this.accounts[uuid] = undefined;
        } finally {
          this.isAccountsLoading = Object.values(this.accounts).some(
            (acc) => acc instanceof Promise
          );
        }
      });
    },
  },
  computed: {
    ...mapGetters("appSearch", { searchParam: "param", filter: "filter" }),
    services() {
      return this.$store.getters["services/all"];
    },
    instances() {
      if (this.noSearch) {
        return this.items;
      }

      const instances = this.items.filter((i) => {
        return Object.keys(this.filter || {})
          .filter((key) => !!this.filter[key])
          .every((key) => {
            let value;
            if (this.headersGetters[key]) {
              value = this.getValue(key, i);
            } else {
              value = getDeepObjectValue(i, key);
            }
            return compareSearchValue(
              value,
              this.filter[key],
              this.searchFields?.find((f) => f.key === key)
            );
          });
      });

      const searchParam = this.searchParam?.toLowerCase();
      if (!searchParam) {
        return instances;
      }

      const searchKeys = [
        "title",
        "uuid",
        "billingPlan.title",
        "account.data.email",
        "price",
        "accountPrice",
        "access",
        "state.meta.networking.public",
      ];

      return instances.filter((item) => {
        item.account = this.getAccount(item);
        const dynamicKeys = [];
        if (item.type === "ovh") {
          dynamicKeys.push(
            `data.${item.config.type}Name`,
            `data.${item.config.type}Id`
          );
        }
        const searchKeysFull = searchKeys.concat(dynamicKeys);
        return searchKeysFull.find((key) => {
          let tempItem = item;
          if (this.headersGetters[key]) {
            tempItem = this.getValue(key, tempItem);
          } else {
            key.split(".").forEach((subkey) => (tempItem = tempItem?.[subkey]));
          }
          if (Array.isArray(tempItem)) {
            return tempItem.some((i) => i.startsWith(searchParam));
          }

          return tempItem?.toString().toLowerCase()?.startsWith(searchParam);
        });
      });
    },
    sp() {
      return this.$store.getters["servicesProviders/all"];
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    instancesHeaders() {
      if (this.headers) return this.headers;
      const headers = [
        { text: "ID", value: "id" },
        { text: "Title", value: "title" },
        { text: "Service", value: "service" },
        { text: "Account", value: "access" },
        { text: "Due date", value: "dueDate" },
        { text: "Status", value: "state" },
        { text: "Tariff", value: "product" },
        { text: "Service provider", value: "sp" },
        { text: "Type", value: "type" },
        { text: "NCU price", value: "price" },
        { text: "Account price", value: "accountPrice" },
        { text: "Period", value: "period" },
        { text: "Email", value: "email" },
        { text: "Created date", value: "date" },
        { text: "UUID", value: "uuid" },
        { text: "Price model", value: "billingPlan.title" },
        { text: "IP", value: "state.meta.networking" },
        { text: "CPU", value: "resources.cpu" },
        { text: "RAM", value: "resources.ram" },
        { text: "Disk", value: "resources.drive_size" },
        { text: "OS", value: "config.template_id" },
        { text: "Domain", value: "resources.domain" },
        { text: "DCV", value: "resources.dcv" },
        { text: "Approver email", value: "resources.approver_email" },
        { text: "Invoice based", value: "config.regular_payment" },
      ];
      return headers;
    },
    headersGetters() {
      return {
        service: this.getService,
        access: (item) => this.getAccount(item)?.title,
        email: this.getEmail,
        state: getState,
        product: (item) => this.getTariff(item) ?? item.product ?? "custom",
        price: this.getNcuPrice,
        accountPrice: this.getAccountPrice,
        period: this.getPeriod,
        date: this.getCreationDate,
        dueDate: (item) => +this.getExpirationDate(item),
        sp: this.getServiceProvider,
        "resources.ram": (item) =>
          +(item?.resources?.ram / 1024).toFixed(2) || 0,
        "resources.drive_size": (item) =>
          +(item?.resources?.drive_size / 1024).toFixed(2) || 0,
        "config.template_id": (item) =>
          this.getOSName(item?.config?.template_id, item.sp),
        "resources.cpu": (item) => item.resources?.cpu || 0,
      };
    },
    isLoading() {
      return this.$store.getters["services/isLoading"];
    },
    priceModelItems() {
      return [...new Set(this.items.map((i) => i.billingPlan?.title))];
    },
    searchFields() {
      return [
        {
          key: "title",
          title: "Title",
          type: "input",
        },
        {
          key: "service",
          items: this.getSearchKeyItems("service"),
          title: "Service",
          type: "select",
        },
        {
          key: "period",
          items: this.getSearchKeyItems("period"),
          title: "Period",
          type: "select",
        },
        {
          key: "sp",
          items: this.getSearchKeyItems("sp"),
          type: "select",
          title: "Service provider",
        },
        {
          key: "access",
          type: "select",
          title: "Account",
          items: this.isAccountsLoading ? [] : this.getSearchKeyItems("access"),
        },
        {
          key: "product",
          items: this.getSearchKeyItems("product"),
          type: "select",
          title: "Product",
        },
        {
          key: "state",
          items: [
            "INIT",
            "RUNNING",
            "STOPPED",
            "PENDING",
            "OPERATION",
            "SUSPENDED",
            "UNKNOWN",
            "DELETED",
            "ERROR",
          ],
          type: "select",
          title: "State",
        },
        {
          key: "type",
          title: "Type",
          type: "select",
          items: this.instancesTypes,
        },
        {
          key: "billingPlan.title",
          type: "select",
          title: "Billing plan",
          items: this.priceModelItems,
        },
        { title: "Due date", key: "dueDate", type: "date" },
        { title: "NCU price", key: "price", type: "number-range" },
        { title: "Account price", key: "accountPrice", type: "number-range" },
        { title: "Email", key: "email", type: "input" },
        { title: "Date", key: "date", type: "date" },
        { title: "IP", key: "state.meta.networking", type: "input" },
        {
          key: "resources.cpu",
          type: "number-range",
          title: "CPU",
        },
        { title: "RAM", key: "resources.ram", type: "number-range" },
        { title: "Disk", key: "resources.drive_size", type: "number-range" },
        { title: "OS", key: "config.template_id", type: "input" },
        { title: "Domain", key: "resources.domain", type: "input" },
        { title: "DCV", key: "resources.dcv", type: "input" },
        {
          title: "Approver email",
          key: "resources.approver_email",
          type: "input",
        },
      ];
    },
    rates() {
      return this.$store.getters["currencies/rates"];
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
  },
  watch: {
    items() {
      this.fetchAccounts();
    },
    namespaces() {
      this.fetchAccounts();
    },
    instances() {
      this.fetchError = "";
    },
    searchFields: {
      deep: true,
      handler() {
        if (!this.noSearch) {
          this.$store.commit("appSearch/setFields", this.searchFields);
        }
      },
    },
  },
};
</script>

<style>
.regular_payment .v-input {
  margin-top: 0px !important;
}
</style>
