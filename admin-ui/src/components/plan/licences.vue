<template>
  <Licences
    v-if="serviceProvider"
    :template="serviceProvider"
    :plan="template.uuid"
  />
  <v-alert v-else type="info" class="mt-4"> No service provider. </v-alert>
</template>

<script setup>
import { computed, toRefs } from "vue";
import Licences from "@/components/ServicesProvider/licences.vue";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const serviceProvider = computed(() =>
  store.getters["servicesProviders/all"].find((sp) =>
    sp.meta?.plans?.includes(template.value?.uuid),
  ),
);
</script>

<script>
export default {
  name: "PlanLicencesTab",
};
</script>
