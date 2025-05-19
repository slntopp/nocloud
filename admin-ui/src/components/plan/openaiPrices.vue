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
          v-if="tab === 'Old prices'"
          :headers="oldPricesHeaders"
          :items="oldPricesResources"
          :show-select="false"
        >
          <template v-slot:[`item.header`]="{ item }">
            <v-text-field type="number" dense v-model.number="item.price" />
          </template>
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field type="number" dense v-model.number="item.price" />
          </template>
        </nocloud-table>

        <div v-else-if="tab === 'Prices'">
          <div class="d-flex justify-space-between align-center">
            <v-text-field
              style="max-width: 400px"
              v-model="searchParam"
              label="Search"
            ></v-text-field>
            <v-dialog v-model="isAddNewModelOpen" width="70%" max-width="800">
              <template v-slot:activator="{ on, attrs }">
                <v-btn color="primary" v-bind="attrs" v-on="on">
                  Add new
                </v-btn>
              </template>

              <v-card class="d-flex pa-5 flex-column">
                <v-card-title>Add new model</v-card-title>
                <v-text-field
                  v-model="newModel"
                  outlined
                  label="Full key"
                ></v-text-field>

                <v-row>
                  <v-col cols="3">
                    <v-text-field
                      :value="newModel.split('|')[0]"
                      outlined
                      disabled
                      label="Specification"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="3">
                    <v-text-field
                      :value="newModel.split('|')[1]"
                      outlined
                      disabled
                      label="Key"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="3">
                    <v-text-field
                      :value="newModel.split('|')[2]"
                      outlined
                      disabled
                      label="Type"
                    ></v-text-field>
                  </v-col>
                </v-row>

                <div class="d-flex justify-end">
                  <v-btn
                    color="primary"
                    class="mr-3"
                    @click="isAddNewModelOpen = false"
                    >Close</v-btn
                  >
                  <v-btn color="primary" @click="addNewModelToNewResources"
                    >Add</v-btn
                  >
                </div>
              </v-card>
            </v-dialog>
          </div>
          <nocloud-table
            :headers="newPricesHeaders"
            :items="newPricesResourcesFiltred"
            :show-select="false"
          >
            <template v-slot:[`item.price`]="{ item }">
              <v-text-field type="number" dense v-model.number="item.price" />
            </template>

            <template v-slot:[`item.actions`]="{ item }">
              <div class="d-flex justify-center">
                <v-btn icon>
                  <v-icon @click="deleteModelFromNewPrices(item)">
                    mdi-delete
                  </v-icon>
                </v-btn>
              </div>
            </template>

            <template v-slot:[`item.title`]="{ item }">
              <v-text-field dense v-model="item.title" />
            </template>

            <template v-slot:[`item.key`]="{ item }">
              <v-text-field dense v-model="item.key" />
            </template>

            <template v-slot:[`item.specification`]="{ item }">
              <v-text-field dense v-model="item.specification" />
            </template>

            <template v-slot:[`item.type`]="{ item }">
              <v-text-field dense v-model="item.type" />
            </template>
          </nocloud-table>
        </div>

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
import { computed, onMounted, ref, toRefs } from "vue";
import NocloudTable from "@/components/table.vue";
import planAddonsTable from "@/components/planAddonsTable.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const tabs = ref(["Prices", "Old prices", "Addons"]);
const tabsIndex = ref(0);

const addons = ref([]);

const oldPricesResources = ref([
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
  },
]);
const newPricesResources = ref([]);

const oldPricesHeaders = [
  { text: "Key", value: "key" },
  { text: "Title", value: "title" },
  {
    text: "Price",
    value: "price",
    width: 200,
  },
];

const newPricesHeaders = [
  { text: "Key", value: "key" },
  { text: "Title", value: "title" },
  { text: "Specification", value: "specification" },
  { text: "Type", value: "type" },
  {
    text: "Price",
    value: "price",
    width: 125,
  },
  {
    text: "Actions",
    value: "actions",
    width: 80,
  },
];
const newModel = ref("");
const searchParam = ref("");
const isAddNewModelOpen = ref(false);

const isSaveLoading = ref(false);

onMounted(() => {
  oldPricesResources.value = oldPricesResources.value.map((resource) => {
    const realResource = template.value.resources.find(
      (realResource) => realResource.key === resource.key
    );

    return { ...resource, price: realResource?.price || 0 };
  });

  template.value.resources.forEach((resource) => {
    if (oldPricesResources.value.find((r) => r.key === resource.key)) {
      return;
    }

    const [specification, key, type] = resource.key.split("|");

    newPricesResources.value.push({
      specification,
      key,
      type,
      price: resource.price,
      title: resource.title,
    });
  });

  addons.value = template.value.addons;
});

const newPricesResourcesFiltred = computed(() => {
  const param = searchParam.value.toLowerCase();
  return newPricesResources.value.filter(
    (r) =>
      r.title.includes(param) ||
      r.key.includes(param) ||
      r.specification.includes(param) ||
      r.type.includes(param)
  );
});

const addNewModelToNewResources = () => {
  const [specification, key, type] = newModel.value.split("|");
  newPricesResources.value.push({
    specification: specification || "",
    key: key || "",
    type: type || "",
    price: 0,
    title: newModel.value,
  });

  newModel.value = "";
  isAddNewModelOpen.value = false;
};

const deleteModelFromNewPrices = (item) => {
  newPricesResources.value = newPricesResources.value.filter(
    (i) =>
      !(
        item.specification === i.specification &&
        item.type === i.type &&
        item.key === i.key
      )
  );
};

const save = async () => {
  isSaveLoading.value = true;
  try {
    const oldPricesResult = JSON.parse(
      JSON.stringify(oldPricesResources.value)
    );
    const imageSize1792x1024 = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1792x1024_quality_standard")
    );
    const imageSize1792x1024HD = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1792x1024_quality_hd")
    );

    const imageSize1024x1792 = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1024x1792_quality_standard")
    );
    const imageSize1024x1792HD = oldPricesResult.find(({ key }) =>
      key.includes("image_size_1024x1792_quality_hd")
    );

    imageSize1792x1024.price = imageSize1024x1792.price;
    imageSize1792x1024HD.price = imageSize1024x1792HD.price;

    const newPricesResult = [];

    for (const value of newPricesResources.value) {
      const key = [];
      if (value.specification) {
        key.push(value.specification);
      }
      if (value.key) {
        key.push(value.key);
      }
      if (value.type) {
        key.push(value.type);
      }
      newPricesResult.push({
        kind: "POSTPAID",
        price: value.price,
        title: value.title,
        public: true,
        key: key.length > 1 ? key.join("|") : key[0],
      });
    }

    await api.plans.update(props.template.uuid, {
      ...props.template,
      products: {},
      addons: addons.value,
      resources: [...oldPricesResult, ...newPricesResult],
    });

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Price model edited successfully",
    });
    store.dispatch("reloadBtn/onclick");
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
