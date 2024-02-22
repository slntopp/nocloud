<template>
  <nocloud-table
    :items="items"
    :headers="headers"
    :show-select="false"
    :sort-by="sortBy"
    sort-desc
  >
    <template v-slot:[`item.amount`]="{ item }">
      <v-text-field
        type="number"
        :rules="!readonly ? generalRule : []"
        :readonly="readonly"
        :suffix="accountCurrency"
        v-model.number="item.amount"
      />
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <v-text-field :rules="generalRule" v-model="item.title" />
    </template>

    <template v-slot:[`item.actions`]="{ index }">
      <div class="d-flex justify-end">
        <v-btn icon @click="emit('click:delete', index)"
          ><v-icon>mdi-delete</v-icon></v-btn
        >
      </div>
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

const generalRule = ref([(v) => !!v || "This field is required!"]);

const headers = computed(() =>
  [
    showDate.value && { text: "Date", value: "date" },
    { text: "Title", value: "title", sortable: false },
    { text: "Amount", value: "amount", sortable: false, width: 200 },
    showDelete.value && {
      text: "Actions",
      value: "actions",
      sortable: false,
      width: 50,
    },
  ].filter((c) => !!c)
);
const accountCurrency = computed(
  () => account.value?.currency || store.getters["currencies/default"]
);
</script>

<style scoped></style>
