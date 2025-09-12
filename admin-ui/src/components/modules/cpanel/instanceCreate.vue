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
          <plans-autocomplete
            :value="billingPlanId"
            :custom-params="{
              filters: { type: ['cpanel'], 'meta.isIndividual': [false] },
              anonymously: false,
            }"
            @input="setValue('billing_plan', $event)"
            return-object
            label="Price model"
            :rules="planRules"
          />
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
        <v-col cols="6" v-if="tarrifAddons.length > 0">
          <v-autocomplete
            @change="(newVal) => setValue('addons', newVal)"
            label="Addons"
            :value="instance.addons"
            :items="isAddonsLoading ? [] : getAvailableAddons()"
            :loading="isAddonsLoading"
            item-value="uuid"
            item-text="title"
            multiple
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="2">
          <v-switch label="Existing" v-model="existing" />
        </v-col>
        <template v-if="existing">
          <v-col>
            <v-text-field
              label="Username"
              :value="instance.data?.username"
              @change="(newVal) => setValue('data.username', newVal)"
            />
          </v-col>
          <v-col>
            <v-text-field
              label="Password"
              :value="instance.data?.password"
              @change="(newVal) => setValue('data.password', newVal)"
            />
          </v-col>
        </template>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import useInstanceAddons from "@/hooks/useInstanceAddons";
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";

const props = defineProps(["instance", "planRules", "spUuid"]);
const { instance, planRules } = toRefs(props);

const emit = defineEmits(["set-instance", "set-value"]);

const { tarrifAddons, setTariffAddons, getAvailableAddons, isAddonsLoading } =
  useInstanceAddons(instance, (key, value) => setValue(key, value));

const rules = ref({
  req: [(v) => !!v || "required field"],
});

const existing = ref(false);

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    domain: "",
    email: "",
    password: "",
  },
  addons: [],
  resources: {
    plan: "",
  },
  data: {},
  billing_plan: {},
});

onMounted(() => {
  emit("set-instance", getDefaultInstance());
});

const billingPlanId = computed(() => {
  return instance.value.billing_plan.uuid;
});
const products = computed(() => {
  const plan = instance.value.billing_plan;
  return Object.keys(plan?.products || {}).map((key) => ({
    ...plan.products[key],
    key,
  }));
});

const setValue = (key, value) => {
  if (key === "resources.plan") {
    setValue("product", value);
    const product = products.value.find((p) => p.key === value);
    Object.keys(product.resources || {}).forEach((key) => {
      setValue("resources." + key, product.resources[key]);
    });
    setTariffAddons();
  }

  emit("set-value", { key, value });
};

watch(existing, () => {
  setValue("config.auto_start", existing.value);
  setValue("data.existing", existing.value);
  setValue("data.username", null);
  setValue("data.password", null);
});
</script>

<script>
export default {
  name: "instance-cpanel-create",
};
</script>
