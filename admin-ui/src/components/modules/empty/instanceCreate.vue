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
          >
          </v-text-field>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-autocomplete
            label="Price model"
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
            label="Product"
            :rules="requiredRule"
            :value="instance.product"
            v-if="products.length > 0"
            :items="products"
            item-text="key"
            item-value="key"
            @change="changeProduct"
          >
            <template v-slot:item="{ item }">
              <div
                style="width: 100%"
                class="d-flex justify-space-between align-center"
              >
                <span>{{ item.key }}</span>
                <span class="ml-4">{{ item.title }}</span>
              </div>
            </template>
          </v-autocomplete>
        </v-col>

        <v-col cols="6" v-if="addons.length">
          <v-autocomplete
            label="Addons"
            :value="instance.config.addons"
            :items="addons"
            multiple
            item-text="title"
            item-value="key"
            @change="setValue('config.addons', $event)"
          />
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script>
const getDefaultInstance = () => ({
  title: "instance",
  data: {},
  config: {
    addons: [],
  },
  billing_plan: {},
});
export default {
  name: "instance-empty-create",
  props: ["plans", "instance", "planRules", "sp-uuid", "is-edit"],
  data: () => ({
    bilingPlan: null,
    products: [],
    product: [],
    requiredRule: [(val) => !!val || "Field required"],
  }),
  mounted() {
    if (!this.isEdit) {
      this.$emit("set-instance", getDefaultInstance());
    } else {
      this.changeBilling(this.instance.billing_plan);
    }
  },
  methods: {
    changeBilling(val) {
      this.bilingPlan = this.plans.list.find((p) => p.uuid === val);
      if (this.bilingPlan) {
        this.products = Object.keys(this.bilingPlan.products).map((key) => ({
          key,
          title: this.bilingPlan.products[key].title,
        }));
      }
      this.setValue("billing_plan", this.bilingPlan);
    },
    changeProduct(val) {
      this.product = val;
      this.setValue("product", this.product);
    },
    setValue(key, value) {
      this.$emit("set-value", { key, value });
    },
  },
  computed: {
    addons() {
      return this.bilingPlan?.products[this.product]?.meta.addons || [];
    },
    autoEnabled() {
      return (
        this.addons.filter((key) => {
          return this.bilingPlan?.resources.find((r) => r.key === key)?.meta
            ?.autoEnable;
        }) || []
      );
    },
  },
  watch: {
    "plans.list"() {
      this.changeBilling(this.instance.billing_plan);
    },
    autoEnabled: {
      handler() {
        this.setValue("config.addons", this.autoEnabled);
      },
      deep: true,
    },
  },
};
</script>

<style scoped></style>
