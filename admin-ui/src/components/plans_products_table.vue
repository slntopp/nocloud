<template>
  <div class="pa-4">
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
          table-name="plans-products"
          item-key="id"
          sort-by="sorter"
          ref="table"
          v-if="tab === 'Products'"
          v-model="selected"
          :show-expand="true"
          :items="productsArray"
          :headers="headers"
          :expanded.sync="expanded"
        >
          <template v-slot:top>
            <v-toolbar flat color="background">
              <v-toolbar-title>Actions</v-toolbar-title>
              <v-divider inset vertical class="mx-4" />
              <v-spacer />

              <v-dialog width="90vw" v-model="isEditOpen">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn
                    :disabled="selected.length < 1"
                    v-bind="attrs"
                    v-on="on"
                    class="mr-2"
                    color="background-light"
                  >
                    Edit
                  </v-btn>
                </template>

                <v-card class="pa-4">
                  <v-subheader class="px-0"> Description: </v-subheader>
                  <rich-editor v-model="newMeta.description" />

                  <v-row class="mt-5" justify="end">
                    <v-btn class="mx-2" @click="isEditOpen = false"
                      >Close</v-btn
                    >
                    <v-btn @click="saveNewMeta" class="mx-2">Save</v-btn>
                  </v-row>
                </v-card>
              </v-dialog>

              <v-btn
                :disabled="selected.length < 1"
                class="mr-2"
                color="background-light"
                @click="copyProducts"
              >
                Copy
              </v-btn>

              <v-btn class="mr-2" color="background-light" @click="addConfig">
                Create
              </v-btn>
              <confirm-dialog @confirm="removeConfig">
                <v-btn color="background-light" :disabled="selected.length < 1">
                  Delete
                </v-btn>
              </confirm-dialog>
            </v-toolbar>
          </template>

          <template v-slot:[`item.key`]="{ item }">
            <v-text-field
              dense
              :value="item.key"
              :rules="[rules.required]"
              @change="(value) => changeProduct('key', value, item.id)"
            />
          </template>
          <template v-slot:[`item.title`]="{ item }">
            <v-text-field
              dense
              :value="item.title"
              :rules="[rules.required]"
              @change="(value) => changeProduct('title', value, item.id)"
            />
          </template>
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field
              dense
              type="number"
              :suffix="defaultCurrency"
              :value="item.price"
              :rules="[rules.price]"
              @change="(value) => changeProduct('price', value, item.id)"
            />
          </template>
          <template v-slot:[`item.period`]="{ item }">
            <date-field
              v-if="!isOneTime(item)"
              :period="fullDate[item.id]"
              @changeDate="(value) => changeDate(value, item.id)"
            />
          </template>
          <template v-slot:[`item.public`]="{ item }">
            <v-switch
              :input-value="item.public"
              @change="(value) => changeProduct('public', value, item.id)"
            />
          </template>
          <template v-slot:[`item.sorter`]="{ item }">
            <v-text-field
              type="number"
              :value="item.sorter"
              @change="(value) => changeProduct('sorter', value, item.id)"
            />
          </template>
          <template v-slot:[`item.group`]="{ item }">
            <div class="d-flex align-center">
              <v-select
                v-if="productId !== item.id"
                :items="[...groups.values()]"
                :value="products[item.key]?.group"
                @change="setGroup($event, item.id)"
              />
              <v-text-field v-else v-model="groupActionPayload" />
              <template v-if="productId !== item.id">
                <v-btn icon @click="setGroupAction('edit', products[item.key])">
                  <v-icon>mdi-pencil</v-icon>
                </v-btn>
                <v-btn icon @click="setGroupAction('add', products[item.key])">
                  <v-icon>mdi-plus</v-icon>
                </v-btn>
              </template>
              <template v-else>
                <v-btn icon @click="setGroupAction('')">
                  <v-icon>mdi-cancel</v-icon>
                </v-btn>
                <v-btn icon @click="invokeGroupAction(item)">
                  <v-icon>mdi-content-save</v-icon>
                </v-btn>
              </template>
            </div>
          </template>
          <template v-slot:[`item.kind`]="{ item }">
            <v-radio-group
              :disabled="isOneTime(item)"
              row
              mandatory
              :value="item.kind"
              @change="(value) => changeProduct('kind', value, item.id)"
            >
              <v-radio
                v-for="(kind, i) of kinds"
                :style="{ marginRight: i === kinds.length - 1 ? 0 : 16 }"
                :key="kind"
                :value="kind"
                :label="kind.toLowerCase()"
              />
            </v-radio-group>
          </template>

          <template v-slot:[`item.meta.oneTime`]="{ item }">
            <v-switch
              :input-value="item.meta?.oneTime"
              @change="changeOneTime(item, $event)"
            />
          </template>

          <template v-slot:[`item.addons`]="{ item }">
            <product-addons-dialog :addons="item.addons" />
          </template>

          <template v-slot:expanded-item="{ headers, item }">
            <td />
            <td :colspan="headers.length - 1">
              <v-text-field
                dense
                class="pt-4"
                label="Image link"
                v-if="type === 'empty'"
                v-model="item.meta.image"
              />

              <v-subheader class="px-0"> Description: </v-subheader>
              <rich-editor
                class="html-editor"
                @input="changeProduct('description', $event, item.id)"
                :value="item.description"
              />

              <template v-if="type === 'empty'">
                <plans-empty-table
                  :rules="rules"
                  :resources="item.meta.resources ?? []"
                  @update:resource="
                    changeMeta($event, item.id, item.meta.resources)
                  "
                />
              </template>

              <template v-else>
                <v-subheader class="px-0"> Amount of resources </v-subheader>
                <json-editor
                  :json="item.resources"
                  @changeValue="changeProduct('amount', $event, item.id)"
                />
              </template>

              <v-subheader class="px-0 pt-4">Installation price:</v-subheader>
              <v-text-field
                dense
                type="number"
                style="width: 150px"
                :value="item.installationFee"
                :suffix="defaultCurrency"
                @input="changeProduct('installationFee', +$event, item.id)"
              />
            </td>
          </template>
        </nocloud-table>

        <plans-resources-table
          v-else-if="tab === 'Resources'"
          :rules="rules"
          :resources="resources.filter((v) => v.virtual === false)"
          :default-virtual="false"
          :type="type"
          @change:resource="changeResource(false, $event)"
        />

        <plan-addons-table v-else-if="tab === 'Addons'" :addons="addons" />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import dateField from "@/components/date.vue";
import JsonEditor from "@/components/JsonEditor.vue";
import nocloudTable from "@/components/table.vue";
import plansResourcesTable from "@/components/plans_resources_table.vue";
import plansEmptyTable from "@/components/plans_empty_table.vue";
import planAddonsTable from "@/components/planAddonsTable.vue";
import productAddonsDialog from "@/components/product_addons_dialog.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { getFullDate } from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import RichEditor from "@/components/ui/richEditor.vue";

const props = defineProps({
  addons: { type: Array, required: true },
  type: { type: String, required: true },
  products: { type: Object, required: true },
  resources: { type: Array, required: true },
  rules: { type: Object },
});
const emits = defineEmits(["change:resource", "change:product", "change:meta"]);
const { products, resources, rules, type } = toRefs(props);

const { defaultCurrency } = useCurrency();

const productsArray = ref([]);
const table = ref();
const fullDate = ref({});
const selected = ref([]);
const expanded = ref([]);
const tabsIndex = ref(0);

const groups = ref(new Map());
const productId = ref("");
const groupAction = ref("");
const groupActionPayload = ref("");

const kinds = ["POSTPAID", "PREPAID"];

const tabs = computed(() => {
  return ["Products", "Resources", "Addons"];
});

const headers = computed(() =>
  [
    { text: "Key", value: "key", width: 250 },
    { text: "Title", value: "title", width: 250 },
    { text: "Price", value: "price", width: 150 },
    {
      text: "One time",
      value: "meta.oneTime",
    },
    { text: "Period", value: "period", width: 220 },
    { text: "Kind", value: "kind", width: 228 },
    { text: "Group", value: "group", width: 300 },
    { text: "Public", value: "public" },
    { text: "Sorter", value: "sorter" },
    {
      text: "Addons",
      value: "addons",
    },
  ].filter((f) => !!f)
);

const isEditOpen = ref(false);
const newMeta = ref({ description: "" });

onMounted(() => {
  setProductsArray();
  setFullDates(products.value);
  setDefaultGroups();
});

function changeDate({ value }, id) {
  fullDate.value[id] = value;
  emits("change:product", { key: "date", value, id });
}

function changeProduct(key, value, id) {
  emits("change:product", { key, value, id });
}

function changeResource(virtual, data) {
  if (data.key === "resources") {
    data.value = [
      ...resources.value.filter((r) => r.virtual !== virtual),
      ...data.value,
    ];
  }
  emits("change:resource", data);
}

function changeMeta(data, id, resources) {
  const value =
    data.key === "resources"
      ? data.value
      : JSON.parse(JSON.stringify(resources));

  if (data.key !== "resources") {
    const i = value.findIndex(({ id }) => id === data.id);

    try {
      value[i][data.key] = JSON.parse(data.value);
    } catch {
      value[i][data.key] = data.value;
    }
  }

  emits("change:meta", { key: "resources", value, id });
}

function changeOneTime(item, value) {
  if (value) {
    changeProduct("kind", "POSTPAID", item.id);
    changeProduct("period", 0, item.id);
  }
  changeProduct("meta", { ...item.meta, oneTime: value }, item.id);
}

function isOneTime(item) {
  return item.meta.oneTime;
}

const setProductsArray = () => {
  productsArray.value = Object.keys(products.value).map((key) => ({
    ...products.value[key],
    key,
  }));
};

const setDefaultGroups = () => {
  for (const key in products.value) {
    let group = products.value[key]?.group;
    if (!group) {
      changeProduct("group", "default", products.value[key].id);
      group = "default";
    }

    groups.value.set(group, group);
  }
};

const setGroupAction = (action, item) => {
  groupAction.value = action;
  groupActionPayload.value = item?.group || "";
  productId.value = item?.id || "";
};

const invokeGroupAction = (item) => {
  if (groupAction.value === "edit") {
    changeGroupName(products.value[item.key].group, groupActionPayload.value);
  } else {
    setGroup(groupActionPayload.value, item.id);
  }

  setGroupAction("");
};

const changeGroupName = (group, newGroup) => {
  for (const key in products.value) {
    if (products.value[key].group === group) {
      changeProduct("group", newGroup, products.value[key].id);
    }
  }

  groups.value.delete(group);
  groups.value.set(newGroup, newGroup);
  setProductsArray();
};

const setGroup = (group, id) => {
  changeProduct("group", group, id);
  groups.value.set(group, group);
  setProductsArray();
};

const copyProducts = () => {
  setProductsArray();
  const copiedProducts = [
    ...selected.value.map((p) => ({
      ...p,
      id: Math.random().toString(16).slice(2),
    })),
  ];

  const newProducts = {};
  for (const product of productsArray.value) {
    const key = product.key;
    delete product.key;
    newProducts[key] = product;
  }
  for (const product of copiedProducts) {
    const key = product.key;
    delete product.key;
    newProducts[key + " 1"] = product;
  }

  changeProduct("products", newProducts);
  setFullDates(newProducts);
  selected.value = [];
};

const saveNewMeta = async () => {
  setProductsArray();

  const newProducts = {};
  for (const product of productsArray.value) {
    const key = product.key;
    delete product.key;
    newProducts[key] = product;
  }

  for (const product of selected.value) {
    const key = product.key;
    delete product.key;
    newProducts[key] = {
      ...product,
      description: newMeta.value.description,
    };
  }

  changeProduct("products", newProducts);
  setFullDates(newProducts);

  selected.value = [];
  isEditOpen.value = false;
  newMeta.value = { description: "" };
};

function addConfig() {
  const result = { ...products.value };
  result[""] = {
    key: "",
    title: "",
    kind: "POSTPAID",
    price: 0,
    group: groups.value.values().next().value,
    period: 0,
    resources: {},
    public: true,
    meta: {},
    sorter: Object.keys(result).length,
    id: Math.random().toString(16).slice(2),
  };

  changeProduct("products", result);
  setProductsArray();
}

function removeConfig() {
  const value = productsArray.value.filter(
    ({ id }) => !selected.value.find((el) => el.id === id)
  );
  const result = {};

  value.forEach((product) => {
    const { key } = product;

    delete product.key;
    result[key] = product;
  });
  changeProduct("products", result);
}

const setFullDates = (products) => {
  Object.values(products).forEach(({ period, id }) => {
    fullDate.value[id] = getFullDate(period);
  });
};

watch(table, (value) => {
  const { rows } = value[0].$el.children[1].children[0];
  const allElements = Object.values(rows).slice(1);
  const height = parseInt(getComputedStyle(allElements[0]).height);

  allElements.forEach((element, i) => {
    element.id = i;
    element.draggable = true;
    element.style.cursor = "grab";
    // element.style.transition = '0.3s';

    element.addEventListener("dragstart", (e) => {
      const img = document.createElement("img");

      e.dataTransfer.dropEffect = "move";
      e.dataTransfer.effectAllowed = "move";

      e.dataTransfer.setDragImage(img, 0, 0);
      e.dataTransfer.setData("text/id", element.id);
      e.dataTransfer.setData("text/y", e.clientY);
    });

    element.addEventListener("dragover", (e) => {
      const i = +e.dataTransfer.getData("text/id");
      const initY = e.dataTransfer.getData("text/y");
      const nextIndex = Math.round((e.clientY - initY) / height) + i;

      allElements[i].style.cssText = `transform: translateY(${
        e.clientY - initY
      }px)`;
      allElements[i].setAttribute("data-i", nextIndex);
      e.preventDefault();
    });

    element.addEventListener("dragend", (e) => {
      allElements.forEach((el) => {
        const j = +e.dataTransfer.getData("text/id");
        let i = +el.getAttribute("data-i");

        if (i >= allElements.length) i = allElements.length - 1;
        if (isFinite(i) && i > -1 && i !== j) {
          const product1 = productsArray.value.find(
            (el) => el.sorter === i
          ).key;
          const product2 = productsArray.value.find(
            (el) => el.sorter === j
          ).key;

          products.value[product1].sorter = j;
          products.value[product2].sorter = i;
          [allElements[i], allElements[j]] = [allElements[j], allElements[i]];
        }

        el.removeAttribute("style");
        el.removeAttribute("data-i");

        el.style.cursor = "grab";
        // el.style.transition = '0.3s';
      });

      for (let i = 0; i < allElements.length; i++) {
        allElements[i].id = i;
      }
    });
  });
});
watch(
  products,
  () => {
    setProductsArray();
  },
  { deep: true }
);
</script>

<style lang="scss">
.mw-20 {
  max-width: 150px;
}
</style>
