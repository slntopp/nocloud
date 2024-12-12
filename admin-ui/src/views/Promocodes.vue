<template>
  <div class="pa-4">
    <div class="d-flex justify-space-between pb-2 mt-4">
      <div>
        <v-btn
          class="mr-2"
          color="background-light"
          :to="{ name: 'Promocode create' }"
        >
          Create
        </v-btn>
      </div>

      <promocode-status-btns :items="selected" @click="changeStatus" />
    </div>

    <promocodes-table v-model="selected" :refetch="refetch" />
  </div>
</template>

<script setup>
import { useStore } from "@/store";
import promocodesTable from "@/components/promocodesTable.vue";
import promocodeStatusBtns from "@/components/promocode/ui/promocodeStatusBtns.vue";
import { onMounted, ref } from "vue";

const store = useStore();

const refetch = ref(false);
const selected = ref([]);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => {
      refetch.value = !refetch.value;
    },
  });
});

const changeStatus = () => {
  refetch.value = !refetch.value;
  selected.value = [];
};
</script>

<script>
export default {
  name: "promocodes-view",
};
</script>
