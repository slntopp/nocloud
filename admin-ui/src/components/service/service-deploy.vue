<template>
  <div>
    <div class="control">
      <template v-if="service.status == 'up' || service.status == 'del'">
        <v-btn :loading="loading.action" @click="down"> down service </v-btn>
      </template>
      <template v-else>
        deploy:
        <v-form ref="deployForm" class="mt-3">
          <v-row>
            <v-col>
              <v-select
                label="instance group"
                :items="instancesGroups"
                :rules="[(v) => !!v || 'required']"
                v-model="deployInstancesGroup"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-select
                label="services provider"
                :items="servicesProviders"
                item-value="uuid"
                item-text="title"
                :rules="[(v) => !!v || 'required']"
                v-model="deployServiceProvider"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-btn
                :loading="loading.action"
                :disabled="!deployServiceProvider || !deployInstancesGroup"
                @click="deploy"
              >
                deploy
              </v-btn>
            </v-col>
          </v-row>
        </v-form>
      </template>
    </div>
  </div>
</template>
<script>
import api from "@/api";
export default {
  name: "service-control",
  props: {
    service: {
      type: Object,
      required: true,
    },
  },
  data: () => ({
    deployServiceProvider: "",
    deployInstancesGroup: "",
    loading: {
      action: false,
    },
  }),
  computed: {
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
    instancesGroups() {
      const result = [];
      for (const group of this.service.instancesGroups) {
        result.push({
          text: group.title,
          value: group.uuid,
        });
      }
      return result;
    },
    serviceId() {
        console.log(this.$route.params.serviceId)
      return this.$route.params.serviceId;
    },
  },
  methods: {
    deploy() {
      if (!this.$refs.deployForm.validate()) return;
      this.loading.action = true;
      api.services
        .up(
          this.serviceId,
          this.deployInstancesGroup,
          this.deployServiceProvider
        )
        .then(() => {
          this.$store.dispatch("services/fetch");
        })
        .finally(() => {
          this.loading.action = false;
        });
    },
    down() {
      this.loading.action = true;
      api.services
        .down(this.serviceId)
        .then(() => {
          this.$store.dispatch("services/fetch");
        })
        .finally(() => {
          this.loading.action = false;
        });
    },
  },
  mounted() {
    if (this.service.status != "up" && this.service.status != "del") {
      this.$store.dispatch("servicesProviders/fetch");
    }
  },
};
</script>
