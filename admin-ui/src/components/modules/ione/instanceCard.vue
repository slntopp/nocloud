<template>
  <div>
    <v-row v-if="!editing">
      <v-col>
        <service-control
          :service="service"
          :instance_uuid="localInstance.uuid"
          :chip-color="chipColor"
          @closePanel="openedlocalInstances = {}"
        />
      </v-col>
    </v-row>
    <v-row v-if="!editing">
      <v-col md="2">
        <v-text-field
          readonly
          :value="localInstance.state && localInstance.state.meta.state_str"
          label="state"
          style="display: inline-block; width: 100px"
        />
      </v-col>
      <v-col md="2">
        <v-text-field
          readonly
          :value="localInstance.state && localInstance.state.meta.lcm_state_str"
          label="lcm state"
          style="display: inline-block; width: 100px"
        />
      </v-col>
      <v-col md="2">
        <v-text-field
          readonly
          :value="localInstance.billingPlan.title"
          label="price model"
          style="display: inline-block; width: 100px"
        />
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col>
        <v-text-field
          v-if="editing"
          v-model="localInstance.title"
          label="title"
          style="display: inline-block; width: 160px"
        />
      </v-col>
      <v-col>
        <v-text-field
          :readonly="!editing"
          :value="localInstance.config.template_id"
          label="template id"
          style="display: inline-block; width: 160px"
          @change="(v) => (localInstance.config.template_id = parseInt(v))"
        />
      </v-col>
      <v-col>
        <v-text-field
          :readonly="!editing"
          v-model="localInstance.config.password"
          label="password"
          style="display: inline-block; width: 160px"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field
          :readonly="!editing"
          :value="localInstance.resources.cpu"
          label="CPU"
          style="display: inline-block; width: 100px"
          @change="(v) => (localInstance.resources.cpu = parseInt(v))"
        />
      </v-col>
      <v-col>
        <v-text-field
          :readonly="!editing"
          :value="localInstance.resources.ram"
          label="RAM"
          style="display: inline-block; width: 100px"
          @change="(v) => (localInstance.resources.ram = parseInt(v))"
        />
      </v-col>
      <v-col>
        <v-text-field
          :readonly="!editing"
          :value="localInstance.resources.drive_size"
          label="drive size"
          style="display: inline-block; width: 100px"
          @change="(v) => (localInstance.resources.drive_size = parseInt(v))"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          v-if="!editing"
          :value="localInstance.resources.drive_type"
          label="drive type"
          style="display: inline-block; width: 100px"
        />
        <v-select
          v-else
          v-model="localInstance.resources.drive_type"
          label="drive type"
          style="display: inline-block; width: 100px"
          :items="['SSD', 'HDD']"
        />
      </v-col>
      <v-col>
        <v-text-field
          :readonly="!editing"
          :value="localInstance.resources.ips_private"
          label="ips private"
          style="display: inline-block; width: 100px"
          @change="(v) => (localInstance.resources.ips_private = parseInt(v))"
        />
      </v-col>
      <v-col>
        <v-text-field
          :readonly="!editing"
          :value="localInstance.resources.ips_public"
          label="ips public"
          style="display: inline-block; width: 100px"
          @change="(v) => (localInstance.resources.ips_public = parseInt(v))"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="localInstance.config.template_id"
          label="template id"
          style="display: inline-block; width: 100px"
          v-if="!editing"
        />
      </v-col>
    </v-row>
    <v-row class="flex-column" v-if="localInstance.state && !editing">
      <v-col>
        <h4 class="mb-2">Snapshots:</h4>
        <v-menu
          bottom
          offset-y
          transition="slide-y-transition"
          v-model="isVisible"
          :close-on-content-click="false"
        >
          <template v-slot:activator="{ on, attrs }">
            <v-btn class="mr-2" v-bind="attrs" v-on="on"> Create </v-btn>
          </template>
          <v-card class="pa-4">
            <v-row>
              <v-col>
                <v-text-field
                  dense
                  label="name"
                  v-model="snapshotName"
                  :rules="[(v) => !!v || 'Required!']"
                />
                <v-btn
                  :loading="isLoading"
                  @click="createSnapshot(localInstance.uuid)"
                >
                  Send
                </v-btn>
              </v-col>
            </v-row>
          </v-card>
        </v-menu>
        <v-btn
          class="mr-2"
          :loading="isDeleteLoading"
          @click="deleteSnapshot(localInstance)"
        >
          Delete
        </v-btn>
        <v-btn
          :loading="isRevertLoading"
          @click="revertToSnapshot(localInstance)"
        >
          Revert
        </v-btn>
      </v-col>
      <v-col>
        <nocloud-table
          single-select
          item-key="ts"
          v-model="selected"
          :items="Object.values(localInstance.state?.meta?.snapshots || {})"
          :headers="headers"
        >
          <template v-slot:[`item.ts`]="{ item }">
            {{ date(item.ts) }}
          </template>
        </nocloud-table>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import serviceControl from "@/components/modules/ione/serviceControls.vue";
import nocloudTable from "@/components/table.vue";
import api from "@/api.js";

import snackbar from "@/mixins/snackbar.js";

export default {
  name: "instance-card",
  components: { serviceControl, nocloudTable },
  mixins: [snackbar],
  props: {
    instance: { type: Object },
    service: { type: Object },
    editing: { type: Boolean },
    type: { type: String },
    chipColor: { type: String },
  },
  data() {
    return {
      localInstance: {},
      selected: [],
      headers: [
        { text: "Name", value: "name" },
        { text: "Time", value: "ts" },
      ],
      isVisible: false,
      snapshotName: "Snapshot",
      isLoading: false,
      isDeleteLoading: false,
      isRevertLoading: false,
    };
  },
  created() {
    console.log(1);
    this.localInstance = this.instance;
  },
  methods: {
    createSnapshot(uuid) {
      this.isLoading = true;

      api.instances
        .action({
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
      api.instances
        .action({
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
        .finally(() => {
          this.isDeleteLoading = false;
        });
    },
    revertToSnapshot({ uuid, state }) {
      const { snapshots } = state.meta;
      console.log(snapshots);
      const [id] = Object.entries(snapshots).find(
        ([, el]) => el.ts === this.selected[0].ts
      );

      this.isRevertLoading = true;
      api.instances
        .action({
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
        .finally(() => {
          this.isRevertLoading = false;
        });
    },
    date(timestamp) {
      const date = new Date(timestamp * 1000);
      const time = date.toUTCString().split(" ")[4];

      const day = date.getUTCDate();
      const month = date.getUTCMonth() + 1;
      const year = date.toUTCString().split(" ")[3];

      return `${day}.${month}.${year} ${time}`;
    },
  },
};
</script>
