<template>
  <div v-if="descriptions.length">
    <div
      v-for="description in descriptions"
      :key="description"
      class="d-flex justify-space-between"
    >
      <span> {{ description.title }} </span>
      <span class="mx-2" style="font-weight: bold">{{
        [description.price, currency].join(" ")
      }}</span>
    </div>

    <div class="d-flex justify-end mt-2">
      <span> Total </span>
      <span class="mr-2 ml-1" style="font-weight: bold">{{
        [descriptions.reduce((acc, d) => d.price + acc, 0), currency].join(" ")
      }}</span>
    </div>
  </div>
  <div v-else>
    <span>No description</span>
  </div>
</template>

<script setup>
import { computed, toRefs } from "vue";

const props = defineProps(["invoice"]);
const { invoice } = toRefs(props);

const currency = computed(() => invoice.value.currency?.code);

const descriptions = computed(() =>
  invoice.value.items
    .map((i) =>
      i.description
        ? {
            title: i.description,
            price: i.price * i.amount,
          }
        : null
    )
    .filter((i) => !!i)
    .slice(0, 5)
);
</script>
