<template>
  <v-menu v-model="isMenuOpen" :close-on-content-click="false">
    <template v-slot:activator="{ on }">
      <div class="d-flex justify-center align-center" :v-ripple="false">
        <v-text-field
          :label="label"
          :value="value"
          clearable
          @input="emit('input', $event)"
          @click:clear="emit('input', '')"
        />
        <v-icon class="ml-3" style="height: 25px" v-on="on">
          mdi-palette
        </v-icon>
      </div>
    </template>
    <v-color-picker
      v-if="isMenuOpen"
      mode="hexa"
      :value="value || '#FFFFFFFF'"
      @input="emit('input', $event.hex || $event)"
      hide-mode-switch
    ></v-color-picker>
  </v-menu>
</template>

<script setup>
import { ref, toRefs } from "vue";

const props = defineProps(["value", "label"]);
const emit = defineEmits(["input"]);

const { value } = toRefs(props);
const isMenuOpen = ref(false);
</script>

<style scoped></style>
