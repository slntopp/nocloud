<template>
  <v-chip class="ma-3 pa-2" outlined color="primary">
    {{
      param.isArray
        ? `${variants[param.key]?.title}:${param?.title}`
        : param.title
    }}
    <v-btn @click="deleteParam" icon small>
      <v-icon small>mdi-close</v-icon>
    </v-btn>
  </v-chip>
</template>

<script setup>
import { useStore } from "@/store";
import { toRefs } from "vue";

const props = defineProps(["param", "variants"]);
const { param, variants } = toRefs(props);

const store = useStore();

const deleteParam = () => {
  store.commit("appSearch/deleteCustomParam", param.value);
};
</script>

<style scoped></style>
