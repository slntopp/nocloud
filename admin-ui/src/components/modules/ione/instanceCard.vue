<template>
  <div>
    <v-row align="center">
      <v-col cols="2">
        <v-text-field
          @click:append="changeTarrifDialog = true"
          :append-icon="
            template.billingPlan.title.toLowerCase() !== 'payg'
              ? 'mdi-pencil'
              : null
          "
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
      <v-col
        v-if="
          template.billingPlan.title.toLowerCase() !== 'payg' ||
          isMonitoringsEmpty
        "
      >
        <v-text-field
          readonly
          label="next payment date"
          style="display: inline-block; width: 200px"
          :value="date"
          :append-icon="!isMonitoringsEmpty ? 'mdi-pencil' : null"
          @click:append="changeDatesDialog = true"
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

    <!-- change tarrif -->
    <change-ione-tarrif
      v-if="availableTarrifs?.length > 0"
      :value="changeTarrifDialog"
      @input="changeTarrifDialog = $event"
      @refresh="refreshInstance"
      :template="template"
      :service="service"
      :sp="sp"
      :available-tarrifs="availableTarrifs"
      :billing-plan="billingPlan"
    />

    <!-- change last monitoring dates -->
    <change-ione-monitorings
      :template="template"
      :service="service"
      :value="changeDatesDialog"
      @input="changeDatesDialog = $event"
      @refresh="refreshInstance"
      v-if="
        template.billingPlan.title.toLowerCase() !== 'payg' ||
        isMonitoringsEmpty
      "
    />
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import { formatSecondsToDate } from "@/functions";
import ChangeIoneMonitorings from "@/components/dialogs/changeIoneMonitorings.vue";
import ChangeIoneTarrif from "@/components/dialogs/changeIoneTarrif.vue";

export default {
  name: "instance-card",
  props: {
    template: { type: Object, required: true },
    dense: { type: Boolean },
  },
  components: { ChangeIoneTarrif, ChangeIoneMonitorings },
  mixins: [snackbar],
  data: () => ({
    isVisible: false,
    dictionary: {
      cpu: "CPU",
      ram: "RAM",
      ips_public: "IP's public",
      ips_private: "IP's private",
    },

    changeTarrifDialog: false,

    changeDatesDialog: false,
  }),
  created() {
    this.$store.dispatch("servicesProviders/fetch");
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
    refreshInstance() {
      this.$emit("refresh");
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

      return this.sp?.publicData.templates[id]?.name;
    },
    tariff() {
      return this.template.product ?? this.getTariff(this.template) ?? "custom";
    },
    billingPlan() {
      return this.$store.getters["plans/one"];
    },
    date() {
      return formatSecondsToDate(this.template?.data?.last_monitoring);
    },
    isMonitoringsEmpty() {
      return this.date === "-";
    },
    availableTarrifs() {
      return Object.keys(this.billingPlan?.products || {}).map((key) => ({
        title: key,
        resources: this.billingPlan.products[key].resources,
      }));
    },
  },
};
</script>
