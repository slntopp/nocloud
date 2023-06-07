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
          table-name="plansProducts"
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
              :rules="generalRule"
              @change="(value) => changeProduct('key', value, item.id)"
            />
          </template>
          <template v-slot:[`item.title`]="{ item }">
            <v-text-field
              dense
              :value="item.title"
              :rules="generalRule"
              @change="(value) => changeProduct('title', value, item.id)"
            />
          </template>
          <template v-slot:[`item.price`]="{ item }">
            <v-text-field
              dense
              type="number"
              :value="item.price"
              :rules="generalRule"
              @change="(value) => changeProduct('price', value, item.id)"
            />
          </template>
          <template v-slot:[`item.period`]="{ item }">
            <date-field
              :period="fullDate[item.id]"
              @changeDate="(value) => changeDate(value, item.id)"
            />
          </template>
          <template v-slot:[`item.kind`]="{ item }">
            <v-radio-group
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
          <template v-slot:expanded-item="{ headers, item }">
            <td />
            <td :colspan="headers.length - 1">
              <v-text-field
                dense
                class="pt-4"
                label="Image link"
                v-if="type === 'virtual'"
                v-model="item.meta.image"
              />
              <v-subheader class="px-0">
                {{ (type === 'virtual') ? 'Description' : 'Amount of resources' }}:
              </v-subheader>

              <vue-editor
                class="html-editor"
                v-if="type === 'virtual'"
                v-model="item.meta.description"
              />
              <json-editor
                v-else
                :json="item.resources"
                @changeValue="
                  (value) => changeProduct('amount', value, item.id)
                "
              />
            </td>
          </template>
        </nocloud-table>

        <plans-resources-table
          v-else-if="tab === 'Resources'"
          :resources="resources"
          @change:resource="changeResource"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import { computed, ref, toRefs, watch } from "vue";
import { VueEditor } from "vue2-editor";
import dateField from "@/components/date.vue";
import JsonEditor from "@/components/JsonEditor.vue";
import nocloudTable from "@/components/table.vue";
import plansResourcesTable from "@/components/plans_resources_table.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { getFullDate } from "@/functions";

const props = defineProps({
  type: { type: String, required: true },
  products: { type: Object, required: true },
  resources: { type: Array, required: true },
});
const emits = defineEmits(["change:resource", "change:product"]);
const { products, resources } = toRefs(props);

const table = ref();
const fullDate = ref({});
const selected = ref([]);
const expanded = ref([]);
const tabsIndex = ref(0);

const generalRule = [(v) => !!v || "This field is required!"];
const kinds = ["POSTPAID", "PREPAID"];
const tabs = ["Products", "Resources"];

const headers = [
  { text: "Key", value: "key" },
  { text: "Title", value: "title" },
  { text: "Price", value: "price" },
  { text: "Period", value: "period" },
  { text: "Kind", value: "kind", width: 228 },
];

function changeDate({ value }, id) {
  fullDate.value[id] = value;
  emits("change:product", { key: "date", value, id });
}

function changeProduct(key, value, id) {
  emits("change:product", { key, value, id });
}

function changeResource(data) {
  emits("change:resource", data);
}

function addConfig() {
  const value = [...productsArray.value];
  const result = {};

  value.push({
    key: "",
    title: "",
    kind: "POSTPAID",
    price: 0,
    period: 0,
    resources: {},
    meta: {},
    sorter: value.length,
    id: Math.random().toString(16).slice(2),
  });

  value.forEach((product, i) => {
    const { key } = product;

    delete product.key;
    product.sorter = i;
    result[key] = product;
  });
  changeProduct("products", result);
}

function removeConfig() {
  const value = productsArray.value.filter(
    ({ id }) => !selected.value.find((el) => el.id === id)
  );
  const result = {};

  value.forEach((product, i) => {
    const { key } = product;

    delete product.key;
    product.sorter = i;
    result[key] = product;
  });
  changeProduct("products", result);
}

Object.values(products.value).forEach(({ period, id }) => {
  fullDate.value[id] = getFullDate(period);
});

const productsArray = computed(() =>
  Object.entries(products.value).map(([key, value]) => ({ key, ...value }))
);

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
</script>

<style lang="scss">
.mw-20 {
  max-width: 150px;
}
.html-editor {
  span.ql-picker-label {
    color: white;
  }
}

.quillWrapper .ql-snow .ql-stroke {
  stroke: rgb(255 255 255 / 95%) !important;
}
.ql-snow .ql-fill {
  fill: white;
}

.quillWrapper .ql-editor {
  color: white;
}
.quillWrapper .ql-editor ul[data-checked="false"] > li::before {
  color: white !important;
}
.quillWrapper .ql-editor ul[data-checked="true"] > li::before {
  color: white !important;
}

.ql-active {
  color: #e06ffe !important;
  fill: #e06ffe !important;
  stroke: #e06ffe !important;
}
</style>
