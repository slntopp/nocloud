<template>
  <div class="module">
    <v-card
      v-if="Object.keys(instance).length > 1"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
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
          <v-autocomplete
            @change="(newVal) => changeOS(newVal)"
            label="template"
            :items="getOsNames"
            :value="getOsTemplates[instance.config.template_id]?.name"
          >
          </v-autocomplete>
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
          <v-autocomplete
            :filter="defaultFilterObject"
            label="price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="plans.list"
            :rules="planRules"
            @change="changeBilling"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="product"
            :value="instance.product"
            :items="products"
            @change="setProduct"
          />
        </v-col>

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
          <v-select
            :items="driveTypes"
            @change="(newVal) => setValue('resources.drive_type', newVal)"
            label="drive type"
            :value="instance.resources.drive_type"
          >
          </v-select>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.drive_size', +newVal)"
            :label="`drive size (minimum ${driveSizeConfig?.minDisk}, maximum ${driveSizeConfig?.maxDisk})`"
            :rules="[driveSizeRule]"
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
      </v-row>
      <v-row>
        <v-col cols="2">
          <v-switch label="Existing" v-model="existing" />
        </v-col>
        <template v-if="existing">
          <v-col>
            <v-text-field
              label="Vm name"
              :value="instance.data?.vm_name"
              @change="(newVal) => setValue('data.vm_name', newVal)"
            />
          </v-col>
          <v-col>
            <v-text-field
              label="Vm id"
              :value="instance.data?.vm_id"
              @change="(newVal) => setValue('data.vm_id', newVal)"
            />
          </v-col>
        </template>
      </v-row>
    </v-card>
  </div>
</template>

<script>
import { defaultFilterObject } from "@/functions";

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    template_id: "",
    password: "",
    auto_renew: true,
  },
  resources: {
    cpu: 1,
    ram: 1024,
    drive_type: null,
    drive_size: 10000,
    ips_public: 0,
    ips_private: 0,
  },
  data: {},
  billing_plan: {},
});
export default {
  name: "instance-ione-create",
  props: ["plans", "instance", "planRules", "sp-uuid", "is-edit"],
  data: () => ({ bilingPlan: null, products: [], existing: false }),
  mounted() {
    if (!this.isEdit) {
      this.$emit("set-instance", getDefaultInstance());
      this.existing = !!(
        this.instance.data?.vm_id || this.instance.data?.vm_name
      );
    } else {
      this.changeBilling(this.instance.billing_plan);
    }
  },
  methods: {
    defaultFilterObject,
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
    changeBilling(val) {
      this.bilingPlan = this.plans.list.find((p) => p.uuid === val);
      if (this.bilingPlan) {
        this.products = Object.keys(this.bilingPlan.products);
      }
      this.setValue("billing_plan", this.bilingPlan);
    },
    setProduct(newVal) {
      const product = this.bilingPlan?.products[newVal].resources;

      Object.keys(product).forEach((key) => {
        this.$emit("set-value", {
          key: "resources." + key,
          value: product[key],
        });
      });
      this.setValue("product", newVal);
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
    driveTypes() {
      return this.instance.billing_plan?.resources
        ?.filter((r) => r.key.includes("drive"))
        .map((k) => k.key.split("_")[1].toUpperCase());
    },
    driveSizeConfig() {
      return this.instance.billing_plan?.meta?.minDisk &&
        this.instance.billing_plan?.meta?.maxDisk
        ? this.instance.billing_plan?.meta
        : { minDisk: 0, maxDisk: 100000000 };
    },
    driveSizeRule() {
      return (val) =>
        (+val >= +this.driveSizeConfig.minDisk &&
          +val <= +this.driveSizeConfig.maxDisk) ||
        "Bad drive size";
    },
  },
  watch: {
    "plans.list"() {
      this.changeBilling(this.instance.billing_plan);
    },
    existing() {
      this.setValue("data.vm_id", null);
      this.setValue("data.vm_name", null);
    },
    driveTypes(newVal) {
      if (newVal && newVal.length > 0) {
        this.setValue("resources.drive_type", newVal[0]);
      }
    },
  },
};
</script>

<style scoped></style>
