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
        <v-col cols="6" style="display: flex; gap: 30px; justify-content: flex-end">
          <v-switch label="Is primary" v-model="showcase.primary" />
          <v-switch label="Enabled" v-model="showcase.public" />
        </v-col>
        <v-col cols="6">
          <icons-autocomplete
            label="Preview icon"
            :value="showcase.icon"
            @input:value="showcase.icon = $event"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            clearable
            item-text="title"
            item-value="id"
            label="Default location"
            v-model="defaultLocation"
            :items="allLocations"
          />
        </v-col>
      </v-row>

      <v-expansion-panels :value="0">
        <v-expansion-panel v-for="(item, i) in showcase.items" :key="i">
          <v-expansion-panel-header color="background">
            {{ getProviderTitle(item.servicesProvider) }}
            - {{ getPlanTitle(item.plan) }}

            <v-icon
              style="flex: 0 0 auto; margin: 0 auto 0 10px"
              color="error"
              v-if="!(i === showcase.items.length - 1 || item.servicesProvider === '')"
              @click="removeItem(i)"
            >
              mdi-close-circle
            </v-icon>
          </v-expansion-panel-header>

          <v-expansion-panel-content color="background">
            <v-row>
              <v-col cols="6">
                <v-autocomplete
                  label="Service provider"
                  item-text="title"
                  item-value="uuid"
                  v-model="item.servicesProvider"
                  :items="serviceProviders"
                  :rules="[requiredRule]"
                  @change="addItem"
                />
              </v-col>
              <v-col cols="6">
                <v-autocomplete
                  clearable
                  label="Price model"
                  item-text="title"
                  item-value="uuid"
                  v-model="item.plan"
                  :items="plans[i]"
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
    servicesProvider: "",
    locations: [],
  }],
  promo: {},
  locations: [],
  public: true
});

const isLoading = ref(false);
const defaultLocation = ref("");
const isSaveLoading = ref(false);

const requiredRule = ref((val) => !!val || "Required field");
const serviceProviders = computed(() => store.getters["servicesProviders/all"]);

const plans = computed(() => {
  const allPlans = store.getters["plans/all"];

  return showcase.value.items.reduce((result, { servicesProvider }, i) => {
    const { meta } = serviceProviders.value.find(
      ({ uuid }) => uuid === servicesProvider
    ) ?? {};

    return { ...result, [i]: allPlans.filter(({ uuid }) => meta?.plans?.includes(uuid)) };
  }, {});
});

const locations = computed(() =>
  showcase.value.items.reduce((result, { servicesProvider }, i) => {
    const { uuid, locations = [] } = serviceProviders.value.find(
      (sp) => sp.uuid === servicesProvider
    ) ?? {};

    return {
      ...result,
      [i]: locations.map((location) => ({
        ...location, sp: uuid, id: getNewLocationKey(location)
      }))
    };
  }, {})
);

const filteredLocations = computed(() => {
  const result = {};

  Object.entries(locations.value).forEach(([i, value]) => {
    const plan = plans.value[i].find(({ uuid }) =>
      uuid === showcase.value.items[i].plan
    );

    if (!plan) return;
    result[i] = value.filter(({ type }) => plan.type === type);
  });

  return result;
});

const allLocations = computed(() =>
  Object.entries(filteredLocations.value).reduce(
    (result, [i, locations]) => [
      ...result,
      ...locations.filter(({ id }) =>
        showcase.value.items[i].locations
          .find((location) => id === (location.id ?? location))
      )
    ], []
  )
);

watch(realShowcase, () => {
  defaultLocation.value = realShowcase.value.promo.main?.default ?? "";
  showcase.value = JSON.parse(JSON.stringify(realShowcase.value));
  showcase.value.newTitle = showcase.value.title;

  if (!Array.isArray(showcase.value.items)) {
    showcase.value.items = [];
  }
  showcase.value.items.push({ plan: "", servicesProvider: "", locations: [] });
});

onMounted(async () => {
  try {
    isLoading.value = true;
    await Promise.all([
      store.dispatch("servicesProviders/fetch",{anonymously:true}),
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
    data.locations = [];
    Object.entries(filteredLocations.value).forEach(([i, value]) => {
      if (value.length < 1) return;
      const item = data.items[i];
      const locs = value.filter(({ id }) =>
        item.locations.find((location) => (location.id ?? location) === id)
      ).map((location) => ({
        ...location,
        sp: undefined,
        id: location.id.replace(
          data.title.replaceAll(' ', '_'),
          data.newTitle.replaceAll(' ', '_')
        )
      }));

      locs.forEach((location) => {
        if (!data.locations.find(({ id }) => id === location.id)) {
          data.locations.push(location);
        }
      });
      item.locations = locs.map(({ id }) => id);
    });

    if (!data.promo.main) data.promo.main = {};
    data.promo.main.default = defaultLocation.value;
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
  if (showcase.value.items.at(-1).servicesProvider !== '') {
    showcase.value.items.push({ plan: '', servicesProvider: '', locations: [] });
  }
}

const removeItem = (i) => {
  showcase.value.items.splice(i, 1);
}

const getPlanTitle = (uuid) => {
  const plans = store.getters['plans/all'] ?? [];

  return plans.find((plan) => plan.uuid === uuid)?.title ?? uuid;
}

const getProviderTitle = (uuid) => {
  return serviceProviders.value.find((provider) =>
    provider.uuid === uuid
  )?. title ?? uuid;
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
