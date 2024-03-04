<template>
  <v-dialog v-model="isOpen" width="90vw">
    <template v-slot:activator="{ on, attrs }">
      <v-btn icon v-bind="attrs" v-on="on">
        <v-icon> mdi-menu-open </v-icon>
      </v-btn>
    </template>

    <v-card color="background-light">
      <addons-table
        :items="productAddons"
        :loading="isAddonsLoading"
        table-name="product-addons-table"
      />
      <div class="d-flex justify-end mt-3 pa-2">
        <v-btn @click="isOpen = false">Close</v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { onMounted, ref, toRefs } from "vue";
import AddonsTable from "@/components/addonsTable.vue";
import { useStore } from "@/store";

const props = defineProps({
  addons: { type: Array, required: true },
});

const { addons } = toRefs(props);

const store = useStore();

const productAddons = ref([]);
const isAddonsLoading = ref(false);
const isOpen = ref(false);

onMounted(async () => {
  try {
    isAddonsLoading.value = true;
    const data = await Promise.all(
      addons.value.map((uuid) =>
        store.getters["addons/addonsClient"].get({ uuid })
      )
    );
    productAddons.value = data;
  } finally {
    isAddonsLoading.value = false;
  }
});
</script>
