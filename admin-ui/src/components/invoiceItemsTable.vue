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
        :suffix="accountCurrency?.title"
        v-model.number="item.price"
      />
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <v-text-field :rules="generalRule" v-model="item.title" />
    </template>

    <template v-slot:[`item.amount`]="{ item }">
      <v-text-field
        type="number"
        v-model.number="item.amount"
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

    <template v-slot:[`item.actions`]="{ index }">
      <div class="d-flex justify-center">
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
const unitItems = ref(["Pcs", "Szt", "Hour`s"]);

const headers = computed(() =>
  [
    showDate.value && { text: "Date", value: "date" },
    { text: "Title", value: "title", width: 125 },
    { text: "Description", value: "description" },
    { text: "Amount", value: "amount", width: 175 },
    showDelete.value && { text: "Actions", value: "actions", width:25 },
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
