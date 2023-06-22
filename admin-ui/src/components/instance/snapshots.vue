<template>
  <v-row class="flex-column mb-5">
    <v-col>
      <v-card-title class="mb-2 px-0">Snapshots:</v-card-title>
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
                @click="createSnapshot(template.uuid)"
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
        @click="deleteSnapshot(template)"
      >
        Delete
      </v-btn>
      <v-btn :loading="isRevertLoading" @click="revertToSnapshot(template)">
        Revert
      </v-btn>
    </v-col>
    <v-col>
      <nocloud-table
        table-name="snapshots"
        single-select
        item-key="ts"
        v-model="selected"
        :items="Object.values(template.state?.meta?.snapshots || {})"
        :headers="headers"
      >
        <template v-slot:[`item.ts`]="{ item }">
          {{ date(item.ts) }}
        </template>
      </nocloud-table>
    </v-col>
  </v-row>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";

export default {
  name: "instance-snapshots",
  components: {
    nocloudTable,
  },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    headers: [
      { text: "Name", value: "name" },
      { text: "Time", value: "ts" },
    ],
    snapshotName: "Snapshot",
    isRevertLoading: false,
    isDeleteLoading: false,
    isLoading: false,
    isVisible: false,
    selected: [],
  }),
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

<style scoped></style>
