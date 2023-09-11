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
            v-if="products.length > 0"
            :items="products"
            item-text="key"
            item-value="key"
            @change="setValue('product', $event)"
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
      </v-row>
    </v-card>
  </div>
</template>

<script>
const getDefaultInstance = () => ({
  title: "instance",
  data: {},
  billing_plan: {},
});
export default {
  name: "instance-virtual-create",
  props: ["plans", "instance", "planRules", "sp-uuid", "is-edit"],
  data: () => ({ bilingPlan: null, products: [] }),
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
    setValue(key, value) {
      this.$emit("set-value", { key, value });
    },
  },
  watch: {
    "plans.list"() {
      this.changeBilling(this.instance.billing_plan);
    },
  },
};
</script>

<style scoped></style>
