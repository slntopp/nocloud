<template>
  <div class="services pa-4">
    <div v-if="isFiltered" class="page__title">
      Used in
      {{ $route.query.provider ? '"' + $route.query.provider + '"' : "" }}
      service provider
      <v-btn small :to="{ name: 'Services' }"> clear </v-btn>
    </div>
    <div class="pb-4 buttons">
      <div class="buttons__inline">
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
      <div>
        <v-text-field
          v-model="searchParam"
          class="service-search mr-2"
          hide-details
          prepend-inner-icon="mdi-magnify"
          placeholder="Search..."
          background-color="background-light"
          dence
          rounded
        />
      </div>
    </div>

    <nocloud-table
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

      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length" style="padding: 0">
          <div v-for="(itemService, index) in services" :key="index">
            <v-expansion-panels
              inset
              multiple
              v-model="opened[index]"
              v-if="item.uuid == itemService.uuid"
              @change="clickColumn(index)"
            >
              <v-expansion-panel
                style="background: var(--v-background-light-base)"
                v-for="(group, i) in itemService.instancesGroups"
                :key="i"
              >
                <v-expansion-panel-header>
                  {{ group.title }} | Type: {{ group.type }} -
                  {{ titleSP(group) }}
                </v-expansion-panel-header>
                <v-expansion-panel-content
                  style="background: var(--v-background-base)"
                >
                  <v-row>
                    <serveces-instances-item
                      v-for="(instance, index) in group.instances"
                      :key="index"
                      :title="instance.title"
                      :state="instance.state ? instance.state.state : 'UNKNOWN'"
                      :cpu="instance.resources.cpu"
                      :drive_type="instance.resources.drive_type"
                      :drive_size="instance.resources.drive_size"
                      :ram="instance.resources.ram"
                      :hash="instance.hash"
                      :index="index"
                      :chipColor="chipColor"
                      :hashTrim="hashTrim"
                    />
                  </v-row>
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
import servecesInstancesItem from "@/components/serveces_instances_item.vue";
import ConfirmDialog from "../components/confirmDialog.vue";

export default {
  name: "Services-view",
  components: {
    nocloudTable,
    servecesInstancesItem,
    ConfirmDialog,
  },
  data: () => ({
    headers: [
      { text: "title", value: "title" },
      { text: "status", value: "status" },
      {
        text: "UUID",
        align: "start",
        value: "uuid",
      },
      { text: "hash", value: "hash" },
    ],
    copyed: -1,
    opened: {},
    expanded: [],
    selected: [],
    fetchError: "",
    searchParam: "",
  }),
  computed: {
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
      if (this.searchParam) {
        return this.services.filter((service) => {
          const ips = this.getPublicIpsFromService(service);
          const isItIpExists = ips.find((ip) => ip.includes(this.searchParam));
          const isTitleIncludes = service.title
            .toLowerCase()
            .includes(this.searchParam.toLowerCase());
          return isTitleIncludes || isItIpExists;
        });
      }
      return this.services;
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
  },
  created() {
    this.$store.dispatch("servicesProviders/fetch");
    this.fetchServices();
  },
  methods: {
    titleSP(group) {
      const data = this.servicesProviders.find((el) =>
        el.uuid == group?.sp
      );
      return data?.title || 'not found';
    },
    fetchServices() {
      this.$store
        .dispatch("services/fetch")
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
    treeview(item) {
      const result = [];
      let index = 0;
      for (const [name, group] of Object.entries(item.instancesGroups)) {
        const temp = {};
        temp.id = ++index;
        temp.name = name;
        const childs = [];
        for (const [ind, inst] of Object.entries(group.instances)) {
          ind;
          const temp = {};
          temp.id = ++index;
          temp.name = inst.title;
          childs.push(temp);
        }
        temp.children = childs;
        result.push(temp);
      }
      return result;
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
        alert('Clipboard is not supported!');
      }
    },
    chipColor(state) {
      const dict = {
        INIT: "orange darken-2",
        UP: "green darken-2",
        DEL: "gray darken-2",
        RUNNING: "green darken-2",
        UNKNOWN: "red darken-2",
        STOPPED: "orange darken-2",
      };
      return dict[state] ?? "blue-grey darken-2";
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
              this.$store.dispatch("services/fetch");

              const ending = deletePromices.length == 1 ? "" : "s";
              this.showSnackbar({
                message: `Service${ending} deleted successfully.`,
              });
            } else {
              this.showSnackbar({
                message: `Can’t delete Services Provider: Has Services deployed.`,
                route: { name: "Home" },
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
    getPublicIpsFromService(service) {
      const ips = [];
      service.instancesGroups.forEach((group) =>
        group.instances.forEach((instance) =>
          instance.state.meta.networking.public.forEach((p) => {
            if (typeof p === "string") {
              ips.push(p);
            }
          })
        )
      );
      return ips;
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });
  },
  watch: {
    services() {
      this.fetchError = '';
    }
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

.buttons {
  display: flex;
  justify-content: space-between;
}

.service-search {
  margin-top: 0px;
  font-size: 1.2rem;
  padding-top: 0px;
}

.v-expansion-panel-content .v-expansion-panel-content__wrap {
  padding: 22px;
}
</style>
