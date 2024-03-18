<template>
  <instance-control-btn hint="reboot">
    <v-dialog v-model="isModalOpen" max-width="500">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
            :disabled="disabled"
            :loading="loading"
            class="ma-1"
            v-bind="attrs"
            v-on="on"
        >
          <v-icon>
            mdi-restart
          </v-icon>
        </v-btn>
      </template>
      <v-card color="background-light">
        <v-card-title>Select reboot type</v-card-title>
        <v-card-actions class="d-flex justify-end">
          <v-btn class="mr-2" @click="reboot('hard')"> Hard </v-btn>
          <v-btn class="mr-2" @click="reboot('soft')"> Soft </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </instance-control-btn>
</template>

<script setup>
import { toRefs, ref } from "vue";
import InstanceControlBtn from "@/components/ui/btnWithHint.vue";

const props = defineProps(["disabled", "loading", "template"]);
const emit = defineEmits(["click"]);

const { disabled, loading } = toRefs(props);

const isModalOpen = ref(false);

const reboot = (type) => {
  emit("click", { type });
  isModalOpen.value=false
};
</script>

<style scoped></style>
