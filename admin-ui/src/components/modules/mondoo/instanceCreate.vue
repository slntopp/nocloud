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
                    <span class="ml-4">
                      {{ item.raw.title }}
                      <span
                        v-if="item.raw.title !== item.raw.kindTitle"
                        class="text-medium-emphasis"
                        >({{ item.raw.kindTitle }})</span
                      >
                    </span>
                  </div>
                </template>
              </v-list-item>
            </template>
          </v-autocomplete>
        </v-col>
      </v-row>
      <v-row v-if="isSsh">
        <v-col cols="12">
          <v-autocomplete
            label="Machine to scan"
            :model-value="instance.config?.instance"
            :items="machines"
            item-title="title"
            item-value="uuid"
            :loading="machinesLoading"
            :rules="requiredRule"
            hint="One of this account's instances. It is scanned over SSH from the scan runner; results appear after the first scan."
            persistent-hint
            no-data-text="This account has no instances to scan"
            @update:model-value="setValue('config.instance', $event)"
          >
            <template v-slot:item="{ internalItem: item, props: itemProps }">
              <v-list-item v-bind="itemProps" :subtitle="item.raw.subtitle" />
            </template>
          </v-autocomplete>
        </v-col>
      </v-row>
      <v-row v-else>
        <v-col cols="12">
          <v-combobox
            label="Domain"
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
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { ListInstancesRequest } from "nocloud-proto/proto/es/instances/instances_pb";
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";
import { kindTitle } from "./productKinds";

const props = defineProps(["instance", "planRules", "spUuid", "accountId"]);
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
    kindTitle: kindTitle(product),
  }))
);
const product = computed(
  () => bilingPlan.value?.products?.[instance.value.product]
);
// A product of kind "ssh" scans one of the account's machines over SSH
// instead of a set of domains.
const isSsh = computed(() => product.value?.meta?.kind === "ssh");
const store = useStore();
const machines = ref([]);
const machinesLoading = ref(false);

async function loadMachines() {
  machines.value = [];
  if (!isSsh.value || !props.accountId) return;
  machinesLoading.value = true;
  try {
    const response = await store.getters["instances/instancesClient"].list(
      ListInstancesRequest.fromJson({
        page: 1,
        limit: 100,
        filters: { account: [props.accountId] },
      })
    );
    machines.value = response.pool
      .map((i) => i.instance.toJson())
      // a scan of a machine is not itself a machine
      .filter((i) => !i.config?.instance && i.state?.state !== "DELETED")
      .map((i) => ({
        uuid: i.uuid,
        title: i.title,
        subtitle: i.config?.host || i.uuid,
      }));
  } catch {
    machines.value = [];
  } finally {
    machinesLoading.value = false;
  }
}
watch([isSsh, () => props.accountId], loadMachines, { immediate: true });
watch(isSsh, (ssh) => {
  if (ssh) setValue("config.domains", []);
  else setValue("config.instance", undefined);
});

const trial = computed(() => product.value?.meta?.trial);
const trialDays = computed(() => {
  const days = (+trial.value?.period || 0) / 86400;
  return days >= 1 ? `${Math.round(days)}-day` : `${trial.value?.period}s`;
});

// One domain per instance: each further domain is another order. Several
// only make sense on a provider that imports: to adopt an existing
// integration that has exactly those domains. The driver checks this.
const domainsHint =
  "One domain per instance; subdomains are discovered by the scan. Enter several only to import an existing integration with exactly those domains.";
const domainRules = computed(() => [
  (val) => (val?.length ?? 0) > 0 || "A domain is required",
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
