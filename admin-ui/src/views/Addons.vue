<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5">
      <v-btn class="mr-2" :to="{ name: 'Addon create' }"> Create </v-btn>
    </div>
    <nocloud-table :loading="isLoading" :items="groups" :headers="headers">
      <template v-slot:[`item.title`]="{ item }">
        <router-link
          :to="{ name: 'Addon page', params: { title: item.title } }"
        >
          {{ item.title }}
        </router-link>
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import NocloudTable from "@/components/table.vue";

const store = useStore();

const headers = ref([{ text: "Title", value: "title" }]);

onMounted(() => {
  store.dispatch("addons/fetch");
});

const isLoading = computed(() => store.getters["addons/isLoading"]);
const groups = computed(() =>
  [...new Set(store.getters["addons/all"].map((a) => a.group)).values()].map(
    (title) => ({ title })
  )
);
</script>

<script>
export default { name: "AddonsView" };
</script>
<style scoped></style>
