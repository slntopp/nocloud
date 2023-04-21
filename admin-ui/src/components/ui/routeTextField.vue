<template>
  <v-text-field
    @click:append="goTo(props.to)"
    readonly
    append-icon="mdi-login"
    :label="props.label"
    :value="props.value"
  />
</template>

<script setup>
import { defineProps } from "vue";
import {useRouter} from "vue-router/composables";


const props = defineProps({
  to: {
    type: Object,
    required: true,
  },
  label: String,
  value: String,
});

const router=useRouter()

const goTo = ({name, params}) => {
  router.push({ name, params });
};
</script>

<style scoped></style>
