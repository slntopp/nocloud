<template>
  <div>
    <v-menu
      v-model="isOpen"
      :close-on-click="true"
      :close-on-content-click="false"
      offset-y
    >
      <template v-slot:activator="{ attrs, on }">
        <v-text-field
          @input="onSearchInput"
          @click="onSearchInput"
          ref="search-input"
          hide-details
          placeholder="Search..."
          single-line
          background-color="background-light"
          dence
          v-model="searchParam"
          rounded
          v-bind="attrs"
          v-on="on"
        >
          <template v-slot:prepend-inner>
            <v-icon
              :style="{
                marginTop: customParamsValues.length ? '12px !important' : 0,
              }"
              class="ma-auto"
              >mdi-magnify
            </v-icon>
          </template>

          <template v-slot:append>
            <search-tag
              v-if="customParamsValues.length"
              :param="customParamsValues[0]"
              :variants="variants"
            />
            <v-menu v-if="customParamsValues.length > 1" offset-y>
              <template v-slot:activator="{ on, attrs }">
                <v-chip
                  class="ma-auto pa-auto"
                  color="primary"
                  v-bind="attrs"
                  outlined
                  v-on="on"
                >
                  +{{ customParamsValues.length - 1 }}
                </v-chip>
              </template>
              <v-card color="background-light" max-width="600px">
                <search-tag
                  v-for="param in customParamsValues.slice(1)"
                  :key="param.key + param.title"
                  :param="param"
                  :variants="variants"
                />
              </v-card>
            </v-menu>
          </template>
        </v-text-field>
      </template>
      <v-card
        color="background-light"
        v-if="searchItems.length || selectedGroupKey"
      >
        <v-card-subtitle v-if="selectedGroupKey">
          <v-btn class="mr-4" icon @click="selectedGroupKey = null">
            <v-icon>mdi-arrow-left</v-icon>
          </v-btn>
          {{ variants[selectedGroupKey].title }}
        </v-card-subtitle>
        <div style="max-height: 600px">
          <v-list ref="searchList" color="background-light">
            <v-list-item-group
              color="primary"
              @change="changeValue"
              :value="selectedGroupKey"
            >
              <template v-if="searchItems.length > 0">
                <v-list-item
                  class="search__list-item"
                  active-class="active"
                  color="primary"
                  :key="item.key"
                  v-for="item in searchItems"
                >
                  <div
                    style="width: 100%"
                    class="d-flex justify-space-between"
                    v-if="!selectedGroupKey"
                  >
                    <span>
                      {{ searchParam || "" }}
                    </span>
                    <span> in {{ item.title }} </span>
                  </div>
                  <span v-else>{{ item.title }}</span>
                </v-list-item>
              </template>
              <div v-else style="width: 100%" class="d-flex justify-center">
                <span class="text-center"> No data available</span>
              </div>
            </v-list-item-group>
          </v-list>
        </div>
      </v-card>
    </v-menu>
  </div>
</template>

<script>
import { mapGetters } from "vuex";
import SearchTag from "@/components/search/searchTag.vue";

export default {
  name: "app-search",
  components: { SearchTag },
  data: () => ({ selectedGroupKey: "", isOpen: false }),
  methods: {
    changeValue(e) {
      const searchItem = this.searchItems[e];
      const variant = this.variants[this.selectedGroupKey || searchItem?.key];
      if (this.selectedGroupKey) {
        this.$store.commit("appSearch/setCustomParam", {
          key: variant?.key || this.selectedGroupKey,
          value: {
            value: searchItem[variant.itemKey || "uuid"],
            title: searchItem[variant.itemTitle || "title"],
            isArray: variant.isArray,
          },
        });
        this.isOpen=false
        this.selectedGroupKey = null;
      } else if (!variant?.items) {
        this.$store.commit("appSearch/setCustomParam", {
          key: variant?.key,
          value: {
            value: this.searchParam,
            title: this.searchParam,
          },
        });
        this.selectedGroupKey = null;
        this.isOpen=false
      } else {
        this.selectedGroupKey = searchItem.key;
      }
    },
    onSearchInput() {
      this.isOpen = true;
      if (this.$refs.searchList?.$el) {
        this.$refs.searchList.$el.focus();
      }
    },
  },
  computed: {
    ...mapGetters("appSearch", {
      isAdvancedSearch: "isAdvancedSearch",
      variants: "variants",
      customParams: "customParams",
    }),
    searchParam: {
      get() {
        return this.$store.getters["appSearch/param"];
      },
      set(newValue) {
        this.$store.commit("appSearch/setSearchParam", newValue);
      },
    },
    searchItems() {
      return (
        this.variants[this.selectedGroupKey]?.items.filter(
          (i) =>
            !this.searchParam ||
            i.title.toLowerCase().includes(this.searchParam.toLowerCase())
        ) ||
        Object.keys(this.variants).map((key) => ({
          key,
          title: this.variants[key]?.title || key,
        }))
      );
    },
    customParamsValues() {
      const values = [];
      Object.keys(this.customParams).forEach((key) => {
        if (Array.isArray(this.customParams[key])) {
          values.push(
            ...this.customParams[key]?.map((v) => ({
              ...v,
              isArray: true,
              key,
            }))
          );
        } else {
          values.push({ ...this.customParams[key], key });
        }
      });

      return values;
    },
  },
};
</script>

<style lang="scss" scoped>
.search__list-item {
  //border:  1px solid #e06ffe;
  //border-radius: 10px;
}
</style>
