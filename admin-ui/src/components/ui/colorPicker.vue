<template>
  <v-menu
    v-model="isMenuOpen"
    :close-on-content-click="false"
    @click:outside="isMenuOpen = false"
  >
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
    <v-card v-if="isMenuOpen" class="pa-0">
      <v-color-picker
        mode="hexa"
        :value="value || '#FFFFFFFF'"
        @input="emit('input', $event.hex || $event)"
        hide-mode-switch
      ></v-color-picker>
      <v-card-actions class="justify-end pt-0">
        <v-btn small text @click="isMenuOpen = false">Close</v-btn>
      </v-card-actions>
    </v-card>
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
