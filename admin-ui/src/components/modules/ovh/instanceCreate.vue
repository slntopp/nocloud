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
        <v-col cols="6">
          <v-select
            label="type"
            :items="ovhTypes"
            :rules="rules.req"
            v-model="ovhType"
            item-value="value"
            item-text="title"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-autocomplete
            label="price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="filtredPlans"
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
            :items="tariffs"
            :rules="rules.req"
            :loading="isFlavorsLoading"
            @change="(value) => setValue('config.planCode', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            label="region"
            :value="instance.config?.configuration[`${ovhType}_datacenter`]"
            :items="regions[instance.config?.planCode]"
            :rules="rules.req"
            :disabled="!instance.config?.planCode"
            @change="
              (value) =>
                setValue(`config.configuration.${ovhType}_datacenter`, value)
            "
          />
        </v-col>
        <v-col cols="6">
          <v-select
            label="OS"
            :value="instance.config?.configuration[`${ovhType}_os`]"
            :items="images[instance.config?.planCode]"
            :rules="rules.req"
            :disabled="!instance.config?.planCode"
            @change="
              (value) => setValue(`config.configuration.${ovhType}_os`, value)
            "
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
            v-if="ovhType"
            :label="`${ovhType} name`"
            :value="instance.data?.[`${ovhType}Name`]"
            :rules="rules.req"
            @change="(value) => setValue(`data.${ovhType}Name`, value)"
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

    ovhTypes: [
      { title: "ovh vps", value: "vps" },
      { title: "ovh cloud", value: "cloud" },
      { title: "ovh dedicated", value: "dedicated" },
    ],
    ovhType: "vps",
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
    setAddons(addons, planCode) {
      const newAddons = {};

      const alowwed = [
        "snapshot",
        "disk",
        "backup",
        "traffic",
        "ram",
        "softraid",
      ];

      addons?.forEach((addon) => {
        const key = alowwed.find((a) => addon.includes(a));
        if (key) {
          newAddons[key] = !newAddons[key]?.length
            ? [addon]
            : [...newAddons[key], addon];
        }
      });

      this.addons[planCode] = newAddons;
    },
    setValue(path, val) {
      const data = JSON.parse(JSON.stringify(this.instance));

      if (path.includes("billing_plan")) {
        const plan = this.plans.list.find(({ uuid }) => val === uuid);
        const title = plan?.title.split(" ");

        title.pop();
        this.flavors[val] = Object.keys(plan.products).map((el) => {
          const [duration, code] = el.split(" ");
          return {
            code,
            duration,
            title: plan.products[el].title,
            key: el,
          };
        });

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
        const flavor = this.flavors[this.instance.billing_plan.uuid].find(
          (f) => f.code === val
        );
        const product = this.instance.billing_plan.products[flavor.key];
        const resources = product.resources;
        this.images[val] = product.meta.os.map((os)=>os.name?os.name:os);
        this.regions[val] = product.meta.datacenter;

        this.setAddons(product.meta.addons, val);

        const savedResources = {
          ips_private: 0,
          ips_public: 1,
        };
        if (this.ovhType === "vps") {
          savedResources.cpu = +resources.cpu;
          savedResources.ram = resources.ram;
          savedResources.drive_size = resources.disk;
          savedResources.drive_type = "SSD";
        }

        this.$emit("set-value", {
          key: "resources",
          value: savedResources,
        });
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
      this.setProduct();
    },
    change(data) {
      this.$emit("update:instances-group", data);
    },
    getAddonValue(addon) {
      return this.instance.config.addons.find((a) => addon.includes(a));
    },
    setProduct() {
      const data = JSON.parse(JSON.stringify(this.instance));
      if (data.billing_plan?.kind?.toLowerCase() === "static") {
        this.$emit("set-value", {
          value: `${data.config?.duration} ${data.config?.planCode}`,
          key: "product",
        });
      }
    },
  },
  computed: {
    filtredPlans() {
      return this.plans?.list?.filter((p) => p.type.includes(this.ovhType));
    },
    tariffs() {
      const tariffs = this.flavors[this.instance?.billing_plan?.uuid];
      if (tariffs && this.instance.billing_plan) {
        return tariffs.filter(
          (t) => t?.duration === this.instance.config?.duration
        );
      }

      return tariffs;
    },
  },
  async created() {
    if (!this.isEdit) {
      this.$emit("set-instance", getDefaultInstance());
      return
    } else if (!this.instance.billing_plan?.uuid) {
      this.ovhType = this.instance.config.type;
      this.setValue("billing_plan", this.instance.billing_plan);
      this.setValue("config.planCode", this.instance.config.planCode);
      this.setValue(
        `config.configuration.${this.ovhType}_datacenter`,
        this.instance.config.configuration.vps_datacenter
      );
      this.setValue(
        `config.configuration.${this.ovhType}_os`,
        this.instance.config.configuration.vps_os
      );
      this.setAddons(
        this.instance.billing_plan.products[
          `${this.instance.config.duration} ${this.instance.config.planCode}`
        ]?.meta?.addons,
        this.instance.config.planCode
      );
    }
    const data = JSON.parse(JSON.stringify(getDefaultInstance()));

    if (data.billingPlan) {
      data.billing_plan = data.billingPlan;
      delete data.billingPlan;
    }
    this.setValue(`billing_plan`, data.billing_plan.uuid);
    data.plan = data.billing_plan.uuid;

    this.change(data);
  },
  watch: {
    ovhType() {
      for (const key of Object.keys(this.instance.config.configuration)) {
        this.setValue("config.configuration." + key, undefined);
      }
      this.setValue("data", { existing: false });
      this.setValue("config.type", this.ovhType);
    },
  },
};
</script>

<style scoped></style>
