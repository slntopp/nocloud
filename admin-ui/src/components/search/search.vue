<template>
  <div>
    <v-menu
      v-model="isOpen"
      :close-on-click="true"
      :close-on-content-click="false"
      offset-y
    >
      <template v-slot:activator="{ on, attrs }">
        <v-text-field
          @input="isOpen = true"
          ref="search-input"
          hide-details
          prepend-inner-icon="mdi-magnify"
          placeholder="Search..."
          single-line
          background-color="background-light"
          dence
          v-model="searchParam"
          rounded
          v-bind="attrs"
          v-on="on"
        >
          <template v-slot:append>
            <v-chip
              outlined
              color="primary"
              v-for="key in Object.keys(customParams)"
              :key="key"
              class="mx-1"
            >
              {{ customParams[key]?.title }}
              <v-btn @click="deleteParam(key)" icon small>
                <v-icon small>mdi-close</v-icon>
              </v-btn>
            </v-chip>
          </template>
        </v-text-field>
      </template>
      <v-card v-if="searchItems.length || selectedGroupKey">
        <v-card-subtitle v-if="selectedGroupKey">
          <v-btn class="mr-4" icon @click="selectedGroupKey = null">
            <v-icon>mdi-arrow-left</v-icon>
          </v-btn>
          {{ variants[selectedGroupKey].title }}
        </v-card-subtitle>
        <div style="max-height: 300px">
          <v-list>
            <v-list-item-group
              color="primary"
              @change="changeValue"
              :value="selectedGroupKey"
            >
              <template v-if="searchItems.length > 0">
                <v-list-item
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

export default {
  name: "app-search",
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
          },
        });
        this.searchParam = "";
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
        this.searchParam = "";
      } else {
        this.selectedGroupKey = searchItem.key;
      }
    },
    deleteParam(key) {
      this.$store.commit("appSearch/deleteCustomParam", key);
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
  },
};
</script>

<style scoped></style>
