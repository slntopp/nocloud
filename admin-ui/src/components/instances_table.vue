<template>
  <nocloud-table
    table-name="instances"
    class="mt-4"
    :value="value"
    :items="instances"
    :headers="headers"
    :loading="isLoading"
    :custom-sort="sortInstances"
    :footer-error="fetchError"
    @input="(value) => $emit('input', value)"
    :default-filtres="defaultFiltres"
    :filters-items="filterItems"
    :filters-values="selectedFilters"
    :show-select="showSelect"
    @input:filter="selectedFilters[$event.key] = $event.value"
  >
    <template v-slot:[`item.id`]="{ index }">
      {{ index + 1 }}
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <div class="d-flex justify-space-between">
        <router-link
          :to="{ name: 'Instance', params: { instanceId: item.uuid } }"
        >
          {{ item.title }}
        </router-link>
        <v-icon @click="goToInstance(item.uuid)">mdi-login</v-icon>
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
      <v-chip small :color="chipColor(item)">
        {{ getValue("state", item) }}
      </v-chip>
    </template>

    <template v-slot:[`item.product`]="{ item }">
      {{ getValue("product", item) }}
    </template>

    <template v-slot:[`item.price`]="{ item }">
      {{ getValue("price", item) }} {{ currency }}
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

    <template v-slot:[`item.resources.cpu`]="{ value }">
      {{ value }} {{ value > 1 ? "cores" : "core" }}
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
      <v-menu
        bottom
        open-on-hover
        v-else
        nudge-top="20"
        nudge-left="15"
        transition="slide-y-transition"
      >
        <template v-slot:activator="{ on, attrs }">
          <span v-bind="attrs" v-on="on">
            {{ item.state.meta.networking.public[0] }}
          </span>
        </template>

        <v-list dense>
          <v-list-item
            v-for="net of item.state.meta.networking.public"
            :key="net"
          >
            <v-list-item-title>{{ net }}</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </template>
  </nocloud-table>
</template>

<script>
import nocloudTable from "@/components/table.vue";
import api from "@/api";
import { getState } from "@/functions";

export default {
  name: "instances-table",
  components: { nocloudTable },
  props: {
    value: { type: Array, required: true },
    selected: { type: Object, default: null },
    showSelect: { type: Boolean, default: true },
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
    selectedFilters: {
      state: [],
      "billingPlan.title": [],
      service: [],
      type: [],
      sp: [],
      access: [],
      "access.namespace": [],
      period: [],
      product: [],
    },
  }),
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
    chipColor(item) {
      if (!item.state) return "error";
      const state =
        item.billingPlan.type === "ione"
          ? item.state.meta?.lcm_state_str
          : item.state.state;

      switch (state) {
        case "RUNNING":
          return "success";
        case "LCM_INIT":
        case "STOPPED":
          return "warning";
        case "SUSPENDED":
        case "UNKNOWN":
          return "error";
        default:
          return "blue-grey darken-2";
      }
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
      const { email } = this.getAccount(inst);

      return email ?? "-";
    },
    getPrice(inst) {
      switch (inst.type) {
        case "goget": {
          const key = `${inst.resources.period} ${inst.resources.id}`;

          return inst.billingPlan.products[key]?.price ?? 0;
        }
        case "ovh": {
          const key = `${inst.config.duration} ${inst.config.planCode}`;

          return inst.billingPlan.products[key]?.price ?? 0;
        }
        case "ione": {
          const initialPrice = inst.billingPlan.products[inst.product]?.price ?? 0;

          return +inst.billingPlan.resources
            .reduce((prev, curr) => {
              if (curr.key === `drive_${inst.resources.drive_type.toLowerCase()}`) {
                return (
                  prev + (curr.price * inst.resources.drive_size) / 1024
                );
              } else if (curr.key === "ram") {
                return (
                  prev + (curr.price * inst.resources.ram) / 1024
                );
              } else if (inst.resources[curr.key]) {
                return (
                  prev + curr.price * inst.resources[curr.key]
                );
              }
              return prev;
            }, initialPrice)?.toFixed(2);
        }
      }
    },
    getPeriod(inst) {
      if (inst.type === "ione" && inst.billingPlan.kind === "DYNAMIC") {
        return "PayG";
      } else if (inst.resources.period) {
        const text = inst.resources.period > 1 ? "months" : "month";

        return `${inst.resources.period} ${text}`;
      }

      const period = (inst.type === "ovh") ? inst.config.duration : this.getIonePeriod(inst);

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
      const day = 3600 * 24
      const month = day * 30;
      const year = day * 365;

      Object.values(inst.billingPlan.products ?? {}).forEach(({ period }) => {
        if (inst.billingPlan.kind === 'DYNAMIC') value.add("P1H");
        if (inst.billingPlan.kind !== 'STATIC') return;

        if (+period === day) value.add("P1D");
        if (+period === month) value.add("P1M");
        if (+period === year) value.add("P1Y");
        if (+period === year * 2) value.add("P2Y");
      });

      return (value.size > 1) ? "PH" : value.keys().next().value;
    },
    getCreationDate(inst) {
      return inst.data.creation ?? "unknown";
    },
    getExpirationDate(inst) {
      if (inst.type === "ovh") return inst.data.expiration;
      if (inst.type === "ione") return this.date(inst.data.last_monitoring);
      return "unknown";
    },
    getService({ service }) {
      return (
        "SRV_" + this.services.find(({ uuid }) => service === uuid)?.title ?? ""
      );
    },
    getServiceProvider({ sp }) {
      return this.sp.find(({ uuid }) => uuid === sp)?.title;
    },
    getOSName(id, sp) {
      if (!id) return;
      return this.sp.find(({ uuid }) => uuid === sp).publicData.templates[id]
        .name;
    },
    getTariff(item) {
      const {
        billingPlan,
        config: { planCode, duration },
      } = item;
      const key = `${duration} ${planCode}`;

      return billingPlan.products[key]?.title;
    },
    getNamespace(id) {
      return "NS_" + this.namespaces.find((n) => n.uuid === id)?.title;
    },
    getValue(key, item) {
      return this.headersGetters[key](item);
    },
    goToInstance(uuid) {
      api.settings.get(["app"]).then((res) => {
        const url = JSON.parse(res["app"]).url;
        window.open(`${url}/#/cloud/${uuid}`, "_blanc");
      });
    },
  },
  computed: {
    services() {
      return this.$store.getters["services/all"];
    },
    instances() {
      const searchKeys = [
        "title",
        "uuid",
        "billingPlan.title",
        "price",
        "state.meta.networking.public.0",
      ];
      const instances = this.items.filter((i) => {
        for (const key of Object.keys(this.selectedFilters)) {
          if (this.selectedFilters[key].length === 0) {
            continue;
          }

          if (!this.headersGetters[key]) {
            let val = i;
            key.split(".").forEach((subkey) => {
              val = val[subkey];
            });
            if (!this.selectedFilters[key].includes(val)) {
              return false;
            }
          } else if (
            !this.selectedFilters[key].includes(this.getValue(key, i))
          ) {
            return false;
          }
        }
        return true;
      });
      if (!this.searchParam) {
        return instances;
      }

      return instances.filter((item) => {
        return searchKeys.some((key) => {
          let tempItem = item;
          if (this.headersGetters[key]) {
            tempItem = this.getValue(key, tempItem);
          } else {
            key.split(".").forEach((subkey) => (tempItem = tempItem?.[subkey]));
          }
          return tempItem?.toString()?.startsWith(this.searchParam);
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
    headers() {
      const headers = [
        { text: "ID", value: "id" },
        { text: "Title", value: "title" },
        { text: "Service", value: "service", customFilter: true },
        { text: "Account", value: "access", customFilter: true },
        {
          text: "Group (NameSpace)",
          value: "access.namespace",
          customFilter: true,
        },
        { text: "Due date", value: "dueDate" },
        { text: "Status", value: "state", customFilter: true },
        { text: "Tariff", value: "product", customFilter: true },
        { text: "Service provider", value: "sp", customFilter: true },
        { text: "Type", value: "type", customFilter: true },
        { text: "Price", value: "price" },
        { text: "Period", value: "period", customFilter: true },
        { text: "Email", value: "email" },
        { text: "Date", value: "date" },
        { text: "UUID", value: "uuid" },
        { text: "Price model", value: "billingPlan.title", customFilter: true },
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
        product: (item) => item.product ?? this.getTariff(item) ?? "custom",
        price: this.getPrice,
        period: this.getPeriod,
        date: this.getCreationDate,
        dueDate: this.getExpirationDate,
        sp: this.getServiceProvider,
        "access.namespace": (item) => this.getNamespace(item.access.namespace),
        "resources.ram": (item) => +item?.resources?.ram / 1024,
        "resources.drive_size": (item) => +item?.resources?.drive_size / 1024,
        "config.template_id": (item) =>
          this.getOSName(item?.config?.template_id, item.sp),
      };
    },
    isLoading() {
      return this.$store.getters["services/isLoading"];
    },
    searchParam() {
      return this.$store.getters["appSearch/param"];
    },
    currency() {
      return this.$store.getters["currencies/default"];
    },
    filterItems() {
      return {
        state: ["RUNNING", "LCM_INIT", "STOPPED", "SUSPENDED", "UNKNOWN"],
        type: ["ione", "ovh", "custom", "opensrs", "goget"],
        "billingPlan.title": this.priceModelItems,
        service: this.serviceItems,
        sp: this.spItems,
        period: this.periodItems,
        access: this.accountsItems,
        "access.namespace": this.namespacesItems,
        product: this.productItems,
      };
    },
    priceModelItems() {
      return new Set(this.items.map((i) => i.billingPlan?.title));
    },
    serviceItems() {
      return new Set([
        ...this.$store.getters["services/all"].map((i) => "SRV_" + i.title),
      ]);
    },
    spItems() {
      const instancesSP = this.items.map((i) => i.sp);

      return new Set(
        this.sp
          .filter((sp) => instancesSP.includes(sp.uuid))
          .map((sp) => sp.title)
      );
    },
    periodItems() {
      const periods = this.items.map((i) => this.getValue("period", i));

      return new Set(periods);
    },
    productItems() {
      const products = this.items.map((i) => this.getValue("product", i));

      return new Set(products);
    },
    accountsItems() {
      const accounts = this.accounts.map((a) => a.title);

      return new Set(accounts);
    },
    namespacesItems() {
      const namespaces = this.namespaces.map((n) => "NS_" + n.title);

      return new Set(namespaces);
    },
  },
  watch: {
    instances() {
      this.fetchError = "";
    },
  },
};
</script>