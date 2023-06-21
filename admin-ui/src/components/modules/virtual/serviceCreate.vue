<template>
  <div class="module">
    <v-card
      v-for="(instance, index) in instances"
      :key="index"
      :id="instance.uuid"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue(index + '.title', newVal)"
            label="title"
            v-model="instance.title"
          >
          </v-text-field>
        </v-col>
        <v-col class="d-flex justify-end">
          <v-btn @click="() => remove(index)"> remove</v-btn>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-select
            @change="(newVal) => changeOS(index, newVal)"
            label="template"
            :items="getOsNames"
            :value="getOsTemplates[instance.config.template_id]?.name"
          >
          </v-select>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue(index + '.config.password', newVal)"
            label="password"
            v-model="instance.config.password"
          >
          </v-text-field>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue(index + '.resources.cpu', +newVal)"
            label="cpu"
            v-model="instance.resources.cpu"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue(index + '.resources.ram', +newVal)"
            label="ram"
            v-model="instance.resources.ram"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.drive_type', newVal)
            "
            label="drive type"
            v-model="instance.resources.drive_type"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.drive_size', +newVal)
            "
            label="drive size"
            v-model="instance.resources.drive_size"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.ips_public', +newVal)
            "
            label="ips public"
            v-model="instance.resources.ips_public"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.ips_private', +newVal)
            "
            label="ips private"
            v-model="instance.resources.ips_private"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="price model"
            item-text="title"
            item-value="uuid"
            v-model="instance.plan"
            :items="plans.list"
            :rules="planRules"
            @change="(newVal) => setValue(index + '.billing_plan', newVal)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            label="product"
            v-model="instance.product"
            v-if="getPlanProducts(index).length > 0"
            :items="getPlanProducts(index)"
            @change="(newVal) => setValue(index + '.product', newVal)"
          />
        </v-col>
      </v-row>
    </v-card>
    <v-row>
      <v-col class="d-flex justify-center">
        <add-instance-btn @click="addInstance" />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import AddInstanceBtn from "@/components/ui/addInstanceBtn.vue";

export default {
  name: "ione-create-service-module",
  components: { AddInstanceBtn },
  props: ["instances-group", "plans", "planRules"],
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
        const productBody = plan.products[val];

        Object.keys(productBody.resources).forEach((key) => {
          data.body.instances[i].resources[key] = productBody.resources[key];
        });

        data.body.instances[i].productTitle = val;
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
      return Object.keys(this.instances[index].billing_plan.products);
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