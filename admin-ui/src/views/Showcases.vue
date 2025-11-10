<template>
  <div class="namespaces pa-4 flex-wrap">
    <div class="d-flex justify-space-between align-center pb-8 pt-4">
      <div class="buttons__inline">
        <v-btn
          color="background-light"
          class="mr-2"
          :to="{ name: 'CreateShowcase' }"
        >
          create
        </v-btn>
        <confirm-dialog
          :disabled="selected.length < 1"
          @confirm="deleteSelected"
        >
          <v-btn
            color="background-light"
            class="mr-8"
            :disabled="selected.length < 1"
            :loading="isDeleteLoading"
          >
            delete
          </v-btn>
        </confirm-dialog>
      </div>

      <div>
        <v-btn
          color="background-light"
          class="mr-2"
          :to="{ name: 'Categories' }"
        >
          Categories
        </v-btn>
      </div>
    </div>
    <showcases_table
      :loading="isLoading"
      :items="showcases"
      v-model="selected"
    />
  </div>
</template>

<script setup>
import ConfirmDialog from "@/components/confirmDialog.vue";
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import Showcases_table from "@/components/showcases_table.vue";
import { filterArrayByTitleAndUuid } from "@/functions";

const store = useStore();

const isDeleteLoading = ref(false);
const selected = ref([]);

const searchParam = computed(() => store.getters["appSearch/param"]);
const showcases = computed(() =>
  filterArrayByTitleAndUuid(store.getters["showcases/all"], searchParam.value)
);
const isLoading = computed(() => store.getters["showcases/isLoading"]);

onMounted(() => {
  fetchShowcases();
  store.commit("reloadBtn/setCallback", { event: fetchShowcases });
});

const fetchShowcases = async () => {
  try {
    await store.dispatch("showcases/fetch", { anonymously: false });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch info",
    });
  }
};

const deleteSelected = async () => {
  try {
    isDeleteLoading.value = true;
    const deletePromises = selected.value.map(({ uuid }) =>
      store.dispatch("showcases/delete", uuid)
    );
    await Promise.all(deletePromises);
    store.commit("snackbar/showSnackbarSuccess", {
      message: `Showcase${
        deletePromises.length === 1 ? "" : "s"
      } deleted successfully.`,
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response.data.message || "Error during delete showcases",
    });
  } finally {
    isDeleteLoading.value = false;
  }
};
</script>

<script>
export default {
  name: "showcases-view",
};
</script>
