<template>
  <div>
    <v-btn
      v-for="status in statuses"
      class="mr-2"
      color="background-light"
      :key="status"
      :loading="loadingStatus === status"
      @click="changeStatus(status)"
      :disabled="
        items.length < 1 || (!!loadingStatus && loadingStatus !== status)
      "
    >
      {{ status }}
    </v-btn>
  </div>
</template>

<script setup>
import {
  Promocode,
  PromocodeStatus,
} from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";
import { computed, ref, toRefs } from "vue";
import { useStore } from "@/store";

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
