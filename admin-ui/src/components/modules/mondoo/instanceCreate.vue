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
            @change="setValue('title', $event.target.value)"
            label="Name"
            :model-value="instance.title"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <plans-autocomplete
            :model-value="bilingPlan"
            :custom-params="{
              filters: { type: ['mondoo'] },
              anonymously: false,
            }"
            @update:model-value="setValue('billing_plan', $event)"
            return-object
            label="Price model"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Product"
            :rules="requiredRule"
            :model-value="instance.product"
            v-if="products.length > 0"
            :items="products"
            item-title="key"
            item-value="key"
            @update:model-value="setValue('product', $event)"
          >
            <template v-slot:item="{ internalItem: item, props: itemProps }">
              <v-list-item v-bind="itemProps">
                <template #title>
                  <div
                    style="width: 100%"
                    class="d-flex justify-space-between align-center"
                  >
                    <span>{{ item.raw.key }}</span>
                    <span class="ml-4">{{ item.raw.title }}</span>
                  </div>
                </template>
              </v-list-item>
            </template>
          </v-autocomplete>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12">
          <v-combobox
            label="Domains"
            :model-value="instance.config?.domains ?? []"
            multiple
            chips
            closable-chips
            :rules="domainRules"
            :hint="domainsHint"
            persistent-hint
            :delimiters="[',', ' ']"
            @update:model-value="changeDomains"
          />
        </v-col>
      </v-row>
      <v-row v-if="trial">
        <v-col cols="12">
          <v-alert type="info" variant="tonal" density="compact">
            This product has a {{ trialDays }} trial ({{ trial.scope || "account" }}
            scope). Billing grants it automatically if it is still unused;
            otherwise a start invoice is issued.
          </v-alert>
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, toRefs } from "vue";
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";

const props = defineProps(["instance", "planRules", "spUuid"]);
const { instance, planRules } = toRefs(props);

const emit = defineEmits(["set-instance", "set-value"]);

const getDefaultInstance = () => ({
  title: "instance",
  data: {},
  config: { domains: [] },
  addons: [],
  billing_plan: {},
});

const requiredRule = [(val) => !!val || "Field required"];
const domainPattern = /^(?!-)[a-z0-9-]{1,63}(?<!-)(\.(?!-)[a-z0-9-]{1,63}(?<!-))+$/;

onMounted(() => {
  emit("set-instance", getDefaultInstance());
});

const bilingPlan = computed(() => instance.value.billing_plan);
const products = computed(() =>
  Object.entries(bilingPlan.value?.products ?? {}).map(([key, product]) => ({
    key,
    title: product.title,
  }))
);
const product = computed(
  () => bilingPlan.value?.products?.[instance.value.product]
);
const limit = computed(() => +product.value?.resources?.domains || 0);
const trial = computed(() => product.value?.meta?.trial);
const trialDays = computed(() => {
  const days = (+trial.value?.period || 0) / 86400;
  return days >= 1 ? `${Math.round(days)}-day` : `${trial.value?.period}s`;
});

const domainsHint = computed(() =>
  limit.value
    ? `Up to ${limit.value} domain(s) on this product. Subdomains are discovered by the scan.`
    : "Domains to scan. Subdomains are discovered by the scan."
);
const domainRules = computed(() => [
  (val) => (val?.length ?? 0) > 0 || "At least one domain is required",
  (val) =>
    !limit.value ||
    (val?.length ?? 0) <= limit.value ||
    `This product allows ${limit.value} domain(s)`,
  (val) =>
    (val ?? []).every((d) => domainPattern.test(d)) || "Not a valid domain name",
]);

const changeDomains = (values) => {
  const domains = [
    ...new Set(
      (values ?? [])
        .map((d) => String(d).trim().toLowerCase().replace(/^https?:\/\//, "").replace(/\/.*$/, ""))
        .filter(Boolean)
    ),
  ];
  setValue("config.domains", domains);
};

const setValue = (key, value) => {
  emit("set-value", { key, value });
};
</script>

<script>
export default {
  name: "instance-mondoo-create",
};
</script>
