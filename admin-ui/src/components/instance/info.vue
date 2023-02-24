<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <instance-actions :template="template" />
      </v-col>
    </v-row>
    <v-row align="center">
      <v-col>
        <v-text-field readonly label="instance uuid" style="display: inline-block; width: 330px" :value="template.uuid"
          :append-icon="copyed === 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(template.uuid, 'rootUUID')" />
      </v-col>
      <v-col v-if="template.state">
        <v-text-field readonly label="state" style="display: inline-block; width: 150px"
          :value="template.state.meta?.state_str || template.state.state" />
      </v-col>
      <v-col v-if="template.state?.meta.lcm_state_str">
        <v-text-field readonly label="lcm state" style="display: inline-block; width: 150px"
          :value="template.state?.meta.lcm_state_str" />
      </v-col>
      <v-col>
        <v-text-field readonly label="price model" style="display: inline-block; width: 150px"
          :value="template.billingPlan.title" />
      </v-col>
    </v-row>

    <component :is="templates[template.type] ?? templates.custom" :template="template" />

    <v-row class="flex-column mb-5" v-if="template.state">
      <v-col>
        <v-card-title class="mb-2 px-0">Snapshots:</v-card-title>
        <v-menu bottom offset-y transition="slide-y-transition" v-model="isVisible" :close-on-content-click="false">
          <template v-slot:activator="{ on, attrs }">
            <v-btn class="mr-2" v-bind="attrs" v-on="on"> Create </v-btn>
          </template>
          <v-card class="pa-4">
            <v-row>
              <v-col>
                <v-text-field dense label="name" v-model="snapshotName" :rules="[(v) => !!v || 'Required!']" />
                <v-btn :loading="isLoading" @click="createSnapshot(template.uuid)">
                  Send
                </v-btn>
              </v-col>
            </v-row>
          </v-card>
        </v-menu>
        <v-btn class="mr-2" :loading="isDeleteLoading" @click="deleteSnapshot(template)">
          Delete
        </v-btn>
        <v-btn :loading="isRevertLoading" @click="revertToSnapshot(template)">
          Revert
        </v-btn>
      </v-col>
      <v-col>
        <nocloud-table single-select item-key="ts" v-model="selected"
          :items="Object.values(template.state?.meta?.snapshots || {})" :headers="headers">
          <template v-slot:[`item.ts`]="{ item }">
            {{ date(item.ts) }}
          </template>
        </nocloud-table>
      </v-col>
    </v-row>

    <v-btn :to="{
      name: 'Service edit', params: {
        serviceId: template.service, sp: template.sp, instance: template.uuid
      }, query: { instance: template.uuid }
    }">
      Edit
    </v-btn>

    <v-snackbar v-model="snackbar.visibility" :timeout="snackbar.timeout" :color="snackbar.color">
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn :color="snackbar.buttonColor" text v-bind="attrs" @click="snackbar.visibility = false">
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import instanceActions from "@/components/instance/controls.vue";
import JsonTextarea from "@/components/JsonTextarea.vue";

export default {
  name: "instance-info",
  components: { nocloudTable, instanceActions, JsonTextarea },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    copyed: null,
    templates: {},
    selected: [],
    headers: [
      { text: "Name", value: "name" },
      { text: "Time", value: "ts" },
    ],
    snapshotName: "Snapshot",

    isVisible: false,
    isLoading: false,
    isDeleteLoading: false,
    isRevertLoading: false,
  }),
  methods: {
    createSnapshot(uuid) {
      this.isLoading = true;

      api.instances.action({
        uuid,
        action: "snapcreate",
        params: { snap_name: this.snapshotName },
      })
        .then(() => {
          this.showSnackbarSuccess({
            message: "Snapshot created successfully",
          });
        })
        .catch((err) => {
          this.showSnackbarError({
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          });
        })
        .finally(() => {
          this.isLoading = false;
          this.isVisible = false;
        });
    },
    deleteSnapshot({ uuid, state }) {
      const { snapshots } = state.meta;
      const [id] = Object.entries(snapshots).find(
        ([, el]) => el.ts === this.selected[0].ts
      );

      this.isDeleteLoading = true;
      api.instances.action({
        uuid,
        action: "snapdelete",
        params: { snap_id: +id },
      })
        .then(() => {
          this.showSnackbarSuccess({
            message: "Snapshot deleted successfully",
          });
        })
        .catch((err) => {
          this.showSnackbarError({
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          });
        })
        .finally(() => { this.isDeleteLoading = false });
    },
    revertToSnapshot({ uuid, state }) {
      const { snapshots } = state.meta;
      const [id] = Object.entries(snapshots).find(
        ([, el]) => el.ts === this.selected[0].ts
      );

      this.isRevertLoading = true;
      api.instances.action({
        uuid,
        action: "snaprevert",
        params: { snap_id: +id },
      })
        .then(() => {
          this.showSnackbarSuccess({
            message: "Snapshot reverted successfully",
          });
        })
        .catch((err) => {
          this.showSnackbarError({
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          });
        })
        .finally(() => { this.isRevertLoading = false });
    },
    date(timestamp) {
      const date = new Date(timestamp * 1000);
      const time = date.toUTCString().split(" ")[4];

      const day = date.getUTCDate();
      const month = date.getUTCMonth() + 1;
      const year = date.toUTCString().split(" ")[3];

      return `${day}.${month}.${year} ${time}`;
    },
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => { this.copyed = index })
          .catch((res) => { console.error(res) });
      } else {
        this.showSnackbarError({
          message: "Clipboard is not supported!",
        });
      }
    },
  },
  created() {
    const types = require.context("@/components/modules/", true, /instanceCard\.vue$/);

    types.keys().forEach((key) => {
      const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/instanceCard\.vue/i);

      if (matched && matched.length > 1) {
        this.templates[matched[1]] = () =>
          import(`@/components/modules/${matched[1]}/instanceCard.vue`);
      }
    });
  },
}
</script>
