<template>
  <nocloud-table
    :items="planAddons"
    :loading="isAddonsLoading"
    table-name="plan-addons-table"
    :headers="headers"
    v-model="selectedToRemove"
  >
    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'Addon page', params: { uuid: item.uuid } }">
        {{ item.title }}
      </router-link>
    </template>

    <template v-slot:top>
      <v-toolbar flat color="background">
        <v-toolbar-title>Actions</v-toolbar-title>
        <v-divider inset vertical class="mx-4" />
        <v-spacer />

        <v-dialog v-model="isAddDialog" width="80vw">
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              v-bind="attrs"
              v-on="on"
              class="mr-2"
              color="background-light"
            >
              Select
            </v-btn>
          </template>
          <v-card color="background-light">
            <addonsTable show-select v-model="selectedToAdd" />
            <div class="d-flex justify-end pa-3">
              <v-btn @click="addAddons" :disabled="selectedToAdd.length < 1"
                >Add</v-btn
              >
            </div>
          </v-card>
        </v-dialog>

        <confirm-dialog
          @confirm="removeAddons"
          :disabled="selectedToRemove.length < 1"
        >
          <v-btn
            color="background-light"
            :disabled="selectedToRemove.length < 1"
            >Disabled</v-btn
          >
        </confirm-dialog>
      </v-toolbar>
    </template>
  </nocloud-table>
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import addonsTable from "@/components/addonsTable.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { onMounted, ref, toRefs } from "vue";
import { useStore } from "@/store";

const props = defineProps(["addons"]);
const { addons } = toRefs(props);

const emit = defineEmits(["change:addons"]);

const store = useStore();

const planAddons = ref([]);
const isAddonsLoading = ref(false);
const selectedToRemove = ref([]);
const selectedToAdd = ref([]);
const isAddDialog = ref(false);

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
]);

onMounted(async () => {
  try {
    isAddonsLoading.value = true;
    const data = await Promise.all(
      addons.value.map((uuid) =>
        store.getters["addons/addonsClient"].get({ uuid })
      )
    );
    planAddons.value = data;
  } finally {
    isAddonsLoading.value = false;
  }
});

const removeAddons = () => {
  planAddons.value = planAddons.value.filter(
    (addon) =>
      !selectedToRemove.value.find(
        (deletedAddon) => deletedAddon.uuid === addon.uuid
      )
  );
  emit(
    "change:addons",
    planAddons.value.map((a) => a.uuid)
  );
};

const addAddons = () => {
  planAddons.value = [...planAddons.value, ...selectedToAdd.value];
  emit(
    "change:addons",
    planAddons.value.map((a) => a.uuid)
  );
  isAddDialog.value = false;
  selectedToAdd.value = [];
};
</script>

<style lang="scss">
.mw-20 {
  max-width: 150px;
}
</style>
