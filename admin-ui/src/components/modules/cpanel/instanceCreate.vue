<template>
  <div class="module">
    <v-card
      class="mb-4 pa-2"
      color="background"
      elevation="0"
      :id="instance.uuid"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="Name"
            :value="instance.title"
            :rules="rules.req"
            @change="setValue('title', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            :items="plans.list"
            :value="billingPlanId"
            @change="setValue('billingPlan', $event)"
            item-text="title"
            return-object
            label="Price model"
            :rules="planRules"
            item-value="uuid"
          ></v-select>
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Product"
            :items="products"
            item-text="title"
            item-value="key"
            :rules="rules.req"
            :value="instance.resources.plan"
            @change="setValue('resources.plan', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="Domain"
            :rules="rules.req"
            :value="instance.config.domain"
            @change="setValue('config.domain', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="Email"
            :rules="rules.req"
            :value="instance.config.email"
            @change="setValue('config.email', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="Password"
            :rules="rules.req"
            :value="instance.config.password"
            @change="setValue('config.password', $event)"
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
    domain: "",
    email: "",
    password: "",
  },
  resources: {
    plan: "",
  },
  billingPlan: {},
});
export default {
  name: "instance-cpanel-create",
  props: ["plans", "instance", "planRules", "sp-uuid", "is-edit"],
  data: () => ({
    rules: {
      req: [(v) => !!v || "required field"],
    },
  }),
  mounted() {
    if (!this.isEdit) {
      this.$emit("set-instance", getDefaultInstance());
    } else {
      const plan = this.plans.list.find(
        (p) => p.uuid == this.instance.billing_plan
      );
      this.setValue("billingPlan", plan);
      this.setValue("billing_plan", undefined);
    }
  },
  computed: {
    billingPlanId() {
      return this.instance.billing_plan || this.instance.billingPlan.uuid;
    },
    products() {
      const plan = this.plans.list.find((p) => p.uuid == this.billingPlanId);
      return Object.keys(plan?.products || {}).map((key) => ({
        ...plan.products[key],
        key,
      }));
    },
  },
  methods: {
    setValue(key, value) {
      if (key === "resources.plan") {
        this.setValue("product", value);
        const product = this.products.find((p) => p.key === value);
        Object.keys(product.resources || {}).forEach((key) => {
          this.setValue("resources." + key, product.resources[key]);
        });
      }

      this.$emit("set-value", { key, value });
    },
  },
};
</script>
