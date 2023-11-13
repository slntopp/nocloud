<template>
  <nocloud-table
    table-name="empty-addons-prices"
    class="pa-4"
    item-key="id"
    v-model="selected"
    :show-expand="true"
    :items="product.meta.addons ?? []"
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

    <template v-slot:[`item.title`]="{ item }">
      <v-text-field
        dense
        v-model="item.title"
        :rules="generalRule"
      />
    </template>

    <template v-slot:[`item.price`]="{ item }">
      <v-text-field
        dense
        type="number"
        :value="item.price"
        :suffix="defaultCurrency"
        :rules="generalRule"
        @input="(value) => item.price = parseFloat(value)"
      />
    </template>

    <template v-slot:[`item.period`]="{ item }">
      <date-field
        :period="fullDate[item.id]"
        @changeDate="(date) => setPeriod(date.value, item.id)"
      />
    </template>

    <template v-slot:[`item.public`]="{ item }">
      <v-switch v-model="item.public" />
    </template>
    
    <template v-slot:expanded-item="{ headers, item }">
      <td />
      <td :colspan="headers.length - 1">
        <v-subheader class="px-0">Description</v-subheader>

        <rich-editor class="html-editor" v-model="item.description" />
      </td>
    </template>
  </nocloud-table>
</template>

<script setup>
import { ref } from "vue";
import useCurrency from "@/hooks/useCurrency.js";
import { getFullDate, getTimestamp } from "@/functions.js";

import dateField from "@/components/date.vue";
import nocloudTable from "@/components/table.vue";
import RichEditor from "@/components/ui/richEditor.vue";
import confirmDialog from "@/components/confirmDialog.vue";

const props = defineProps({
  product: { type: Object, required: true }
})
const emits = defineEmits(["update:addons"])

const { defaultCurrency } = useCurrency();
const fullDate = ref({})
const selected = ref([])
const expanded = ref([])

const addonsHeaders = [
  { text: "Title", value: "title" },
  { text: "Price", value: "price" },
  { text: "Period", value: "period", width: 400 },
  { text: "Public", value: "public" },
];
const generalRule = [(v) => !!v || "This field is required!"];

function addConfig() {
  const addons = JSON.parse(JSON.stringify(props.product.meta.addons ?? []))

  addons.push({
    id: Math.random().toString(16).slice(2),
    description: "",
    title: "",
    price: 0,
    period: 0,
    public: true
  })

  emits("update:addons", addons)
}

function removeConfig() {
  const addons = JSON.parse(JSON.stringify(props.product.meta.addons ?? []))
    .filter(({ id }) => !selected.value.find((el) => el.id === id))

  selected.value = []
  emits("update:addons", addons)
}

function setPeriod(value, id) {
  const item = props.product.meta.addons.find((addon) => addon.id === id)

  fullDate.value[id] = value
  item.period = getTimestamp(value)
}

props.product.meta.addons?.forEach(({ period, id }) => {
  fullDate.value[id] = getFullDate(period);
});
</script>
