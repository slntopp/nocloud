<template>
  <div class="module">
    <instance-create-card
      v-for="(instance, index) in instances"
      :key="index"
      @set-value="setValue(`${index}.${$event.key}`, $event.value)"
      @change-os="changeOS(index, $event)"
      @remove="remove(index)"
      :instance="instance"
      :plan-rules="planRules"
      :os-name="getOsTemplates[instance.config.template_id]?.name"
      :os-names="getOsNames"
      :plans="plans"
      :plans-products="getPlanProducts(index)"
    />
    <v-row>
      <v-col class="d-flex justify-center">
        <add-instance @add="addInstance" />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import InstanceCreateCard from "@/components/modules/ione/instanceCreateCard.vue";
import AddInstance from "@/components/ui/addInstance.vue";

export default {
  name: "ione-create-service-module",
  components: {AddInstance, InstanceCreateCard },
  props: {
    "instances-group": {},
    plans: {},
    planRules: {},
  },
  data: () => ({
    defaultItem: {
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
    },
    types: ["SSD", "HDD"],
    // instances: []
  }),
  methods: {
    addProducts(instance) {
      const { plan, billing_plan } = instance;
      const products =
        this.plans.list.find((el) => el.uuid.includes(plan.uuid))?.products ||
        {};

      if (billing_plan.kind === "STATIC") {
        instance.products = [];
        Object.values(products).forEach(({ title }) => {
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
      item.title += "#" + (data.body.instances.length + 1);

      data.body.instances.push(item);
      this.change(data);
    },
    remove(index) {
      const data = JSON.parse(this.instancesGroup);

      data.body.instances.splice(index, 1);
      this.change(data);
    },
    setValue(path, val) {
      const data = JSON.parse(this.instancesGroup);
      const i = path.split(".")[0];

      if (path.includes("plan")) {
        const plan = this.plans.list.find(({ uuid }) => val.includes(uuid));
        const j = plan.title.length - 14;

        data.body.instances[i].plan = val;
        val = { ...plan, title: plan.title.slice(0, j) };
      }
      if (path.includes("product")) {
        const plan = data.body.instances[i].billing_plan;
        const [product] = Object.entries(plan.products).find(
          ([, prod]) => prod.title === val
        );

        data.body.instances[i].productTitle = val;
        val = product;
      }

      setToValue(data.body.instances, val, path);
      if (path.includes("plan")) this.addProducts(data.body.instances[i]);
      this.change(data);
    },
    change(data) {
      this.$emit("update:instances-group", JSON.stringify(data));
    },
    changeOS(index, newVal) {
      let osId = null;

      for (const [key, value] of Object.entries(this.getOsTemplates)) {
        if (value.name === newVal) {
          osId = key;
          break;
        }
      }

      this.setValue(index + ".config.template_id", +osId);
    },
    getPlanProducts(index) {
      if (!this.instances[index].billing_plan?.products) {
        return [];
      }
      return Object.values(this.instances[index].billing_plan.products).map(
        (p) => p.title
      );
    },
  },
  computed: {
    instances() {
      const data = JSON.parse(this.instancesGroup);
      return data.body.instances;
    },

    getOsTemplates() {
      const data = JSON.parse(this.instancesGroup);

      const sp = this.$store.getters["servicesProviders/all"].filter(
        (el) => el.uuid === data.sp
      )[0];

      if (!sp) return {};

      return sp.publicData.templates;
    },

    getOsNames() {
      if (!this.getOsTemplates) return [];

      return Object.values(this.getOsTemplates).map((os) => os.name);
    },
  },
  created() {
    const data = JSON.parse(this.instancesGroup);

    if (!data.body.instances) {
      data.body.instances = [];
    } else {
      data.body.instances.forEach((inst, i, arr) => {
        if (inst.billingPlan) {
          arr[i].billing_plan = inst.billingPlan;
          delete arr[i].billingPlan;
        }
        if (inst.product) arr[i].productTitle = inst.product;
        arr[i].plan = inst.billing_plan.uuid;
      });
    }

    this.change(data);
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

<style></style>