<template>
  <div>
    <v-row align="center">
      <v-col cols="2">
        <v-text-field
          @click:append="changeTarrifDialog = true"
          append-icon="mdi-pencil"
          readonly
          label="Tarrif"
          :value="tariff"
        ></v-text-field>
      </v-col>
      <v-col cols="2">
        <v-text-field readonly label="price" :value="getPrice()"></v-text-field>
      </v-col>
    </v-row>
    <h3 v-if="dense">Data:</h3>
    <v-card-title v-else class="px-0">Data:</v-card-title>
    <v-row align="center">
      <v-col>
        <v-text-field
          readonly
          label="id"
          style="display: inline-block; width: 200px"
          :value="template.data.vmid"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="next payment date"
          style="display: inline-block; width: 200px"
          :value="date"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="OS"
          style="display: inline-block; width: 200px"
          :value="osName"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="password"
          style="display: inline-block; width: 200px"
          :type="isVisible ? 'text' : 'password'"
          :value="template.config.password"
          :append-icon="isVisible ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append="isVisible = !isVisible"
        />
      </v-col>
    </v-row>

    <h3 v-if="dense">Resources:</h3>
    <v-card-title v-else class="px-0">Resources:</v-card-title>
    <v-row align="center">
      <v-col v-for="(item, key) in template.resources" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key.replace('_', ' ')"
          :value="item"
        />
      </v-col>
    </v-row>

    <v-dialog v-model="changeTarrifDialog" max-width="60%">
      <v-card class="pa-5">
        <v-row>
          <v-col cols="3">
            <v-card-title>Tarrif:</v-card-title>
            <v-select
              v-model="selectedTarrif"
              :items="availableTarrifs"
              item-text="title"
              return-object
            ></v-select>
          </v-col>
          <v-col cols="9">
            <v-card-title>Tarrif resources:</v-card-title>
            <v-text-field
              v-for="resource in Object.keys(selectedTarrif?.resources)"
              :key="resource"
              :value="selectedTarrif?.resources?.[resource]"
              :label="resource"
            ></v-text-field>
          </v-col>
        </v-row>
        <v-row justify="end">
          <v-btn class="mx-3" @click="changeTarrifDialog = false">Close</v-btn>
          <v-btn
            class="mx-3"
            @click="changeTarrif"
            :disabled="selectedTarrif.title === template.product"
            :loading="changeTarrifLoading"
            >Change tarrif</v-btn
          >
        </v-row>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";

export default {
  name: "instance-card",
  props: {
    template: { type: Object, required: true },
    dense: { type: Boolean },
  },
  mixins: [snackbar],
  data: () => ({
    isVisible: false,
    dictionary: {
      cpu: "CPU",
      ram: "RAM",
      ips_public: "IP's public",
      ips_private: "IP's private",
    },

    selectedTarrif: {},
    changeTarrifDialog: false,
    changeTarrifLoading: false,
  }),
  created() {
    this.$store.dispatch("servicesProviders/fetch");
    this.$store.dispatch("plans/fetch");

    this.selectedTarrif = {
      title: this.template.product,
      resources:
        this.template.billingPlan.products[this.template.product].resources,
    };
  },
  methods: {
    getTariff() {
      const {
        billingPlan,
        config: { planCode, duration },
      } = this.template;
      const key = `${duration} ${planCode}`;

      return billingPlan.products[key]?.title;
    },
    getPrice() {
      const initialPrice =
        this.template.billingPlan.products[this.template.product]?.price ?? 0;

      return +this.template.billingPlan.resources
        .reduce((prev, curr) => {
          if (
            curr.key ===
            `drive_${this.template.resources.drive_type.toLowerCase()}`
          ) {
            return (
              prev + (curr.price * this.template.resources.drive_size) / 1024
            );
          } else if (curr.key === "ram") {
            return prev + (curr.price * this.template.resources.ram) / 1024;
          } else if (this.template.resources[curr.key]) {
            return prev + curr.price * this.template.resources[curr.key];
          }
          return prev;
        }, initialPrice)
        ?.toFixed(2);
    },
    changeTarrif() {
      const service = this.service;
      const igIndex = service.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === this.template.uuid)
      );
      const instanceIndex = service.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === this.template.uuid);
      service.instancesGroups[igIndex].instances[instanceIndex].product =
        this.selectedTarrif.title;
      Object.keys(this.selectedTarrif.resources).forEach((k) => {
        service.instancesGroups[igIndex].instances[instanceIndex].resources[k] =
          this.selectedTarrif.resources[k];
      });
      console.log(
        service.instancesGroups[igIndex].instances[instanceIndex].billingPlan
      );

      service.instancesGroups[igIndex].instances[instanceIndex].billingPlan =
        this.billingPlan;

      console.log(
        service.instancesGroups[igIndex].instances[instanceIndex].billingPlan
      );

      this.changeTarrifLoading = true;

      api.services
        ._update(this.service)
        .then(() => {
          this.showSnackbarSuccess("Instance tarrif changed successfully");
        })
        .finally(() => {
          this.changeTarrifLoading = false;
          this.changeTarrifDialog = false;
        });
    },
  },
  computed: {
    sp() {
      return this.$store.getters["servicesProviders/all"].find(
        ({ uuid }) => uuid === this.template.sp
      );
    },
    plans() {
      return this.$store.getters["plans/all"];
    },
    service() {
      return this.$store.getters["services/all"].find(
        (s) => s.uuid === this.template.service
      );
    },
    osName() {
      const id = this.template.config.template_id;

      return this.sp?.publicData.templates[id].name;
    },
    tariff() {
      return this.template.product ?? this.getTariff(this.template) ?? "custom";
    },
    billingPlan() {
      return this.plans.find((p) => p.uuid === this.template.billingPlan.uuid);
    },
    availableTarrifs() {
      return Object.keys(this.billingPlan?.products || {}).map((key) => ({
        title: key,
        resources: this.billingPlan.products[key].resources,
      }));
    },
    date() {
      if (!this.template.data.last_monitoring) return "-";
      const date = new Date(this.template.data.last_monitoring * 1000);

      const year = date.toUTCString().split(" ")[3];
      let month = date.getUTCMonth() + 1;
      let day = date.getUTCDate();

      if (`${month}`.length < 2) month = `0${month}`;
      if (`${day}`.length < 2) day = `0${day}`;

      return `${day}.${month}.${year}`;
    },
  },
};
</script>
