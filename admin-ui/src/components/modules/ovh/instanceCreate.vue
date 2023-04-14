<template>
  <div class="module">
    <v-card
      v-if="Object.keys(instance).length > 0"
      class="mb-4 pa-2"
      color="background"
      elevation="0"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="title"
            :value="instance.title"
            :rules="rules.req"
            @change="(value) => setValue('title', value)"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-select
            label="price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="plans.list"
            :rules="planRules"
            @change="(value) => setValue('billing_plan', value)"
          />
        </v-col>
        <v-col cols="6" v-if="instance.products?.length > 0">
          <v-select
            label="product"
            :value="instance.productTitle"
            :items="instance.products"
            @change="(value) => setValue('product', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            label="tariff"
            item-text="title"
            item-value="code"
            :value="instance.config?.planCode"
            :items="flavors[instance?.billing_plan?.uuid]"
            :rules="rules.req"
            :loading="isFlavorsLoading"
            @change="(value) => setValue('config.planCode', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            label="region"
            :value="instance.config?.configuration.vps_datacenter"
            :items="regions[instance.config?.planCode]"
            :rules="rules.req"
            :disabled="!instance.config?.planCode"
            @change="
              (value) => setValue('config.configuration.vps_datacenter', value)
            "
          />
        </v-col>
        <v-col cols="6">
          <v-select
            label="OS"
            :value="instance.config?.configuration.vps_os"
            :items="images[instance.config?.planCode]"
            :rules="rules.req"
            :disabled="!instance.config?.planCode"
            @change="(value) => setValue('config.configuration.vps_os', value)"
          />
        </v-col>
        <v-col cols="6" class="d-flex align-center">
          Payment:
          <v-switch
            class="d-inline-block ml-2"
            true-value="P1Y"
            false-value="P1M"
            :value="instance.config?.duration"
            :label="instance.config?.duration === 'P1Y' ? 'yearly' : 'monthly'"
            @change="(value) => setValue('config.duration', value)"
          />
        </v-col>
        <v-col cols="6" class="d-flex align-center">
          Existing:
          <v-switch
            class="d-inline-block ml-2"
            :input-value="instance.data?.existing"
            @change="(value) => setValue('data.existing', value)"
          />
        </v-col>
        <v-col
          cols="6"
          class="d-flex align-center"
          v-if="instance.data?.existing"
        >
          <v-text-field
            label="VPS name"
            :value="instance.data?.vpsName"
            :rules="rules.req"
            @change="(value) => setValue('data.vpsName', value)"
          />
        </v-col>
      </v-row>

      <template
        v-if="Object.values(addons[instance.config?.planCode] || {}).length > 0"
      >
        <v-card-title class="px-0 pb-0">Addons:</v-card-title>
        <v-row>
          <v-col
            cols="6"
            v-for="(addon, key) in addons[instance.config?.planCode]"
            :key="key"
          >
            <v-select
              :label="key"
              :items="addon"
              :value="getAddonValue(addon)"
              @change="(value) => setValue('config.addons', value)"
            />
          </v-col>
        </v-row>
      </template>
    </v-card>
  </div>
</template>

<script>
import api from "@/api";

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    type: "vps",
    planCode: null,
    configuration: {
      vps_datacenter: null,
      vps_os: null,
    },
    duration: "P1M",
    pricingMode: "default",
    addons: [],
  },
  data: { existing: false },
  resources: {},
  billing_plan: {},
});
export default {
  name: "instance-ovh-create",
  props: ["plans", "instance", "planRules", "sp-uuid", "meta", "is-edit"],
  data: () => ({
    rules: {
      req: [(v) => !!v || "required field"],
    },

    isFlavorsLoading: false,
    flavors: {},
    regions: {},
    images: {},
    addons: {},
  }),
  methods: {
    addProducts(instance) {
      const { plan, billing_plan } = instance;
      const { products } =
        this.plans.list.find((el) => el.uuid === plan.uuid) || {};

      if (billing_plan.kind === "STATIC") {
        instance.products = [];
        Object.values(products || {}).forEach(({ title }) => {
          instance.products.push(title);
        });
      } else {
        delete instance.products;
        delete instance.product;
      }
    },
    fetchPlans() {
      if (this.regions.length > 0) return;
      if (this.meta && "catalog" in this.meta) return;

      this.isFlavorsLoading = true;
      api
        .post(`/sp/${this.spUuid}/invoke`, { method: "get_plans" })
        .then(({ meta }) => {
          this.$emit("set-meta", meta);
          this.setAddons(meta);
        })
        .finally(() => {
          this.isFlavorsLoading = false;
        });
    },
    setAddons(meta) {
      this.plans.list.forEach(({ products, resources }) => {
        for (let key in products) {
          key = key.split(" ")[1];
          if (key in this.addons) continue;

          const plans = meta ? meta : this.meta;
          const plan = plans?.catalog.plans.find(
            ({ planCode }) => planCode === key
          );

          plan?.configurations.forEach((el) => {
            el.values.sort();
            if (el.name.includes("os")) {
              this.$set(this.images, key, el.values);
            }
            if (el.name.includes("datacenter")) {
              this.$set(this.regions, key, el.values);
            }
          });

          plan?.addonFamilies.forEach((el) => {
            if (!this.addons[key]) {
              this.addons[key] = {};
            }
            if (el.name === "snapshot") {
              this.addons[key].snapshot = el.addons.filter((addon) =>
                resources.find(({ key }) => key.includes(addon))
              );
            }
            if (el.name === "additionalDisk") {
              this.addons[key].disk = el.addons.filter((addon) =>
                resources.find(({ key }) => key.includes(addon))
              );
            }
            if (el.name === "automatedBackup") {
              this.addons[key].backup = el.addons.filter((addon) =>
                resources.find(({ key }) => key.includes(addon))
              );
            }
          });
        }
      });
    },
    setValue(path, val) {
      if (!val) {
        return;
      }

      const data = JSON.parse(JSON.stringify(this.instance));

      if (path.includes("billing_plan")) {
        const plan = this.plans.list.find(({ uuid }) => val === uuid);
        const title = plan.title.split(" ");

        title.pop();
        this.flavors[val] = Object.keys(plan.products).map((el) => ({
          code: el.split(" ")[1],
          title: plan.products[el].title,
        }));

        data.plan = val;
        val = { ...plan, title: title.join(" ") };
      }

      if (path.includes("product")) {
        const plan = data.billing_plan;
        const [product] = Object.entries(plan.products).find(
          ([, prod]) => prod.title === val
        );

        data.productTitle = val;
        val = product;
      }

      if (path.includes("planCode")) {
        const plan = this.meta?.catalog.plans.find(
          ({ planCode }) => planCode === val
        );
        const resources = val.split("-");

        plan?.configurations.forEach((el) => {
          el.values.sort();
          if (el.name.includes("os")) {
            this.$set(this.images, val, el.values);
          }
          if (el.name.includes("datacenter")) {
            this.$set(this.regions, val, el.values);
          }
        });

        this.$emit("set-value", { key: "resources", value: {
          cpu: +resources.at(-3),
          ram: resources.at(-2) * 1024,
          drive_size: resources.at(-1) * 1024,
          drive_type: "SSD",
          ips_private: 0,
          ips_public: 1,
        } });
      }

      if (path.includes("duration")) {
        data.config.pricingMode = val === "P1M" ? "default" : "upfront12";
      }

      if (path.includes("addons")) {
        const { addons } = data.config;

        val = [...addons, val];
      }

      this.$emit("set-value", { value: val, key: path });
      if (path.includes("billing_plan")) this.addProducts(data);
      this.change(data);
    },
    change(data) {
      this.$emit("update:instances-group", data);
    },
    getAddonValue(addon){
      return this.instance.config.addons.find(a=>addon.includes(a))
    }
  },
  async created() {
    if (!this.isEdit) {
      this.$emit("set-instance", getDefaultInstance());
    } else if (!this.instance.billing_plan?.uuid) {
      await this.fetchPlans()
      this.setValue("billing_plan", this.instance.billing_plan);
      this.setValue("config.planCode", this.instance.config.planCode);
      this.setValue(
        "config.configuration.vps_datacenter",
        this.instance.config.configuration.vps_datacenter
      );
      this.setValue(
        "config.configuration.vps_os",
        this.instance.config.configuration.vps_os
      );
      this.setAddons()
    }
    const data = JSON.parse(JSON.stringify(getDefaultInstance()));

    if (data.billingPlan) {
      data.billing_plan = data.billingPlan;
      delete data.billingPlan;
    }
    this.setValue(`billing_plan`, data.billing_plan.uuid);
    data.plan = data.billing_plan.uuid;

    if (this.meta && "catalog" in this.meta) this.setAddons();
    this.change(data);
  },
  watch: {
    instance() {
      this.fetchPlans();
    },
  },
};
</script>

<style scoped></style>
