<template>
  <div class="dns-manager pa-4 flex-wrap">
    <div class="buttons__inline pb-8 pt-4">
      <v-menu
        offset-y
        transition="slide-y-transition"
        bottom
        :close-on-content-click="false"
        v-model="createMenuVisible"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn color="background-light" class="mr-2" v-bind="attrs" v-on="on">
            create
          </v-btn>
        </template>
        <v-card class="d-flex pa-4">
          <v-text-field
            dense
            hide-details
            v-model="newHost.title"
            @keypress.enter="createZone"
          >
          </v-text-field>
          <v-btn :loading="newHost.loading" @click="createZone"> send </v-btn>
        </v-card>
      </v-menu>

      <confirm-dialog
        @confirm="deleteSelectedZones"
        :disabled="selected.length < 1"
      >
        <v-btn
          color="background-light"
          class="mr-8"
          :disabled="selected.length < 1"
        >
          delete
        </v-btn>
      </confirm-dialog>
    </div>

    <zones-table :searchParam="searchParam" v-model="selected" single-select>
    </zones-table>

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
import zones from "@/components/zones_table.vue";
import api from "@/api.js";

import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";
import ConfirmDialog from "../components/confirmDialog.vue";

export default {
  name: "dns-manager",
  components: {
    "zones-table": zones,
    ConfirmDialog,
  },
  mixins: [snackbar, search],
  data() {
    return {
      createMenuVisible: false,
      selected: [],
      newHost: {
        title: "",
        loading: false,
        modalVisible: false,
      },
    };
  },
  computed: {
    searchParam() {
      return this.$store.getters["appSearch/param"];
    },
  },
  methods: {
    createZone() {
      if (this.newHost.title.length < 3) return;
      this.newHost.loading = true;
      let title = this.newHost.title;
      title = title[title.length - 1] == "." ? title : title + ".";
      const template = {
        locations: {
          "@": {
            txt: [{ text: "nocloud internal host", ttl: 360 }],
          },
        },
        name: title,
      };
      api.dns
        .setZone(template)
        .then(() => {
          this.createMenuVisible = false;
          this.newHost.title = "";
          this.$store.dispatch("dns/fetch");
        })
        .finally(() => {
          this.newHost.loading = false;
        });
    },
    deleteSelectedZones() {
      if (this.selected.length > 0) {
        const deletePromices = this.selected.map((el) =>
          api.dns.delete(el.original)
        );
        Promise.all(deletePromices)
          .then((res) => {
            if (res.every((el) => el.result)) {
              console.log("all ok");
            }
            this.selected = [];
            this.$store.dispatch("dns/fetch");

            this.snackbar.message = `dns${
              deletePromices.length == 1 ? "" : "es"
            } deleted successfully.`;
            this.snackbar.visibility = true;
          })
          .catch((err) => {
            console.error(err);
          });
      }
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "dns/fetch",
    });
  },
};
</script>

<style></style>
