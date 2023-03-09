<template>
  <nocloud-table
    table-name="instances"
    class="mt-4"
    :value="value"
    :custom-sort="sortInstances"
    :items="instances"
    :headers="headers"
    :loading="isLoading"
    :footer-error="fetchError"
    @input="(value) => $emit('input', value)"
    :filter="filters"
  >
    <template v-slot:[`item.id`]="{ index }">
      {{ index + 1 }}
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <router-link
        :to="{ name: 'Instance', params: { instanceId: item.uuid } }"
      >
        {{ item.title }}
      </router-link>
    </template>

    <template v-slot:[`item.access`]="{ item }">
      <router-link :to="{ name: 'Account', params: { accountId: getAccount(item)?.uuid } }">
        {{ getAccount(item)?.title }}
      </router-link>
    </template>

    <template v-slot:[`item.email`]="{ item }">
      {{ getEmail(item) }}
    </template>

    <template v-slot:[`item.state`]="{ item }">
      <v-chip small :color="chipColor(item)">
        {{ getState(item) }}
      </v-chip>
    </template>

    <template v-slot:[`item.product`]="{ item, value }">
      {{ value ?? getTariff(item) ?? "custom" }}
    </template>

    <template v-slot:[`item.price`]="{ item }">
      {{ getPrice(item) }} {{ currency }}
    </template>

    <template v-slot:[`item.period`]="{ item }">
      {{ getPeriod(item) }}
    </template>

    <template v-slot:[`item.date`]="{ item }">
      {{ getCreationDate(item) }}
    </template>

    <template v-slot:[`item.dueDate`]="{ item }">
      {{ getExpirationDate(item) }}
    </template>

    <template v-slot:[`item.service`]="{ item, value }">
      <router-link :to="{ name: 'Service', params: { serviceId: value } }">
        {{ "SRV_" + getService(item) }}
      </router-link>
    </template>

    <template v-slot:[`item.sp`]="{ item, value }">
      <router-link :to="{ name: 'ServicesProvider', params: { uuid: value } }">
        {{ getServiceProvider(item) }}
      </router-link>
    </template>

    <template v-slot:[`item.access.namespace`]="{ item }">
      <router-link
        :to="{
          name: 'NamespacePage',
          params: { namespaceId: item.access.namespace },
        }"
      >
        {{ getNamespace(item.access.namespace) }}
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

    <template v-slot:[`item.resources.ram`]="{ value }">
      {{ value / 1024 }} GB
    </template>

    <template v-slot:[`item.resources.drive_size`]="{ value }">
      {{ value / 1024 }} GB
    </template>

    <template v-slot:[`item.config.template_id`]="{ item, value }">
      {{ getOSName(value, item.sp) }}
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
import { filterArrayIncludes } from "@/functions.js";

export default {
  name: "instances-table",
  components: { nocloudTable },
  props: {
    value: { type: Array, required: true },
    type: { type: String, required: true },
    column: { type: String, required: true },
    filters: { type: Array, required: true },
    selected: { type: Object, required: true },
    getState: { type: Function, required: true },
    changeFilters: { type: Function, required: true },
  },
  data: () => ({ fetchError: "" }),
  methods: {
    changeIcon() {
      setTimeout(() => {
        const headers = document.querySelectorAll(".groupable");

        headers.forEach(({ firstElementChild, children }) => {
          if (!children[1]?.className.includes("group-icon")) {
            const element = document.querySelector(".group-icon");
            const icon = element.cloneNode(true);

            firstElementChild.after(icon);
            icon.style = "display: inline-flex";

            icon.addEventListener("click", (e) => {
              const menu = document.querySelector(".v-menu__content");
              const { x, y } = icon.getBoundingClientRect();

              if (menu.className.includes("menuable__content__active")) return;

              this.$emit("changeColumn", firstElementChild.innerText);
              element.dispatchEvent(new Event("click"));
              e.stopPropagation();

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
    fetchServices() {
      this.$store
        .dispatch("services/fetch")
        .then(() => {
          this.fetchError = "";
          this.$emit("getHeaders", this.headers);

          this.changeFilters();
          this.changeIcon();
        })
        .catch((err) => {
          console.log(err);
          this.fetchError = "Can't reach the server";
          if (err.response) {
            this.fetchError += `: [ERROR]: ${err.response.data.message}`;
          } else {
            this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
          }
        });
    },
    sortInstances(items, sortBy, sortDesc) {
      return items.sort((a, b) => {
        for (let i = 0; i < sortBy.length; i++) {
          if (sortDesc[i]) [a, b] = [b, a];

          let valueA = a;
          let valueB = b;

          sortBy[i].split(".").forEach((key) => {
            valueA = valueA[key];
            valueB = valueB[key];
          });

          if (sortBy[i] === "state") {
            return this.getState(a) < this.getState(b);
          }

          if (sortBy[i] === "service") {
            return this.getService(a) < this.getService(b);
          }

          if (typeof valueA === "string") {
            return valueA.toLowerCase() < valueB.toLowerCase();
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
      const state = (item.billingPlan.type === 'ione')
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
      const { access: { namespace } } = this.namespaces.find(({ uuid }) =>
        uuid === access.namespace) ??
        { access: {} };

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
        case 'ione': {
          const initialPrice = inst.billingPlan.products[inst.product]?.price ?? 0

          return +inst.billingPlan.resources.reduce((prev, curr) => {
            if (curr.key === `drive_${inst.resources.drive_type.toLowerCase()}`) {
              return prev + curr.price / curr.period * 3600 * 24 * 30 * inst.resources.drive_size / 1024;
            } else if (curr.key === "ram") {
              return prev + curr.price / curr.period * 3600 * 24 * 30 * inst.resources.ram / 1024;
            } else if (inst.resources[curr.key]) {
              return prev + curr.price / curr.period * 3600 * 24 * 30 * inst.resources[curr.key];
            }
            return prev;
          }, initialPrice)?.toFixed(2);
        }
      }
    },
    getPeriod(inst) {

      if (inst.billingPlan.kind === "STATIC") return "monthly";
      else if (inst.type === "ione") return "PayG";

      else if (inst.resources.period) {
        const text = inst.resources.period > 1 ? "months" : "month";

        return `${inst.resources.period} ${text}`;
      }

      switch (inst.duration) {
        case "P1M":
          return "monthly";
        case "P1Y":
          return "yearly";

        default:
          return "unknown";
      }
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
      return this.services.find(({ uuid }) => service === uuid)?.title ?? "";
    },
    getServiceProvider({ sp }) {
      return this.sp.find(({ uuid }) => uuid === sp).title;
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
  },
  created() {
    this.fetchServices();
  },
  computed: {
    services() {
      return this.$store.getters["services/all"];
    },
    instances() {
      const instances = this.$store.getters["services/getInstances"]?.filter(
        (inst) => {
          const result = [];

          Object.entries(this.selected).forEach(([key, filters]) => {
            const { value, text } =
              this.headers.find(({ text }) => text === key) || {};
            let filter = inst;

            if (!value) return [];
            value.split(".").forEach((key) => {
              filter = filter[key];
            });

            switch (text) {
              case "Service":
                filter = this.getService(inst);
                break;
              case "Status":
                filter = this.getState(inst).toLowerCase();
                break;
              case "OS":
                filter = this.getOSName(inst.config.template_id, inst.sp);
                break;
              case "RAM":
              case "Disk":
                filter = filter / 1024;
            }

            if (filters.includes(`${filter}`)) result.push(true);
            else result.push(false);
          });

          return result.every((el) => el);
        }
      );

      const filtered = filterArrayIncludes(instances, {
        keys: ["uuid", "service", "title", "billingPlan", "state"],
        value: this.searchParam,
        params: {
          billingPlan: "title",
          service: this.getService,
          state: this.getState,
        },
      });

      if (this.type === "all") return filtered;
      return filtered?.filter(({ type }) => type === this.type);
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
        { text: "Service", value: "service", class: "groupable" },
        { text: "Account", value: "access" },
        { text: "Group (NameSpace)", value: "access.namespace" },
        { text: "Due date", value: "dueDate" },
        { text: "Status", value: "state", class: "groupable" },
        { text: "Tariff", value: "product" },
        { text: "Service provider", value: "sp" },
        { text: "Type", value: "type" },
        { text: "Price", value: "price" },
        { text: "Period", value: "period" },
        { text: "Email", value: "email" },
        { text: "Date", value: "date" },
        { text: "UUID", value: "uuid" },
        { text: "Price model", value: "billingPlan.title", class: "groupable" },
      ];

      if (this.type !== "all") headers.splice(1, 1);
      switch (this.type) {
        case "ione":
          headers.push(
            { text: "IP", value: "state.meta.networking" },
            { text: "CPU", value: "resources.cpu", class: "groupable" },
            { text: "RAM", value: "resources.ram", class: "groupable" },
            { text: "Disk", value: "resources.drive_size", class: "groupable" },
            { text: "OS", value: "config.template_id", class: "groupable" }
          );
          break;
        case "ovh":
          headers.push({ text: "IP", value: "state.meta.networking" });
          break;
        case "goget":
        case "opensrs":
          headers.push({
            text: "Domain",
            value: "resources.domain",
            class: "groupable",
          });

          if (this.type === "opensrs") break;
          else
            headers.push(
              { text: "DCV", value: "resources.dcv", class: "groupable" },
              {
                text: "Approver email",
                value: "resources.approver_email",
                class: "groupable",
              }
            );
      }

      return headers;
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
  },
  watch: {
    instances() {
      this.fetchError = "";
    },
    headers() {
      const headers = document.querySelectorAll(".groupable");

      headers.forEach(({ children }) => {
        children[1].remove();
      });

      this.changeIcon();
      this.changeFilters();
    },
  },
};
</script>
