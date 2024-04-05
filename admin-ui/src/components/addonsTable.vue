<template>
  <nocloud-table
    v-model="value"
    @input="emit('input', value)"
    @update:options="emit('update:options', $event)"
    :loading="loading"
    :items="items"
    :headers="headers"
    :show-select="showSelect"
    :table-name="tableName"
    :server-items-length="serverItemsLength"
    :server-side-page="serveSidePage"
    :sort-desc="sortDesc"
    :sort-by="sortBy"
    :footer-error="fetchError"
  >
    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'Addon page', params: { uuid: item.uuid } }">
        {{ item.title }}
      </router-link>
    </template>

    <template v-slot:[`item.public`]="{ item }">
      <div class="change_public">
        <v-switch
          :readonly="!editable"
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
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import { ref, toRefs } from "vue";
import { useStore } from "@/store";

const props = defineProps({
  loading: { type: Boolean, default: false },
  items: {},
  value: {},
  tableName: { type: String, default: "addons-table" },
  showSelect: { type: Boolean, default: false },
  sortBy: {},
  sortDesc: {},
  serverItemsLength: {},
  serveSidePage: {},
  fetchError: {},
  editable: { type: Boolean, default: false },
});
const {
  loading,
  items,
  showSelect,
  tableName,
  value,
  sortBy,
  sortDesc,
  serverItemsLength,
  serveSidePage,
} = toRefs(props);

const emit = defineEmits(["input", "update:options"]);

const store = useStore();

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Group", value: "group" },
  { text: "Public", value: "public" },
]);
const updatingAddonUuid = ref(false);

const updateAddon = async (item, { key, value }) => {
  try {
    updatingAddonUuid.value = item.uuid;
    console.log(item);
    await store.getters["addons/addonsClient"].update({
      ...item,
      [key]: value,
    });
    item.public = value;
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    updatingAddonUuid.value = "";
  }
};
</script>

<style scoped></style>
