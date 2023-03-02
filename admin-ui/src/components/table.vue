<template>
  <components
    :is="VDataTable"
    v-sortable-table="{ onEnd: sortTheHeadersAndUpdateTheKey }"
    :item-key="itemKey"
    class="elevation-0 background-light rounded-lg"
    :loading="loading"
    loading-text="Loading... Please wait"
    color="background-light"
    :items="items"
    :show-select="showSelect"
    :value="selected"
    @input="handleSelect"
    :single-select="singleSelect"
    :headers="headers"
    :sort-by="sortByTable"
    :sort-desc="sortDesc"
    :expanded="expanded"
    @update:expanded="(nw) => $emit('update:expanded', nw)"
    :show-expand="showExpand"
    :page="serverSidePage"
    @update:page="serverSidePage ? _ : (page = $event)"
    :items-per-page.sync="itemsPerPage"
    :show-group-by="showGroupBy"
    :group-by="groupBy"
    :custom-sort="customSort"
    :key="anIncreasingNumber"
    :server-items-length="serverItemsLength"
    :options="options"
    @update:options="$emit('update:options', $event)"
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

    <template v-if="footerError.length > 0" v-slot:footer>
      <v-toolbar class="mt-2" color="error" dark flat>
        <v-toolbar-title class="subheading">
          {{ footerError }}
        </v-toolbar-title>
      </v-toolbar>
    </template>
  </components>
</template>

<script>
import { VDataTable } from "vuetify/lib";
import Sortable from "sortablejs";

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
    items: {
      type: Array,
      default: () => [],
    },
    value: {
      type: Array,
      default: () => [],
    },
    headers: {
      type: Array,
      default: () => defaultHeaders,
    },
    "single-select": {
      type: Boolean,
      default: false,
    },
    "item-key": {
      type: String,
      default: "uuid",
    },
    "no-hide-uuid": {
      type: Boolean,
      default: false,
    },
    "server-items-length": Number,
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
    "footer-error": {
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
  },
  data() {
    return {
      selected: this.value,
      showed: [],
      copyed: -1,
      VDataTable,
      page: 1,
      anIncreasingNumber: 0,
      itemsPerPage: 10,
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
      this.showed = this.showed.filter((i) => i !== index);
    },
    isIdShort(id) {
      return id.length <= 8;
    },
    makeIdShort(id) {
      if (this.isIdShort(id)) {
        return id;
      }
      return id.slice(0, 8) + "...";
    },
    sortTheHeadersAndUpdateTheKey(evt) {
      const headersTmp = this.headers;
      const oldIndex = evt.oldIndex - 1;
      const newIndex = evt.newIndex - 1;
      headersTmp.splice(newIndex, 0, headersTmp.splice(oldIndex, 1)[0]);
      this.table = headersTmp;
      this.anIncreasingNumber += 1;
    },
  },
  computed: {
    sortByTable() {
      return this.sortBy || "title";
    },
  },
  watch: {
    page(value) {
      localStorage.setItem("page", value);
      localStorage.setItem("url", this.$route.path);
    },
    itemsPerPage(value) {
      localStorage.setItem("itemsPerPage", value);
      localStorage.setItem("url", this.$route.path);
    },
  },
  mounted() {
    const page = localStorage.getItem("page");
    const items = localStorage.getItem("itemsPerPage");
    if (items) this.itemsPerPage = +items;
    if (page)
      setTimeout(() => {
        this.page = +page;
      }, 100);
  },
  destroyed() {
    const url = localStorage.getItem("url");

    if (this.$route.path.includes(url)) return;
    localStorage.removeItem("page");
    localStorage.removeItem("itemsPerPage");
  },
  directives: {
    "sortable-table": {
      inserted: (el, binding) => {
        el.querySelectorAll("th").forEach((draggableEl) => {
          // Need a class watcher because sorting v-data-table rows asc/desc removes the sortHandle class
          watchClass(draggableEl, "sortHandle");
          draggableEl.classList.add("sortHandle");
        });
        Sortable.create(
          el.querySelector("tr"),
          binding.value ? { ...binding.value, handle: ".sortHandle" } : {}
        );
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
  background-color: rgb(13, 16, 60);
}
</style>
