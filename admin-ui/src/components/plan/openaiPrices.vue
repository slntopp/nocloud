<template>
  <div class="pa-5">
    <v-tabs
      class="rounded-t-lg"
      v-model="tabsIndex"
      background-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab">{{ tab }}</v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabsIndex"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item v-for="tab in tabs" :key="tab">
        <nocloud-table
          v-if="tab === 'Prices'"
          table-name="openai-prices_table"
          :headers="headers"
          :items="resources"
          :show-select="false"
        >
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field type="number" dense v-model.number="item.price" />
          </template>
        </nocloud-table>

        <div class="os-tab__card" v-else>
          <plan-addons-table
            @change:addons="addons = $event"
            :addons="template.addons"
          />
        </div>
      </v-tab-item>
    </v-tabs-items>
    <div class="d-flex justify-end">
      <v-btn :loading="isSaveLoading" @click="save">Save</v-btn>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, toRefs } from "vue";
import NocloudTable from "@/components/table.vue";
import planAddonsTable from "@/components/planAddonsTable.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const tabs = ref(["Prices", "Addons"]);
const tabsIndex = ref(0);

const addons = ref([]);

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
    key: "image_size_1024x1024_quality_standard",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1024",
    public: true,
  },
  {
    key: "image_size_1024x1024_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1024 HD",
    public: true,
  },
  {
    key: "image_size_1024x1792_quality_standard",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024",
    public: true,
  },
  {
    key: "image_size_1024x1792_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024 HD",
    public: true,
  },
  {
    key: "image_size_1792x1024_quality_standard",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024",
    public: true,
  },
  {
    key: "image_size_1792x1024_quality_hd",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    title: "Image 1024*1792 or 1792*1024 HD",
    public: true,
  }
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

  addons.value = template.value.addons;
});

const save = async () => {
  isSaveLoading.value = true;
  try {
    const result = JSON.parse(JSON.stringify(resources.value))
    const imageSize1792x1024 = result.find(({ key }) =>
      key.includes('image_size_1792x1024_quality_standard')
    )
    const imageSize1792x1024HD = result.find(({ key }) =>
      key.includes('image_size_1792x1024_quality_hd')
    )

    const imageSize1024x1792 = result.find(({ key }) =>
      key.includes('image_size_1024x1792_quality_standard')
    )
    const imageSize1024x1792HD = result.find(({ key }) =>
      key.includes('image_size_1024x1792_quality_hd')
    )

    imageSize1792x1024.price = imageSize1024x1792.price
    imageSize1792x1024HD.price = imageSize1024x1792HD.price

    await api.plans.update(props.template.uuid, {
      ...props.template,
      products: {},
      addons: addons.value,
      resources: result,
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
