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

      <v-expansion-panels :value="0">
        <v-expansion-panel v-for="(item, i) in showcase.items" :key="i">
          <v-expansion-panel-header color="indigo darken-4">
            {{ item.serviceProvider }} - {{ item.plan }}

            <v-icon
              style="flex: 0 0 auto; margin: 0 auto 0 10px"
              color="error"
              v-if="!(i === showcase.items.length - 1 || item.serviceProvider === '')"
              @click="removeItem(i)"
            >
              mdi-close-circle
            </v-icon>
          </v-expansion-panel-header>

          <v-expansion-panel-content color="indigo darken-4">
            <v-row>
              <v-col cols="6">
                <v-autocomplete
                  label="Service provider"
                  item-text="title"
                  item-value="uuid"
                  v-model="item.serviceProvider"
                  :items="serviceProviders"
                  @change="addItem"
                />
              </v-col>
              <v-col cols="6">
                <v-autocomplete
                  label="Price model"
                  item-text="title"
                  item-value="uuid"
                  v-model="item.plan"
                  :items="plans[i]"
                />
              </v-col>
              <v-col cols="6">
                <v-autocomplete
                  multiple
                  label="Allowed types"
                  v-model="allowedTypes[i]"
                  :items="locationsTypes[i]"
                />
              </v-col>
              <v-col cols="6">
                <locations-autocomplete
                  label="Locations"
                  v-model="item.locations"
                  :locations="filteredLocations[i]"
                />
              </v-col>
            </v-row>
          </v-expansion-panel-content>
        </v-expansion-panel>
      </v-expansion-panels>

      <v-btn
        style="display: block; margin: 10px 0 0 auto"
        :loading="isSaveLoading"
        @click="save"
      >
        {{ isEdit ? "Save" : "Create" }}
      </v-btn>
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
  items: [{
    plan: "",
    serviceProvider: "",
    locations: [],
  }],
  promo: {},
});

const isLoading = ref(false);
const allowedTypes = ref({});
const isSaveLoading = ref(false);

const requiredRule = ref((val) => !!val || "Required field");
const serviceProviders = computed(() => store.getters["servicesProviders/all"]);

const plans = computed(() => {
  const allPlans = store.getters["plans/all"];

  return showcase.value.items.reduce((result, { serviceProvider }, i) => {
    const { meta } = serviceProviders.value.find(
      ({ uuid }) => uuid === serviceProvider
    ) ?? {};

    return { ...result, [i]: allPlans.filter(({ uuid }) => meta?.plans?.includes(uuid)) };
  }, {});
});

const locations = computed(() => {
  return showcase.value.items.reduce((result, { serviceProvider }, i) => {
    const { uuid, locations = [] } = serviceProviders.value.find(
      (sp) => sp.uuid === serviceProvider
    ) ?? {};

    return {
      ...result,
      [i]: locations.map((location) => ({
        ...location, sp: uuid, id: getNewLocationKey(location)
      }))
    };
  }, {});
});

const filteredLocations = computed(() => {
  const result = {};

  Object.entries(locations.value).forEach(([i, value]) => {
    result[i] = value.filter(({ type }) =>
      allowedTypes.value[i].includes(type)
    );
  });

  return result;
});

const locationsTypes = computed(() => {
  const result = {};

  Object.entries(locations.value).forEach(([i, value]) => {
    result[i] = [...new Set(value.map(({ type }) => type))];
  });

  return result;
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

  if (!Array.isArray(showcase.value.items)) {
    showcase.value.items = [];
  }
  showcase.value.items.push({ plan: "", serviceProvider: "", locations: [] });

  allowedTypes.value = [
    ...new Set(showcase.value.locations.map((location) => location.type))
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
    const data = JSON.parse(JSON.stringify(showcase.value));

    data.items.pop();
    Object.entries(filteredLocations.value).forEach(([i, value]) => {
      if (value.length < 1) return;
      data.items[i].locations = value.map((location) => ({
        ...location,
        sp: undefined,
        id: location.id.replace(
          data.title.replaceAll(' ', '_'),
          data.newTitle.replaceAll(' ', '_')
        )
      }));
    });

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
    console.log(e);
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

const addItem = () => {
  if (showcase.value.items.at(-1).serviceProvider !== '') {
    showcase.value.items.push({ plan: '', serviceProvider: '', locations: [] });
  }
}

const removeItem = (i) => {
  showcase.value.items.splice(i, 1);
}
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
