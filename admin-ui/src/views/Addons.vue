<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5">
      <v-btn class="mr-2" :to="{ name: 'Addon create' }"> Create </v-btn>
    </div>
    <nocloud-table
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
        <div class="d-flex justify-center align-center change_public">
          <v-skeleton-loader
            v-if="updatingAddonUuid === item.uuid"
            type="text"
          />
          <v-switch
            v-else
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

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Group", value: "group" },
  { text: "Public", value: "public" },
]);

onMounted(() => {
  store.dispatch("addons/fetch");
});

const isLoading = computed(() => store.getters["addons/isLoading"]);
const addons = computed(() => store.getters["addons/all"]);

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
</script>

<script>
export default { name: "AddonsView" };
</script>
<style scoped></style>
