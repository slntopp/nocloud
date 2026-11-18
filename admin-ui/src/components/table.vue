<template>
  <component
    :is="tableComponent"
    v-sortable-table="{
      onEnd: sortTheHeadersAndUpdateTheKey,
      disabled: !tableName,
    }"
    :item-value="itemKey"
    class="elevation-0 bg-background-light rounded-lg"
    :loading="loading"
    loading-text="Loading... Please wait"
    :items="tableItems"
    v-bind="serverProps"
    :show-select="showSelect"
    :select-strategy="singleSelect ? 'single' : 'page'"
    return-object
    :model-value="selected"
    @update:model-value="handleSelect"
    :hide-default-footer="hideDefaultFooter"
    :headers="tableHeaders"
    :sort-by="tableSortByV3"
    @update:sort-by="onSortByUpdate"
    :custom-key-sort="customKeySort"
    :expanded="expanded"
    @update:expanded="(nw) => $emit('update:expanded', nw)"
    :show-expand="showExpand"
    v-model:items-per-page="itemsPerPage"
    :group-by="groupByV3"
    :key="anIncreasingNumber"
    :items-per-page-options="itemsPerPageOptions"
    @update:options="updateOptions"
    @update:items-per-page="saveItemsPerPage"
  >
    <!--
      This cell is rendered for every row, so it uses plain markup with
      Vuetify's chip/button classes instead of v-chip/v-btn components:
      same look, much cheaper to render for long tables.
    -->
    <template v-if="!noHideUuid" v-slot:[`item.${itemKey}`]="props">
      <template v-if="showed.includes(props.index)">
        <span v-if="isKeyInCircle" :class="lightChipClass">
          <span class="v-chip__underlay"></span>
          <div class="v-chip__content">{{ props.value }}</div>
        </span>
        <template v-else>
          {{ props.value }}
        </template>
        <button
          type="button"
          :class="lightIconBtnClass"
          @click="hideID(props.index)"
        >
          <span class="v-btn__overlay"></span>
          <span class="v-btn__underlay"></span>
          <span class="v-btn__content">
            <i :class="lightIconClass('mdi-close-circle-outline')"></i>
          </span>
        </button>
      </template>
      <template v-else>
        <template v-if="!isKeyOnlyAfterClick">
          <span v-if="isKeyInCircle" :class="lightChipClass">
            <span class="v-chip__underlay"></span>
            <div class="v-chip__content">{{ makeIdShort(props.value) }}</div>
          </span>
          <template v-else>
            {{ makeIdShort(props.value) }}
          </template>
        </template>
        <button
          v-if="!isIdShort(props.value) || isKeyOnlyAfterClick"
          type="button"
          :class="lightIconBtnClass"
          @click="showID(props.index)"
        >
          <span class="v-btn__overlay"></span>
          <span class="v-btn__underlay"></span>
          <span class="v-btn__content">
            <i :class="lightIconClass('mdi-eye-outline')"></i>
          </span>
        </button>
      </template>

      <button
        type="button"
        :class="lightIconBtnClass"
        @click="addToClipboard(props.value, props.index)"
      >
        <span class="v-btn__overlay"></span>
        <span class="v-btn__underlay"></span>
        <span class="v-btn__content">
          <i
            :class="
              lightIconClass(
                copyed == props.index ? 'mdi-check' : 'mdi-content-copy',
              )
            "
          ></i>
        </span>
      </button>

      <slot name="uuid-actions" v-bind="{ item: props.item }" />
    </template>

    <!-- row checkbox as plain markup (a VCheckboxBtn per row is expensive) -->
    <template v-if="showSelect" v-slot:[`item.data-table-select`]="{ props }">
      <button
        type="button"
        :class="[lightIconBtnClass, { 'text-primary': props.modelValue }]"
        :disabled="props.disabled"
        :aria-checked="props.modelValue"
        role="checkbox"
        @click="props.onClick"
      >
        <span class="v-btn__overlay"></span>
        <span class="v-btn__underlay"></span>
        <span class="v-btn__content">
          <i
            :class="
              lightIconClass(
                props.modelValue
                  ? 'mdi-checkbox-marked'
                  : 'mdi-checkbox-blank-outline',
              )
            "
          ></i>
        </span>
      </button>
    </template>

    <template v-slot:[`item.titleLink`]="{ item }">
      <router-link :to="item.route">
        {{ item.titleLink }}
      </router-link>
    </template>

    <!-- pass consumer slots through; Vuetify 2 slot names/props are mapped -->
    <template
      v-for="(_, slotName) in passThroughSlots"
      v-slot:[slotName]="slotData"
    >
      <slot :name="slotName" v-bind="slotData || {}" />
    </template>

    <template v-if="$slots['expanded-item']" v-slot:expanded-row="slotData">
      <slot
        name="expanded-item"
        v-bind="{ ...slotData, headers: slotData.columns }"
      />
    </template>

    <template v-slot:[`body.append`] v-if="isEdit || $slots['body.append']">
      <slot name="body.append" />
      <tr v-if="isEdit">
        <td />
        <td
          :style="header.editable ? 'min-width: 200px;' : ''"
          v-for="(header, index) in filtredHeaders"
          :key="index"
        >
          <component
            v-if="header.editable"
            :is="getEditComponent(header)"
            @update:model-value="changeEditableValue(header, $event)"
            :model-value="editableValues[header.value]"
            v-bind="getEditComponentProps(header)"
          />
        </td>
      </tr>
    </template>

    <template v-slot:[`footer.prepend`]>
      <v-toolbar
        v-if="footerError.length > 0"
        class="mr-2"
        color="error"
        density="compact"
        flat
      >
        <v-toolbar-title class="text-subtitle-1">
          {{ footerError }}
        </v-toolbar-title>
      </v-toolbar>

      <div
        v-if="editable"
        class="d-flex justify-end align-center"
        style="min-width: calc(100% - 450px)"
      >
        <v-btn
          :disabled="modelValue.length < 1 && !isEdit"
          class="mx-1"
          @click="isEdit = !isEdit"
          >{{ isEdit ? "Cancel" : "Edit" }}</v-btn
        >
        <v-btn
          :disabled="modelValue.length < 1"
          @click="saveEditableValues"
          class="mx-1"
          v-if="isEdit"
          >Save</v-btn
        >
      </div>

      <v-spacer />

      <div v-if="tableName">
        <v-dialog
          @click:outside="changeFiltres"
          max-width="60%"
          v-model="settingsDialog"
          width="500"
        >
          <template v-slot:activator="{ props }">
            <v-icon v-bind="props" size="23" class="mr-1"
              >mdi-cog-outline</v-icon
            >
          </template>
          <v-card
            color="background-light"
            style="overflow: hidden"
            max-width="100%"
          >
            <v-card-title>Table settings</v-card-title>
            <v-row class="pa-5">
              <v-col
                v-for="header in headers.filter((h) => h.text)"
                :key="header.value"
                cols="4"
              >
                <v-checkbox
                  @click.stop
                  :label="header.text"
                  v-model="filter"
                  :value="header.value"
                />
              </v-col>
            </v-row>
          </v-card>
        </v-dialog>
      </div>
    </template>
  </component>
</template>

<script>
import Sortable from "sortablejs";
import {
  VBtn,
  VChip,
  VDataTable,
  VDataTableServer,
  VIcon,
  VSelect,
  VTextField,
} from "vuetify/components";
import DatePicker from "@/components/ui/dateTimePicker.vue";

function watchClass(targetNode, classToWatch) {
  let lastClassState = targetNode.classList.contains(classToWatch);
  const observer = new MutationObserver((mutationsList) => {
    for (let i = 0; i < mutationsList.length; i++) {
      const mutation = mutationsList[i];
      if (
        mutation.type === "attributes" &&
        mutation.attributeName === "class"
      ) {
        const currentClassState =
          mutation.target.classList.contains(classToWatch);
        if (lastClassState !== currentClassState) {
          lastClassState = currentClassState;
          if (!currentClassState) {
            mutation.target.classList.add("sortHandle");
          }
        }
      }
    }
  });
  observer.observe(targetNode, { attributes: true });
}

const defaultHeaders = [
  { text: "title", value: "title" },
  {
    text: "UUID",
    align: "start",
    sortable: true,
    value: "uuid",
  },
];

export default {
  name: "nocloud-table",
  props: {
    sortBy: { type: String },
    sortDesc: { type: Boolean },
    customSort: { type: Function },
    loading: Boolean,
    editable: Boolean,
    items: {
      type: Array,
      default: () => [],
    },
    tableName: {
      type: String,
      default: "",
    },
    modelValue: {
      type: Array,
      default: () => [],
    },
    headers: {
      type: Array,
      default: () => defaultHeaders,
    },
    singleSelect: {
      type: Boolean,
      default: false,
    },
    itemKey: {
      type: String,
      default: "uuid",
    },
    noHideUuid: {
      type: Boolean,
      default: false,
    },
    serverItemsLength: Number,
    options: Object,
    expanded: {
      type: Array,
      default: () => [],
    },
    showSelect: {
      type: Boolean,
      default: true,
    },
    checkboxColor: String,
    showExpand: Boolean,
    showGroupBy: Boolean,
    height: [Number, String],
    hideDefaultHeader: Boolean,
    hideDefaultFooter: Boolean,
    groupBy: String,
    caption: String,
    dense: Boolean,
    headerProps: Object,
    calculateWidths: Boolean,
    fixedHeader: Boolean,
    headersLength: Number,
    expandIcon: {
      type: String,
      default: "$expand",
    },
    itemClass: {
      type: [String, Function],
      default: () => "",
    },
    loaderHeight: {
      type: [Number, String],
      default: 4,
    },
    footerError: {
      type: String,
      default: "",
    },
    isKeyInCircle: {
      type: Boolean,
      default: true,
    },
    isKeyOnlyAfterClick: {
      type: Boolean,
      default: false,
    },
    serverSidePage: Number,
    itemsPerPageOptions: {
      type: Array,
      default: () => [10, 25, 50, 100, 500, 1000],
    },
  },
  // registered so their styles are loaded for the plain-markup uuid cell
  components: { VBtn, VChip, VIcon },
  emits: [
    "update:modelValue",
    "update:expanded",
    "update:options",
    "update:edit-values",
  ],
  data() {
    return {
      selected: [],
      showed: [],
      copyed: -1,
      page: 1,
      serverPage: this.serverSidePage || 1,
      anIncreasingNumber: 0,
      itemsPerPage: 10,
      columns: {},
      filter: [],
      filtredHeaders: [],
      tableSortBy: "title",
      tableSortDesc: false,
      settingsDialog: false,
      isEdit: false,
      editableValues: {},
    };
  },
  computed: {
    lightChipClass: () =>
      "v-chip text-gray v-chip--density-default v-chip--size-default v-chip--variant-tonal",
    lightIconBtnClass: () =>
      "v-btn v-btn--icon v-btn--density-default v-btn--size-default v-btn--variant-text",
    // server-side pagination needs VDataTableServer in Vuetify 3
    tableComponent() {
      return this.serverItemsLength !== undefined &&
        this.serverItemsLength !== null
        ? VDataTableServer
        : VDataTable;
    },
    // Vuetify 3 headers use title/key; callers still pass Vuetify 2 text/value
    // Server tables: the page is controlled here and reported to the parent
    // via update:options (it refetches and passes serverSidePage back).
    // Client tables keep Vuetify's own page state, as in Vuetify 2, where the
    // page stored in localStorage was never applied to them.
    serverProps() {
      return this.tableComponent === VDataTableServer
        ? {
            // a missing count used to show "0-0 of 0" in Vuetify 2, not NaN
            itemsLength: Number.isFinite(+this.serverItemsLength)
              ? +this.serverItemsLength
              : 0,
            page: this.serverPage,
            "onUpdate:page": this.onServerPageUpdate,
          }
        : { "onUpdate:page": (page) => (this.page = page) };
    },
    tableHeaders() {
      const headers = this.filtredHeaders.filter(Boolean).map((header) => ({
        ...header,
        title: header.title ?? header.text,
        key: header.key ?? header.value,
      }));
      // Vuetify 3+ appends the expand column at the end of the row; Vuetify 2
      // put it first, followed by the checkbox. Keep the old order.
      const leading = [];
      if (this.showExpand) leading.push({ key: "data-table-expand" });
      if (this.showSelect) leading.push({ key: "data-table-select" });
      return [...leading, ...headers];
    },
    tableSortByV3() {
      return this.tableSortBy
        ? [{ key: this.tableSortBy, order: this.tableSortDesc ? "desc" : "asc" }]
        : [];
    },
    groupByV3() {
      return this.groupBy ? [{ key: this.groupBy, order: "asc" }] : [];
    },
    // Vuetify 3 has no whole-table custom sort: sort here and neutralise
    // the built-in key comparison so the order is kept
    tableItems() {
      if (!this.customSort || this.tableComponent === VDataTableServer) {
        return this.items;
      }
      return this.customSort(
        [...this.items],
        this.tableSortBy ? [this.tableSortBy] : [],
        [this.tableSortDesc],
      );
    },
    customKeySort() {
      if (!this.customSort) return undefined;
      return Object.fromEntries(
        this.tableHeaders.map((header) => [header.key, () => 0]),
      );
    },
    passThroughSlots() {
      // expanded-item is mapped to Vuetify 3's expanded-row separately,
      // body.append is merged with the inline edit row
      const skip = ["expanded-item", "body.append"];
      return Object.fromEntries(
        Object.entries(this.$slots).filter(([name]) => !skip.includes(name)),
      );
    },
  },
  methods: {
    lightIconClass(icon) {
      return `${icon} mdi v-icon notranslate v-icon--size-default`;
    },
    handleSelect(items) {
      this.selected = items;
      this.$emit("update:modelValue", items);
    },
    onSortByUpdate(sortBy) {
      this.tableSortBy = sortBy[0]?.key ?? "";
      this.tableSortDesc = sortBy[0]?.order === "desc";
    },
    onServerPageUpdate(page) {
      this.serverPage = page;
      this.page = page;
    },
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((res) => {
            console.error(res);
          });
      } else {
        alert("Clipboard is not supported!");
      }
    },
    showID(index) {
      this.showed.push(index);
    },
    hideID(index) {
      this.showed = this.showed?.filter((i) => i !== index);
    },
    isIdShort(id) {
      return id?.length <= 8;
    },
    makeIdShort(id) {
      if (this.isIdShort(id)) {
        return id;
      }
      return id?.slice(0, 8) + "...";
    },
    saveColumnPosition(headers) {
      if (!headers) {
        return;
      }

      headers.forEach(({ value }, index) => {
        this.columns[value] = index;
      });

      const columnJson = localStorage.getItem("columns");
      const allColumnsSetting = columnJson
        ? JSON.parse(columnJson)
        : { [this.tableName]: {} };
      allColumnsSetting[this.tableName] = this.columns;

      localStorage.setItem("columns", JSON.stringify(allColumnsSetting));
    },
    sortTheHeadersAndUpdateTheKey(evt) {
      const originalHeaders = JSON.parse(JSON.stringify(this.filtredHeaders));
      this.filtredHeaders = [];
      let oldIndex = evt.oldIndex;
      let newIndex = evt.newIndex;

      if (this.showExpand) {
        oldIndex--;
        newIndex--;
      }

      if (this.showSelect) {
        oldIndex--;
        newIndex--;
      }

      for (const header of originalHeaders) {
        if (header) {
          this.filtredHeaders.push(header);
        }
      }

      if (newIndex === 0) {
        this.filtredHeaders = [
          this.filtredHeaders[oldIndex],
          ...this.filtredHeaders,
        ];
      } else {
        this.filtredHeaders.splice(
          oldIndex > newIndex ? newIndex : newIndex + 1,
          0,
          this.filtredHeaders[oldIndex],
        );
      }

      this.filtredHeaders.splice(
        oldIndex > newIndex ? oldIndex + 1 : oldIndex,
        1,
      );

      this.table = this.filtredHeaders;
      this.anIncreasingNumber += 1;
    },
    setHeadersBy(columns) {
      const tempHeaders = [];
      const originalHeaders = JSON.parse(JSON.stringify(this.headers));
      for (const [key, value] of Object.entries(columns)) {
        const el = originalHeaders.find((h) => h.value === key);
        let index = value;
        if (tempHeaders[value]) {
          index++;
        }
        if (el) {
          tempHeaders[index] = el;
        }
      }
      if (
        Object.keys(tempHeaders).length ===
        Object.keys(this.filtredHeaders).length
      ) {
        let isSame = true;
        for (const key of Object.keys(tempHeaders)) {
          if (!isSame) {
            continue;
          }
          isSame = tempHeaders[key]?.value === this.filtredHeaders[key]?.value;
        }
        if (isSame) {
          return;
        }
      }

      this.filtredHeaders = tempHeaders.filter((h) => !!h);

      this.table = tempHeaders;
      this.anIncreasingNumber += 1;
    },
    setFilterBy(columns) {
      Object.keys(columns).forEach((col) => {
        this.filter.push(col);
      });
    },
    changeFiltres() {
      const newColumns = {};
      for (const [key, value] of Object.entries(this.columns)) {
        const col = this?.filter.find((f) => f === key);
        if (col) {
          newColumns[key] = value;
        }
      }
      //add new columns
      const newColumnsKeys = Object.keys(newColumns);
      this.filter
        ?.filter((f) => newColumnsKeys.findIndex((nc) => nc === f) === -1)
        .forEach((key, index) => {
          newColumns[key] = newColumnsKeys.length + index;
        });
      this.setHeadersBy(newColumns);
      this.columns = newColumns;
      this.settingsDialog = false;
    },
    setDefaultHeaders() {
      this.filtredHeaders.forEach(({ value }, index) => {
        this.columns[value] = index;
      });
    },
    setDefaultFiltres() {
      if (this.defaultFiltres && this.defaultFiltres.length) {
        this.defaultFiltres?.forEach((value) => this?.filter.push(value));
      } else {
        this.headers.forEach((h) => {
          this.filter?.push(h.value);
        });
      }
    },
    saveTableData() {
      this.saveColumnPosition(this.filtredHeaders);

      const url = localStorage.getItem("url");

      if (this.$route.path.includes(url)) return;
      localStorage.removeItem("page");
    },
    saveItemsPerPage(val) {
      let itemsPerPageSettings = JSON.parse(
        localStorage.getItem("itemsPerPage"),
      );

      if (!itemsPerPageSettings) {
        itemsPerPageSettings = { [this.tableName]: val };
      } else {
        itemsPerPageSettings[this.tableName] = val;
      }

      localStorage.setItem(
        "itemsPerPage",
        JSON.stringify(itemsPerPageSettings),
      );
    },
    configureColumns() {
      if (this.tableName) {
        const columnsString = localStorage.getItem("columns");
        if (columnsString) {
          this.columns = JSON.parse(columnsString)?.[this.tableName];
        }
        if (Object.keys(this.columns || {}).length === 0) {
          this.columns = {};
          this.setDefaultHeaders();
          this.setDefaultFiltres();
        } else {
          this.setHeadersBy(this.columns);
          this.setFilterBy(this.columns);
        }
      }
    },
    configureItemsPerPage() {
      const storageData = localStorage.getItem("itemsPerPage");
      if (storageData) {
        const itemsPerPage = JSON.parse(storageData);
        this.itemsPerPage = +itemsPerPage[this.tableName] || 15;
      }
    },
    updateOptions(v3Options) {
      const options = {
        page: v3Options.page,
        itemsPerPage: v3Options.itemsPerPage,
        sortBy: v3Options.sortBy.map((s) => s.key),
        sortDesc: v3Options.sortBy.map((s) => s.order === "desc"),
        groupBy: v3Options.groupBy.map((g) => g.key),
        groupDesc: v3Options.groupBy.map((g) => g.order === "desc"),
      };
      if (this.tableName) {
        this.saveSortBy(options);
      }
      this.$emit("update:options", options);
    },
    saveSortBy(options) {
      const sorts = JSON.parse(localStorage.getItem("sorts") || "{}");
      if (!options.sortBy[0] && sorts[this.tableName]) {
        const [sortBy] = sorts[this.tableName].split("$");
        this.tableSortBy = sortBy;
        this.tableSortDesc = false;
        options.sortBy.push(sortBy);
        options.sortDesc.push(false);
      }

      sorts[this.tableName] = !options.sortBy[0]
        ? undefined
        : options.sortBy[0] + "$" + options.sortDesc[0];
      localStorage.setItem("sorts", JSON.stringify(sorts));
    },
    configureSortBy() {
      const sorts = JSON.parse(localStorage.getItem("sorts") || "{}");
      if (sorts[this.tableName]) {
        const [sortBy, sortDesc] = sorts[this.tableName].split("$");
        this.tableSortBy = sortBy;
        this.tableSortDesc = sortDesc === "true";
      } else if (this.sortBy) {
        this.tableSortBy = this.sortBy;
        this.tableSortDesc = !!this.sortDesc;
      }
    },
    getEditComponent(header) {
      switch (header.editable.type) {
        case "date":
          return DatePicker;
        case "logic-select":
          return VSelect;
        default:
          return VTextField;
      }
    },
    getEditComponentProps(header) {
      const baseProps = { ...header.editable, density: "compact" };
      switch (header.editable.type) {
        case "date": {
          return {
            ...baseProps,
            clearable: false,
          };
        }

        case "logic-select": {
          if (!this.editableValues[header.value]) {
            this.changeEditableValue(header, "False");
          }

          return {
            ...baseProps,
            items: ["True", "False"],
          };
        }

        default:
          return baseProps;
      }
    },
    changeEditableValue(header, value) {
      this.editableValues = { ...this.editableValues, [header.value]: value };
    },
    saveEditableValues() {
      this.isEdit = false;
      this.$emit("update:edit-values", this.editableValues);
      this.editableValues = {};
    },
  },
  beforeUnmount() {
    this.saveTableData();
    window.removeEventListener("beforeunload", this.saveTableData);
  },
  created() {
    this.configureItemsPerPage();
    this.configureSortBy();
  },
  mounted() {
    window.addEventListener("beforeunload", this.saveTableData);
    this.filtredHeaders = this.headers;
    const page = localStorage.getItem("page");
    if (page)
      setTimeout(() => {
        this.page = +page;
      }, 100);

    this.configureColumns();
  },
  watch: {
    serverSidePage(value) {
      if (value) this.serverPage = value;
    },
    page(value) {
      localStorage.setItem("page", value);
      localStorage.setItem("url", this.$route.path);
    },
    modelValue: {
      handler(value) {
        this.selected = value;
      },
      immediate: true,
    },
  },
  directives: {
    "sortable-table": {
      mounted: (el, binding) => {
        if (binding.value.disabled) {
          return;
        }

        setTimeout(() => {
          el.querySelectorAll("th").forEach((draggableEl) => {
            // Need a class watcher because sorting v-data-table rows asc/desc removes the sortHandle class
            watchClass(draggableEl, "sortHandle");
            draggableEl.classList.add("sortHandle");
          });
          Sortable.create(
            el.querySelector("tr"),
            binding.value ? { ...binding.value, handle: ".sortHandle" } : {},
          );
        }, 0);
      },
    },
  },
};
</script>

<style>
.v-data-table .v-table__wrapper > table > tbody > tr > td,
.v-data-table .v-table__wrapper > table > thead > tr:last-child > th {
  white-space: nowrap;
}

.sortable-drag {
  color: salmon;
  background-color: rgb(var(--v-theme-background-light));
}
</style>
