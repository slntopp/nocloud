<template>
  <v-row>
    <v-col cols="3">
      <route-text-field
        v-if="!isLoading"
        :to="{ name: 'Account', params: { accountId: user?.uuid } }"
        label="User"
        :value="user?.title"
      />
      <v-skeleton-loader type="text" v-else />
    </v-col>
    <v-col> </v-col>
  </v-row>
</template>

<script setup>
import { toRefs, defineProps, ref, onMounted } from "vue";
import RouteTextField from "@/components/ui/routeTextField.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const user = ref([]);
const isLoading = ref(false);

onMounted(async () => {
  try {
    isLoading.value = true;
    user.value = await api.accounts.get(template.value.config.user);
  } catch (err) {
    store.commit("snackbar/showSnackbarError", { message: err.message });
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped></style>
