<template>
  <nocloud-table
    :value="value"
    @input="emits('input', $event)"
    :headers="headers"
    item-key="uuid"
    table-name="showcases"
    :items="items"
    :loading="loading"
  >
    <template v-slot:[`item.preview`]="{ item }">
      <icon-title-preview is-mdi :icon="item.icon" :title="item.title" />
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'ShowcasePage', params: { uuid: item.uuid } }">
        {{ item.title }}
      </router-link>
    </template>

    <template v-slot:[`item.public`]="{ item }">
      <v-skeleton-loader v-if="updatedShowcase === item.uuid" type="text" />
      <v-switch
        v-else
        :readonly="!!updatedShowcase"
        :input-value="item.public"
        @change="changeEnabled(item, $event)"
      />
    </template>
  </nocloud-table>
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";
import { ref, toRefs } from "vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["items", "loading", "value"]);
const emits = defineEmits(["input"]);

const { value, loading, items } = toRefs(props);

const store = useStore();

const updatedShowcase = ref("");

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Preview", value: "preview" },
  { text: "Enabled", value: "public" },
]);

const changeEnabled = async (item, value) => {
  try {
    updatedShowcase.value = item.uuid;
    const data = { ...item, public: value };
    await api.showcases.update(data);
    store.commit("showcases/replaceShowcase", data);
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during update showcase enabled",
    });
  } finally {
    updatedShowcase.value = "";
  }
};
</script>

<style scoped></style>
