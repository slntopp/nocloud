<template>
  <div class="pa-5">
    <div
      class="d-flex justify-center align-center"
      style="flex-direction: column"
    >
      <v-icon size="30">mdi-information-outline</v-icon>
      <pre
        style="
          max-width: 800px;
          font-size: 1rem;
          color: var(--v-primary-base);
          text-align: center;
        "
      >
        {{ playbookTip }}
      </pre>
    </div>

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
import { Action, Addon } from "nocloud-proto/proto/es/billing/addons/addons_pb";

const props = defineProps({
  addon: {},
});
const { addon } = toRefs(props);

const addonAction = ref({ playbook: "", vars: {} });
const isSaveLoading = ref(false);

const store = useStore();

const playbookTip = ` Access to internal fields of the instance or service provider is carried
      out through the template syntax: {{ .Instance }} and {{ .SP }} Example of
      obtaining an instance identifier: {{ .Instance.Uuid }}`;

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

    await store.getters["addons/addonsClient"].update(
      Addon.fromJson(dto, { ignoreUnknownFields: true })
    );

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
