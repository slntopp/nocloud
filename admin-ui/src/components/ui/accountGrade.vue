<template>
  <v-tooltip v-if="grade" bottom>
    <template v-slot:activator="{ on, attrs }">
      <v-chip
        v-bind="attrs"
        v-on="on"
        :color="grade.color || 'grey'"
        :outlined="grade.manual"
        :text-color="grade.manual ? grade.color : 'white'"
        class="account-grade font-weight-bold"
        x-small
        label
      >
        {{ grade.title }}
      </v-chip>
    </template>
    <div>{{ spendingText }}</div>
    <div v-if="grade.manual">Grade is set manually</div>
  </v-tooltip>
</template>

<script setup>
import { computed, toRefs } from "vue";
import { useStore } from "@/store";
import { getAccountGrade } from "@/accountGrades.js";

const props = defineProps({ account: { type: Object, default: null } });
const { account } = toRefs(props);

const store = useStore();

const grade = computed(() =>
  getAccountGrade(account.value, store.getters["settings/accountGrades"]),
);

const spendingText = computed(() => {
  const spending = grade.value?.spending;
  if (!spending) return "Not counted yet";
  const amount = (+spending.amount || 0).toFixed(2);
  const code = spending.currency?.code || spending.currency?.title || "";
  return `Paid ${amount} ${code} in the last ${spending.periodDays || 90} days`;
});
</script>

<style scoped>
.account-grade {
  margin: 0 4px;
  vertical-align: middle;
}
</style>
