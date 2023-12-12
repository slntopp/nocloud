<template>
  <div class="services pa-4">
    <div v-if="isFiltered" class="page__title">
      Used in
      {{ $route.query.provider ? '"' + $route.query.provider + '"' : "" }}
      service provider
      <v-btn small :to="{ name: 'Services' }"> clear </v-btn>
    </div>
    <div class="pb-8 pt-4 buttons__inline">
      <v-btn
        color="background-light"
        class="mr-2"
        :to="{ name: 'Service create' }"
      >
        create
      </v-btn>
      <confirm-dialog
        :disabled="selected.length < 1"
        @confirm="deleteSelectedServices"
      >
        <v-btn
          :disabled="selected.length < 1"
          color="background-light"
          class="mr-2"
        >
          delete
        </v-btn>
      </confirm-dialog>
      <v-btn
        color="background-light"
        class="mr-2"
        @click="upServices"
        :disabled="selected.length < 1"
      >
        UP
      </v-btn>
      <v-btn
        color="background-light"
        class="mr-2"
        @click="downServices"
        :disabled="selected.length < 1"
      >
        DOWN
      </v-btn>
    </div>

    <nocloud-table
      table-name="services"
      show-expand
      v-model="selected"
      :items="filteredServices"
      :headers="headers"
      :loading="isLoading"
      :expanded.sync="expanded"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.hash`]="{ item, index }">
        <v-btn icon @click="addToClipboard(item.hash, index)">
          <v-icon v-if="copyed == index"> mdi-check </v-icon>
          <v-icon v-else> mdi-content-copy </v-icon>
        </v-btn>
        {{ hashTrim(item.hash) }}
      </template>

      <template v-slot:[`item.title`]="{ item }">
        <router-link
          :to="{ name: 'Service', params: { serviceId: item.uuid } }"
        >
          {{ item.title }}
        </router-link>
      </template>

      <template v-slot:[`item.status`]="{ value }">
        <v-chip small :color="chipColor(value)">
          {{ value }}
        </v-chip>
      </template>
      <template v-slot:[`item.access`]="{ item }">
        <v-chip :color="accessColor(item.access?.level)">
          {{ getName(item.access?.namespace) }} ({{
            item.access?.level ?? "NONE"
          }})
        </v-chip>
      </template>

      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length" style="padding: 0">
          <div v-for="(itemService, index) in services" :key="index">
            <v-expansion-panels
              inset
              multiple
              v-model="opened[index]"
              v-if="item.uuid == itemService.uuid"
            >
              <v-expansion-panel
                style="background: var(--v-background-light-base)"
                v-for="(group, i) in itemService.instancesGroups"
                :key="i"
                :disabled="!group.instances.length"
              >
                <v-expansion-panel-header>
                  {{ group.title }} | Type: {{ group.type }} -
                  {{ titleSP(group) }}
                  <v-chip
                    small
                    class="instance-group-status"
                    :color="instanceCountColor(group)"
                  >
                    {{ group.instances.length }}
                  </v-chip>
                </v-expansion-panel-header>
                <v-expansion-panel-content
                  style="background: var(--v-background-base)"
                >
                  <service-instances-item
                    :instances="group.instances"
                    :spId="group.sp"
                    :type="group.type"
                  />
                </v-expansion-panel-content>
              </v-expansion-panel>
            </v-expansion-panels>
          </div>
          <!-- <v-card class="pa-4" color="background">
            <v-treeview :items="treeview(item)"> </v-treeview>
          </v-card> -->
        </td>
      </template>
    </nocloud-table>
  </div>
</template>

<script>
import api from "@/api";
import nocloudTable from "@/components/table.vue";
import serviceInstancesItem from "@/components/service_instances_item.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import search from "@/mixins/search.js";
import snackbar from "@/mixins/snackbar.js";
import {
  compareSearchValue,
  filterArrayByTitleAndUuid,
  getDeepObjectValue,
} from "@/functions";
import { mapGetters } from "vuex";

export default {
  name: "Services-view",
  components: {
    nocloudTable,
    serviceInstancesItem,
    confirmDialog,
  },
  mixins: [snackbar, search({name:"services"})],
  data: () => ({
    headers: [
      { text: "Title", value: "title" },
      { text: "Status", value: "status" },
      { text: "UUID", value: "uuid", align: "start" },
      { text: "Hash", value: "hash" },
      { text: "Access", value: "access" },
    ],
    copyed: -1,
    opened: {},
    expanded: [],
    selected: [],
    fetchError: "",

    accessColorsMap: {
      ROOT: "info",
      ADMIN: "success",
      MGMT: "warning",
      READ: "gray",
      NONE: "error",
    },
    stateColorMap: {
      INIT: "orange darken-2",
      SUS: "orange darken-2",
      UP: "green darken-2",
      DEL: "gray darken-2",
      RUNNING: "green darken-2",
      UNKNOWN: "red darken-2",
      STOPPED: "orange darken-2",
    },
  }),
  computed: {
    ...mapGetters("appSearch", { searchParam: "param", filter: "filter" }),
    services() {
      const items = this.$store.getters["services/all"];

      if (this.isFiltered) {
        return items.filter((item) =>
          this.$route.query["items[]"].includes(item.uuid)
        );
      }
      return items;
    },
    filteredServices() {
      const services = this.services.filter((s) =>
        Object.keys(this.filter).every((key) => {
          const data = getDeepObjectValue(s, key);

          return compareSearchValue(
            data,
            this.filter[key],
            this.searchFields.find((f) => f.key === key)
          );
        })
      );

      if (this.searchParam) {
        return filterArrayByTitleAndUuid(services, this.searchParam);
      }
      return services;
    },
    isFiltered() {
      return this.$route.query.filter == "uuid" && this.$route.query["items[]"];
    },
    isLoading() {
      return this.$store.getters["services/isLoading"];
    },
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    searchFields() {
      return [
        {
          type: "input",
          title: "Title",
          key: "title",
        },
        {
          items: Object.keys(this.stateColorMap),
          type: "select",
          title: "Status",
          key: "status",
        },
        {
          key: "access.level",
          items: Object.keys(this.accessColorsMap),
          type: "select",
          title: "Access",
        },
        {
          key: "hash",
          type: "input",
          title: "Hash",
        },
      ];
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetch",{anonymously:true});
    this.$store.dispatch("namespaces/fetch");
    this.fetchServices();
  },
  methods: {
    titleSP(group) {
      const data = this.servicesProviders.find((el) => el.uuid == group?.sp);
      return data?.title || "not found";
    },
    getName(namespace) {
      return (
        this.namespaces.find(({ uuid }) => namespace === uuid)?.title ?? ""
      );
    },
    fetchServices() {
      this.$store
        .dispatch("services/fetch",{showDeleted:true})
        .then(() => {
          this.fetchError = "";
        })
        .catch((err) => {
          console.log(`err`, err);
          this.fetchError = "Can't reach the server";
          if (err.response) {
            this.fetchError += `: [ERROR]: ${err.response.data.message}`;
          } else {
            this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
          }
        });
    },
    hashTrim(hash) {
      if (hash) return hash.slice(0, 8) + "...";
      else return "XXXXXXXX...";
    },
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((res) => {
            console.error(res);
          });
      } else {
        alert("Clipboard is not supported!");
      }
    },
    chipColor(state) {
      return this.stateColorMap[state] ?? "blue-grey darken-2";
    },
    accessColor(level) {
      return this.accessColorsMap[level];
    },

    instanceCountColor(group) {
      if (group.instances.length) {
        return this.chipColor(group.status);
      }
      return this.chipColor("DEL");
    },

    deleteSelectedServices() {
      if (this.selected.length > 0) {
        const deletePromices = this.selected.map((el) =>
          api.services.delete(el.uuid)
        );
        Promise.all(deletePromices)
          .then((res) => {
            if (res.every((el) => el.result)) {
              console.log("all ok");
              this.$store.dispatch("services/fetch",{showDeleted:true});

              const ending = deletePromices.length == 1 ? "" : "s";
              this.showSnackbar({
                message: `Service${ending} deleted successfully.`,
              });
            } else {
              this.showSnackbarError({
                message: `Can’t delete Services Provider: Has Services deployed.`,
              });
            }
          })
          .catch((err) => {
            if (err.response.status >= 500 || err.response.status < 600) {
              const opts = {
                message: `Service Unavailable: ${
                  err?.response?.data?.message ?? "Unknown"
                }.`,
                timeout: 0,
              };
              this.showSnackbarError(opts);
            } else {
              const opts = {
                message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
              };
              this.showSnackbarError(opts);
            }
          });
      }
    },
    upServices() {
      const servicesToUp = this.getSelectedServicesWithStatus("INIT");
      if (!servicesToUp) {
        return;
      }
      Promise.all(
        servicesToUp.map((s) => {
          return api.services.up(s.uuid);
        })
      );

      this.fetchAfterTimeout();
    },
    downServices() {
      const servicesToDown = this.getSelectedServicesWithStatus("UP");
      if (!servicesToDown) {
        return;
      }
      Promise.all(
        servicesToDown.map((s) => {
          return api.services.down(s.uuid);
        })
      );
      this.fetchAfterTimeout();
    },
    getSelectedServicesWithStatus(status) {
      const services = this.selected.filter(
        (service) => service.status === status
      );
      if (!services.length) {
        return;
      }
      return services;
    },
    fetchAfterTimeout(ms = 500) {
      setTimeout(this.fetchServices, ms);
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });

    this.$store.commit("appSearch/setFields", this.searchFields);
  },
  watch: {
    services() {
      this.fetchError = "";
    },
  },
};
</script>

<style>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}

.v-expansion-panel-content .v-expansion-panel-content__wrap {
  padding: 22px;
}

.instance-group-status {
  max-width: 30px;
  align-items: center;
  margin-left: 25px;
}
</style>
