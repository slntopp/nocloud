<template>
  <v-card class="pa-5" color="background-light">
    <v-autocomplete
      label="Playbook"
      :items="playbooks"
      v-model="playbookUuid"
      item-value="uuid"
      item-text="title"
      clearable
    />

    <v-data-table
      class="elevation-0 background-light rounded-lg"
      color="background-light"
      :headers="varsHeaders"
      :items="playbookVars"
    >
      <template v-slot:footer>
        <div class="d-flex justify-end pa-3">
          <v-btn small @click="addNewVar"> Add new </v-btn>
        </div>
      </template>

      <template v-slot:item.value="{ item }">
        <v-text-field v-model="item.value" />
      </template>

      <template v-slot:item.key="{ item }">
        <v-text-field v-model="item.key" />
      </template>

      <template v-slot:item.actions="{ index }">
        <div class="d-flex justify-center">
          <v-btn icon>
            <v-icon @click="() => deleteVar(index)"> mdi-delete </v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>

    <div class="d-flex justify-end">
      <v-btn :loading="isSaveLoading" @click="savePlaybookAction">Save</v-btn>
    </div>
  </v-card>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import yaml from "yaml";

const props = defineProps(["playbooks", "playbook", "vars", "isSaveLoading"]);
const { playbooks, playbook, vars } = toRefs(props);
const emit = defineEmits(["save"]);

const varsHeaders = [
  { text: "Key", value: "key" },
  { text: "Value", value: "value" },
  { text: "Actions", value: "actions" },
];

const playbookVars = ref([]);
const playbookUuid = ref("");

onMounted(() => {
  setData();
});

const fullPlaybook = computed(() =>
  playbooks.value.find((p) => p.uuid === playbookUuid.value)
);

const defaultVars = computed(() => {
  if (!fullPlaybook.value) {
    return;
  }

  const content = yaml.parse(fullPlaybook.value.content);
  return content?.[0]?.environment;
});

function setData() {
  setVars();
  playbookUuid.value = props.playbook;
}

function addNewVar() {
  playbookVars.value.push({ key: "", value: "" });
}

function deleteVar(index) {
  playbookVars.value = playbookVars.value.filter((_, ind) => ind != index);
}

function setVars() {
  playbookVars.value = Object.keys(props.vars || {}).map((key) => ({
    key,
    value: props.vars[key],
  }));
}

function savePlaybookAction() {
  var vars = {};

  playbookVars.value.forEach((v) => {
    if (!v.key || !v.value) {
      return;
    }
    vars[v.key] = v.value;
  });

  if (!playbookUuid.value) {
    vars = {};
  }

  emit("save", { playbook: playbookUuid.value, vars });
}

watch(defaultVars, () => {
  setTimeout(() => {
    Object.keys(defaultVars.value || {}).map((key) => {
      if (playbookVars.value.find((v) => v.key === key)) {
        return;
      }

      playbookVars.value.push({ key: key, value: "" });
    });
  }, 0);
});

watch(playbookUuid, () => setVars(playbookVars));
watch([playbook, vars], () => setData());
</script>
