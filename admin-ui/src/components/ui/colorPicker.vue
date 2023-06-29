<template>
  <v-menu>
    <template v-slot:activator="{ on }">
      <div class="d-flex justify-center align-center" :v-ripple="false">
        <v-text-field
          :label="label"
          :value="value"
          @input="emit('input', $event)"
        />
        <v-icon class="ml-3" style="height: 25px" v-on="on">
          mdi-palette
        </v-icon>
      </div>
    </template>
    <v-color-picker
      mode="hexa"
      :value="value"
      @input="emit('input', $event.hex || $event)"
      hide-mode-switch
    ></v-color-picker>
  </v-menu>
</template>

<script setup>
import { toRefs } from "vue";

const props = defineProps(["value", "label"]);
const emit = defineEmits(["input"]);

const { value } = toRefs(props);
</script>

<style scoped></style>
