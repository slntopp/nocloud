<template>
  <v-dialog persistent v-model="isOpen" width="90vw">
    <template v-slot:activator="{ on, attrs }">
      <v-btn icon v-bind="attrs" v-on="on">
        <v-icon> mdi-menu-open </v-icon>
      </v-btn>
    </template>

    <v-card color="background-light">
      <nocloud-table
        table-name="empty-addons-prices"
        class="pa-4"
        item-key="id"
        :show-select="false"
        :items="addons ?? []"
        :headers="addonsHeaders"
      >
        <template v-slot:top>
          <v-toolbar flat color="background">
            <v-toolbar-title>Addons</v-toolbar-title>
            <v-divider inset vertical class="mx-4" />
            <v-spacer />
          </v-toolbar>
        </template>

        <template v-slot:[`item.period`]="{ item }">
          <span>{{ getBillingPeriod(item.period) }}</span>
        </template>

        <template v-slot:[`item.sell`]="{ item }">
          <v-switch
            :input-value="isSell(item)"
            @change="changeIsSell(item, $event)"
          />
        </template>
      </nocloud-table>
      <div class="d-flex justify-end mt-3 pa-2">
        <v-btn @click="isOpen = false">Close</v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, toRefs } from "vue";
import nocloudTable from "@/components/table.vue";
import { getBillingPeriod } from "../functions";

const props = defineProps({
  product: { type: String, required: true },
  item: { type: Object, required: true },
  addons: { type: Array, required: true },
  rules: { type: Object },
});

const { addons, item } = toRefs(props);
const emits = defineEmits(["update:addons"]);
//
// const store = useStore();

const isOpen = ref(false);

const addonsHeaders = [
  { text: "Key", value: "key" },
  { text: "Title", value: "title" },
  { text: "Price", value: "price" },
  { text: "Period", value: "period" },
  { text: "Sell", value: "sell" },
];

const isSell = (addon) => {
  return item.value?.meta?.addons?.includes(addon.key);
};

const changeIsSell = (addon, value) => {
  emits(
    "update:addons",
    addons.value
      .filter((a) => (addon.key === a.key ? value : isSell(a)))
      .map((a) => a.key)
  );
};
</script>
