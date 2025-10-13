<template>
  <div class="pa-10 h-100 w-100">
    <h1 class="page__title" v-if="!isEdit">Create showcase</h1>
    <v-form ref="showcaseForm" align="center">
      <v-row>
        <v-col cols="4">
          <v-text-field
            :rules="[requiredRule]"
            v-model="showcase.newTitle"
            label="Name"
          />
        </v-col>
        <v-col cols="4">
          <v-text-field
            :rules="[requiredRule]"
            @input="setTitle"
            :value="showcase.promo?.[currentLang]?.title || ''"
            label="Title"
          />
        </v-col>
        <v-col cols="1">
          <v-autocomplete :items="langs" v-model="currentLang"></v-autocomplete>
        </v-col>
        <v-col
          cols="3"
          style="display: flex; gap: 30px; justify-content: flex-end"
        >
          <v-switch label="Is primary" v-model="showcase.primary" />
          <v-switch label="Enabled" v-model="showcase.public" />
        </v-col>
        <v-col cols="3">
          <icons-autocomplete
            label="Preview icon"
            :value="showcase.icon"
            @input:value="showcase.icon = $event"
          />
        </v-col>
        <v-col cols="2">
          <color-picker label="Color" v-model="showcase.meta.iconColor" />
        </v-col>
        <v-col cols="3">
          <v-autocomplete
            clearable
            item-text="title"
            item-value="id"
            label="Default location"
            v-model="defaultLocation"
            :items="allLocations"
          />
        </v-col>

        <v-col cols="2">
          <v-select v-model="showcase.meta.type" label="Type" :items="types" />
        </v-col>

        <v-col cols="2">
          <v-switch v-model="showcase.meta.isNew" label="Is new?" />
        </v-col>
      </v-row>

      <v-expansion-panels :value="0">
        <v-expansion-panel v-for="(item, i) in showcase.items" :key="i">
          <v-expansion-panel-header color="background">
            {{ getProviderTitle(item.servicesProvider) }}
            -
            {{ getPlan(item.servicesProvider, item.plan)?.title || item.plan }}

            <v-icon
              style="flex: 0 0 auto; margin: 0 auto 0 10px"
              color="error"
              v-if="
                !(
                  i === showcase.items.length - 1 ||
                  item.servicesProvider === ''
                )
              "
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
                  :loading="isPlansLoading"
                  :items="
                    isPlansLoading
                      ? []
                      : plansBySpMap.get(item.servicesProvider)
                  "
                />
              </v-col>
              <v-col cols="6">
                <locations-autocomplete
                  label="Locations"
                  v-model="item.locations"
                  :loading="isPlansLoading"
                  :locations="isPlansLoading ? [] : filteredLocations[i]"
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
import ColorPicker from "@/components/ui/colorPicker.vue";

const props = defineProps({
  realShowcase: {},
  isEdit: { type: Boolean, default: false },
});
const { realShowcase, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const types = [
  "cloud",
  "custom",
  "virtual",
  "openai",
  "vpn",
  "ione-vpn",
  "bots",
  "domains",
  "acronis",
  "ssl",
  "b24-apps",
];

const showcase = ref({
  primary: false,
  title: "",
  newTitle: "",
  icon: "",
  items: [
    {
      plan: "",
      servicesProvider: "",
      locations: [],
    },
  ],
  promo: {},
  locations: [],
  public: true,
  meta: {
    type: "",
    iconColor: "",
  },
});

const currentLang = ref("en");
const langs = ["en", "ru", "pl"];

const isLoading = ref(false);
const defaultLocation = ref("");
const isSaveLoading = ref(false);
const isPlansLoading = ref(false);

const plansBySpMap = ref(new Map());

const requiredRule = ref((val) => !!val || "Required field");
const serviceProviders = computed(() => store.getters["servicesProviders/all"]);

const locations = computed(() =>
  showcase.value.items.reduce((result, { servicesProvider }, i) => {
    const { uuid, locations = [] } =
      serviceProviders.value?.find((sp) => sp.uuid === servicesProvider) ?? {};

    return {
      ...result,
      [i]: locations.map((location) => ({
        ...location,
        sp: uuid,
        id: getNewLocationKey(location),
      })),
    };
  }, {})
);

const filteredLocations = computed(() => {
  if (isPlansLoading.value || isLoading.value) {
    return {};
  }

  const result = {};

  Object.entries(locations.value).forEach(([i, value]) => {
    const plan = getPlan(
      showcase.value.items[i].servicesProvider,
      showcase.value.items[i].plan
    );

    if (!plan) return;
    result[i] = value.filter(({ type }) => plan.type.split("-")[0] === type);
  });

  return result;
});

const allLocations = computed(() =>
  Object.entries(filteredLocations.value).reduce(
    (result, [i, locations]) => [
      ...result,
      ...locations.filter(({ id }) =>
        showcase.value.items[i].locations?.find(
          (location) => id === (location.id ?? location)
        )
      ),
    ],
    []
  )
);

watch(realShowcase, () => {
  defaultLocation.value = realShowcase.value.main?.default ?? "";
  showcase.value = JSON.parse(JSON.stringify(realShowcase.value));
  showcase.value.newTitle = showcase.value.title;

  if (!showcase.value.meta) {
    showcase.value.meta = {};
  }

  if (!Array.isArray(showcase.value.items)) {
    showcase.value.items = [];
  }
  showcase.value.items.push({ plan: "", servicesProvider: "", locations: [] });
});

onMounted(async () => {
  try {
    isLoading.value = true;
    await Promise.all([
      store.dispatch("servicesProviders/fetch", { anonymously: true }),
    ]);
  } catch (e) {
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
      const locs = value
        .filter(({ id }) =>
          item.locations?.find((location) => (location.id ?? location) === id)
        )
        .map((location) => ({
          ...location,
          sp: undefined,
          id: location.id.replace(
            data.title.replaceAll(" ", "_"),
            data.newTitle.replaceAll(" ", "_")
          ),
        }));

      locs.forEach((location) => {
        if (!data.locations?.find(({ id }) => id === location.id)) {
          data.locations.push(location);
        }
      });
      item.locations = locs.map(({ id }) => id);
    });

    if (!data.main) data.main = {};
    data.main.default = defaultLocation.value;
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

    if (!isEdit.value) {
      router.push({ name: "Showcases" });
    }
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

const setTitle = (value) => {
  if (!showcase.value.promo[currentLang.value]) {
    showcase.value.promo[currentLang.value] = {};
  }

  showcase.value.promo[currentLang.value].title = value;
};

const addItem = () => {
  if (showcase.value.items.at(-1).servicesProvider !== "") {
    showcase.value.items.push({
      plan: "",
      servicesProvider: "",
      locations: [],
    });
  }
};

const removeItem = (i) => {
  showcase.value.items.splice(i, 1);
};

const getPlan = (sp, uuid) => {
  const plans = plansBySpMap.value.get(sp) ?? [];

  if (Array.isArray(plans)) {
    return plans?.find((plan) => plan.uuid === uuid);
  }
};

const getProviderTitle = (uuid) => {
  return (
    serviceProviders.value?.find((provider) => provider.uuid === uuid)?.title ??
    uuid
  );
};

const fetchPlans = async () => {
  isPlansLoading.value = true;

  try {
    await Promise.all(
      showcase.value.items.map(async (item) => {
        let sp = item.servicesProvider;

        try {
          if (sp && !plansBySpMap.value.has(sp)) {
            plansBySpMap.value.set(
              sp,
              store.getters["plans/plansClient"].listPlans({
                spUuid: sp,
              })
            );

            const data = await plansBySpMap.value.get(sp);
            plansBySpMap.value.set(sp, data.toJson().pool);
          }
        } catch {
          plansBySpMap.value.delete(sp);
        }
      })
    );
  } finally {
    setTimeout(() => {
      isPlansLoading.value = false;
    }, 1);
  }
};

watch(() => showcase.value.items, fetchPlans, { deep: true });
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
