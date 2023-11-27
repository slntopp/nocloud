<template>
  <nocloud-table
    table-name="empty-addons-prices"
    class="pa-4"
    item-key="id"
    v-model="selected"
    :show-expand="true"
    :items="filteredAddons ?? []"
    :headers="addonsHeaders"
    :expanded.sync="expanded"
  >
    <template v-slot:top>
      <v-toolbar flat color="background">
        <v-toolbar-title>Addons</v-toolbar-title>
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
        :value="item.key.split(`; product: ${product}`)[0]"
        :rules="generalRule"
        @input="setKey($event, item)"
      />
    </template>

    <template v-slot:[`item.price`]="{ item }">
      <v-text-field
        dense
        type="number"
        :value="item.price"
        :suffix="defaultCurrency"
        :rules="generalRule"
        @input="(value) => (item.price = parseFloat(value))"
      />
    </template>

    <template v-slot:[`item.auto`]="{ item }">
      <v-switch
        :input-value="getAutoEnable(item)"
        @change="setAutoEnable(item, $event)"
      />
    </template>

    <template v-slot:[`item.period`]="{ item }">
      <date-field
        :period="fullDate[item.id]"
        @changeDate="(date) => setPeriod(date.value, item)"
      />
    </template>

    <template v-slot:expanded-item="{ headers, item }">
      <td />
      <td :colspan="headers.length - 1">
        <v-subheader class="px-0">Description</v-subheader>

        <rich-editor class="html-editor" v-model="item.title" />
      </td>
    </template>
  </nocloud-table>
</template>

<script setup>
import { computed, ref } from "vue";
import useCurrency from "@/hooks/useCurrency.js";
import { getFullDate, getTimestamp } from "@/functions.js";

import dateField from "@/components/date.vue";
import nocloudTable from "@/components/table.vue";
import RichEditor from "@/components/ui/richEditor.vue";
import confirmDialog from "@/components/confirmDialog.vue";

const props = defineProps({
  product: { type: String, required: true },
  item: { type: Object, required: true },
  addons: { type: Array, required: true },
});
const emits = defineEmits(["update:addons"]);

const { defaultCurrency } = useCurrency();
const fullDate = ref({});
const selected = ref([]);
const expanded = ref([]);

const addonsHeaders = [
  { text: "Key", value: "key" },
  { text: "Price", value: "price" },
  { text: "Auto", value: "auto" },
  { text: "Period", value: "period", width: 400 },
];
const generalRule = [(v) => !!v || "This field is required!"];

const filteredAddons = computed(() =>
  props.addons.filter(
    ({ key }) => key.split("; product: ")[1] === props.product
  )
);

function addConfig() {
  const addons = JSON.parse(JSON.stringify(props.addons ?? []));

  addons.push({
    id: Math.random().toString(16).slice(2),
    key: `key; product: ${props.product}`,
    title: "",
    price: 0,
    period: 0,
    kind: "PREPAID",
  });

  emits("update:addons", addons);
}

function removeConfig() {
  const addons = JSON.parse(JSON.stringify(props.addons ?? [])).filter(
    ({ id }) => !selected.value.find((el) => el.id === id)
  );

  selected.value = [];
  emits("update:addons", addons);
}

function setKey(value, item) {
  item.key = `${value}; product: ${props.product}`;
  emits("update:addons", JSON.parse(JSON.stringify(props.addons)));
}

function setAutoEnable(item, value) {
  item.auto = !!value;
  emits("update:addons", JSON.parse(JSON.stringify(props.addons)));
}

function getAutoEnable(item) {
  return props.item.meta.autoEnabled.includes(item.key);
}

function setPeriod(value, item) {
  fullDate.value[item.id] = value;
  item.period = getTimestamp(value);
}

props.addons?.forEach((a) => {
  fullDate.value[a.id] = getFullDate(a.period);
  a.auto = getAutoEnable(a);
});
</script>
