<template>
  <div v-if="descriptions.length">
    <p v-for="description in descriptions" :key="description">
      {{ description }}
    </p>
  </div>
  <div v-else>
    <span>No description</span>
  </div>
</template>

<script setup>
import { computed, toRefs } from "vue";

const props = defineProps(["invoice"]);
const { invoice } = toRefs(props);

const descriptions = computed(() =>
  invoice.value.items
    .map((i) =>
      i.description
        ? `${i.description} ${i.price * i.amount} ${
            invoice.value.currency?.code
          }`
        : null
    )
    .filter((i) => !!i)
    .slice(0, 5)
);
</script>
