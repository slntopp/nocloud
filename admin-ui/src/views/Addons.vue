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
    <nocloud-table
      v-model="selectedAddons"
      :loading="isLoading"
      :items="addons"
      :headers="headers"
      table-name="addons-table"
    >
      <template v-slot:[`item.title`]="{ item }">
        <router-link :to="{ name: 'Addon page', params: { uuid: item.uuid } }">
          {{ item.title }}
        </router-link>
      </template>

      <template v-slot:[`item.public`]="{ item }">
        <div class="change_public">
          <v-switch
            :loading="updatingAddonUuid === item.uuid"
            dense
            hide-details
            :disabled="!!updatingAddonUuid"
            :input-value="item.public"
            @change="updateAddon(item, { key: 'public', value: $event })"
          />
        </div>
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import NocloudTable from "@/components/table.vue";
import api from "@/api";

const store = useStore();

const updatingAddonUuid = ref("");
const selectedAddons = ref([]);
const isDeleteLoading = ref(false);

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Group", value: "group" },
  { text: "Public", value: "public" },
]);

onMounted(() => {
  fetchAddons();
});

const isLoading = computed(() => store.getters["addons/isLoading"]);
const searchParam = computed(() => store.getters["appSearch/param"]);
const addons = computed(() =>
  store.getters["addons/all"].filter(
    (a) =>
      !searchParam.value ||
      a.title.toLowerCase().includes(searchParam.value.toLowerCase()) ||
      a.group.toLowerCase().includes(searchParam.value.toLowerCase())
  )
);

const fetchAddons = () => {
  store.dispatch("addons/fetch");
};

const updateAddon = async (item, { key, value }) => {
  try {
    updatingAddonUuid.value = item.uuid;
    await api.patch("/addons/" + item.uuid, { ...item, [key]: value });
    item.public = value;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    updatingAddonUuid.value = "";
  }
};

const deleteSelectedAddons = async () => {
  try {
    isDeleteLoading.value = true;
    await Promise.all(
      selectedAddons.value.map((addon) => api.delete("/addons/" + addon.uuid))
    );
    fetchAddons();
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isDeleteLoading.value = false;
  }
};
</script>

<script>
export default { name: "AddonsView" };
</script>

<style>
.change_public .v-input {
  margin-top: 0px !important;
}
</style>
