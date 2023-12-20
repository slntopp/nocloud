<template>
  <admin-notes
    :template="template"
    :on-add="onAdd"
    :on-delete="onDelete"
    :on-update="onUpdate"
  />
</template>

<script setup>
import adminNotes from "@/components/adminNotes.vue";
import { toRefs } from "vue";
import api from "@/api";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const onAdd = (...data) => api.instances.addNote(...data);
const onDelete = (...data) => api.instances.removeNote(...data);
const onUpdate = (...data) => api.instances.updateNote(...data);
</script>

<script>
export default {
  name: "instances-notes",
};
</script>

<style scoped></style>
