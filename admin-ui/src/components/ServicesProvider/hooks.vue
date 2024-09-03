<template>
  <v-card color="background-light" class="pa-5">
    <json-editor :json="newHooks" @changeValue="newHooks = $event" />

    <div class="d-flex jsutify-end align-center">
      <v-btn :loading="isSaveLoading" @click="updateSp">Update</v-btn>
    </div>
  </v-card>
</template>

<script setup>
import { ref, toRefs } from "vue";
import JsonEditor from "@/components/JsonEditor.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const newHooks = ref({ ...template.value.hooks });
const isSaveLoading = ref(false);

const updateSp = async () => {
  console.log(newHooks.value);
  isSaveLoading.value = true;
  try {
    await api.servicesProviders.update(template.value.uuid, {
      ...template.value,
      hooks: { ...newHooks.value },
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save offering items",
    });
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<script>
export default {
  name: "sp-hooks",
};
</script>
