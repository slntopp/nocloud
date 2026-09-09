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
      <v-col>
        <v-text-field
          readonly
          :value="service.access?.namespace"
          label="namespace"
          style="display: inline-block; width: 150px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="service.access?.level"
          label="level"
          style="display: inline-block; width: 150px"
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
    <v-row justify="center" class="px-2 pb-2">
      <v-expansion-panels inset v-model="opened" multiple>
        <v-expansion-panel
          v-for="(group, i) in service.instancesGroups"
          :key="i"
          style="background: var(--v-background-base)"
        >
          <v-expansion-panel-header>
            {{ group.title }} | Type: {{ group.type }} |
            <v-chip class="instance-group-status" small color="grey">
              {{ group.instances.length }}
            </v-chip>
            <v-icon
              class="instance-group-button"
              @click.stop="openMove(group.uuid)"
              >mdi-arrow-up-bold</v-icon
            >
            <v-icon
              class="instance-group-button"
              @click.stop="openDeleteIg(i)"
              >mdi-delete</v-icon
            >
          </v-expansion-panel-header>
          <v-expansion-panel-content>
            <v-row>
              <v-col>
                <v-text-field
                  readonly
                  :value="provider(group)"
                  label="service provider"
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
                    copyed == `${group}-UUID` ? 'mdi-check' : 'mdi-content-copy'
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
                    copyed == `${group}-hash` ? 'mdi-check' : 'mdi-content-copy'
                  "
                  @click:append="addToClipboard(group.hash, `${group}-hash`)"
                >
                </v-text-field>
              </v-col>
            </v-row>
            <template v-if="group.instances.length > 0">Instances:</template>
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
                        :color="stateColor(instance.state?.meta.state_str)"
                      />
                    </v-expansion-panel-header>
                    <v-expansion-panel-content>
                      <div class="mb-4">
                        <span class="mr-2">Instance uuid:</span>
                        <router-link
                          :to="{
                            name: 'Instance',
                            params: { instanceId: instance.uuid },
                          }"
                        >
                          <v-chip class="mr-2" style="cursor: pointer">{{
                            instance.uuid
                          }}</v-chip>
                        </router-link>
                        <span class="mr-2"
                          >Location: {{ location(instance, group.sp) }}</span
                        >
                      </div>

                      <component
                        dense
                        :is="getInstanceCardComponent(group.type)"
                        :template="instance"
                        :provider="group.sp"
                      />
                    </v-expansion-panel-content>
                  </v-expansion-panel>
                </v-expansion-panels>
              </v-col>
            </v-row>
          </v-expansion-panel-content>
        </v-expansion-panel>
      </v-expansion-panels>
    </v-row>

    <v-dialog persistent v-model="deleteIgDialog" max-width="400">
      <v-card>
        <div class="confirm-card">
          <v-card-title class="text-h6">
            Are you sure you want to do this?
          </v-card-title>
          <v-card-text style="white-space: break-spaces">
            {{ deleteIgWarning(pendingDeleteGroup) }}
          </v-card-text>
          <v-card-actions class="buttons">
            <v-btn color="red darken-1" :disabled="isLoading" @click="closeDeleteIg">
              Cancel
            </v-btn>
            <v-btn
              color="primary darken-1"
              :loading="isLoading"
              @click="confirmDeleteIg"
            >
              Confirm
            </v-btn>
          </v-card-actions>
        </div>
      </v-card>
    </v-dialog>

    <v-dialog style="box-shadow: none" v-model="changeIGDialog" width="40%">
      <v-card class="ma-auto pa-5">
        <v-card-title>Move service</v-card-title>
        <v-form ref="addServiceCard">
          <v-select
            :rules="requiredRule"
            v-model="selectedService"
            item-text="title"
            item-value="uuid"
            label="service"
            :items="allAvailableServices"
          />
        </v-form>
        <v-card-actions class="d-flex justify-center">
          <v-btn @click="changeIGDialog = false">Close</v-btn>
          <v-btn
            :loading="isMoveLoading"
            class="ml-5"
            @click="moveInstanceGroup"
            >Accert</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-row>
      <v-col>
        <v-btn
          :to="{ name: 'Service edit', params: { serviceId: service.uuid } }"
        >
          Edit
        </v-btn>
      </v-col>
    </v-row>
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import ServiceDeploy from "@/components/service/service-deploy.vue";

export default {
  name: "service-info",
  components: {
    ServiceDeploy,
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
    isLoading: false,
    isMoveLoading: false,

    instancesGroup: { uuid: "", type: "" },
    types: [],
    templates: {},

    changeIGDialog: false,
    deleteIgDialog: false,
    pendingDeleteIndex: null,
    selectedService: null,
    igUUID: null,

    requiredRule: [(val) => !!val || "Required field"],
  }),
  computed: {
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
    allServices() {
      return this.$store.getters["services/all"];
    },
    allAvailableServices() {
      if (this.allServices.length === 0) {
        return [];
      }

      return this.allServices.filter((s) => s.uuid !== this.service.uuid);
    },
    pendingDeleteGroup() {
      const index = this.pendingDeleteIndex;
      if (index == null) {
        return null;
      }
      return this.service.instancesGroups[index] || null;
    },
  },
  methods: {
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
    provider(group) {
      return (
        this.servicesProviders.find((el) => el.uuid === group?.sp)?.title ??
        "not found"
      );
    },
    location(inst, uuid) {
      const sp = this.servicesProviders.find((el) => el.uuid === uuid);
      const locationItem = sp?.locations.find(
        ({ extra }) => extra.region === inst.config.datacenter
      );

      return locationItem?.title || sp?.locations[0]?.title || "not found";
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
    getInstanceCardComponent(type) {
      return this.templates[type] ?? this.templates.custom;
    },
    moveInstanceGroup() {
      if (!this.$refs.addServiceCard.validate()) {
        return;
      }
      this.isMoveLoading = true;
      api.instanceGroupService
        .move(this.igUUID, this.selectedService)
        .then(() => {
          this.changeIGDialog = false;
          this.$emit("refresh");
        })
        .finally(() => {
          this.isMoveLoading = false;
        });
    },
    openMove(uuid) {
      this.changeIGDialog = true;
      this.igUUID = uuid;
    },
    hypervisorRef(inst) {
      const data = inst?.data || {};
      return data.vmid || data.vm_id;
    },
    deleteIgWarning(group) {
      const n = group?.instances?.length || 0;
      if (group?.type === "proxmox") {
        return `This will destroy ${n} VM(s) in Proxmox, then remove the group from NoCloud. VMs that stay on the hypervisor cannot be deleted from NoCloud later.`;
      }
      return `Instances will be marked deleted so the driver can destroy them on the hypervisor. The group is removed from NoCloud only after that.`;
    },
    openDeleteIg(index) {
      this.pendingDeleteIndex = index;
      this.deleteIgDialog = true;
    },
    closeDeleteIg() {
      if (this.isLoading) {
        return;
      }
      this.deleteIgDialog = false;
      this.pendingDeleteIndex = null;
    },
    confirmDeleteIg() {
      const index = this.pendingDeleteIndex;
      if (index == null) {
        return;
      }
      this.deleteIg(index);
    },
    async deleteIg(index) {
      const group = this.service.instancesGroups[index];
      if (!group) {
        this.closeDeleteIg();
        return;
      }
      this.isLoading = true;
      try {
        const instances = group.instances || [];
        if (group.type === "proxmox") {
          for (const inst of instances) {
            if (inst.status === "DEL" && !this.hypervisorRef(inst)) {
              continue;
            }
            try {
              await api.instances.action({
                uuid: inst.uuid,
                action: "destroy",
              });
            } catch (err) {
              const msg = err?.response?.data?.message ?? "";
              if (!/not declared/i.test(msg)) {
                throw err;
              }
              for (const item of instances) {
                if (item.status !== "DEL") {
                  await api.delete(`/instances/${item.uuid}`);
                }
              }
              this.showSnackbarSuccess({
                message:
                  "Instances marked for deletion. Deploy the updated Proxmox driver to destroy VMs immediately, then remove the group.",
              });
              this.$store.dispatch("reloadBtn/onclick");
              return;
            }
            if (inst.status !== "DEL") {
              await api.delete(`/instances/${inst.uuid}`);
            }
          }
        } else {
          for (const inst of instances) {
            if (inst.status !== "DEL") {
              await api.delete(`/instances/${inst.uuid}`);
            }
          }
          if (instances.some((inst) => this.hypervisorRef(inst))) {
            this.showSnackbarSuccess({
              message:
                "Instances marked for deletion. Remove the group after the hypervisor destroys the VMs.",
            });
            this.$store.dispatch("reloadBtn/onclick");
            return;
          }
        }

        const template = JSON.parse(JSON.stringify(this.service));
        template.instancesGroups = template.instancesGroups.filter(
          (_, ind) => ind !== index
        );
        await api.services._update(template);
        this.showSnackbarSuccess({
          message: "Service edited successfully",
        });
        this.$store.dispatch("reloadBtn/onclick");
      } catch (err) {
        this.showSnackbarError({
          message: `Error: ${err?.response?.data?.message ?? "Unknown"}. Group was not removed from NoCloud.`,
        });
      } finally {
        this.isLoading = false;
        this.deleteIgDialog = false;
        this.pendingDeleteIndex = null;
      }
    },
  },
  created() {
    const types = require.context(
      "@/components/modules/",
      true,
      /instanceCard\.vue$/
    );

    types.keys().forEach((key) => {
      const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/instanceCard\.vue/i);
      if (matched && matched.length > 1) {
        const type = matched[1];
        this.templates[type] = () =>
          import(`@/components/modules/${type}/instanceCard.vue`);
      }
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

.instance-group-button {
  max-width: 30px;
  margin-left: 25px;
}

.confirm-card {
  background-color: var(--v-background-base);
  padding: 30px;
}

.confirm-card .buttons {
  margin-top: 10px;
  display: flex;
  justify-content: center;
}

.confirm-card .buttons >>> button {
  margin: 0 20px;
}
</style>
