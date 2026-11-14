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
            @change="setValue('title', $event.target.value)"
            label="name"
            :model-value="instance.title"
          />
        </v-col>

        <v-col cols="6">
          <plans-autocomplete
            :value="instance.billing_plan"
            :custom-params="{
              filters: { type: ['opensrs'], 'meta.isIndividual': [false] },
              anonymously: false,
            }"
            @input="changeBilling"
            return-object
            label="Price model"
            :rules="planRules"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.first_name', $event.target.value)"
            label="First name"
            :model-value="instance.resources.user.first_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.last_name', $event.target.value)"
            label="Last name"
            :model-value="instance.resources.user.last_name"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.org_name', $event.target.value)"
            label="Organization name"
            :model-value="instance.resources.user.org_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.address1', $event.target.value)"
            label="Address 1"
            :model-value="instance.resources.user.address1"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.address2', $event.target.value)"
            label="Address 2"
            :model-value="instance.resources.user.address2"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.city', $event.target.value)"
            label="City"
            :model-value="instance.resources.user.city"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.country', $event.target.value)"
            label="Country"
            :model-value="instance.resources.user.country"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.state', $event.target.value)"
            label="State"
            :model-value="instance.resources.user.state"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.postal_code', $event.target.value)"
            label="Postal code"
            :model-value="instance.resources.user.postal_code"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.phone', $event.target.value)"
            label="Phone"
            :model-value="instance.resources.user.phone"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.user.email', $event.target.value)"
            label="Email"
            :model-value="instance.resources.user.email"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.reg_username', $event.target.value)"
            label="Reg username"
            :model-value="instance.resources.reg_username"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.reg_password', $event.target.value)"
            label="Reg password"
            :model-value="instance.resources.reg_password"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="3">
          <v-switch
            @update:model-value="(newVal) => setValue('resources.auto_renew', newVal)"
            :value="instance.resources.auto_renew"
            label="Auto_renew"
          />
        </v-col>
        <v-col cols="3">
          <v-switch
            @update:model-value="(newVal) => setValue('resources.who_is_privacy', newVal)"
            :value="instance.resources.who_is_privacy"
            label="Who_is_privacy"
          />
        </v-col>
        <v-col cols="3">
          <v-switch
            @update:model-value="(newVal) => setValue('resources.lock_domain', newVal)"
            :value="instance.resources.lock_domain"
            label="Lock_domain"
          />
        </v-col>
        <v-col cols="3">
          <v-switch
            @update:model-value="(newVal) => setValue('data.existing', newVal)"
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
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";

const props = defineProps(["instance", "planRules", "spUuid", "isEdit"]);
const { instance, planRules, spUuid } = toRefs(props);

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
    if (instance.value.billing_plan?.uuid) {
      const planCopy = JSON.parse(JSON.stringify(instance.value.billing_plan));

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
  billing_plan: {},
});

export default {
  name: "instance-opensrs-create",
};
</script>

<style scoped></style>
