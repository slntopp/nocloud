<template>
  <div>
    <v-row>
      <v-col>
        <v-btn
          v-for="(btn, index) in vmControlBtns"
          :key="btn.action"
          @click="sendVmAction(btn.action)"
          :class="{ 'mr-2': index !== vmControlBtns.lenght - 1 }"
          :disabled="actionLoading && actualAction != btn.action"
          :loading="actionLoading && actualAction == btn.action"
        >
          {{ btn.title || btn.action }}
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
    instance_uuid: {
      type: String,
      required: true,
    },
    "chip-color": {
      type: String,
      required: true,
    },
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
      api.services
        .action(this.instance_uuid, action)
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
  },
};
</script>
