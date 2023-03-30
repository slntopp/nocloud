<template>
  <div>
    <v-row align="center">
      <v-col cols="2">
        <v-text-field readonly label="Tarrif" :value="tariff"></v-text-field>
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
  </div>
</template>

<script>
export default {
  name: "instance-card",
  props: {
    template: { type: Object, required: true },
    dense: { type: Boolean },
  },
  data: () => ({
    isVisible: false,
    dictionary: {
      cpu: "CPU",
      ram: "RAM",
      ips_public: "IP's public",
      ips_private: "IP's private",
    },
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
  },
  computed: {
    sp() {
      return this.$store.getters["servicesProviders/all"].find(
        ({ uuid }) => uuid === this.template.sp
      );
    },
    osName() {
      const id = this.template.config.template_id;

      return this.sp?.publicData.templates[id].name;
    },
    tariff() {
      return this.template.product ?? this.getTariff(this.template) ?? "custom";
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
