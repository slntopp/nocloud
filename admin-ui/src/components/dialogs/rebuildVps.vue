<template>
  <instance-control-btn hint="Rebuild">
    <v-dialog v-model="isModalOpen" max-width="500">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          :disabled="disabled"
          :loading="loading"
          class="ma-1"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon>mdi-account-convert</v-icon>
        </v-btn>
      </template>
      <v-card class="pa-5" color="background-light">
        <v-row class="mt-3">
          <v-card-title>Choose os</v-card-title>
          <v-autocomplete
            label="OS"
            v-model="selectedOs"
            :items="images"
            :loading="isImagesLoading"
            item-text="name"
            item-value="id"
            return-object
          />
        </v-row>
        <v-card-actions class="d-flex justify-end mt-5">
          <v-btn class="mr-2" @click="isModalOpen = false"> cancel </v-btn>
          <v-btn class="mr-2" @click="rebuild" :disabled="!selectedOs">
            rebuild
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </instance-control-btn>
</template>

<script setup>
import { toRefs, ref, onMounted } from "vue";
import { useStore } from "@/store";
import InstanceControlBtn from "@/components/ui/instanceControlBtn.vue";
import api from "@/api";

const props = defineProps(["disabled", "loading", "template"]);
const emit = defineEmits(["click"]);

const store = useStore();

const { disabled, loading, template } = toRefs(props);

const isModalOpen = ref(false);
const images = ref([]);
const isImagesLoading = ref(false);
const selectedOs = ref();

onMounted(async () => {
  isImagesLoading.value = true;
  try {
    const { meta } = await api.instances.action({
      uuid: template.value.uuid,
      action: "images",
    });
    images.value = meta.images;
    const os = images.value.find(
      (os) => os.name === template.value.config.configuration.vps_os
    );
    if (os) {
      selectedOs.value = os;
    }
  } catch (err) {
    store.commit("snackbar/showSnackbarError", { message: err });
  } finally {
    isImagesLoading.value = false;
  }
});

const rebuild = async () => {
  isModalOpen.value = false;
  emit("click", selectedOs.value);
};
</script>

<style scoped></style>
