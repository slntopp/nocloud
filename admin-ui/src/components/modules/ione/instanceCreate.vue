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
            label="Name"
            :value="instance.title"
            :rules="requiredRule"
          >
          </v-text-field>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="Price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="plans"
            :rules="planRules"
            @change="changeBilling"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Product"
            :disabled="isDynamicPlan"
            :rules="!isDynamicPlan ? requiredRule : []"
            :value="instance.product"
            :items="products"
            @change="setProduct"
          />
        </v-col>

        <v-col cols="6">
          <v-autocomplete
            @change="(newVal) => changeOS(newVal)"
            label="Template"
            :rules="requiredRule"
            :items="osNames"
            :value="selectedTemplate?.name"
          >
          </v-autocomplete>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.password', newVal)"
            label="Password"
            :rules="requiredRule"
            :value="instance.config?.password"
          >
          </v-text-field>
        </v-col>

        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.cpu', +newVal)"
            label="CPU"
            :value="instance.resources.cpu"
            type="number"
            :rules="requiredRule"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            :rules="requiredRule"
            @change="(newVal) => setValue('resources.ram', +newVal)"
            label="RAM"
            :value="instance.resources.ram"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-select
            :items="driveTypes"
            :rules="requiredRule"
            @change="(newVal) => setValue('resources.drive_type', newVal)"
            label="Drive type"
            :value="instance.resources.drive_type"
          >
          </v-select>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue('resources.drive_size', +newVal * 1024)
            "
            :label="`Drive size (minimum ${driveSizeConfig?.minDisk} GB, maximum ${driveSizeConfig?.maxDisk} GB)`"
            :rules="[driveSizeRule]"
            :value="instance.resources.drive_size / 1024"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.ips_public', +newVal)"
            label="IPs public"
            :value="instance.resources.ips_public"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.ips_private', +newVal)"
            label="IPs private"
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
              label="VM name"
              :value="instance.data?.vm_name"
              @change="(newVal) => setValue('data.vm_name', newVal)"
            />
          </v-col>
          <v-col>
            <v-text-field
              label="VM id"
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
    drive_size: null,
    ips_public: 0,
    ips_private: 0,
  },
  data: {},
  billing_plan: {},
});
export default {
  name: "instance-ione-create",
  props: ["plans", "instance", "planRules", "sp-uuid", "is-edit"],
  data: () => ({
    bilingPlan: null,
    products: [],
    existing: false,
    requiredRule: [(val) => !!val || "Field required"],
  }),
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

      for (const [key, value] of Object.entries(this.osTemplates)) {
        if (value.name === newVal) {
          osId = key;
          break;
        }
      }

      this.setValue("config.template_id", +osId);
    },
    changeBilling(val) {
      this.bilingPlan = this.plans.find((p) => p.uuid === val);
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
    osTemplates() {
      const sp = this.$store.getters["servicesProviders/all"].filter(
        (el) => el.uuid === this.spUuid
      )[0];

      if (!sp) return {};

      const osTemplates = {};

      Object.keys(sp.publicData.templates || {}).forEach((key) => {
        if (!this.instance?.billing_plan?.meta?.hidedOs?.includes(key)) {
          osTemplates[key] = sp.publicData.templates[key];
        }
      });

      return osTemplates;
    },
    osNames() {
      if (!this.osTemplates) return [];

      return Object.values(this.osTemplates).map((os) => os.name);
    },
    driveTypes() {
      return this.instance.billing_plan?.resources
        ?.filter((r) => r.key.includes("drive"))
        .map((k) => k.key.split("_")[1].toUpperCase());
    },
    driveSizeConfig() {
      let minDisk, maxDisk;
      if (this.instance.billing_plan?.meta?.minDisk) {
        minDisk = this.instance.billing_plan.meta.minDisk;
      }
      if (this.instance.billing_plan?.meta?.maxDisk) {
        maxDisk = this.instance.billing_plan.meta.maxDisk;
      }

      if (this.selectedTemplate?.min_size) {
        minDisk = this.selectedTemplate?.min_size;
      }

      return {
        minDisk: (minDisk || 0) / 1024,
        maxDisk: (maxDisk || 1024000000) / 1024,
      };
    },
    selectedTemplate() {
      return this.osTemplates[this.instance.config.template_id];
    },
    driveSizeRule() {
      return (val) =>
        (+val >= +this.driveSizeConfig.minDisk &&
          +val <= +this.driveSizeConfig.maxDisk) ||
        "Bad drive size";
    },
    isDynamicPlan() {
      return this.instance.billing_plan?.kind === "DYNAMIC";
    },
  },
  watch: {
    "plans"() {
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
