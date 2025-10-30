<template>
  <div class="account-groups pa-4 flex-wrap mt-4">
    <div class="d-flex justify-space-between align-center">
      <v-btn
        color="background-light"
        class="mr-2 mt-2"
        @click="goToAccountGroupCreate"
      >
        Create
      </v-btn>

      <confirm-dialog
        :disabled="selected.length < 1 || isDeleteLoading"
        @confirm="deleteAccountGroups"
      >
        <v-btn
          color="background-light"
          class="mr-2 mt-2"
          :disabled="selected.length < 1 || isDeleteLoading"
          :loading="isDeleteLoading"
        >
          Delete
        </v-btn>
      </confirm-dialog>
    </div>
    <div class="mt-3">
      <nocloud-table
        key="uuid"
        table-name="account-groups"
        :headers="accountsGroupsHeaders"
        :items="accountGroups"
        :loading="isLoading"
        :footer-error="fetchError"
        v-model="selected"
        sort-by="title"
        sort-desc
        show-select
      >
        <template v-slot:[`item.title`]="{ item }">
          <router-link
            :to="{ name: 'AccountGroupPage', params: { uuid: item.uuid } }"
          >
            {{ getShortName(item.title) }}
          </router-link>
        </template>

         <template v-slot:[`item.color`]="{ item }">
            <div :style="{ backgroundColor: item.color }" class="color-box"></div>
        </template>
      </nocloud-table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import { getShortName } from "../functions";
import { useRouter } from "vue-router/composables";
import ConfirmDialog from "../components/confirmDialog.vue";

const store = useStore();

const router = useRouter();

const selected = ref([]);
const isDeleteLoading = ref(false);
const fetchError = ref("");

const accountsGroupsHeaders = [
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Color", value: "color" },
];

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => {
      fetchAccountGroups();
    },
  });

  fetchAccountGroups();
});

const accountGroups = computed(() => {
  return store.getters["accountGroups/all"];
});

const isLoading = computed(() => {
  return store.getters["accountGroups/isLoading"];
});

const fetchAccountGroups = async () => {
  try {
    await store.dispatch("accountGroups/fetch");
  } catch (e) {
    fetchError.value = e.message;
  }
};

const deleteAccountGroups = async () => {
  isDeleteLoading.value = true;
  try {
    await Promise.all(
      selected.value.map((accountGroup) => store.dispatch("accountGroups/delete", accountGroup.uuid))
    );

    selected.value = [];
  } finally {
    isDeleteLoading.value = false;
  }
};

const goToAccountGroupCreate = () => {
  router.push({ name: "AccountGroupCreate" });
};
</script>

<script>
export default {
  name: "account-groups-page",
};
</script>

<style scoped lang="scss">
.color-box {
  width: 24px;
  height: 24px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
</style>
