<template>
  <v-autocomplete :loading="isIconsLoading" :items="icons" :value="value" @input="emits('input:value',$event)">
    <template v-slot:prepend>
      <v-icon class="ml-3">{{ `mdi-${value}` }}</v-icon>
    </template>
    <template v-slot:item="{ item }">
      <icon-title-preview :icon="item" :title="item" is-mdi/>
    </template>
  </v-autocomplete>
</template>

<script setup>

import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";
import {onMounted, ref, toRefs} from "vue";
import {fetchMDIIcons} from "@/functions";

const props = defineProps(['value'])
const emits = defineEmits(['input:value'])

const {value} = toRefs(props)
const icons = ref([])
const isIconsLoading = ref(false)

onMounted(() => {
  fetchIcons()
})

const fetchIcons = () => {
  isIconsLoading.value = true
  fetchMDIIcons()
      .then((data) => {
        icons.value = data.map((icon) => icon.name);
      }).finally(() => isIconsLoading.value = false);
}
</script>

<style scoped>

</style>