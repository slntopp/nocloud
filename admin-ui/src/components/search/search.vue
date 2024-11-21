<template>
  <div class="search_container">
    <v-menu
      @input="setLayoutMode('preview')"
      content-class="search"
      :close-on-content-click="false"
      v-model="isOpen"
      offset-y
    >
      <template v-slot:activator="{ attrs, on }">
        <v-text-field
          hide-details
          placeholder="Search..."
          single-line
          background-color="background-light"
          dence
          rounded
          class="search__input"
          @keydown.enter="hideSearch"
          v-model="param"
        >
          <template v-slot:append>
            <div class="d-flex" style="margin-top: 2px">
              <v-btn v-if="!isResetAllHide" icon small @click="resetAll">
                <v-icon size="23">mdi-close</v-icon>
              </v-btn>
              <v-btn
                v-if="searchName"
                icon
                small
                @click="isOpen ? hideSearch() : showSearch()"
              >
                <v-icon size="30">
                  {{ !isOpen ? "mdi-chevron-down" : "mdi-chevron-up" }}
                </v-icon>
              </v-btn>
            </div>
          </template>
          <template v-slot:prepend-inner>
            <v-chip
              v-bind="searchName ? attrs : undefined"
              v-on="searchName ? on : undefined"
              class="px-2"
              small
              outlined
              color="primary"
              v-if="currentLayout"
              >{{ currentLayout.title }}
              <v-btn icon x-small @click.stop="setCurrentLayout('')">
                <v-icon small>mdi-close</v-icon>
              </v-btn>
            </v-chip>
            <template v-else>
              <filter-tags @click="isOpen = true" />
            </template>
          </template>
        </v-text-field>
      </template>
      <v-card
        style="height: 100%"
        class="pa-3 search__card"
        color="background-light"
      >
        <v-row style="height: 100%" class="search__content">
          <v-col
            cols="3"
            :class="{
              layouts: true,
              dark: theme === 'dark',
              light: theme !== 'dark',
            }"
          >
            <v-list dense color="background-light">
              <v-subheader>LAYOUTS</v-subheader>
              <v-card
                style="box-shadow: none"
                :class="{ 'pa-3': isLayoutModePreview }"
                v-for="layout in layouts"
                :key="layout.id"
                :disabled="isLayoutModeAdd"
                @click="setCurrentLayout(layout)"
              >
                <v-list-item-content class="ma-0 pa-0">
                  <v-list-item-title
                    :style="{
                      color:
                        layout.id === currentLayout?.id
                          ? 'var(--v-primary-base) !important'
                          : undefined,
                    }"
                    v-if="isLayoutModePreview || isLayoutModeAdd"
                  >
                    <div class="d-flex justify-space-between">
                      <span>
                        {{ layout.title }}
                      </span>
                      <v-icon v-if="isPinned(layout)" small color="primary"
                        >mdi-pin</v-icon
                      >
                    </div>
                  </v-list-item-title>
                  <v-text-field
                    v-else-if="isLayoutModeEdit"
                    v-model="layout.title"
                    dense
                    class="pa-0 ma-0"
                  >
                    <template v-slot:append>
                      <v-btn
                        @click="setPinned(layout.id)"
                        small
                        icon
                        :color="isPinned(layout) ? 'primary' : undefined"
                      >
                        <v-icon small>{{
                          isPinned(layout) ? "mdi-pin" : "mdi-pin-off"
                        }}</v-icon>
                      </v-btn>
                      <v-btn @click="deleteLayout(layout.id)" small icon>
                        <v-icon small>mdi-close</v-icon>
                      </v-btn>
                    </template>
                  </v-text-field>
                </v-list-item-content>
              </v-card>

              <v-card v-if="isLayoutModeAdd" style="box-shadow: none">
                <v-list-item-content>
                  <v-form ref="addNewLayoutForm" v-model="isNewLayoutValid">
                    <v-text-field
                      :rules="newLayoutRules"
                      v-model="newLayoutName"
                      dense
                      class="pa-0 ma-0"
                    >
                      <template v-slot:append>
                        <v-btn
                          :disabled="!isNewLayoutValid"
                          @click="saveNewLayout"
                          small
                          icon
                        >
                          <v-icon small>mdi-content-save</v-icon>
                        </v-btn>
                      </template>
                    </v-text-field>
                  </v-form>
                </v-list-item-content>
              </v-card>

              <v-card
                style="box-shadow: none"
                :disabled="isLayoutModeAdd"
                @click="onAddClick"
              >
                <v-list-item-content>
                  <v-btn
                    :disabled="isLayoutModeAdd"
                    small
                    outlined
                    color="primary"
                  >
                    Add <v-icon small>mdi-plus</v-icon></v-btn
                  >
                </v-list-item-content>
              </v-card>
            </v-list>
          </v-col>
          <v-col cols="9" class="filter" v-if="allFields.length">
            <v-row
              v-for="fieldKey in currentFieldsKeys"
              :key="fieldKey"
              class="px-2"
            >
              <v-col class="d-flex align-center">
                <component
                  v-if="currentFields[fieldKey].custom"
                  :is="currentFields[fieldKey].component"
                  v-bind="currentFields[fieldKey]"
                  :disabled="isFieldsDisabled"
                  v-model="localFilter[fieldKey]"
                />
                <component
                  v-else
                  clearable
                  :disabled="isFieldsDisabled"
                  dense
                  :multiple="
                    currentFields[fieldKey]?.type === 'select'
                      ? !currentFields[fieldKey]?.single
                      : false
                  "
                  :label="currentFields[fieldKey].title"
                  :item-value="currentFields[fieldKey].item?.value"
                  :item-text="currentFields[fieldKey].item?.title"
                  :items="currentFields[fieldKey].items"
                  :is="getFieldComponent(currentFields[fieldKey])"
                  v-model="localFilter[fieldKey]"
                  range
                />
                <v-btn
                  :disabled="isFieldsDisabled"
                  icon
                  small
                  @click="changeFields(currentFields[fieldKey], false)"
                >
                  <v-icon size="20">mdi-delete-outline</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </v-col>
        </v-row>
        <v-row justify="space-between">
          <v-col
            cols="3"
            :class="{
              layout__controls: true,
              dark: theme === 'dark',
              light: theme !== 'dark',
            }"
          >
            <v-btn
              :disabled="isLayoutsOptionsDisabled"
              plain
              color="primary"
              @click="onLayoutsOptionsClick"
              small
            >
              <span v-if="isLayoutModeEdit">Save</span>
              <span v-else-if="isLayoutModeAdd">Cancel</span>
              <span v-else-if="isLayoutModePreview">Edit layouts</span>
            </v-btn>
          </v-col>
          <v-col class="d-flex justify-space-between align-end">
            <v-menu
              :disabled="isFieldsDisabled"
              :close-on-content-click="false"
              offset-y
              top
            >
              <template v-slot:activator="{ attrs, on }">
                <v-btn
                  small
                  color="primary"
                  plain
                  :disabled="isFieldsDisabled"
                  v-bind="attrs"
                  v-on="on"
                  >Add fields</v-btn
                >
              </template>
              <v-card class="pa-5">
                <v-row align="center" style="max-width: 35vw">
                  <v-col
                    v-for="field in allFields"
                    :key="field.key"
                    class="ma-0 pa-0 mx-2"
                    cols="4"
                  >
                    <v-checkbox
                      dense
                      :input-value="
                        currentFieldsKeys.find((key) => key === field.key)
                      "
                      @change="changeFields(field, $event)"
                      :label="field.title"
                    />
                  </v-col>
                </v-row>
              </v-card>
            </v-menu>
            <div class="buttons">
              <v-btn
                class="mx-2"
                @click="resetFilter"
                :disabled="isResetDisabled"
                color="primary"
                >Reset</v-btn
              >
              <v-btn
                class="mx-2"
                @click="saveFilter"
                :disabled="isSaveDisabled"
                color="primary"
                >Search</v-btn
              >
            </div>
          </v-col>
        </v-row>
      </v-card>
    </v-menu>
  </div>
</template>

<script setup>
import { computed, watch, ref, onMounted } from "vue";
import { useStore } from "@/store";
import { VAutocomplete, VTextField } from "vuetify/lib";
import DatePicker from "@/components/ui/datePicker.vue";
import LogickSelect from "@/components/ui/logickSelect.vue";
import FromToNumberField from "@/components/ui/fromToNumberField.vue";
import FilterTags from "@/components/search/filterTags.vue";
import { debounce } from "@/functions";

const store = useStore();

const isOpen = ref(false);
const localFilter = ref({});
const pinnedLayout = ref();
const currentFieldsKeys = ref([]);
function getBlankLayout() {
  return { filter: {}, fields: {}, id: "blank" };
}
const blankLayout = ref(getBlankLayout());
const layoutMode = ref("preview");
const newLayoutName = ref("New layout");

const addNewLayoutForm = ref();
const isNewLayoutValid = ref(false);
const newLayoutRules = ref([
  (val) => {
    if (!val) {
      return false;
    }

    return layouts.value.findIndex((l) => l.title === val) === -1;
  },
]);

onMounted(() => {
  window.addEventListener("beforeunload", () => saveSearchData());

  window.addEventListener("keydown", function (event) {
    const { key } = event;
    if (key === "Escape" && isOpen.value) {
      hideSearch();
    }
  });
});

const theme = computed(() => store.getters["app/theme"]);

const allFields = computed(() => store.getters["appSearch/fields"]);
const currentFields = computed(() => {
  const fields = {};

  currentFieldsKeys.value.forEach((key) => {
    fields[key] = allFields.value?.find((f) => f.key === key);
  });

  return fields;
});
const layouts = computed({
  get: () => store.getters["appSearch/layouts"] || [],
  set: (val) => store.commit("appSearch/setLayouts", val),
});
const currentLayout = computed({
  get: () =>
    layouts.value.find(
      (l) => l.id === store.getters["appSearch/currentLayout"]
    ),
  set: (val) => store.commit("appSearch/setCurrentLayout", val?.id),
});
const visibleLayout = computed(() => currentLayout.value || blankLayout.value);
const searchName = computed(() => store.getters["appSearch/searchName"]);
const defaultLayout = computed(
  () => store.getters["appSearch/defaultLayout"] || { title: "Default" }
);
const param = computed({
  get: () => store.getters["appSearch/param"],
  set: debounce((val) => store.commit("appSearch/setParam", val), 300),
});
const filter = computed({
  get: () => store.getters["appSearch/filter"],
  set: (filter) => store.commit("appSearch/setFilter", filter),
});

const isLayoutModePreview = computed(() => layoutMode.value === "preview");
const isLayoutModeEdit = computed(() => layoutMode.value === "edit");
const isLayoutModeAdd = computed(() => layoutMode.value === "add");

const isFieldsDisabled = computed(() => layoutMode.value !== "preview");
const isLayoutsOptionsDisabled = computed(() => false);
const isSaveDisabled = computed(() => {
  return JSON.stringify(localFilter.value) === JSON.stringify(filter.value);
});
const isResetDisabled = computed(() => {
  return JSON.stringify(localFilter.value) === JSON.stringify(filter.value);
});
const isResetAllHide = computed(() => {
  return (
    !currentLayout.value && !param.value && !Object.keys(filter.value).length
  );
});

const getFieldComponent = (field) => {
  switch (field.type) {
    case "input": {
      return VTextField;
    }
    case "select": {
      return VAutocomplete;
    }
    case "date": {
      return DatePicker;
    }
    case "logic-select": {
      return LogickSelect;
    }
    case "number-range": {
      return FromToNumberField;
    }
  }
};

const getSearchKey = (name) => `searchFilter-${name}`;

const saveSearchData = (name) => {
  name = name || searchName.value;
  if (!name) {
    return;
  }
  const data = {
    current: currentLayout.value?.id,
    layouts: layouts.value,
    pinned: pinnedLayout.value,
  };
  const key = getSearchKey(name);
  localStorage.setItem(key, JSON.stringify(data));

  const localKey = `${key}-local`;
  if (
    !currentLayout.value?.id &&
    JSON.stringify(filter.value) !== JSON.stringify("{}")
  ) {
    localStorage.setItem(localKey, JSON.stringify(filter.value));
  } else {
    localStorage.removeItem(localKey);
  }
};

const loadSearchData = (name) => {
  name = name || searchName.value;
  if (!name) {
    return;
  }
  const key = getSearchKey(name);
  const data = JSON.parse(localStorage.getItem(key) || `{}`);
  if (!data.layouts) {
    return;
  }

  if (data.layouts.length) {
    layouts.value = data.layouts.map((l) => {
      if (l.filter?.filter && typeof l.filter.filter === "object") {
        return { ...l, filter: { ...l.filter.filter } };
      }
      return l;
    });
  }

  if (data?.current) {
    currentLayout.value = layouts.value.find((l) => l.id === data.current);
    filter.value = currentLayout.value.filter;
  }

  pinnedLayout.value = data?.pinned;
  if (pinnedLayout.value && !currentLayout.value) {
    currentLayout.value = layouts.value.find((l) => isPinned(l));
  }

  const localKey = `${key}-local`;
  const local = JSON.parse(localStorage.getItem(localKey) || `{}`);
  if (!data?.current && Object.keys(local).length > 0) {
    filter.value = local;
    blankLayout.value.filter = local;
    currentLayout.value = null;
  }
};

const saveFilter = () => {
  filter.value = { ...localFilter.value };
  if (visibleLayout.value.id === blankLayout.value.id) {
    blankLayout.value.fields = [...currentFieldsKeys.value];
  } else {
    const layoutIndex = layouts.value.findIndex(
      (l) => l.id === currentLayout.value?.id
    );
    if (layoutIndex !== -1) {
      layouts.value[layoutIndex].fields = [...currentFieldsKeys.value];
      layouts.value[layoutIndex].filter = { ...filter.value };
    }
  }
  hideSearch();
};

const hideSearch = () => {
  isOpen.value = false;
};

const showSearch = () => {
  isOpen.value = true;
};

const resetFilter = () => {
  localFilter.value = { ...visibleLayout.value };
};

const resetAll = () => {
  setLayoutMode("preview");
  setCurrentLayout();
  param.value = "";
  setTimeout(() => (filter.value = {}));
};

const setCurrentFieldsKeys = () => {
  if (allFields.value.length === 0) {
    return;
  }
  const newCurrentFields = [];
  if (currentLayout.value?.filter) {
    newCurrentFields.push(...Object.keys(currentLayout.value.filter));
  }
  if (currentLayout.value?.fields) {
    newCurrentFields.push(...currentLayout.value.fields);
  } else {
    let i = newCurrentFields.length;
    while (
      newCurrentFields.length < 5 &&
      newCurrentFields.length !== allFields.value.length
    ) {
      const key = allFields.value[i].key;
      if (newCurrentFields.findIndex((f) => f.key === key) === -1) {
        newCurrentFields.push(key);
      }
      i++;
    }
  }

  currentFieldsKeys.value = [
    ...new Set(
      newCurrentFields.filter(
        (key) => !!allFields.value?.find((f) => f.key === key)
      )
    ),
  ];
};
const changeFields = ({ key }, value) => {
  if (value) {
    currentFieldsKeys.value.push(key);
  } else {
    currentFieldsKeys.value = currentFieldsKeys.value.filter((f) => f !== key);
    const newFilter = { ...localFilter.value };
    delete newFilter[key];
    localFilter.value = newFilter;
  }
};

const addNewLayout = (data) => {
  layouts.value.push({
    ...data,
    filter: data.filter || {},
    id: Date.now(),
  });
};

const onAddClick = () => {
  setLayoutMode("add");
  setTimeout(() => addNewLayoutForm.value.validate(), 300);
};
const saveNewLayout = () => {
  if (blankLayout.value.id === visibleLayout.value.id) {
    addNewLayout({
      title: newLayoutName.value,
      filter: { ...filter.value },
      fields: [...currentFieldsKeys.value],
    });
    blankLayout.value = getBlankLayout();
  } else {
    addNewLayout({ title: newLayoutName.value });
  }
  newLayoutName.value = "New layout";
  currentLayout.value = layouts.value[layouts.value.length - 1];
  setLayoutMode("preview");
};
const deleteLayout = (id) => {
  layouts.value = layouts.value.filter((l) => l.id !== id);
};

const isPinned = (layout) => pinnedLayout.value === layout.id;

const setPinned = (id) => {
  if (pinnedLayout.value === id) {
    pinnedLayout.value = undefined;
  } else {
    pinnedLayout.value = id;
  }
};

const setCurrentLayout = (layout) => {
  if (isLayoutModePreview.value) {
    if (
      currentLayout.value &&
      pinnedLayout.value &&
      !isPinned(currentLayout.value) &&
      !layout
    ) {
      layout = layouts.value.find((l) => isPinned(l));
    }
    currentLayout.value = layout;
    if (!layout) {
      blankLayout.value.filter = {};
      blankLayout.value.fields = [];
    }
  }
};
const setLayoutMode = (mode) => {
  layoutMode.value = mode;
};

const onLayoutsOptionsClick = () => {
  if (isLayoutModePreview.value) {
    setLayoutMode("edit");
  } else if (isLayoutModeEdit.value) {
    setLayoutMode("preview");
  } else {
    newLayoutName.value = "New layout";
    setLayoutMode("preview");
  }
};

watch(searchName, (value, oldValue) => {
  if (oldValue) {
    saveSearchData(oldValue);
  }
  currentLayout.value = undefined;
  param.value = "";
  filter.value = {};
  blankLayout.value = getBlankLayout();

  layouts.value = [];
  if (value) {
    loadSearchData(value);
  }

  if (layouts.value.length === 0 && value) {
    addNewLayout(JSON.parse(JSON.stringify(defaultLayout.value)));
    setCurrentLayout(layouts.value[0]);
    setPinned(layouts.value[0].id);
  }
});
watch(allFields, (fields) => {
  if (fields?.length) {
    layouts.value = layouts.value.map((layout) => {
      Object.keys(layout.filter || {}).map((key) => {
        const filterValue = layout.filter[key];
        const field = fields.find((f) => f.key === key);

        if (Array.isArray(filterValue) && field && field.items?.length) {
          const { items, item: options } = field;
          const correctFilter = [];
          filterValue.forEach((value) => {
            if (
              items.find(
                (item) => (item[options?.value || "value"] || item) === value
              )
            ) {
              correctFilter.push(value);
            }
          });

          if (correctFilter.length !== filterValue.length) {
            layout.filter[key] = correctFilter;
          }
        }
      });

      return layout;
    });
  }

  setCurrentFieldsKeys();
});
watch(visibleLayout, (_, prevLayout) => {
  if (prevLayout?.id === blankLayout.value.id) {
    blankLayout.value.filter = { ...filter.value };
    blankLayout.value.fields = [...currentFieldsKeys.value];
  } else {
    const prevLayoutIndex =
      prevLayout && layouts.value.findIndex((l) => l.id === prevLayout.id);

    if (typeof prevLayoutIndex === "number" && prevLayoutIndex !== -1) {
      layouts.value[prevLayoutIndex].filter = filter.value;
    }
  }

  filter.value = visibleLayout.value?.filter || {};
  setCurrentFieldsKeys();
});
watch(
  filter,
  (newValue) => {
    localFilter.value = { ...newValue };
  },
  { deep: true }
);
</script>

<script>
export default {
  name: "app-search",
};
</script>

<style scoped lang="scss">
.search {
  max-width: 50vw !important;
  max-height: 80vh !important;
  margin-top: 5px;
  .search__card {
    min-height: 40vh !important;
    padding-bottom: 100px;

    .search__content {
      min-height: calc(40vh - 55px) !important;
      .layouts {
        min-height: calc(40vh - 55px) !important;
        * {
          background-color: inherit !important;
        }
      }
    }

    .fields__controls {
      position: absolute;
      bottom: 10px;
      right: 10px;
    }
    .layout__controls {
      background-color: var(--v-background-base);
      * {
        background-color: inherit !important;
      }
    }

    .dark {
      filter: brightness(110%);
      background-color: var(--v-background-base);
    }

    .light {
      filter: brightness(95%);
      background-color: var(--v-background-base);
    }
  }
}
</style>

<style>
.search__input .v-input__control .v-input__slot {
  padding: 0 10px;
}

.search__input .v-input__append-inner {
  margin-top: 0px;
}

.search__input .v-input__prepend-inner {
  margin-right: unset !important;
}
</style>

<style scoped lang="scss">
.search_container {
  width: 50vw;
  min-width: 100pxpx;

  @media (max-width: 1100px) {
    width: 40vw;
  }

  @media (max-width: 900px) {
    width: 30vw;
  }
}
</style>
