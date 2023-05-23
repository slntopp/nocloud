<template>
  <div>
    <v-row>
      <v-col>
        <v-btn @click="changeDNS" class="mr-2"> dns </v-btn>
        <v-btn :loading="actionLoading" @click="deleteInstance"> Delete </v-btn>
      </v-col>
    </v-row>
  </div>
</template>
<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";

export default {
  name: "service-state",
  mixins: [snackbar],
  props: {
    service: { type: Object, required: true },
    instance_uuid: { type: String, required: true },
    "chip-color": { type: String, required: true },
  },
  data() {
    return {
      actualAction: "",
      actionLoading: false,
    };
  },
  methods: {
    changeDNS() {
      this.$router.push({
        name: "InstanceDns",
        params: { instanceId: this.instance_uuid },
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

      this.actualAction = "delete";
      this.actionLoading = true;
      api.services
        ._update(newService)
        .then(() => {
          this.$emit("closePanel");
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
          this.actualAction = "";
          this.actionLoading = false;
        });
    },
  },
};
</script>
