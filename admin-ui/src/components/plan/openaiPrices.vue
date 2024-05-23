<template>
  <div class="pa-5">
    <nocloud-table
      table-name="openai-prices_table"
      :headers="headers"
      :items="resources"
      :show-select="false"
    >
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field type="number" dense v-model.number="item.price" />
      </template>
    </nocloud-table>
    <div class="d-flex justify-end">
      <v-btn :loading="isSaveLoading" @click="save">Save</v-btn>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, toRefs } from "vue";
import NocloudTable from "@/components/table.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const resources = ref([
  {
    key: "input_kilotoken",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Input kilotoken",
    public: true,
  },
  {
    key: "output_kilotoken",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Output kilotoken",
    public: true,
  },
  {
    key: "image_size_1024_1024_quality_standart",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1024",
    public: true,
  },
  {
    key: "image_size_1024_1024_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1024 HD",
    public: true,
  },
  {
    key: "image_size_1024_1792_quality_standart",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024",
    public: true,
  },
  {
    key: "image_size_1024_1792_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024 HD",
    public: true,
  },
]);
const headers = [
  { text: "Key", value: "key" },
  { text: "Title", value: "title" },
  {
    text: "Price",
    value: "price",
    width: 200,
  },
];

const isSaveLoading = ref(false);

onMounted(() => {
  resources.value = resources.value.map((resource) => {
    const realResource = template.value.resources.find(
      (realResource) => realResource.key === resource.key
    );
    console.log(realResource?.price);

    return { ...resource, price: realResource?.price || 0 };
  });
});

const save = async () => {
  isSaveLoading.value = true;
  try {
    await api.plans.update(props.template.uuid, {
      ...props.template,
      products: {},
      resources: JSON.parse(JSON.stringify(resources.value)),
    });

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Price model edited successfully",
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save prices",
    });
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<style scoped></style>
