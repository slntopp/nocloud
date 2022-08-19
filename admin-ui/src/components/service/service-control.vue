<template>
  <div>
    <v-row>
      <v-col>
        <v-btn
          class="mr-2"
          v-for="btn in vmControlBtns"
          :key="btn.action"
          :disabled="actionLoading && actualAction != btn.action"
          :loading="actionLoading && actualAction == btn.action"
          @click="sendVmAction(btn.action)"
        >
          {{ btn.title || btn.action }}
        </v-btn>
        <v-btn
          :loading="actionLoading"
          @click="deleteInstance"
        >
          Delete
        </v-btn>
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
  </div>
</template>
<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";

export default {
  name: "service-state",
  mixins: [snackbar],
  props: {
    service: { type: Object, required: true, },
    instance_uuid: { type: String, required: true, },
    "chip-color": { type: String, required: true, },
  },
  data: () => ({
    actualAction: "",
    actionLoading: false,
    vmControlBtns: [
      {
        action: "poweroff",
        title: "poweroff", //not reqired, use 'action' for a name if not found
      },
      {
        action: "resume",
      },
      {
        action: "suspend",
      },
      {
        action: "reboot",
      },
    ],
  }),
  methods: {
    sendVmAction(action) {
      this.actualAction = action;
      this.actionLoading = true;
      api.instances
        .action({ uuid: this.instance_uuid, action })
        .then(() => {
          this.showSnackbarSuccess({ message: `Done!` });
        })
        .catch((err) => {
          const opts = {
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          };
          this.showSnackbarError(opts);
        })
        .finally(() => {
          this.actualAction = "";
          this.actionLoading = false;
        });
    },
    deleteInstance() {
      const newService = JSON.parse(JSON.stringify(this.service));

      newService.instancesGroups.forEach((group, i, groups) => {
        group.instances.forEach(({ uuid }, j) => {
          if (uuid === this.instance_uuid) {
            groups[i].instances.splice(j, 1);
          }
        });
      });

      this.actualAction = 'delete';
      this.actionLoading = true;
      api.services._update(newService)
        .then(() => {
          this.$emit('closePanel');
          this.service.instancesGroups.forEach((group, i, groups) => {
            group.instances.forEach(({ uuid }, j) => {
              if (uuid === this.instance_uuid) {
                groups[i].instances.splice(j, 1);
              }
            });
            groups[i].resources.ips_public = groups[i].instances.length;
          });

          setTimeout(() => {
            this.showSnackbarSuccess({ message: `Done!` });
          }, 100);
        })
        .catch((err) => {
          this.showSnackbarError({
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          });
        })
        .finally(() => {
          this.actualAction = '';
          this.actionLoading = false;
        });
    }
  },
};
</script>
