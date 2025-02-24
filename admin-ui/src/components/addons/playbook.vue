<template>
  <div class="pa-5">
    <playbook-item
      :playbooks="playbooks"
      :vars="addonAction.vars"
      :playbook="addonAction.playbook"
      :loading="isSaveLoading"
      @save="saveAddonAction"
    />
  </div>
</template>

<script setup>
import { useStore } from "@/store";
import playbookItem from "@/components/ui/playbookItem.vue";
import { computed, onMounted, ref, toRefs } from "vue";
import { Action } from "nocloud-proto/proto/es/billing/addons/addons_pb";

const props = defineProps({
  addon: {},
});
const { addon } = toRefs(props);

const addonAction = ref({ playbook: "", vars: {} });
const isSaveLoading = ref(false);

const store = useStore();

onMounted(() => {
  store.dispatch("playbooks/fetch");

  if (addon.value.action) {
    addonAction.value = addon.value.action;
  }
});

const playbooks = computed(() => store.getters["playbooks/all"]);

const saveAddonAction = async (data) => {
  isSaveLoading.value = true;

  try {
    const dto = { ...addon.value };
    dto.action = Action.fromJson(data);
    await store.getters["addons/addonsClient"].update(dto);

    store.commit("snackbar/showSnackbarSuccess", { message: "Done" });
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<script>
export default { name: "addon-products" };
</script>

<style scoped></style>
