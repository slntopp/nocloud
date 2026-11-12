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
  // NCU (or no currency) means the default currency, as elsewhere in the app.
  let code = spending.currency?.code || spending.currency?.title || "";
  if (!code || code === "NCU") {
    code = store.getters["currencies/default"]?.code || code;
  }
  const months = Math.round(((spending.periodDays || 395) * 12) / 365);
  return `Pays ${amount} ${code} per month on average (payments of the last ${months} months)`;
});
</script>

<style scoped>
.account-grade {
  margin: 0 4px;
  vertical-align: middle;
}
</style>
