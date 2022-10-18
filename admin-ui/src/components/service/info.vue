<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <!-- <v-row>
      <v-col>
        <div>
          service state:
          <v-chip x-small :color="chipColor">
            {{ this.service.status }}
          </v-chip>
        </div>
      </v-col>
    </v-row> -->
    <v-row align="center">
      <v-col>
        <v-text-field
          readonly
          :value="service.uuid"
          label="service uuid"
          style="display: inline-block; width: 330px"
          :append-icon="copyed == 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(service.uuid, 'rootUUID')"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="service && service.status"
          label="state"
          style="display: inline-block; width: 150px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="hashpart(service.hash)"
          label="service hash"
          style="display: inline-block; width: 150px"
          :append-icon="copyed == 'rootHash' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(service.hash, 'rootHash')"
        >
        </v-text-field>
      </v-col>
    </v-row>
    <v-row class="mb-5">
      <v-col>
        <service-deploy :service="service"></service-deploy>
      </v-col>
    </v-row>

    groups:
    <v-switch label="editing" class="ml-2 d-inline-block" v-model="editing" />

    <v-row justify="center" class="px-2 pb-2">
      <v-expansion-panels inset v-model="opened" multiple>
        <v-expansion-panel
          v-for="(group, i) in service.instancesGroups"
          :key="i"
          style="background: var(--v-background-base)"
        >
          <v-expansion-panel-header>
            {{ group.title }} | Type: {{ group.type }} |
            <v-chip
              class="instance-group-status"
              small
              color="grey"
            >
              {{ group.instances.length }}
            </v-chip>
          </v-expansion-panel-header>
          <v-expansion-panel-content>
            <v-row>
              <template v-if="!editing">
                <v-col>
                  <v-text-field
                    readonly
                    :value="location(group)"
                    label="location"
                    style="display: inline-block; width: 330px"
                  >
                  </v-text-field>
                </v-col>
                <v-col>
                  <v-text-field
                    readonly
                    :value="group.uuid"
                    label="group uuid"
                    style="display: inline-block; width: 330px"
                    :append-icon="
                      copyed == `${group}-UUID`
                        ? 'mdi-check'
                        : 'mdi-content-copy'
                    "
                    @click:append="addToClipboard(group.uuid, `${group}-UUID`)"
                  >
                  </v-text-field>
                </v-col>
                <v-col>
                  <v-text-field
                    readonly
                    :value="hashpart(group.hash)"
                    label="group hash"
                    style="display: inline-block; width: 150px"
                    :append-icon="
                      copyed == `${group}-hash`
                        ? 'mdi-check'
                        : 'mdi-content-copy'
                    "
                    @click:append="addToClipboard(group.hash, `${group}-hash`)"
                  >
                  </v-text-field>
                </v-col>
              </template>
              <template v-else>
                <v-col>
                  <v-text-field
                    label="title"
                    style="display: inline-block; width: 330px"
                    v-model="group.title"
                  />
                </v-col>
                <v-col v-if="!group.type">
                  <v-select
                    label="type"
                    style="display: inline-block; width: 330px"
                    v-model="instancesGroup.type"
                    :items="types"
                    @change="instancesGroup.uuid = group.uuid"
                  />
                </v-col>
              </template>
            </v-row>
            Instances:
            <v-row>
              <v-col>
                <v-expansion-panels inset v-model="openedInstances[i]" multiple>
                  <v-expansion-panel
                    v-for="(instance, index) in group.instances"
                    :key="index"
                    style="background: var(--v-background-light-base)"
                  >
                    <v-expansion-panel-header>
                      {{ instance.title }}
                      <v-chip
                        x-small
                        class="ml-2"
                        style="max-width: 10px; max-height: 10px; padding: 0"
                        :color="
                          stateColor(
                            instance.state && instance.state.meta.state_str
                          )
                        "
                      />
                    </v-expansion-panel-header>
                    <v-expansion-panel-content>
                      <v-row v-if="!editing">
                        <v-col>
                          <service-control
                            :service="service"
                            :instance_uuid="instance.uuid"
                            :chip-color="chipColor"
                            @closePanel="openedInstances = {}"
                          />
                        </v-col>
                      </v-row>
                      <v-row v-if="!editing">
                        <template v-if="group.type === 'ovh'">
                          <v-col md="2">
                            <v-text-field
                              readonly
                              :value="
                                instance.state && instance.state.meta.state
                              "
                              label="state"
                              style="display: inline-block; width: 100px"
                            />
                          </v-col>
                          <v-col md="2">
                            <v-text-field
                              readonly
                              :value="instance.state && instance.state.state"
                              label="lcm state"
                              style="display: inline-block; width: 100px"
                            />
                          </v-col>
                        </template>
                        <template v-else>
                          <v-col md="2">
                            <v-text-field
                              readonly
                              :value="
                                instance.state && instance.state.meta.state_str
                              "
                              label="state"
                              style="display: inline-block; width: 100px"
                            />
                          </v-col>
                          <v-col md="2">
                            <v-text-field
                              readonly
                              :value="
                                instance.state &&
                                instance.state.meta.lcm_state_str
                              "
                              label="lcm state"
                              style="display: inline-block; width: 100px"
                            />
                          </v-col>
                        </template>
                        <v-col md="2">
                          <v-text-field
                            readonly
                            :value="instance.billingPlan.title"
                            label="billing plan"
                            style="display: inline-block; width: 100px"
                          />
                        </v-col>
                      </v-row>
                      <v-row v-else>
                        <v-col>
                          <v-text-field
                            v-if="editing"
                            v-model="instance.title"
                            label="title"
                            style="display: inline-block; width: 160px"
                          />
                        </v-col>
                        <v-col v-if="group.type === 'ione'">
                          <v-text-field
                            :readonly="!editing"
                            :value="instance.config.template_id"
                            label="template id"
                            style="display: inline-block; width: 160px"
                            @change="
                              (v) => (instance.config.template_id = parseInt(v))
                            "
                          />
                        </v-col>
                        <v-col v-if="group.type === 'ione'">
                          <v-text-field
                            :readonly="!editing"
                            v-model="instance.config.password"
                            label="password"
                            style="display: inline-block; width: 160px"
                          />
                        </v-col>
                      </v-row>
                      <v-row v-if="group.type !== 'ione'">
                        <json-editor
                          v-if="editing"
                          :json="instance.config"
                          @changeValue="(data) => (instance.config = data)"
                        />
                        <json-textarea
                          v-else
                          :json="instance.config"
                          :readonly="true"
                        />
                      </v-row>
                      <v-row v-else>
                        <v-col>
                          <v-text-field
                            :readonly="!editing"
                            :value="instance.resources.cpu"
                            label="CPU"
                            style="display: inline-block; width: 100px"
                            @change="
                              (v) => (instance.resources.cpu = parseInt(v))
                            "
                          />
                        </v-col>
                        <v-col>
                          <v-text-field
                            :readonly="!editing"
                            :value="instance.resources.ram"
                            label="RAM"
                            style="display: inline-block; width: 100px"
                            @change="
                              (v) => (instance.resources.ram = parseInt(v))
                            "
                          />
                        </v-col>
                        <v-col>
                          <v-text-field
                            :readonly="!editing"
                            :value="instance.resources.drive_size"
                            label="drive size"
                            style="display: inline-block; width: 100px"
                            @change="
                              (v) =>
                                (instance.resources.drive_size = parseInt(v))
                            "
                          />
                        </v-col>
                        <v-col>
                          <v-text-field
                            readonly
                            v-if="!editing"
                            :value="instance.resources.drive_type"
                            label="drive type"
                            style="display: inline-block; width: 100px"
                          />
                          <v-select
                            v-else
                            v-model="instance.resources.drive_type"
                            label="drive type"
                            style="display: inline-block; width: 100px"
                            :items="['SSD', 'HDD']"
                          />
                        </v-col>
                        <v-col>
                          <v-text-field
                            :readonly="!editing"
                            :value="instance.resources.ips_private"
                            label="ips private"
                            style="display: inline-block; width: 100px"
                            @change="
                              (v) =>
                                (instance.resources.ips_private = parseInt(v))
                            "
                          />
                        </v-col>
                        <v-col>
                          <v-text-field
                            :readonly="!editing"
                            :value="instance.resources.ips_public"
                            label="ips public"
                            style="display: inline-block; width: 100px"
                            @change="
                              (v) =>
                                (instance.resources.ips_public = parseInt(v))
                            "
                          />
                        </v-col>
                        <v-col>
                          <v-text-field
                            readonly
                            :value="instance.config.template_id"
                            label="template id"
                            style="display: inline-block; width: 100px"
                            v-if="!editing"
                          />
                        </v-col>
                      </v-row>
                      <v-row
                        class="flex-column"
                        v-if="isIone(i) && instance.state && !editing"
                      >
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
                              <v-btn class="mr-2" v-bind="attrs" v-on="on">
                                Create
                              </v-btn>
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
                                    @click="createSnapshot(instance.uuid)"
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
                            @click="deleteSnapshot(instance)"
                          >
                            Delete
                          </v-btn>
                          <v-btn
                            :loading="isRevertLoading"
                            @click="revertToSnapshot(instance)"
                          >
                            Revert
                          </v-btn>
                        </v-col>
                        <v-col>
                          <nocloud-table
                            single-select
                            item-key="ts"
                            v-model="selected"
                            :items="
                              Object.values(
                                instance.state?.meta?.snapshots || {}
                              )
                            "
                            :headers="headers"
                          >
                            <template v-slot:[`item.ts`]="{ item }">
                              {{ date(item.ts) }}
                            </template>
                          </nocloud-table>
                        </v-col>
                      </v-row>
                    </v-expansion-panel-content>
                  </v-expansion-panel>
                </v-expansion-panels>
              </v-col>
            </v-row>
          </v-expansion-panel-content>
        </v-expansion-panel>
      </v-expansion-panels>
    </v-row>
    <v-row v-if="editing">
      <v-col>
        <v-btn :loading="isLoading" @click="editService">Edit</v-btn>
      </v-col>
    </v-row>

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
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import ServiceDeploy from "@/components/service/service-deploy.vue";
import ServiceControl from "@/components/service/service-control.vue";
import nocloudTable from "@/components/table.vue";
import JsonEditor from "../JsonEditor.vue";
import JsonTextarea from "../JsonTextarea.vue";

export default {
  name: "service-info",
  components: {
    ServiceDeploy,
    ServiceControl,
    nocloudTable,
    JsonEditor,
    JsonTextarea,
  },
  mixins: [snackbar],
  props: {
    service: {
      type: Object,
      required: true,
    },
    chipColor: {
      type: String,
      required: true,
    },
  },
  data: () => ({
    copyed: null,
    opened: [],
    openedInstances: {},
    editing: false,
    isLoading: false,
    isDeleteLoading: false,
    isRevertLoading: false,

    headers: [
      { text: "Name", value: "name" },
      { text: "Time", value: "ts" },
    ],
    instancesGroup: { uuid: "", type: "" },
    types: [],
    selected: [],
    isVisible: false,
    snapshotName: "Snapshot",
  }),
  computed: {
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
  },
  methods: {
    isIone(index) {
      return this.service.instancesGroups[index].type === "ione";
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
        this.showSnackbarError({
          message: "Clipboard is not supported!",
        });
      }
    },
    hashpart(hash) {
      if (hash) return hash.slice(0, 8);
      return "WWWWWWWW";
    },
    location(group) {
      const lc = this.servicesProviders.find((el) => el.uuid === group?.sp);

      return lc?.title || "not found";
    },
    stateColor(state) {
      const dict = {
        INIT: "orange darken-2",
        ACTIVE: "green darken-2",
        UNKNOWN: "gray darken-2",
        POWEROFF: "red darken-2",
        SUSPENDED: "orange darken-2",
        OPERATION: "gray darken-2",
      };

      return dict[state] ?? "blue-grey darken-2";
    },
    editService() {
      this.isLoading = true;
      if (this.instancesGroup.uuid) {
        this.service.instancesGroups.forEach((el, i, arr) => {
          if (el.uuid === this.instancesGroup.uuid) {
            arr[i].type = this.instancesGroup.type;
          }
        });
      }

      api.services
        ._update(this.service)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Service edited successfully",
          });
        })
        .catch((err) => {
          this.showSnackbarError({
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          });
        })
        .finally(() => {
          this.isLoading = false;
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
  },
  created() {
    this.$store.dispatch("namespaces/fetch").catch((err) => {
      this.showSnackbarError({
        message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
      });
    });
  },
  mounted() {
    const types = require.context(
      "@/components/modules/",
      true,
      /serviceProviders\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(
        /\.\/([A-Za-z0-9-_,\s]*)\/serviceProviders\.vue/i
      );
      if (matched && matched.length > 1) {
        this.types.push(matched[1]);
      }
    });
    // Object.keys(this.service.instancesGroups).forEach((key) => {
    //   this.$set(this.openedInstances, key, [0]);
    // });
  },
};
</script>

<style scoped>
.instance-group-status {
  max-width: 30px;
  align-items: center;
  margin-left: 25px;
}
</style>