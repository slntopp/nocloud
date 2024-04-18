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
        :suffix="accountCurrency?.title"
        v-model.number="item.amount"
      />
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <v-text-field :rules="generalRule" v-model="item.title" />
    </template>

    <template v-slot:[`item.instance`]="{ item }">
      <v-text-field readonly disabled :value="getInstance(item.instance)" />
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
  instances: {},
});
const { items, account, showDelete, showDate, readonly, sortBy, instances } =
  toRefs(props);

const emit = defineEmits("click:delete");

const store = useStore();

const generalRule = ref([(v) => !!v || "This field is required!"]);

const headers = computed(() =>
  [
    showDate.value && { text: "Date", value: "date" },
    { text: "Instance", value: "instance", sortable: false, width: 200 },
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

const getInstance = (uuid) => {
  return instances.value?.find((i) => i.uuid === uuid)?.title;
};
</script>

<style scoped></style>
