<template>
  <accounts-table :namespace="namespace?.uuid" />
</template>

<script setup>
import AccountsTable from "@/components/accounts_table.vue";
import { toRefs } from "vue";

const props = defineProps(["namespace"]);
const { namespace } = toRefs(props);
</script>

<script>
export default {
  name: "namespace-accounts",
};
</script>

<style scoped></style>
