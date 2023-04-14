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
            label="title"
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
            item-value="uuid"
          ></v-select>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="domain"
            :rules="rules.req"
            :value="instance.config.domain"
            @change="setValue('config.domain', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="mail"
            :rules="rules.req"
            :value="instance.config.mail"
            @change="setValue('config.mail', $event)"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="password"
            :rules="rules.req"
            :value="instance.config.password"
            @change="setValue('config.password', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="plan"
            :rules="rules.req"
            :value="instance.resources.plan"
            @change="setValue('resources.plan', $event)"
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
    mail: "",
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
      const plan = this.plans.list.find((p) => {
        return p.uuid == this.instance.billing_plan;
      });
      this.setValue("billingPlan", plan);
      this.setValue("billing_plan", undefined);
    }
  },
  computed: {
    billingPlanId() {
      return this.instance.billing_plan || this.instance.billingPlan.uuid;
    },
  },
  methods: {
    setValue(key, value) {
      this.$emit("set-value", { key, value });
    },
  },
};
</script>
