<template>
  <div class="container">
    <v-progress-circular v-if="!allRegions.length" class="spinner" size="40" color="primary" indeterminate/>
    <v-list v-else flat dark color="rgba(12, 12, 60, 0.9)">
      <v-list-item-group v-model="selectedRegion" color="primary">
        <v-list-item v-for="(item, i) in allRegions" :key="i">
          {{ item }}
        </v-list-item>
      </v-list-item-group>
    </v-list>
    <support-map :multiSelect="true" @save="onSavePin" :template="template" />
  </div>
</template>

<script>
import supportMap from "./map.vue";
import api from "@/api.js";

export default {
  name: "ovh-map",
  data() {
    return { selectedRegion: "", allRegions: [] };
  },
  components: {
    supportMap,
  },
  props: { template: { required: true, type: Object } },
  methods: {
    onSavePin(item) {
      if (this.selectedRegion) {
        item.locations[item.locations.length - 1].extra = {
          region: this.allRegions[this.selectedRegion],
        };
      }
    },
  },
  mounted() {
    api
      .post(`/sp/${this.template.uuid}/invoke`, { method: "flavors" })
      .then(({ meta }) => {
        this.allRegions = meta.result
          .map((el) => el.region)
          .filter((element, index, arr) => {
            return arr.indexOf(element) === index;
          });
      })
      .catch(() => {
        this.showSnackbarError({
          message: "Error: Cannot download regions",
        });
      });
  },
};
</script>

<style scoped>
.container {
  display: grid;
  grid-template-columns: 100px 1fr;
  grid-column-gap: 20px;
}
.spinner{
  margin: auto;
  margin-top: 150px;
}
</style>
