<template>
  <nocloud-table
    :items="items"
    :headers="headers"
    :show-select="false"
    :sort-by="sortBy"
    sort-desc
  >
    <template v-slot:[`item.price`]="{ item }">
      <v-text-field
        type="number"
        :rules="!readonly ? generalRule : []"
        :readonly="readonly"
        :suffix="accountCurrency?.code"
        v-model.number="item.price"
      />
    </template>

    <template v-slot:[`item.amount`]="{ item }">
      <v-text-field
        type="number"
        v-model.number="item.amount"
        :rules="generalRule"
      />
    </template>

    <template v-slot:[`item.price`]="{ item }">
      <v-text-field
        type="number"
        v-model.number="item.price"
        :rules="generalRule"
      />
    </template>

    <template v-slot:[`item.unit`]="{ item }">
      <v-select :rules="generalRule" v-model="item.unit" :items="unitItems" />
    </template>

    <template v-slot:[`item.description`]="{ item }">
      <v-textarea
        :readonly="readonly"
        :rules="!readonly ? generalRule : []"
        no-resize
        label="Items description"
        v-model="item.description"
        rows="1"
        auto-grow
      />
    </template>

    <template v-if="showDelete" v-slot:[`item.actions`]="{ index }">
      <div class="d-flex justify-center">
        <v-btn icon @click="emit('click:delete', index)"
          ><v-icon>mdi-delete</v-icon></v-btn
        >
      </div>
    </template>

    <template v-if="showDelete" v-slot:[`item.total`]="{ item }">
      <span>{{
        [
          (item.price * item.amount || 0).toFixed(2),
          account?.currency?.code,
        ].join(" ")
      }}</span>
    </template>

    <template v-if="items.length" v-slot:body.append>
      <tr>
        <td></td>
        <td></td>
        <td></td>
        <td></td>
        <td>
          {{
            [
              items
                .reduce(
                  (acc, item) => acc + (item.amount || 0) * (item.price || 0),
                  0
                )
                .toFixed(2),
              account?.currency?.code,
            ].join(" ")
          }}
        </td>
        <td></td>
      </tr>
    </template>
  </nocloud-table>
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import { computed, ref, toRefs } from "vue";
import { useStore } from "@/store";

const props = defineProps({
  items: { required: true },
  account: { required: true },
  showDelete: { type: Boolean, default: false },
  showDate: { type: Boolean, default: false },
  readonly: { type: Boolean, default: false },
  sortBy: {},
});
const { items, account, showDelete, showDate, readonly, sortBy } =
  toRefs(props);

const emit = defineEmits("click:delete");

const store = useStore();

const generalRule = ref([(v) => !!v || v === 0 || "This field is required!"]);
const unitItems = ref(["Pcs", "Szt", "Hour`s"]);

const headers = computed(() =>
  [
    showDate.value && { text: "Date", value: "date" },
    { text: "Description", value: "description", width: "40%" },
    { text: "Unit", value: "unit", width: 100 },
    { text: "Amount", value: "amount", width: 100 },
    { text: "Price", value: "price", width: 100 },
    { text: "Total", value: "total", width: 75 },
    showDelete.value ? { text: "Actions", value: "actions", width: 25 } : null,
  ].filter((c) => !!c)
);
const accountCurrency = computed(
  () => account.value?.currency || store.getters["currencies/default"]
);
</script>

<style scoped></style>
