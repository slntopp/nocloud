<template>
  <div class="module">
    <instance-create-card
      v-for="(instance, index) of instances"
      :key="index"
      :instance="instance || {}"
      :plans="plans"
      :plan-rules="planRules"
      :flavors="flavors"
      :images="images"
      :regions="regions"
      :addons="addons"
      :is-flavors-loading="isFlavorsLoading"
      @remove="remove(index)"
      @set-value="setValue(`${index}.${$event.key}`, $event.value)"
    />

    <v-row>
      <v-col class="d-flex justify-center">
        <add-instance :disabled="isDisabled" @add="addInstance" />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import api from "@/api.js";
import InstanceCreateCard from "@/components/modules/ovh/instanceCreateCard.vue";
import AddInstance from "@/components/ui/addInstance.vue";

export default {
  name: "ovh-create-service-module",
  components: {AddInstance, InstanceCreateCard },
  props: ["instances-group", "plans", "planRules", "meta"],
  data: () => ({
    defaultItem: {
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
      billing_plan: {},
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
        this.plans.list.find((el) => el?.uuid === plan?.uuid) || {};

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
    addInstance() {
      const item = JSON.parse(JSON.stringify(this.defaultItem));
      const data = JSON.parse(this.instancesGroup);
      const i = data.body.instances.length;

      item.title += "#" + (i + 1);
      data.body.instances.push(item);
      this.change(data);
    },
    remove(index) {
      const data = JSON.parse(this.instancesGroup);

      data.body.instances.splice(index, 1);
      this.change(data);
    },
    fetchPlans() {
      const data = JSON.parse(this.instancesGroup);

      if (data.body.type !== "ovh") return;
      if (this.regions.length > 0) return;
      if ("catalog" in this.meta) return;

      this.isFlavorsLoading = true;
      return api
        .post(`/sp/${data.sp}/invoke`, { method: "get_plans" })
        .then(({ meta }) => {
          this.$emit("changeMeta", meta);
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
          const plan = plans.catalog.plans.find(
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
      const data = JSON.parse(this.instancesGroup);
      const i = path.split(".")[0];

      if (path.includes("billing_plan")) {
        const plan = this.plans.list.find(({ uuid }) => val === uuid);
        console.log(this.flavors, this.plans.list, plan);
        const title = plan?.title.split(" ");

        title?.pop();
        this.flavors[val] = Object.keys(plan.products).map((el) => ({
          code: el.split(" ")[1],
          title: plan.products[el]?.title,
        }));

        data.body.instances[i].plan = val;
        val = { ...plan, title: title.join(" ") };
      }

      if (path.includes("product")) {
        const plan = data.body.instances[i].billing_plan;
        const [product] = Object.entries(plan.products).find(
          ([, prod]) => prod?.title === val
        );

        data.body.instances[i].productTitle = val;
        val = product;
      }

      if (path.includes("planCode")) {
        const plan = this.meta.catalog.plans.find(
          ({ planCode }) => planCode === val
        );
        const resources = val.split("-");

        plan.configurations.forEach((el) => {
          el.values.sort();
          if (el.name.includes("os")) {
            this.$set(this.images, val, el.values);
          }
          if (el.name.includes("datacenter")) {
            this.$set(this.regions, val, el.values);
          }
        });

        data.body.instances[i].resources = {
          cpu: +resources.at(-3),
          ram: resources.at(-2) * 1024,
          drive_size: resources.at(-1) * 1024,
          drive_type: "SSD",
          ips_private: 0,
          ips_public: 1,
        };
      }

      if (path.includes("duration")) {
        data.body.instances[i].config.pricingMode =
          val === "P1M" ? "default" : "upfront12";
      }

      if (path.includes("addons")) {
        const { addons } = data.body.instances[i].config;

        val = [...addons, val];
      }

      setToValue(data.body.instances, val, path);
      if (path.includes("billing_plan"))
        this.addProducts(data.body.instances[i]);
      this.change(data);
    },
    change(data) {
      this.$emit("update:instances-group", JSON.stringify(data));
    },
  },
  computed: {
    instances() {
      return JSON.parse(this.instancesGroup).body.instances;
    },
    isDisabled() {
      const group = JSON.parse(this.instancesGroup);

      return group.body.type === "ovh" && !group.sp;
    },
  },
  created() {
    const data = JSON.parse(this.instancesGroup);

    if (!data.body.instances) data.body.instances = [];
    else {
      data.body.instances.forEach((inst, i, arr) => {
        if (inst.billingPlan) {
          arr[i].billing_plan = inst.billingPlan;
          delete arr[i].billingPlan;
        }
        if (this.plans.list.length > 0) {
          console.log(1);
          this.setValue(`${i}.billing_plan`, inst.billing_plan?.uuid);
        }
        arr[i].plan = inst.billing_plan?.uuid;
      });
    }

    if ("catalog" in this.meta) this.setAddons();
    this.change(data);
  },
  watch: {
    async instances() {
      await this.fetchPlans();
      const data = JSON.parse(this.instancesGroup);

      if (!data.body.instances) data.body.instances = [];
      else {
        data.body.instances.forEach((inst, i, arr) => {
          if (inst.billingPlan) {
            arr[i].billing_plan = inst.billingPlan;
            delete arr[i].billingPlan;
          }
          console.log(2);
          console.log(`${i}.billing_plan`, inst.billing_plan?.uuid, this.plans);
          this.setValue(`${i}.billing_plan`, inst.billing_plan?.uuid);
          arr[i].plan = inst.billing_plan?.uuid;
        });
      }
    },
  },
};

function setToValue(obj, value, path) {
  path = path.split(".");
  let i;
  for (i = 0; i < path.length - 1; i++) {
    if (path[i] === "__proto__" || path[i] === "constructor")
      throw new Error("Can't use that path because of: " + path[i]);
    obj = obj[path[i]];
  }

  obj[path[i]] = value;
}
</script>
