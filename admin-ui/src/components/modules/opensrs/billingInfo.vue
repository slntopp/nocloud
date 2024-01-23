<template>
  <v-row>
    <v-col>
      <v-text-field
        readonly
        label="price model"
        :value="template.billingPlan.title"
      />
    </v-col>

    <v-col>
      <v-text-field
        readonly
        label="Product name"
        :value="template.resources.domain"
      />
    </v-col>

    <v-col>
      <v-text-field readonly label="Price" :value="price" :suffix="account.currency"/>
    </v-col>

    <v-col>
      <v-text-field
        readonly
        label="Date (create)"
        :value="formatSecondsToDate(template.data.creation)"
      />
    </v-col>

    <v-col>
      <v-text-field
        readonly
        label="Due to date/next payment"
        :value="template.data.expiry.expiredate"
      />
    </v-col>
  </v-row>
</template>

<script setup>
import { computed, defineProps, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import { useStore } from "@/store";

const props = defineProps(["template", "plans", "service", "sp"]);
const { template } = toRefs(props);

const { convertTo } = useCurrency();
const store = useStore();

const account = computed(() => {
  const namespace = store.getters["namespaces/all"]?.find(
    (n) => n.uuid === template.value?.access.namespace
  );
  const account = store.getters["accounts/all"].find(
    (a) => a.uuid === namespace?.access.namespace
  );
  return account;
});

const price = computed(() => {
  return convertTo(
    template.value.billingPlan.resources[0]?.price || 0,
    account.value.currency
  );
});
</script>

<style lang="scss">
.ione-billing {
  .v-expansion-panel-content__wrap {
    padding: 0px !important;
  }
}
</style>
