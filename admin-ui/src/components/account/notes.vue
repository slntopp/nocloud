<template>
  <admin-notes
    :template="account"
    :on-add="onAdd"
    :on-delete="onDelete"
    :on-update="onUpdate"
  />
</template>

<script setup>
import adminNotes from "@/components/adminNotes.vue";
import { toRefs } from "vue";
import api from "@/api";

const props = defineProps(["account"]);
const { account } = toRefs(props);

const onAdd = (...data) => api.accounts.addNote(...data);
const onDelete = (...data) => api.accounts.removeNote(...data);
const onUpdate = (...data) => api.accounts.updateNote(...data);
</script>

<script>
export default {
  name: "account-notes",
};
</script>

<style scoped></style>