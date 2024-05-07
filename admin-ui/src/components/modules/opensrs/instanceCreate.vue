<template>
  <div class="module">
    <v-card
      v-if="Object.keys(instance).length > 0"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('title', newVal)"
            label="name"
            :value="instance.title"
          />
        </v-col>

        <v-col cols="6">
          <v-autocomplete
            label="Price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="plans"
            :rules="planRules"
            return-object
            @change="changeBilling"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.first_name', newVal)"
            label="First name"
            :value="instance.resources.user.first_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.last_name', newVal)"
            label="Last name"
            :value="instance.resources.user.last_name"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.org_name', newVal)"
            label="Organization name"
            :value="instance.resources.user.org_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.address1', newVal)"
            label="Address 1"
            :value="instance.resources.user.address1"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.address2', newVal)"
            label="Address 2"
            :value="instance.resources.user.address2"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.city', newVal)"
            label="City"
            :value="instance.resources.user.city"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.country', newVal)"
            label="Country"
            :value="instance.resources.user.country"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.state', newVal)"
            label="State"
            :value="instance.resources.user.state"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.postal_code', newVal)"
            label="Postal code"
            :value="instance.resources.user.postal_code"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.phone', newVal)"
            label="Phone"
            :value="instance.resources.user.phone"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.email', newVal)"
            label="Email"
            :value="instance.resources.user.email"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.reg_username', newVal)"
            label="Reg username"
            :value="instance.resources.reg_username"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.reg_password', newVal)"
            label="Reg password"
            :value="instance.resources.reg_password"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="3">
          <v-switch
            @change="(newVal) => setValue('resources.auto_renew', newVal)"
            :value="instance.resources.auto_renew"
            label="Auto_renew"
          />
        </v-col>
        <v-col cols="3">
          <v-switch
            @change="(newVal) => setValue('resources.who_is_privacy', newVal)"
            :value="instance.resources.who_is_privacy"
            label="Who_is_privacy"
          />
        </v-col>
        <v-col cols="3">
          <v-switch
            @change="(newVal) => setValue('resources.lock_domain', newVal)"
            :value="instance.resources.lock_domain"
            label="Lock_domain"
          />
        </v-col>
        <v-col cols="3">
          <v-switch
            @change="(newVal) => setValue('data.existing', newVal)"
            :value="instance.data.existing"
            label="Existing"
          />
        </v-col>
      </v-row>
      <domains-table
        :sp-uuid="spUuid"
        @input:period="setValue('resources.period', $event)"
        @input:domain="setValue('resources.domain', $event)"
        @input:price="setValue('resources.price', $event)"
      />
    </v-card>
  </div>
</template>

<script setup>
import DomainsTable from "@/components/domains_table.vue";
import { onMounted, toRefs, watch } from "vue";
import useCurrency from "@/hooks/useCurrency";
import { getMarginedValue } from "@/functions";

const props = defineProps([
  "plans",
  "instance",
  "planRules",
  "spUuid",
  "isEdit",
]);
const { instance, planRules, plans, spUuid } = toRefs(props);

const emit = defineEmits(["set-instance", "set-value"]);

const { convertFrom } = useCurrency();

onMounted(() => {
  emit("set-instance", getDefaultInstance());
});

const setValue = (key, value) => {
  emit("set-value", { key, value });
};
const changeBilling = (val) => {
  setValue("billing_plan", val);
};

watch(
  () => instance.value.resources.price,
  () => {
    if (instance.value.billing_plan.uuid) {
      const planCopy = JSON.parse(
        JSON.stringify(
          plans.value.find((p) => p.uuid === instance.value.billing_plan.uuid)
        )
      );

      const domain = instance.value.resources.domain;
      const price = convertFrom(+instance.value.resources.price, "USD");

      planCopy.products[domain] = {
        period: (instance.value.resources.period || 0) * 86400 * 365,
        price: getMarginedValue(planCopy.fee, price),
        kind: "PREPAID",
        title: `Domain ${domain}`,
        meta: {
          basePrice: price,
        },
      };

      changeBilling(planCopy);
      setValue("product", domain);
    }
  }
);
</script>

<script>
const getDefaultInstance = () => ({
  title: "instance",
  config: {},
  data: {},
  resources: {
    user: {
      first_name: "",
      last_name: "",
      org_name: "",
      address1: "",
      address2: "",
      city: "",
      country: "",
      state: "",
      postal_code: "",
      phone: "",
      email: "",
    },
    reg_username: "",
    reg_password: "",
    domain: "",
    period: 1,
    price: 0,
    auto_renew: true,
    who_is_privacy: false,
    lock_domain: true,
  },
  data: {},
  config: {},
  billing_plan: {},
});

export default {
  name: "instance-opensrs-create",
};
</script>

<style scoped></style>
