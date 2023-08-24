<template>
  <nocloud-table :headers="reportsHeaders" :items="reports"> </nocloud-table>
</template>

<script setup>
import { onMounted, toRefs, ref, computed, watch } from "vue";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import useRate from "@/hooks/useRate";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const reports = ref([]);

const userCurrency = ref("");
const { rate, convertTo } = useRate(userCurrency);

const reportsHeaders = [
  { text: "Duration", value: "duration" },
  { text: "Executed date", value: "date" },
  { text: "Total", value: "totalPreview" },
  { text: "Total in default currency", value: "totalDefaultPreview" },
  { text: "Product or resource", value: "item" },
];

onMounted(async () => {
  const { records: result } = await api.reports.list(template.value.uuid);
  if (result.length) {
    userCurrency.value = result[0].currency;
  }

  reports.value = result.map((r) => {
    return {
      totalPreview: `${r.total} ${r.currency}`,
      total: r.total,
      duration: `${new Date(r.start * 1000).toLocaleString()} - ${new Date(
        r.end * 1000
      ).toLocaleString()}`,
      date: new Date(r.exec * 1000).toLocaleString(),
      item: r.product || r.resource,
    };
  });
});

const defaultCurrency = computed(() => store.getters["currencies/default"]);

watch(rate, () => {
  console.log(rate.value);
  reports.value = reports.value.map((r) => ({
    ...r,
    totalDefaultPreview: `${convertTo(r.total)} ${defaultCurrency.value}`,
  }));
});
</script>

<script>
export default {
  name: "instance-report",
};
</script>

<style scoped lang="scss"></style>
