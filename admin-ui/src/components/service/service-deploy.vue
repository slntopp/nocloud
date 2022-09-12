<template>
  <div>
    <div class="control">
      <template v-if="service.status === 'UP' || service.status === 'DEL'">
        <v-btn :loading="loading.action" @click="down"> down service </v-btn>
      </template>
      <v-btn v-else :loading="loading.action" @click="deploy">
        deploy
      </v-btn>
      <!-- <template v-else>
        deploy:
        <v-form ref="deployForm" class="mt-3">
          <v-row>
            <v-col>
              <v-select
                label="instance group"
                :items="instancesGroups"
                item-value="uuid"
                item-text="title"
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
      </template> -->
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
    loading: {
      action: false,
    },
  }),
  computed: {
    serviceId() {
      return this.$route.params.serviceId;
    },
  },
  methods: {
    deploy() {
      this.loading.action = true;
      api.services
        .up(this.serviceId)
        .then(() => {
          this.$store.dispatch("services/fetchById", this.serviceId);
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
          this.$store.dispatch("services/fetchById", this.serviceId);
        })
        .finally(() => {
          this.loading.action = false;
        });
    },
  },
  // mounted() {
  //   if (this.service.status != "up" && this.service.status != "del") {
  //     this.$store.dispatch("servicesProviders/fetch");
  //   }
  // },
};
</script>
