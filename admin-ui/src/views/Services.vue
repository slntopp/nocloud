<template>
  <div class="services pa-4">
    <div v-if="isFiltered" class="page__title">
      Used in
      {{ $route.query.provider ? '"' + $route.query.provider + '"' : "" }}
      service provider
      <v-btn small :to="{ name: 'Services' }"> clear </v-btn>
    </div>
    <div class="buttons__inline pb-4">
      <v-btn
        color="background-light"
        class="mr-2"
        :to="{ name: 'Service create' }"
      >
        create
      </v-btn>
      <v-btn
        color="background-light"
        class="mr-2"
        @click="deleteSelectedServices"
      >
        delete
      </v-btn>
    </div>

    <nocloud-table
      :items="services"
      :headers="headers"
      :expanded.sync="expanded"
      @input="(v) => (selected = v)"
      :value="selected"
      show-expand
      :loading="isLoading"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.hash`]="{ item, index }">
        <v-btn icon @click="addToClipboard(item.hash, index)">
          <v-icon v-if="copyed == index"> mdi-check </v-icon>
          <v-icon v-else> mdi-content-copy </v-icon>
        </v-btn>
        {{ hashTrim(item.hash) }}
      </template>

      <template v-slot:[`item.title`]="slotData">
        <router-link
          :to="{ name: 'Service', params: { serviceId: slotData.item.uuid } }"
        >
          {{ slotData.item.title }}
        </router-link>
      </template>

      <template v-slot:[`item.status`]="{ value }">
        <v-chip small :color="chipColor(value)">
          {{ value }}
        </v-chip>
      </template>
      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length" style="padding: 0">
          <v-expansion-panels inset v-model="opened" multiple>
            <v-expansion-panel
              style="background: var(--v-background-light-base)"
              v-for="(group, title) in item.instancesGroups"
              :key="title"
            >
              <v-expansion-panel-header>
                {{ title }} | Type: {{ group.type }} -
                {{ titleSP(item) }}
                </v-expansion-panel-header
              >
              <v-expansion-panel-content
                style="background: var(--v-background-base)"
              >
                <v-row>
                  <serveces-instances-item
                    v-for="(elem, index) in group.instances"
                    :key="index"
                    :title="elem.title"
                    :state="elem.state ? elem.state.state : 'UNKNOWN'"
                    :cpu="elem.resources.cpu"
                    :drive_type="elem.resources.drive_type"
                    :drive_size="elem.resources.drive_size"
                    :ram="elem.resources.ram"
                    :hash="elem.hash"
                    :index="index"
                    :chipColor="chipColor"
                    :hashTrim="hashTrim"
                  >
                  </serveces-instances-item>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
        </td>
      </template>
    </nocloud-table>
  </div>
</template>

<script>
import noCloudTable from "@/components/table.vue";
import servecesInstancesItem from "@/components/serveces_instances_item.vue";
import api from "@/api";

export default {
  name: "Services-view",
  components: {
    "nocloud-table": noCloudTable,
    servecesInstancesItem,
  },
  data: () => ({
    opened: [0],
    headers: [
      { text: "title", value: "title" },
      { text: "status", value: "status" },
      {
        text: "UUID",
        align: "start",
        value: "uuid",
      },
      {
        text: "hash",
        value: "hash",
      },
    ],
    copyed: -1,
    expanded: [],
    selected: [],
    fetchError: "",
  }),
  computed: {
    services() {
      const items = this.$store.getters["services/all"];
      if (this.isFiltered) {
        return items.filter((item) => {
          return this.$route.query["items[]"].includes(item.uuid);
        });
      }
      console.log(items)
      return items;
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
  methods: {
    titleSP(item) {
      for (const group of Object.values(item.instancesGroups)) {
        let data = this.servicesProviders.find(
          (el) => el.uuid == item.provisions[group.uuid]
        );
        return data.title;
      }
    },
    // treeview(item) {
    //   const result = [];
    //   let index = 0;
    //   for (const [name, group] of Object.entries(item.instancesGroups)) {
    //     const temp = {};
    //     temp.id = ++index;
    //     let sp = this.servicesProviders.find(
    //       (el) => el.uuid == item.provisions[group.uuid]
    //     );
    //     temp.name = name;
    //     if (sp) {
    //       temp.name += ` [location]: ${sp.title}`;
    //     }
    //     const childs = [];
    //     for (const [ind, inst] of Object.entries(group.instances)) {
    //       ind;
    //       const temp = {};
    //       temp.id = ++index;
    //       temp.name = inst.title;
    //       if (sp) {
    //         temp.name += ` [state]: ${inst.state.meta.lcm_state_str}`;
    //       }
    //       childs.push(temp);
    //     }
    //     temp.children = childs;
    //     result.push(temp);
    //     console.log(result);
    //   }
    //   return result;
    // },
    hashTrim(hash) {
      if (hash) return hash.slice(0, 8) + "...";
      else return "XXXXXXXX...";
    },
    addToClipboard(text, index) {
      navigator.clipboard
        .writeText(text)
        .then(() => {
          this.copyed = index;
        })
        .catch((res) => {
          console.error(res);
        });
    },
    clickColumn(slotData) {
      // const indexRow = slotData.index;
      const indexExpanded = this.expanded.findIndex((i) => i === slotData.item);
      if (indexExpanded > -1) {
        this.expanded.splice(indexExpanded, 1);
      } else {
        this.expanded.push(slotData.item);
      }
    },
    chipColor(state) {
      const dict = {
        init: "orange darken-2",
        up: "green darken-2",
        del: "gray darken-2",
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
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });
    this.opened.push(0);
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
</style>
