<template>
  <div class="zone-manager pa-4 flex-wrap">
    <div class="buttons__inline pb-4">
      <div class="page__title">
        <router-link :to="{ name: 'DNS manager' }"> zones</router-link> /
        {{ $route.params.dnsname }}
      </div>

      <v-menu
        offset-y
        transition="slide-y-transition"
        bottom
        :close-on-content-click="false"
        v-model="newHost.modalVisible"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn color="background-light" class="mr-2" v-bind="attrs" v-on="on">
            create
          </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row>
            <v-col cols="12">
              <v-text-field
                dense
                hide-details
                v-model="newHost.location"
                label="location"
              >
              </v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-select
                dense
                v-model="newHost.type"
                label="type"
                :items="Object.keys(types)"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                dense
                hide-details
                v-model="newHost.value"
                label="value"
              >
              </v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                dense
                hide-details
                v-model="newHost.ttl"
                label="ttl"
                type="number"
                @keypress.enter="createZone"
                min="0"
              >
              </v-text-field>
            </v-col>
          </v-row>
          <v-row justify="end">
            <v-col class="d-flex justify-end">
              <v-btn :loading="newHost.loading" @click="createZone">
                send
              </v-btn>
            </v-col>
          </v-row>
        </v-card>
      </v-menu>

      <v-btn
        color="background-light"
        class="mr-8"
        :disabled="selected.length < 1"
        @click="deleteSelectedZones"
        :loading="deleteLoading"
      >
        delete
      </v-btn>
    </div>

    <zone-hosts-table v-model="selected"> </zone-hosts-table>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script>
import hosts from "@/components/single_zone_table.vue";
import api from "@/api.js";

import snackbar from "@/mixins/snackbar.js";

export default {
  name: "zone-manager",
  components: {
    "zone-hosts-table": hosts,
  },
  mixins: [snackbar],
  data() {
    return {
      selected: [],
      deleteLoading: false,
      newHost: {
        location: "",
        type: "",
        value: "",
        ttl: "",
        loading: false,
        modalVisible: false,
      },
      types: {
        A: "ip",
        AAAA: "ip",
        CNAME: "host",
        TXT: "text",
      },
    };
  },
  methods: {
    createZone() {
      if (this.newHost.location.length < 1) return;
      if (this.newHost.type.length < 1) return;
      if (this.newHost.value.length < 1) return;

      const host = {
        [this.types[this.newHost.type]]: this.newHost.value,
        ttl: this.newHost.ttl,
      };

      const location = this.newHost.location;
      const type = this.newHost.type.toLowerCase();

      const template = JSON.parse(JSON.stringify(this.tableData));
      if (template[location]) {
        if (template[location][type]) {
          template[location][type].push(host);
        } else {
          template[location][type] = [host];
        }
      } else {
        template[location] = {};
        template[location][type] = [host];
      }
      const result = {
        locations: template,
        name: this.$route.params.dnsname,
      };
      api.dns
        .setZone(result)
        .then(() => {
          (this.newHost = {
            location: "",
            type: "",
            value: "",
            ttl: "",
            loading: false,
            modalVisible: false,
          }),
            this.$store.dispatch("dns/fetchHosts", this.$route.params.dnsname);
        })
        .catch((err) => {
          if (
            err.response.status == 501 ||
            err.response.status == 502 ||
            err.response.status == 500
          ) {
            const opts = {
              message: `Service Unavailable: ${err.response.data.message}.`,
              timeout: 0,
            };
            this.showSnackbarError(opts);
          }
        })
        .finally(() => {
          this.newHost.loading = false;
        });
    },
    deleteSelectedZones() {
      if (this.selected.length > 0) {
        const template = JSON.parse(JSON.stringify(this.tableData));
        this.selected.forEach((el) => {
          const elementIndex = template[el.location][el.type].findIndex(
            (rm) => {
              let result = rm.ttl == el.ttl;

              if (rm.host) {
                result = result && rm.host == el.value;
              }

              if (rm.text) {
                result = result && rm.text == el.value;
              }

              return result;
            }
          );
          // delete template[el.location][el.type][]
          template[el.location][el.type].splice(elementIndex, 1);

          const result = {
            locations: JSON.parse(JSON.stringify(template)),
            name: this.$route.params.dnsname,
          };
          this.deleteLoading = true;
          api.dns
            .setZone(result)
            .then(() => {
              this.snackbar.message = `host${
                this.selected.length == 1 ? "" : "s"
              } deleted successfully.`;
              this.snackbar.visibility = true;

              this.selected = [];
              this.$store.dispatch(
                "dns/fetchHosts",
                this.$route.params.dnsname
              );
            })
            .catch((err) => {
              if (
                err.response.status == 501 ||
                err.response.status == 502 ||
                err.response.status == 500
              ) {
                const opts = {
                  message: `Service Unavailable: ${err.response.data.message}.`,
                  timeout: 0,
                };
                this.showSnackbarError(opts);
              }
            })
            .finally(() => {
              this.deleteLoading = false;
            });
        });
      }
    },
  },
  computed: {
    tableData() {
      return this.$store.getters["dns/getHost"](this.$route.params.dnsname);
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "dns/fetchHosts",
      params: this.$route.params.dnsname,
    });
  },
};
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand";
  line-height: 1em;
  margin-bottom: 15px;
}
</style>
