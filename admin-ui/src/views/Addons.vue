<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5">
      <v-btn class="mr-2" :to="{ name: 'Addon create' }"> Create </v-btn>
      <v-btn
        class="mr-2"
        :loading="isDeleteLoading"
        :disabled="!selectedAddons.length"
        @click="deleteSelectedAddons"
      >
        Delete
      </v-btn>
    </div>

    <addons-table
      :refetch="refetch"
      editable
      show-select
      v-model="selectedAddons"
      sort-by="exec"
      sort-desc
    />
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import AddonsTable from "@/components/addonsTable.vue";
import { useStore } from "@/store";

const store = useStore();

const selectedAddons = ref([]);
const isDeleteLoading = ref(false);
const refetch = ref(false);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => (refetch.value = !refetch.value),
  });
});

const deleteSelectedAddons = async () => {
  try {
    isDeleteLoading.value = true;
    await Promise.all(
      selectedAddons.value.map((addon) =>
        store.getters["addons/addonsClient"].delete(addon)
      )
    );
    selectedAddons.value = [];
    refetch.value = !refetch.value;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isDeleteLoading.value = false;
  }
};
</script>

<script>
export default {
  name: "addons-view",
};
</script>

<style>
.change_public .v-input {
  margin-top: 0px !important;
}
</style>
