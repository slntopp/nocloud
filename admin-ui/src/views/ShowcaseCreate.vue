<template>
  <div class="pa-10 h-100 w-100">
    <h1 class="page__title" v-if="!isEdit">Create showcase</h1>
    <v-form ref="showcaseForm" align="center">
      <v-row>
        <v-col cols="6">
          <v-text-field
            :rules="[requiredRule]"
            v-model="showcase.newTitle"
            label="Title"
          />
        </v-col>
        <v-col cols="4">
          <icons-autocomplete
            label="Preview icon"
            :value="showcase.icon"
            @input:value="showcase.icon = $event"
          />
        </v-col>
        <v-col cols="2">
          <v-switch label="Is primary" v-model="showcase.primary" />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-autocomplete
            item-text="title"
            v-model="showcase.servicesProviders"
            item-value="uuid"
            multiple
            label="Service providers"
            :items="serviceProviders"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            item-text="title"
            item-value="uuid"
            multiple
            label="Plans"
            v-model="showcase.plans"
            :items="plans"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-autocomplete
            multiple
            label="Allowed types"
            v-model="allowedTypes"
            :items="locationsTypes"
          />
        </v-col>
        <v-col cols="6">
          <locations-autocomplete
            label="Locations"
            :locations="filtredLocations"
            v-model="showcase.locations"
          />
        </v-col>
      </v-row>
      <v-row justify="end">
        <v-btn @click="save" :loading="isSaveLoading">{{
          isEdit ? "Save" : "Create"
        }}</v-btn>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import { onMounted, ref, computed, toRefs, watch } from "vue";
import IconsAutocomplete from "@/components/ui/iconsAutocomplete.vue";
import { useStore } from "@/store";
import api from "@/api";
import { useRouter } from "vue-router/composables";
import LocationsAutocomplete from "@/components/ui/locationsAutocomplete.vue";

const props = defineProps({
  realShowcase: {},
  isEdit: { type: Boolean, default: false },
});
// const emits=defineEmits(['input'])

const { realShowcase, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const showcase = ref({
  primary: false,
  title: "",
  newTitle: "",
  icon: "",
  servicesProviders: [],
  plans: [],
  promo: {},
  locations: [],
});
const isLoading = ref(false);
const allowedTypes = ref([]);
const isSaveLoading = ref(false);

const requiredRule = ref((val) => !!val || "Required field");
const serviceProviders = computed(() => store.getters["servicesProviders/all"]);
const plans = computed(() => store.getters["plans/all"]);
const locations = computed(() => {
  const sps = serviceProviders.value.filter((sp) =>
    showcase.value.servicesProviders?.includes(sp.uuid)
  );
  const locations = [];
  sps.forEach((sp) => {
    locations.push(
      ...sp.locations.map((l) => ({
        ...l,
        sp: sp.title,
        id: getNewLocationKey(l),
      }))
    );
  });
  return locations;
});

const filtredLocations = computed(() => {
  return locations.value.filter((l) => allowedTypes.value.includes(l.type));
});
const locationsTypes = computed(() => {
  return [...new Set(locations.value.map((l) => l.type))];
});

watch(locationsTypes, () => {
  if (!isEdit.value) {
    allowedTypes.value = locationsTypes.value;
  } else if (allowedTypes.value.length === 0) {
    allowedTypes.value = locationsTypes.value;
  }
});

watch(realShowcase, () => {
  showcase.value = realShowcase.value;
  showcase.value.newTitle = showcase.value.title;
  allowedTypes.value = [
    ...new Set(showcase.value.locations.map((l) => l.type)),
  ];
});

onMounted(async () => {
  try {
    isLoading.value = true;
    await Promise.all([
      store.dispatch("servicesProviders/fetch"),
      store.dispatch("plans/fetch"),
    ]);
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch info",
    });
  } finally {
    isLoading.value = false;
  }
});

const save = async () => {
  try {
    const data = {
      ...showcase.value,
    };
    data.locations = data.locations
      .filter((l) => filtredLocations.value.find((l2) => l2.id === l.id))
      .map((l) => ({
        ...l,
        id: l.id.replace(data.title.replaceAll(' ','_'), data.newTitle.replaceAll(' ','_')),
        sp: undefined,
      }));
    data.title = data.newTitle;
    delete data.newTitle;
    
    isSaveLoading.value = true;
    if (isEdit.value) {
      await api.showcases.update(data);
    } else {
      await api.showcases.create(data);
    }
    store.commit("snackbar/showSnackbarSuccess", {
      message: `Showcase successfully ${isEdit.value ? "saved" : "created"}`,
    });
    router.push({ name: "Showcases" });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save showcase",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const getNewLocationKey = (l) => {
  return `${showcase.value.title.replaceAll(" ", `_`)}-${l.id}`;
};
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
