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
      <icon-title-preview
        type="nocloud"
        :icon="item.icon"
        :title="item.title"
      />
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'ShowcasePage', params: { uuid: item.uuid } }">
        {{ getShortName(item.title, 45) }}
      </router-link>
    </template>

    <template v-slot:[`item.sorter`]="{ item }">
      <v-skeleton-loader v-if="updatedShowcase === item.uuid" type="text" />
      <v-text-field
        v-else
        dense
        hide-details
        style="max-width: 50px"
        type="number"
        :disabled="!!updatedShowcase"
        :value="item.sorter"
        @change="updateShowcase(item, { key: 'sorter', value: +$event })"
      />
    </template>

    <template v-slot:[`item.public`]="{ item }">
      <div class="d-flex justify-center align-center change_public">
        <v-skeleton-loader v-if="updatedShowcase === item.uuid" type="text" />
        <v-switch
          v-else
          dense
          hide-details
          :readonly="!!updatedShowcase"
          :input-value="item.public"
          @change="updateShowcase(item, { key: 'public', value: $event })"
        />
      </div>
    </template>
  </nocloud-table>
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";
import { ref, toRefs } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import { getShortName } from "@/functions";

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
  { text: "Sorter", value: "sorter" },
]);

const updateShowcase = async (item, { key, value }) => {
  try {
    updatedShowcase.value = item.uuid;
    const data = { ...item, [key]: value };
    await api.showcases.update(data);
    store.commit("showcases/replaceShowcase", data);
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during update showcase",
    });
  } finally {
    updatedShowcase.value = "";
  }
};
</script>

<style>
.change_public .v-input {
  margin-top: 0px !important;
}
</style>
