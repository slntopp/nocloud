<template>
  <components
    :is="VDataTable"
    v-sortable-table="{
      onEnd: sortTheHeadersAndUpdateTheKey,
      disabled: !tableName,
    }"
    :item-key="itemKey"
    class="elevation-0 background-light rounded-lg"
    :loading="loading"
    loading-text="Loading... Please wait"
    color="background-light"
    :items="items"
    :show-select="showSelect"
    :value="selected"
    @input="handleSelect"
    :hide-default-footer="hideDefaultFooter"
    :single-select="singleSelect"
    :headers="filtredHeaders"
    :sort-by="tableSortBy"
    :sort-desc="tableSortDesc"
    :expanded="expanded"
    @update:expanded="(nw) => $emit('update:expanded', nw)"
    :show-expand="showExpand"
    :page="serverSidePage"
    @update:page="serverSidePage ? () => 1 : (page = $event)"
    :items-per-page.sync="itemsPerPage"
    :show-group-by="showGroupBy"
    :group-by="groupBy"
    :custom-sort="customSort"
    :key="anIncreasingNumber"
    :server-items-length="serverItemsLength"
    :options="options"
    :footer-props="{ 'items-per-page-options': itemsPerPageOptions }"
    @update:options="updateOptions"
    @update:items-per-page="saveItemsPerPage"
  >
    <template v-if="!noHideUuid" v-slot:[`item.${itemKey}`]="props">
      <template v-if="showed.includes(props.index)">
        <v-chip v-if="isKeyInCircle" color="gray">
          {{ props.value }}
        </v-chip>
        <template v-else>
          {{ props.value }}
        </template>
        <v-btn icon @click="hideID(props.index)">
          <v-icon>mdi-close-circle-outline</v-icon>
        </v-btn>
      </template>
      <template v-else>
        <template v-if="!isKeyOnlyAfterClick">
          <v-chip v-if="isKeyInCircle" color="gray">
            {{ makeIdShort(props.value) }}
          </v-chip>
          <template v-else>
            {{ makeIdShort(props.value) }}
          </template>
        </template>
        <v-btn
          v-if="!isIdShort(props.value) || isKeyOnlyAfterClick"
          icon
          @click="showID(props.index)"
        >
          <v-icon>mdi-eye-outline</v-icon>
        </v-btn>
      </template>

      <v-btn icon @click="addToClipboard(props.value, props.index)">
        <v-icon v-if="copyed == props.index"> mdi-check </v-icon>
        <v-icon v-else> mdi-content-copy </v-icon>
      </v-btn>

      <slot name="uuid-actions" v-bind="{ item: props.item }" />
    </template>

    <template v-slot:[`item.titleLink`]="{ item }">
      <router-link :to="item.route">
        {{ item.titleLink }}
      </router-link>
    </template>

    <template
      v-for="(_, scopedSlotName) in $scopedSlots"
      v-slot:[scopedSlotName]="slotData"
    >
      <slot :name="scopedSlotName" v-bind="slotData" />
    </template>
    <template v-for="(_, slotName) in $slots" v-slot:[slotName]>
      <slot :name="slotName" />
    </template>

    <template
      v-for="header in filtredHeaders"
      v-slot:[`header.${header?.value}`]
    >
      {{ header.text }}
    </template>

    <template v-if="!showSelect" v-slot:[`item.data-table-select`]>
      <div></div>
    </template>

    <template v-slot:footer>
      <v-toolbar
        v-if="footerError.length > 0"
        class="mt-2"
        color="error"
        dark
        flat
      >
        <v-toolbar-title class="subheading">
          {{ footerError }}
        </v-toolbar-title>
      </v-toolbar>
    </template>

    <template v-slot:body.append v-if="isEdit">
      <tr>
        <td />
        <td v-for="(header, index) in filtredHeaders" :key="index">
          <component
            v-if="header.editable"
            :is="getEditComponent(header)"
            v-bind="getEditComponentProps(header)"
            @input="changeEditableValue(header, $event)"
            :value="editableValues[header.value]"
          />
        </td>
      </tr>
    </template>

    <template v-slot:[`footer.prepend`] v-if="editable">
      <div
        class="d-flex justify-end align-center"
        style="min-width: calc(100% - 450px)"
      >
        <v-btn
          :disabled="value.length < 1 && !isEdit"
          class="mx-1"
          @click="isEdit = !isEdit"
          >{{ isEdit ? "Cancel" : "Edit" }}</v-btn
        >
        <v-btn
          :disabled="value.length < 1"
          @click="saveEditableValues"
          class="mx-1"
          v-if="isEdit"
          >Save</v-btn
        >
      </div>
    </template>

    <template
      v-slot:[`footer.page-text`]="{ pageStart, pageStop, itemsLength }"
    >
      <div class="d-flex align-center">
        <div v-if="tableName">
          <v-dialog
            @click:outside="changeFiltres"
            max-width="60%"
            v-model="settingsDialog"
            width="500"
          >
            <template v-slot:activator="{ on, attrs }">
              <v-icon v-bind="attrs" v-on="on" size="23" class="mr-1"
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
        <span class="ml-3">
          {{ pageStart }}-{{ pageStop }} of {{ itemsLength }}
        </span>
      </div>
    </template>
  </components>
</template>

<script>
import { VDataTable } from "vuetify/lib";
import Sortable from "sortablejs";
import { VSelect, VTextField } from "vuetify/lib";
import DatePicker from "@/components/ui/datePicker";
import { formatSecondsToDateString } from "@/functions";

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
    value: {
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
    itemsPerPageOptions: { type: Array, default: () => [5, 10, 15, 25, 50] },
  },
  data() {
    return {
      selected: [],
      showed: [],
      copyed: -1,
      VDataTable,
      page: 1,
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
  methods: {
    handleSelect(item) {
      this.$emit("input", item);
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
          this.filtredHeaders[oldIndex]
        );
      }

      this.filtredHeaders.splice(
        oldIndex > newIndex ? oldIndex + 1 : oldIndex,
        1
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
        localStorage.getItem("itemsPerPage")
      );

      if (!itemsPerPageSettings) {
        itemsPerPageSettings = { [this.tableName]: val };
      } else {
        itemsPerPageSettings[this.tableName] = val;
      }

      localStorage.setItem(
        "itemsPerPage",
        JSON.stringify(itemsPerPageSettings)
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
    updateOptions(options) {
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
      const baseProps = { ...header.editable, dense: true };
      switch (header.editable.type) {
        case "date": {
          if (!this.editableValues[header.value]) {
            this.changeEditableValue(
              header,
              formatSecondsToDateString(Date.now() / 1000)
            );
          }
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
  beforeDestroy() {
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
    page(value) {
      localStorage.setItem("page", value);
      localStorage.setItem("url", this.$route.path);
    },
    value() {
      this.selected = this.value;
    },
  },
  directives: {
    "sortable-table": {
      inserted: (el, binding) => {
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
            binding.value ? { ...binding.value, handle: ".sortHandle" } : {}
          );
        }, 0);
      },
    },
  },
};
</script>

<style>
.v-data-table > .v-data-table__wrapper > table > tbody > tr > td,
.theme--dark.v-data-table
  > .v-data-table__wrapper
  > table
  > thead
  > tr:last-child
  > th {
  white-space: nowrap;
}

.sortable-drag {
  color: salmon;
  background-color: var(--v-background-light-base);
}
</style>
