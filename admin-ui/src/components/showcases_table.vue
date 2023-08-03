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
  </nocloud-table>
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";
import { ref, toRefs } from "vue";

const props = defineProps(["items", "loading", "value"]);
const emits = defineEmits(["input"]);

const { value, loading, items } = toRefs(props);

const headers = ref([
  { text: "uuid", value: "uuid" },
  { text: "title", value: "title" },
  { text: "preview", value: "preview" },
]);
</script>

<style scoped></style>
