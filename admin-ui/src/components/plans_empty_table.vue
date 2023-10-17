<template>
  <nocloud-table
    table-name="plans-meta-resources"
    item-key="id"
    class="mt-4"
    v-model="selected"
    :items="resources"
    :headers="headers"
  >
    <template v-slot:top>
      <v-toolbar flat color="background">
        <v-toolbar-title>Resources</v-toolbar-title>
        <v-divider inset vertical class="mx-4" />
        <v-spacer />

        <v-btn class="mr-2" color="background-light" @click="addConfig">
          Create
        </v-btn>
        <confirm-dialog @confirm="removeConfig">
          <v-btn color="background-light" :disabled="selected.length < 1">
            Delete
          </v-btn>
        </confirm-dialog>
      </v-toolbar>
    </template>

    <template v-slot:[`item.key`]="{ item }">
      <v-text-field
        dense
        :value="item.key"
        :rules="generalRule"
        @change="(value) => changeResource('key', value, item.id)"
      />
    </template>

    <template v-slot:[`item.value`]="{ item }">
      <v-text-field
        dense
        type="number"
        :value="item.value"
        :rules="generalRule"
        @input="(value) => changeResource('value', value, item.id)"
      />
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <v-text-field
        dense
        :value="item.title"
        :rules="generalRule"
        @change="(value) => changeResource('title', value, item.id)"
      />
    </template>
  </nocloud-table>
</template>

<script setup>
import { ref, toRefs } from "vue";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";

const props = defineProps({
  resources: { type: Array, required: true },
});
const emits = defineEmits(["update:resource"]);
const { resources } = toRefs(props);

const selected = ref([]);
const generalRule = [(v) => !!v || "This field is required!"];

const headers = [
  { text: "Key", value: "key" },
  { text: "Value", value: "value" },
  { text: "Title", value: "title" },
];

function changeResource(key, value, id) {
  emits("update:resource", { key, value, id });
}

function addConfig() {
  const value = JSON.parse(JSON.stringify(resources.value));

  value.push({
    key: "",
    title: "",
    value: 0,
    id: Math.random().toString(16).slice(2)
  });
  changeResource("resources", value);
}

function removeConfig() {
  const value = resources.value.filter(
    ({ id }) => !selected.value.find((el) => el.id === id)
  );
  changeResource("resources", value);
}
</script>