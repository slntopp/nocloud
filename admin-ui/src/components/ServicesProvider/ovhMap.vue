<template>
  <div class="container">
    <div>
      <v-select
        label="type"
        :disabled="isRegionLoading"
        :loading="isRegionLoading"
        :items="types"
        v-model="selectedType"
      ></v-select>
      <div v-if="isRegionLoading" class="spinner">
        <v-progress-circular size="40" color="primary" indeterminate />
      </div>
      <div v-else class="hint">
        Click on a country to add a location, then pick its OVH region in the
        popup settings.
      </div>
    </div>
    <support-map
      :multiSelect="true"
      :error="mapError"
      :template="template"
      :regions="allRegions"
      :type="selectedType"
      @set-locations="setLocations"
    />
  </div>
</template>

<script setup>
import supportMap from "./map.vue";
import api from "@/api.js";
import { ref, defineProps, toRefs, watch, onMounted } from "vue";

const props = defineProps({ template: { required: true, type: Object } });
const { template } = toRefs(props);

const allRegions = ref([]);
const types = ref(["ovh vps", "ovh cloud", "ovh dedicated"]);
const mapError = ref();
const selectedType = ref("ovh vps");
const isRegionLoading = ref(false);

onMounted(() => {
  fetchRegions();
});

watch(selectedType, () => {
  fetchRegions();
});

const fetchRegions = async () => {
  try {
    isRegionLoading.value = true;
    const { meta } = await api.servicesProviders.action({
      action: "regions",
      uuid: template.value.uuid,
      params: {
        projectId: template.value.vars?.projectId?.value?.default,
        type: selectedType.value,
      },
    });
    allRegions.value = meta.datacenters;
  } catch {
    mapError.value = "Error: Cannot download regions";
  } finally {
    isRegionLoading.value = false;
  }
};

const setLocations = (locations) => {
  template.value.locations = locations;
};
</script>

<style scoped>
.container {
  display: grid;
  grid-template-columns: 150px 1fr;
  grid-column-gap: 20px;
}
.hint {
  margin-top: 10px;
  font-size: 12px;
  opacity: 0.7;
}
.spinner {
  margin-top: 150px;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
