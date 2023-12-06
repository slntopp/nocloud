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
        :to="{ name: 'Account', params: { accountId: getAccount(item)?.uuid } }"
      >
        {{ getValue("access", item) }}
      </router-link>
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
      {{ getValue("date", item) }}
    </template>

    <template v-slot:[`item.dueDate`]="{ item }">
      {{ getValue("dueDate", item) }}
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

    <template v-slot:[`item.access.namespace`]="{ item }">
      <router-link
        :to="{
          name: 'NamespacePage',
          params: { namespaceId: item.access.namespace },
        }"
      >
        {{ getValue("access.namespace", item) }}
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
  </nocloud-table>
</template>

<script>
import nocloudTable from "@/components/table.vue";
import instanceIpMenu from "./ui/instanceIpMenu.vue";
import {
  compareSearchValue,
  getDeepObjectValue,
  getOvhPrice,
  getState,
} from "@/functions";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import searchMixin from "@/mixins/search";
import InstanceState from "@/components/ui/instanceState.vue";
import { mapGetters } from "vuex";

export default {
  name: "instances-table",
  components: {
    InstanceState,
    LoginInAccountIcon,
    nocloudTable,
    instanceIpMenu,
  },
  mixins: [searchMixin("instances-table")],
  props: {
    value: { type: Array, required: false },
    headers: { type: Array, default: null },
    selected: { type: Object, default: null },
    showSelect: { type: Boolean, default: true },
    openInNewTab: { type: Boolean, default: false },
    items: { type: Array, default: () => [] },
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
    instancesTypes: [],
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
              valueA = valueA[subkey];
              valueB = valueB[subkey];
            });
          }

          if (typeof valueA === "string") {
            return valueA.toLowerCase().localeCompare(valueB.toLowerCase());
          } else if (typeof valueA === "number") {
            return valueA - valueB;
          } else {
            return valueA > valueB;
          }
        }
      });
    },
    date(timestamp) {
      if (!timestamp) return "PayG";
      const date = new Date(timestamp * 1000);

      const year = date.toUTCString().split(" ")[3];
      let month = date.getUTCMonth() + 1;
      let day = date.getUTCDate();

      if (`${month}`.length < 2) month = `0${month}`;
      if (`${day}`.length < 2) day = `0${day}`;

      return `${year}-${month}-${day}`;
    },
    getAccount({ access }) {
      const {
        access: { namespace },
      } = this.namespaces.find(({ uuid }) => uuid === access.namespace) ?? {
        access: {},
      };

      return this.accounts.find(({ uuid }) => uuid === namespace) ?? {};
    },
    getEmail(inst) {
      const account = this.getAccount(inst);

      return account?.data?.email ?? "-";
    },
    getPrice(inst) {
      switch (inst.type) {
        case "goget": {
          const key = `${inst.resources.period} ${inst.resources.id}`;

          return inst.billingPlan.products[key]?.price ?? 0;
        }
        case "ovh": {
          return getOvhPrice(inst);
        }
        case "empty": {
          const initialPrice =
            inst.billingPlan.products[inst.product]?.price ?? 0;
          return inst.billingPlan.resources
            .filter(({ key }) => inst.config?.addons?.find((a) => a === key))
            .reduce((acc, r) => acc + +r?.price, initialPrice);
        }
        case "keyweb": {
          const key = inst.product;
          const tariff = inst.billingPlan.products[key];

          const getAddonKey = (key, metaKey) =>
            tariff.meta?.[metaKey].find(
              (a) =>
                key === a.type &&
                a.key.startsWith(inst.config?.configurations[key])
            )?.key;

          const addons = Object.keys(inst.config?.configurations || {}).map(
            (key) =>
              inst.billingPlan?.resources?.find((r) => {
                return (
                  r.key === getAddonKey(key, "addons") ||
                  r.key === getAddonKey(key, "os")
                );
              })
          );

          return (
            (+tariff.price || 0) +
            (addons.reduce((acc, a) => acc + a.price, 0) || 0)
          );
        }
        case "ione":
        case "cpanel": {
          const initialPrice =
            inst.billingPlan.products[inst.product]?.price ?? 0;

          return +inst.billingPlan.resources.reduce((prev, curr) => {
            if (
              curr.key === `drive_${inst.resources.drive_type?.toLowerCase()}`
            ) {
              return prev + (curr.price * inst.resources.drive_size) / 1024;
            } else if (curr.key === "ram") {
              return prev + (curr.price * inst.resources.ram) / 1024;
            } else if (inst.resources[curr.key]) {
              return prev + curr.price * inst.resources[curr.key];
            }
            return prev;
          }, initialPrice);
        }
      }
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
      return this.rates.find(
        (r) => r.to === currency && r.from === this.defaultCurrency
      )?.rate;
    },
    getPeriod(inst) {
      if (inst.type === "ione" && inst.billingPlan.kind === "DYNAMIC") {
        return "PayG";
      } else if (inst.resources.period) {
        const text = inst.resources.period > 1 ? "months" : "month";

        return `${inst.resources.period} ${text}`;
      }

      const period =
        inst.type === "ovh" ? inst.config.duration : this.getIonePeriod(inst);

      switch (period) {
        case "P1H":
          return "hourly";
        case "P1D":
          return "daily";
        case "P1M":
          return "monthly";
        case "P1Y":
          return "yearly";
        case "P2Y":
          return "2-yearly";
        case "PH":
          return "hybrid";
        default:
          return "unknown";
      }
    },
    getIonePeriod(inst) {
      const value = new Set();
      const day = 3600 * 24;
      const month = day * 30;
      const year = day * 365;

      Object.values(inst.billingPlan.products ?? {}).forEach(({ period }) => {
        if (inst.billingPlan.kind === "DYNAMIC") value.add("P1H");
        if (inst.billingPlan.kind !== "STATIC") return;

        if (+period === day) value.add("P1D");
        if (+period === month) value.add("P1M");
        if (+period === year) value.add("P1Y");
        if (+period === year * 2) value.add("P2Y");
      });

      return value.size > 1 ? "PH" : value.keys().next().value;
    },
    getCreationDate(inst) {
      return inst.data.creation ?? "unknown";
    },
    getExpirationDate(inst) {
      if (inst.data.next_payment_date)
        return this.date(inst.data.next_payment_date);
      return "-";
    },
    getService({ service }) {
      return (
        this.services.find(({ uuid }) => service === uuid)?.title ?? service
      );
    },
    getServiceProvider({ sp }) {
      return this.sp.find(({ uuid }) => uuid === sp)?.title;
    },
    getOSName(id, sp) {
      if (!id) return;
      return this.sp.find(({ uuid }) => uuid === sp)?.publicData.templates[id]
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
    getNamespace(id) {
      return this.namespaces.find((n) => n.uuid === id)?.title;
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
  },
  computed: {
    ...mapGetters("appSearch", { searchParam: "param", filter: "filter" }),
    services() {
      return this.$store.getters["services/all"];
    },
    instances() {
      const instances = this.items.filter((i) => {
        return Object.keys(this.filter || {}).every((key) => {
          let value;
          if (this.headersGetters[key]) {
            value = this.getValue(key, i);
          } else {
            value = getDeepObjectValue(i, key);
          }
          return compareSearchValue(
            value,
            this.filter[key],
            this.searchFields.find((f) => f.key === key)
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
        return searchKeysFull.some((key) => {
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
    accounts() {
      return this.$store.getters["accounts/all"];
    },
    instancesHeaders() {
      if (this.headers) return this.headers;
      const headers = [
        { text: "ID", value: "id" },
        { text: "Title", value: "title" },
        { text: "Service", value: "service" },
        { text: "Account", value: "access" },
        {
          text: "Group (NameSpace)",
          value: "access.namespace",
        },
        { text: "Due date", value: "dueDate" },
        { text: "Status", value: "state" },
        { text: "Tariff", value: "product" },
        { text: "Service provider", value: "sp" },
        { text: "Type", value: "type" },
        { text: "NCU price", value: "price" },
        { text: "Account price", value: "accountPrice" },
        { text: "Period", value: "period" },
        { text: "Email", value: "email" },
        { text: "Date", value: "date" },
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
        dueDate: this.getExpirationDate,
        sp: this.getServiceProvider,
        "access.namespace": (item) => this.getNamespace(item.access.namespace),
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
          items: this.getSearchKeyItems("access"),
          title: "Account",
        },
        {
          key: "access.namespace",
          items: this.getSearchKeyItems("access.namespace"),
          type: "select",
          title: "Namespace",
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
    instances() {
      this.fetchError = "";
    },
    searchFields: {
      deep: true,
      handler() {
        this.$store.commit("appSearch/setFields", this.searchFields);
      },
    },
  },
};
</script>
