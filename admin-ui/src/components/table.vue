<template>
  <components
    :is="VDataTable"
    :item-key="itemKey"
    class="elevation-0 background-light rounded-lg"
    :loading="loading"
    loading-text="Loading... Please wait"
    color="background-light"
    :items="itemsWithShortId"
    show-select
    :value="selected"
    @input="handleSelect"
    :single-select="singleSelect"
    :headers="headers"
    :sort-by.sync="sortByTable"
    :sort-desc="sortDesc"
    :expanded="expanded"
    @update:expanded="(nw) => $emit('update:expanded', nw)"
    :show-expand="showExpand"
    :page.sync="page"
    :items-per-page.sync="itemsPerPage"
  >
    <template v-if="!noHideUuid" v-slot:[`item.${itemKey}`]="props">
      <template v-if="showed.includes(props.index)">
        <v-chip color="gray">
          {{ props.value }}
        </v-chip>
        <v-btn icon @click="hideID(props.index)">
          <v-icon>mdi-close-circle-outline</v-icon>
        </v-btn>
      </template>
      <v-btn v-else icon @click="showID(props.index)">
        <v-icon>mdi-eye-outline</v-icon>
      </v-btn>
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
    expanded: {
      type: Array,
      default: () => [],
    },
    showSelect: Boolean,
    checkboxColor: String,
    showExpand: Boolean,
    showGroupBy: Boolean,
    height: [Number, String],
    hideDefaultHeader: Boolean,
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
  },
  data() {
    return {
      selected: this.value,
      showed: [],
      copyed: -1,
      VDataTable,
      page: 1,
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
  },
  computed: {
    sortByTable() {
      return this.sortBy || "title";
    },
    itemsWithShortId() {
      return this.items.map((i) => {
        return {
          ...i,
          [this.itemKey]: i[this.itemKey].slice(0, 8) + "...",
        };
      });
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
};
</script>
