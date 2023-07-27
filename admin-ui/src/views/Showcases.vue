<template>
  <div class="namespaces pa-4 flex-wrap">
    <div class="buttons__inline pb-8 pt-4">
      <v-btn
        color="background-light"
        class="mr-2"
        :to="{ name: 'CreateShowcase' }"
      >
        create
      </v-btn>
      <confirm-dialog :disabled="selected.length < 1" @confirm="deleteSelected">
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
    <nocloud-table
      v-model="selected"
      :headers="headers"
      item-key="uuid"
      table-name="showcases"
      :items="showcases"
      :loading="isLoading"
    >
      <template v-slot:[`item.preview`]="{ item }">
        <icon-title-preview is-mdi :icon="item.icon" :title="item.title" />
      </template>

      <template v-slot:[`item.title`]="{ item }">
        <router-link :to="{ name: 'ShowcasePage', params: { uuid: item.uuid } }">
          {{ item.title }}
        </router-link>
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import ConfirmDialog from "@/components/confirmDialog.vue";
import NocloudTable from "@/components/table.vue";
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";

const store = useStore();

const headers = ref([
  { text: "uuid", value: "uuid" },
  { text: "title", value: "title" },
  { text: "preview", value: "preview" },
]);
const isDeleteLoading = ref(false);
const selected = ref([]);

const showcases = computed(() => store.getters["showcases/all"]);
const isLoading = computed(() => store.getters["showcases/isLoading"]);

onMounted(async () => {
  try{
    await store.dispatch("showcases/fetch");
  }catch (e){
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch info",
    });
  }
});

const deleteSelected = async () => {
  try {
    isDeleteLoading.value = true;
    const deletePromises = selected.value.map(({ uuid }) =>
      store.dispatch("showcases/delete", uuid)
    );
    await Promise.all(deletePromises);
    store.commit("snackbar/showSnackbarSuccess", {
      message: `Showcase${
        deletePromises.length == 1 ? "" : "s"
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