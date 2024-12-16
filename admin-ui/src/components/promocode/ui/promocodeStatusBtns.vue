<template>
  <div>
    <confirm-dialog
      v-for="status in statuses"
      :key="status"
      :disabled="
        items.length < 1 || (!!loadingStatus && loadingStatus !== status)
      "
      @confirm="changeStatus(status)"
      :loading="loadingStatus === status"
    >
      <v-btn
        :disabled="
          items.length < 1 || (!!loadingStatus && loadingStatus !== status)
        "
        class="mr-2"
        color="background-light"
        :loading="loadingStatus === status"
      >
        {{ status }}
      </v-btn>
    </confirm-dialog>
  </div>
</template>

<script setup>
import {
  Promocode,
  PromocodeStatus,
} from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";
import { computed, ref, toRefs } from "vue";
import { useStore } from "@/store";
import confirmDialog from "@/components/confirmDialog.vue";

const props = defineProps(["items"]);
const { items } = toRefs(props);

const emit = defineEmits(["click"]);

const store = useStore();

const loadingStatus = ref(false);

const statuses = computed(() =>
  Object.keys(PromocodeStatus).filter(
    (value) => !Number.isInteger(+value) && value !== "STATUS_UNKNOWN"
  )
);

const changeStatus = async (status) => {
  loadingStatus.value = status;

  try {
    await Promise.all(
      items.value.map((item) =>
        store.getters["promocodes/promocodesClient"].update(
          Promocode.fromJson({
            ...item,
            status,
          })
        )
      )
    );

    emit("click");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    loadingStatus.value = "";
  }
};
</script>
