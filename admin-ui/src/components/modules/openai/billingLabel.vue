<template>
  <v-container>
    <v-row class="mt-0" align="start" justify="end">
      <v-col class="d-flex justify-end px-1 py-1">
        <instance-state :template="template" />
      </v-col>
      <v-col class="d-flex justify-end px-1 py-1">
        <v-chip
          class="mx-2"
          color="primary"
          outlined
          v-for="price in pricesPerToken"
          :key="price.name"
          >{{ price.name }} : {{ price.price }} {{ accountCurrency }}</v-chip
        >
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { computed, toRefs } from "vue";
import { useStore } from "@/store";
import InstanceState from "@/components/ui/instanceState.vue";
import useCurrency from "@/hooks/useCurrency";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();
const { convertFrom } = useCurrency();

const account = computed(() => {
  const namespace = store.getters["namespaces/all"]?.find(
      (n) => n.uuid === template.value?.access.namespace
  );
  const account = store.getters["accounts/all"].find(
      (a) => a.uuid === namespace?.access.namespace
  );
  return account;
});

const accountCurrency = computed(() => {
  return account.value.currency;
});

const pricesPerToken = computed(() => {
  const items = [];

  const acceptedResources = ["input_kilotoken", "output_kilotoken"];

  template.value.billingPlan.resources.forEach((r) => {
    if (acceptedResources.includes(r.key)) {
      items.push({
        name: r.key.replace("_kilotoken", ""),
        price: r.price,
        accountPrice: convertFrom(r.price),
      });
    }
  });

  return items;
});
</script>

<style scoped></style>
