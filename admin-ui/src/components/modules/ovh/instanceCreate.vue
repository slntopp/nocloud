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
            label="Name"
            :value="instance.title"
            :rules="rules.req"
            @change="(value) => setValue('title', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="Type"
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
            :filter="defaultFilterObject"
            label="Price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="filtredPlans"
            :rules="planRules"
            @change="(value) => setValue('billing_plan', value)"
          />
        </v-col>
        <v-col cols="6" v-if="instance.products?.length > 0">
          <v-autocomplete
            label="Product"
            :value="instance.productTitle"
            :items="instance.products"
            @change="(value) => setValue('product', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            :items="durationItems"
            :value="instance.config?.duration"
            item-value="value"
            item-text="title"
            label="Payment:"
            @change="(value) => setValue('config.duration', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Tariff"
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
          <v-autocomplete
            label="Region"
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
          <v-autocomplete
            label="OS"
            :value="instance.config?.configuration[`${ovhType}_os`]"
            :items="images[instance.config?.planCode]"
            item-text="title"
            item-value="id"
            :rules="rules.req"
            :disabled="!instance.config?.planCode"
            @change="
              (value) => setValue(`config.configuration.${ovhType}_os`, value)
            "
          />
        </v-col>
        <v-col v-if="ovhType === 'cloud'" cols="6">
          <v-text-field
            label="SSH"
            :value="instance?.config?.ssh"
            @change="setValue('config.ssh', $event)"
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
            <v-autocomplete
              :label="key"
              item-text="title"
              return-object
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
import { defaultFilterObject } from "@/functions";

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    auto_renew: false,
    type: "vps",
    planCode: null,
    configuration: {
      vps_datacenter: null,
      vps_os: null,
    },
    duration: "",
    pricingMode: "",
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
    defaultFilterObject,
    addProducts(instance) {
      const { plan, billing_plan } = instance;
      const { products } = this.plans.find((el) => el.uuid === plan.uuid) || {};

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
        "vrack",
        "storage",
        "system-storage",
        "bandwidth",
        "memory",
      ];
      addons?.forEach((addonKey) => {
        const key = alowwed.find((a) => addonKey.includes(a));

        const realAddon = this.instance.billing_plan.resources.find(
          (a) => a.key === addonKey
        );

        const addonId =
          this.ovhType === "vps"
            ? addonKey.split(" ")[1]
            : addonKey.split(" ")[2];

        if (key && realAddon?.public) {
          if (!newAddons[key]) {
            newAddons[key] = [];
          }
          newAddons[key].push({ ...realAddon, type: key, id: addonId });
        }
      });

      this.addons[planCode] = newAddons;
    },
    setValue(path, val) {
      const data = JSON.parse(JSON.stringify(this.instance));

      if (path.includes("billing_plan")) {
        const plan = this.plans.find(({ uuid }) => val === uuid);
        const title = plan?.title.split(" ");

        title.pop();
        this.flavors[val] = Object.keys(plan.products)
          .map((el) => {
            const [duration, code] = el.split(" ");
            return {
              public: !!plan.products[el].public,
              code,
              duration,
              title: plan.products[el].title,
              key: el,
            };
          })
          .filter((el) => el.public);

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
        this.images[val] = product.meta.os.map((os) => ({
          title: os.name ? os.name : os,
          id: os.id ? os.id : os,
        }));
        this.regions[val] = product.meta.datacenter;

        this.setAddons(product.meta.addons, val);
        let savedResources = {
          ips_private: 0,
          ips_public: 1,
        };
        switch (this.ovhType) {
          case "vps": {
            savedResources.cpu = +resources.cpu;
            savedResources.ram = resources.ram;
            savedResources.drive_size = resources.disk;
            savedResources.drive_type = "SSD";
            break;
          }
          case "cloud": {
            this.setValue(
              "config.monthlyBilling",
              this.instance.config?.duration === "P1M"
            );
            savedResources = { ...savedResources, ...resources };
            break;
          }
        }

        this.$emit("set-value", {
          key: "resources",
          value: savedResources,
        });
      }

      if (path.includes("duration")) {
        this.$emit("set-value", {
          value: val === "P1Y" ? "upfront12" : "default",
          key: "config.pricingMode",
        });
      }

      if (path.includes("addons")) {
        let { addons } = data.config;

        addons = (addons || []).filter((a) => !a.includes(val.type));

        if (this.ovhType === "dedicated") {
          const dedicatedKeys = ["ram", "softraid"];
          const newAddonKey = dedicatedKeys.find((key) => val.id.includes(key));
          addons = addons.filter((a) => !a.includes(newAddonKey));

          const resources = {};
          for (let addonKey of [...addons, val.id]) {
            if (addonKey.includes("ram")) {
              resources.ram = parseInt(addonKey?.split("-")[1] ?? 0);
            }
            if (addonKey.includes("softraid")) {
              const [count, size] = addonKey?.split("-")[1].split("x") ?? [
                "0",
                "0",
              ];

              resources.drive_size = count * parseInt(size) * 1024;
              if (addonKey?.includes("hybrid"))
                resources.drive_type = "SSD + HDD";
              else if (size.includes("sa")) resources.drive_type = false;
              else resources.drive_type = "SSD";
            }
          }
          this.setValue("resources", {
            ...this.instance.resources,
            ...resources,
          });
        }

        val = [...addons, val.id];
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
      this.$emit("set-value", {
        value: this.product,
        key: "product",
      });
    },
  },
  computed: {
    filtredPlans() {
      return this.plans?.filter((p) => p.type.includes(this.ovhType));
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
    product() {
      return [
        this.instance.config.duration,
        this.instance.config.planCode,
      ].join(" ");
    },
    durationItems() {
      const annotations = {
        P1M: "Monthly",
        P1Y: "Yearly",
        P1D: "Daily",
        P1H: "Hourly",
      };

      return [
        ...new Set(
          this.flavors[this.instance.billing_plan?.uuid]?.map((item) => ({
            value: item.duration,
            title: annotations[item.duration] || item.duration,
          }))
        ),
      ];
    },
  },
  async created() {
    if (!this.isEdit) {
      this.$emit("set-instance", getDefaultInstance());
      return;
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
        this.instance.billing_plan.products[this.product]?.meta?.addons,
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
      const title = this.instance.title;
      this.$emit("set-instance", getDefaultInstance());
      this.setValue("config.type", this.ovhType);
      this.setValue("title", title);
    },
  },
};
</script>

<style scoped></style>
