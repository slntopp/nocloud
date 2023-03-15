<template>
  <div class="module">
    <v-card v-if="Object.keys(instance).length>1" class="mb-4 pa-2" elevation="0" color="background">
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('title', newVal)"
            label="title"
            :value="instance.title"
          >
          </v-text-field>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-select
            @change="(newVal) => changeOS(newVal)"
            label="template"
            :items="getOsNames"
            :value="getOsTemplates[instance.config.template_id]?.name"
          >
          </v-select>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.password', newVal)"
            label="password"
            :value="instance.config?.password"
          >
          </v-text-field>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.cpu', +newVal)"
            label="cpu"
            :value="instance.resources.cpu"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.ram', +newVal)"
            label="ram"
            :value="instance.resources.ram"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.drive_type', newVal)"
            label="drive type"
            :value="instance.resources.drive_type"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.drive_size', +newVal)"
            label="drive size"
            :value="instance.resources.drive_size"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.ips_public', +newVal)"
            label="ips public"
            :value="instance.resources.ips_public"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.ips_private', +newVal)"
            label="ips private"
            :value="instance.resources.ips_private"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-select
            label="price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="plans.list"
            :rules="planRules"
            @change="(newVal) => setValue('billing_plan', newVal)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            label="product"
            :value="instance.productTitle"
            v-if="getPlanProducts().length > 0"
            :items="getPlanProducts()"
            @change="(newVal) => setValue('product', newVal)"
          />
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script>
const getDefaultInstance = () => ({
  title: "instance",
  config: {
    template_id: "",
    password: "",
  },
  resources: {
    cpu: 1,
    ram: 1024,
    drive_type: "SSD",
    drive_size: 10000,
    ips_public: 0,
    ips_private: 0,
  },
  billing_plan: {},
});
export default {
  name: "instance-ione-create",
  props: ["plans", "instance", "planRules", "sp-uuid"],
  mounted() {
    this.$emit("set-instance", getDefaultInstance());
  },
  methods: {
    getPlanProducts() {
      if (!this.instance.billing_plan?.products) {
        return [];
      }
      return Object.values(this.instance.billing_plan.products).map(
        (p) => p.title
      );
    },
    changeOS(newVal) {
      let osId = null;

      for (const [key, value] of Object.entries(this.getOsTemplates)) {
        if (value.name === newVal) {
          osId = key;
          break;
        }
      }

      this.setValue("config.template_id", +osId);
    },
    setValue(key, value) {
      this.$emit("set-value", { key, value });
    },
  },
  computed: {
    getOsTemplates() {
      const sp = this.$store.getters["servicesProviders/all"].filter(
        (el) => el.uuid === this.spUuid
      )[0];

      if (!sp) return {};

      return sp.publicData.templates;
    },
    getOsNames() {
      if (!this.getOsTemplates) return [];

      return Object.values(this.getOsTemplates).map((os) => os.name);
    },
  },
};
</script>

<style scoped></style>
