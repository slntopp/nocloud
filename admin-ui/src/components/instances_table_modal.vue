<template>
  <component
    offset-y
    max-width="1280"
    :is="type === 'dialog' ? VDialog : VMenu"
    :value="visible"
    :close-on-content-click="false"
    @input="emits('close')"
  >
    <template v-slot:activator="{ on, attrs }">
      <slot v-on="on" v-bind="attrs" name="activator" />
    </template>

    <v-card
      class="pa-4"
      elevation="0"
      color="background"
      @mouseenter="emits('hover')"
      @mouseleave="emits('close')"
    >
      <instances-table
        open-in-new-tab
        style="margin-top: 0 !important"
        :value="[]"
        :headers="headers"
        no-search
        :show-select="false"
        :custom-filter="{ account: [uuid] }"
      />
    </v-card>
  </component>
</template>

<script setup>
import { VDialog, VMenu } from "vuetify/lib/components";
import instancesTable from "@/components/instancesTable.vue";
import { onMounted, toRefs } from "vue";

const props = defineProps({
  uuid: { type: String, required: true },
  visible: { type: Boolean, required: true },
  type: { type: String, default: "dialog" },
});
const emits = defineEmits(["hover", "close"]);

const { uuid } = toRefs(props);

onMounted(() => console.log(2332));

const headers = [
  { text: "Title", value: "title" },
  { text: "Due date", value: "dueDate" },
  { text: "Status", value: "state" },
  { text: "Tariff", value: "product" },
  { text: "Price", value: "accountPrice" },
];
</script>
