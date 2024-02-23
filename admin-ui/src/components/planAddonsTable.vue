<template>
  <addons-table
    :items="planAddons"
    :loading="isAddonsLoading"
    table-name="plan-addons-table"
  />
</template>

<script setup>
import AddonsTable from "@/components/addonsTable.vue";
import { onMounted, ref, toRefs } from "vue";
import api from "@/api";

const props = defineProps(["addons"]);
const { addons } = toRefs(props);

const planAddons = ref([]);
const isAddonsLoading = ref(false);

onMounted(async () => {
  try {
    isAddonsLoading.value = true;
    const data = await Promise.all(
      addons.value.map((uuid) => api.get("/billing/addons/" + uuid))
    );
    planAddons.value = data;
  } finally {
    isAddonsLoading.value = false;
  }
});
</script>

<style lang="scss">
.mw-20 {
  max-width: 150px;
}
</style>
