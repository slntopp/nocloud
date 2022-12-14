<template>
  <div class="container">
    <v-progress-circular
      v-if="!allRegions.length"
      class="spinner"
      size="40"
      color="primary"
      indeterminate
    />
    <v-list v-else flat dark color="rgba(12, 12, 60, 0.9)">
      <v-subheader>REGIONS</v-subheader>
      <v-list-item-group v-model="selectedRegion" color="primary">
        <v-list-item v-for="(item, i) in allRegions" :key="i">
          {{ item }}
        </v-list-item>
      </v-list-item-group>
    </v-list>
    <support-map
      @errorAddPin="errorAddPin"
      :activePinTitle="activePinTitle"
      :canAddPin="canAddPin"
      :multiSelect="true"
      :error="mapError"
      :template="template"
      :region="allRegions[selectedRegion]"
      @save="onSavePin"
      @pinHover="onPinHover"
    />
  </div>
</template>

<script>
import supportMap from "./map.vue";
import api from "@/api.js";

export default {
  name: "ovh-map",
  components: { supportMap },
  props: { template: { required: true, type: Object } },
  data: () => ({ selectedRegion: "", allRegions: [], mapError: "" }),
  methods: {
    onPinHover(id) {
      if (this.allRegions) {
        const location = this.template.locations.find((el) => el.id === id);

        this.selectedRegion = this.allRegions.indexOf(location.extra.region);
      }
    },
    errorAddPin() {
      if (this.selectedLocation) {
        this.mapError = "Error: This region alredy taken";
      } else {
        this.mapError = "Error: Choose the region";
      }
    },
    onSavePin(item) {
      item.locations[item.locations.length - 1].extra = {
        region: this.allRegions[this.selectedRegion],
      };
      this.selectedRegion = null;
      this.mapError = "";
    },
  },
  computed: {
    activePinTitle() {
      return this.selectedLocation?.title || "";
    },
    selectedLocation() {
      return this.template.locations.find(
        (l) =>
          l.extra?.region &&
          this.allRegions[this.selectedRegion] &&
          l.extra?.region === this.allRegions[this.selectedRegion]
      );
    },
    canAddPin() {
      return (
        !this.selectedLocation &&
        (!!this.selectedRegion || this.selectedRegion === 0)
      );
    },
  },
  mounted() {
    api.post(`/sp/${this.template.uuid}/invoke`, { method: "regions" })
      .then(({ meta }) => {
        this.allRegions = meta.datacenters;
      })
      .catch(() => {
        this.mapError = "Error: Cannot download regions";
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
.spinner {
  margin: auto;
  margin-top: 150px;
}
</style>
