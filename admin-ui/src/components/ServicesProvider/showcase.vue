<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <showcases_table
      :items="showcases"
      :loading="isLoading"
      v-model="selected"
    />
  </v-card>
</template>

<script setup>
import { useStore } from "@/store";
import { computed, onMounted, ref, toRefs } from "vue";
import Showcases_table from "@/components/showcases_table.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const selected = ref([]);

const showcases = computed(() =>
  store.getters["showcases/all"].filter((sh) =>
    sh.items.find((i) => i.servicesProvider === template.value.uuid)
  )
);
const isLoading = computed(() => store.getters["showcases/isLoading"]);

onMounted(async () => {
  try {
    await store.dispatch("showcases/fetch", {anonymously:false});
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch info",
    });
  }
});
</script>

<script>
export default {
  name: "showcase-tab",
};
</script>
