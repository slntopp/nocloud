<template>
  <nocloud-table
    :loading="isPromocodesLoading"
    table-name="instance-promocodes"
    :items="items"
    :headers="headers"
    sort-by="exec"
    sort-desc
  >
    <template v-slot:[`item.exec`]="{ item }">
      {{ formatSecondsToDate(item.exec, true) }}
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <router-link
        :to="{ name: 'Promocode page', params: { uuid: item.promocode.uuid } }"
      >
        {{ item.promocode.title }}
      </router-link>
    </template>

    <template v-slot:[`item.code`]="{ item }">
      {{ item.promocode.code }}
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { ListPromocodesRequest } from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";
import { computed, onMounted, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const promocodes = ref([]);
const isPromocodesLoading = ref(false);

const headers = ref([
  { text: "Promocode title", value: "title" },
  { text: "Promocode code", value: "code" },
  { text: "Execution", value: "exec" },
]);

const items = computed(() => {
  const items = [];

  promocodes.value.map((promocode) => {
    (promocode.uses || [])

      .filter((use) => use.instance === template.value.uuid)
      .forEach((use) => {
        items.push({
          exec: use.exec,
          promocode: {
            title: promocode.title,
            uuid: promocode.uuid,
            code: promocode.code,
          },
        });
      });
  });

  return items;
});

onMounted(() => {
  fetchInstancePromocodes();
});

const fetchInstancePromocodes = async () => {
  try {
    isPromocodesLoading.value = true;

    const data = await store.getters["promocodes/promocodesClient"].list(
      ListPromocodesRequest.fromJson({
        limit: 20,
        filters: {
          resources: [`instances/${template.value.uuid}`],
        },
      })
    );

    promocodes.value = data.toJson().promocodes || [];
  } finally {
    isPromocodesLoading.value = false;
  }
};
</script>

<script>
export default {
  name: "instance-promocodes",
};
</script>

<style scoped></style>
