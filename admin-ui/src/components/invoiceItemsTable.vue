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

    <template v-slot:[`item.type`]="{ item }">
      <v-select v-model="item.type" :items="types"></v-select>
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <v-select
        v-if="item.type === 'instance'"
        :rules="generalRule"
        :items="instances"
        item-value="title"
        item-text="title"
        :value="item.title"
        @change="onInstanceChange(item, $event)"
      />
      <v-text-field v-else :rules="generalRule" v-model="item.title" />
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
import { getInstancePrice } from "@/functions";

const props = defineProps({
  items: { required: true },
  instances: { type: Array, default: () => [] },
  account: { required: true },
  showDelete: { type: Boolean, default: false },
  showDate: { type: Boolean, default: false },
  readonly: { type: Boolean, default: false },
  sortBy: {},
});
const { items, account, showDelete, showDate, readonly, sortBy, instances } =
  toRefs(props);

const emit = defineEmits("click:delete");

const store = useStore();

const generalRule = ref([(v) => !!v || "This field is required!"]);
const types = ref(["default", "instance"]);

const headers = computed(() =>
  [
    showDate.value && { text: "Date", value: "date" },
    { text: "Type", value: "type" },
    { text: "Title", value: "title" },
    { text: "Amount", value: "amount" },
    showDelete.value && { text: "Actions", value: "actions" },
  ].filter((c) => !!c)
);
const accountCurrency = computed(
  () => account.value?.currency || store.getters["currencies/default"]
);

const onInstanceChange = (item, value) => {
  item.title = value;
  console.log(
    getInstancePrice(instances.value.find((i) => i.title === value)),
    instances.value.find((i) => i.title === value)
  );
  item.amount = getInstancePrice(
    instances.value.find((i) => i.title === value)
  );
};
</script>

<style scoped></style>
