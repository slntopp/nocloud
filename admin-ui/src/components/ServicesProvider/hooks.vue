<template>
  <v-card color="background-light" class="pa-5">
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

    <div class="d-flex justify-end my-5">
      <v-btn @click="addHook">Add hook</v-btn>
    </div>

    <div class="d-flex justify-center" v-if="newHooks.length === 0">
      <v-card-title>No hooks</v-card-title>
    </div>
    <v-expansion-panels v-else>
      <v-expansion-panel v-for="(hook, index) in newHooks" :key="index">
        <v-expansion-panel-header color="background-light">
          <div>
            Key {{ hook.key }}

            <confirm-dialog @confirm="updateSp(hook, undefined)">
              <v-btn icon>
                <v-icon> mdi-delete </v-icon>
              </v-btn>
            </confirm-dialog>
          </div>
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background-light">
          <v-text-field label="key" v-model="hook.key" />

          <playbook-item
            :playbooks="playbooks"
            :playbook="hook.value.playbook"
            :vars="hook.value.vars"
            @save="updateSp(hook, $event)"
            :loading="isSaveLoading"
          />
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
  </v-card>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import api from "@/api";
import { useStore } from "@/store";
import confirmDialog from "@/components/confirmDialog.vue";
import PlaybookItem from "@/components/ui/playbookItem.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const playbookTip = `Access to internal fields of the instance or service provider is carried out through the template syntax: {{ .Instance }} and {{ .SP }}
Example of obtaining an instance identifier: {{ .Instance.Uuid }}`;

const newHooks = ref([]);
const isSaveLoading = ref(false);

onMounted(() => {
  store.dispatch("playbooks/fetch");

  newHooks.value = Object.keys(template.value.hooks || {}).map((key) => ({
    key,
    oldKey: key,
    value: template.value.hooks[key],
  }));
});

const playbooks = computed(() => store.getters["playbooks/all"]);

const addHook = async () => {
  newHooks.value.push({ key: "", value: {} });
};

const updateSp = async (hook, data) => {
  isSaveLoading.value = true;
  try {
    const hooks = { ...template.value.hooks };

    hooks[hook.oldKey] = undefined;
    hooks[hook.key] = data;

    await api.servicesProviders.update(template.value.uuid, {
      ...template.value,
      hooks: hooks,
    });

    store.dispatch("reloadBtn/onclick");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save offering items",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

watch(
  template,
  () =>
    (newHooks.value = Object.keys(template.value.hooks || {}).map((key) => ({
      key,
      value: template.value.hooks[key],
    })))
);
</script>

<script>
export default {
  name: "sp-hooks",
};
</script>
