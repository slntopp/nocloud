<template>
  <div>
    <v-btn
      class="mr-2"
      v-for="btn in actions"
      :key="btn.action"
      :disabled="btn.disabled"
      :loading="isLoading"
      @click="sendVmAction(btn.action)"
    >
      {{ btn.title || btn.action }}
    </v-btn>
    <v-btn :loading="isLoading" @click="deleteInstance"> Delete </v-btn>

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
  name: "instance-actions",
  mixins: [snackbar],
  props: {
    uuid: { type: String, required: true },
    actions: { type: Array, default: () => [] }
  },
  data: () => ({ isLoading: false }),
  methods: {
    sendVmAction(action) {
      if (action === "vnc") {
        this.openVnc();
        return;
      }
      if (action === "dns") {
        this.openDns();
        return;
      }

      this.isLoading = true;
      api.instances.action({ uuid: this.uuid, action })
        .then(() => {
          this.showSnackbarSuccess({ message: 'Done!' });
        })
        .catch((err) => {
          const opts = {
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          };
          this.showSnackbarError(opts);
        })
        .finally(() => { this.isLoading = false });
    },
    openVnc() {
      this.$router.push({ name: "Vnc", params: { instanceId: this.uuid } });
    },
    openDns() {
      this.$router.push({ name: "InstanceDns", params: { instanceId: this.instance_uuid } });
    },
    deleteInstance() {
      this.isLoading = true;
      api.delete(`/instances/${this.uuid}`)
        .then(() => {
          this.showSnackbarSuccess({ message: 'Done!' })

          setTimeout(() => {
            this.$router.push({ name: "Instances" });
          }, 100);
        })
        .catch((err) => {
          this.showSnackbarError({
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          });
        })
        .finally(() => { this.isLoading = false });
    },
  },
};
</script>
