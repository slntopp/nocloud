<template>
  <div class="container">
    <div>
      <v-list flat dark color="rgba(12, 12, 60, 0.9)">
        <v-subheader>TYPE</v-subheader>
        <v-list-item-group mandatory v-model="selectedType" color="primary">
          <v-list-item
            :disabled="isRegionLoading"
            v-for="(item, i) in types"
            :key="i"
          >
            {{ item }}
          </v-list-item>
        </v-list-item-group>
      </v-list>
      <div v-if="isRegionLoading" class="spinner">
        <v-progress-circular size="40" color="primary" indeterminate />
      </div>

      <v-list v-else flat dark color="rgba(12, 12, 60, 0.9)">
        <v-subheader>REGIONS</v-subheader>
        <v-list-item-group mandatory v-model="selectedRegion" color="primary">
          <v-list-item v-for="(item, i) in allRegions" :key="i">
            {{ item }}
          </v-list-item>
        </v-list-item-group>
      </v-list>
    </div>
    <support-map
      @errorAddPin="errorAddPin"
      :activePinTitle="activePinTitle"
      :canAddPin="canAddPin"
      :multiSelect="true"
      :error="mapError"
      :template="template"
      :type="types[selectedType]"
      :region="allRegions[selectedRegion]"
      @save="onSavePin"
      @pinHover="onPinHover"
    />
  </div>
</template>

<script setup>
import supportMap from "./map.vue";
import api from "@/api.js";
import { ref, defineProps, toRefs, computed, watch, onMounted } from "vue";

const props = defineProps({ template: { required: true, type: Object } });
const { template } = toRefs(props);

const selectedRegion = ref("");
const allRegions = ref([]);
const types = ref(["ovh vps", "ovh cloud", "ovh dedicated"]);
const mapError = ref("");
const selectedType = ref();
const isRegionLoading = ref(false);

const onPinHover = (id) => {
  if (allRegions.value) {
    const location = template.value.locations.find((el) => el.id === id);

    selectedRegion.value = allRegions.value.indexOf(location.extra.region);
  }
};
const errorAddPin = () => {
  mapError.value = "";
  if (selectedLocation.value) {
    mapError.value = "Error: This region alredy taken";
  } else {
    mapError.value = "Error: Choose the region";
  }
};
const onSavePin = () => {
  selectedRegion.value = null;
  mapError.value = "";
};
const activePinTitle = computed(() => selectedLocation.value?.title || "");
const selectedLocation = computed(() =>
  template.value.locations.find(
    (l) =>
      l.extra?.region &&
      allRegions.value[selectedRegion.value] &&
      l.extra?.region === allRegions.value[selectedRegion.value]
  )
);
const canAddPin = computed(
  () =>
    !selectedLocation.value &&
    (!!selectedRegion.value || selectedRegion.value == 0)
);

onMounted(() => {
  selectedType.value = 0;
});

watch(selectedType, async () => {
  try {
    isRegionLoading.value = true;
    const { meta } = await api.servicesProviders.action({
      action: "regions",
      uuid: template.value.uuid,
      params: {
        projectId: template.value.vars?.projectId?.value?.default,
        type: types.value[selectedType.value],
      },
    });
    allRegions.value = meta.datacenters;
  } catch {
    mapError.value = "Error: Cannot download regions";
  } finally {
    isRegionLoading.value = false;
  }
});
</script>

<style scoped>
.container {
  display: grid;
  grid-template-columns: 150px 1fr;
  grid-column-gap: 20px;
}
.spinner {
  margin-top: 150px;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
