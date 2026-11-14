<template>
  <v-menu
    v-model="isMenuOpen"
    :close-on-content-click="false"
    @click:outside="isMenuOpen = false"
  >
    <template v-slot:activator="{ props }">
      <div class="d-flex justify-center align-center" :v-ripple="false">
        <v-text-field
          :label="label"
          :model-value="value"
          clearable
          @update:model-value="emit('update:modelValue', $event)"
          @click:clear="emit('update:modelValue', '')"
        />
        <v-icon class="ml-3" style="height: 25px" v-bind="props">
          mdi-palette
        </v-icon>
      </div>
    </template>
    <v-card v-if="isMenuOpen" class="pa-0">
      <v-color-picker
        mode="hexa"
        :model-value="value || '#FFFFFFFF'"
        @update:model-value="emit('update:modelValue', $event.hex || $event)"
        :modes="['hexa']"
      ></v-color-picker>
      <v-card-actions class="justify-end pt-0">
        <v-btn size="small" variant="text" @click="isMenuOpen = false">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup>
import { ref, toRefs } from "vue";

const props = defineProps(["modelValue", "label"]);
const emit = defineEmits(["update:modelValue"]);

const { modelValue: value } = toRefs(props);
const isMenuOpen = ref(false);
</script>

<style scoped></style>
