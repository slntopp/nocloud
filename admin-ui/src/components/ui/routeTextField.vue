<template>
  <v-text-field
    @click:append-inner="goTo(props.to)"
    readonly
    append-inner-icon="mdi-login"
    :label="props.label"
    :model-value="props.modelValue"
  />
</template>

<script setup>
import {useRouter} from "vue-router";


const props = defineProps({
  to: {
    type: Object,
    required: true,
  },
  label: String,
  modelValue: String,
});

const router=useRouter()

const goTo = ({name, params}) => {
  router.push({ name, params });
};
</script>

<style scoped></style>
