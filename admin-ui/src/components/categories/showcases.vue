<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <div class="d-flex justify-end align-center pb-4">
      <v-btn
        @click="changeShowcasesCategory"
        :disabled="!selected.length"
        color="primary"
        class="mr-2"
      >
        Enabled for selected
      </v-btn>
    </div>

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
import api from "../../api";

const props = defineProps(["category"]);
const { category } = toRefs(props);

const store = useStore();

const selected = ref([]);

const showcases = computed(() => store.getters["showcases/all"]);
const isLoading = computed(() => store.getters["showcases/isLoading"]);

onMounted(async () => {
  try {
    await store.dispatch("showcases/fetch", { anonymously: false });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch info",
    });
  }
});

const changeShowcasesCategory = async () => {
  try {
    await Promise.all(
      selected.value.map(async (sh) => {
        const data = {
          ...sh,
          meta: {
            ...sh.meta,
            category: JSON.parse(JSON.stringify(category.value)),
          },
        };

        await api.showcases.update(data);
        store.commit("showcases/replaceShowcase", data);
      })
    );

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Showcases updated",
    });
    selected.value = [];
  } catch (e) {
    console.log(e);

    store.commit("snackbar/showSnackbarError", {
      message: "Error during update showcases",
    });
  }
};
</script>

<script>
export default {
  name: "CategoriesShowcases",
};
</script>
