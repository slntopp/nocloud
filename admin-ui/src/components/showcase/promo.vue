<template>
  <div class="pa-10">
    <v-row align="center" justify="space-between">
      <v-col cols="6">
        <v-autocomplete
          multiple
          label="acceptable languages"
          v-model="promo.languages"
          :items="languages"
        />
      </v-col>
      <v-col cols="6">
        <v-select
          label="current language"
          v-model="language"
          :items="promo.languages"
        />
      </v-col>
    </v-row>
    <template v-if="language && Object.keys(currentLocation).length">
      <v-tabs
        class="rounded-t-lg"
        background-color="background-light"
        v-model="tabsIndex"
      >
        <v-tab>App</v-tab>
        <v-tab>Widget</v-tab>
      </v-tabs>
      <v-tabs-items
        class="rounded-b-lg"
        style="background: var(--v-background-light-base)"
        v-model="tabsIndex"
      >
        <v-tab-item>
          <v-text-field
            class="mt-3"
            hide-details
            v-model.trim="currentLocation.title"
            outlined
            label="Title"
          />
          <v-card-title class="d-flex align-center"
            >Service preview
            <v-switch
              class="ml-4"
              v-model.trim="currentLocation.previewEnable"
              outlined
            />
          </v-card-title>
          <rich-editor v-model="currentLocation.preview" />
        </v-tab-item>
        <v-tab-item>
          <div class="d-flex">
            <div class="widget__template">
              <div class="main">
                <div class="map">
                  <v-img height="190px" src="/admin/img/promo/map.svg" />
                </div>
                <div class="right-menus">
                  <div
                    :class="{
                      service: true,
                      active: activeWidgetPlace === 'service',
                    }"
                    @click="setActiveWidgetPlace('service')"
                  ></div>
                  <div
                    :class="{
                      location: true,
                      active: activeWidgetPlace === 'location',
                    }"
                    @click="setActiveWidgetPlace('location')"
                  ></div>
                </div>
              </div>
              <div class="footer">
                <div
                  :class="{
                    rewards: true,
                    active: activeWidgetPlace === 'rewards',
                  }"
                  @click="setActiveWidgetPlace('rewards')"
                ></div>
                <div
                  :class="{
                    offer: true,
                    active: activeWidgetPlace === 'offer',
                  }"
                  @click="setActiveWidgetPlace('offer')"
                ></div>
              </div>
            </div>
            <v-list color="background-light" class="px-4">
              <v-list-item
                v-for="item in widgetPlaces"
                :key="item.value"
                @click="setActiveWidgetPlace(item.value)"
                :class="{
                  active: activeWidgetPlace === item.value,
                }"
              >
                <v-list-item-content>
                  <v-list-item-title>{{ item.title }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </div>
          <template v-if="activeWidgetPlace === 'service'">
            <v-card-title class="text-center">Service settings</v-card-title>
            <v-text-field
              v-model.trim="currentLocation.service.title"
              outlined
              label="Title"
            />
            <v-text-field
              v-model.trim="currentLocation.service.btn"
              outlined
              label="Btn title"
            />
            <v-card-subtitle>Description:</v-card-subtitle>
            <rich-editor v-model="currentLocation.service.description" />
          </template>
          <template v-else-if="activeWidgetPlace === 'location'">
            <v-card-title class="text-center"
              >Location default settings</v-card-title
            >
            <v-text-field
              label="Title"
              outlined
              v-model="currentLocation.location.title"
            />
            <v-card-subtitle>Description:</v-card-subtitle>
            <rich-editor v-model="currentLocation.location.description" />
            <v-card-title>Individual location setting</v-card-title>
            <v-expansion-panels>
              <v-expansion-panel
                v-for="location in template?.locations"
                :key="getLocationKey(location)"
              >
                <v-expansion-panel-header color="background-light"
                  >{{ location.title }}
                </v-expansion-panel-header>
                <v-expansion-panel-content color="background-light">
                  <v-card-subtitle>Description:</v-card-subtitle>
                  <rich-editor
                    v-model="
                      currentLocation.locations[getLocationKey(location)]
                        .description
                    "
                  />
                </v-expansion-panel-content>
              </v-expansion-panel>
            </v-expansion-panels>
          </template>
          <template v-else-if="activeWidgetPlace === 'offer'">
            <v-card-title class="text-center">Offer settings</v-card-title>
            <v-text-field
              class="mt-5"
              outlined
              label="Media src"
              v-model="currentLocation.offer.src"
            ></v-text-field>
            <v-text-field
              outlined
              label="Media src link"
              v-model="currentLocation.offer.link"
            />
            <v-card-subtitle class="mt-3">Description:</v-card-subtitle>
            <rich-editor v-model="currentLocation.offer.text" />
          </template>
          <template v-if="activeWidgetPlace === 'rewards'">
            <v-card-title class="text-center">Rewards settings</v-card-title>
            <v-text-field
              label="Title"
              outlined
              v-model.trim="currentLocation.rewards.title"
            />
            <v-card-subtitle>Description:</v-card-subtitle>
            <rich-editor v-model="currentLocation.rewards.description" />
            <v-card-title>Icons:</v-card-title>
            <v-row>
              <v-col
                heigh="100%"
                xs="4"
                md="3"
                lg="2"
                xl="2"
                v-for="icon in currentLocation.icons"
                :key="icon.id"
              >
                <v-card color="background-light">
                  <v-img :src="icon.src" />
                  <v-divider />
                  <div class="d-flex flex-row-reverse">
                    <v-btn color="primary" @click="deleteIcon(icon.id)" icon>
                      <v-icon>mdi-delete</v-icon>
                    </v-btn>
                  </div>
                </v-card>
              </v-col>
              <v-col class="d-flex justify-center align-center" cols="1">
                <v-dialog max-width="600" v-model="addIconDialog" persistent>
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn
                      v-on="on"
                      v-bind="attrs"
                      block
                      width="50"
                      height="50"
                      color="background-light"
                    >
                      <v-icon size="64">mdi-plus</v-icon>
                    </v-btn>
                  </template>
                  <v-card
                    color="background-light"
                    class="pa-5 ma-auto"
                    max-width="600"
                  >
                    <v-card-title class="text-h5"> Add new icon:</v-card-title>
                    <v-text-field label="icon link" v-model="newIcon.file" />
                    <v-card-actions class="flex-row-reverse">
                      <v-btn class="mx-5" color="red" @click="closeAddIcon">
                        close
                      </v-btn>
                      <v-btn class="mx-5" color="primary" @click="addIcon">
                        add</v-btn
                      >
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-col>
            </v-row>
          </template>
        </v-tab-item>
      </v-tabs-items>
      <v-row class="mt-3 justify-end">
        <v-btn :loading="isSaveLoading" @click="save" color="primary"
          >Save</v-btn
        >
      </v-row>
    </template>
  </div>
</template>

<script>
export default {
  name: "promo-tab",
};
</script>

<script setup>
import api from "@/api";
import RichEditor from "@/components/ui/richEditor.vue";
import { onBeforeMount, toRefs, watch, ref } from "vue";
import { useStore } from "@/store";

const props = defineProps({ template: { type: Object, required: true } });

const store = useStore();

const { template } = toRefs(props);
const language = ref(null);
const languages = ref(["en", "ru", "pl"]);
const addIconDialog = ref(false);
const newIcon = ref({ file: null });
const isSaveLoading = ref(false);
const currentLocation = ref({});
const tabsIndex = ref(0);
const promo = ref({
  languages: ["en"],
});
const widgetPlaces = ref([
  { value: "service", title: "Service settings" },
  { value: "location", title: "Location settings" },
  { value: "rewards", title: "Rewards settings" },
  { value: "offer", title: "Special offer settings" },
]);
const activeWidgetPlace = ref("service");

onBeforeMount(() => {
  if (template.value.promo?.languages?.length) {
    promo.value.languages = template.value.promo.languages;
    language.value = promo.value.languages[0];
  } else {
    language.value = "en";
  }

  if (!template.value.promo.main) {
    promo.value.main = {};
  } else {
    promo.value.main = template.value.promo?.main;
  }
});

watch(language, (newValue, prevValue) => {
  if (!newValue) {
    return;
  }
  promo.value[prevValue] = currentLocation.value;

  if (!promo.value[language.value]) {
    promo.value[language.value] = {
      icons: [],
      service: {},
      location: {},
      locations: {},
      offer: { text: "", src: "", link: "" },
      rewards: { description: "", title: "" },
      preview: "",
      previewEnable: false,
      ...template.value.promo?.[language.value],
    };
  }

  currentLocation.value = promo.value[newValue];

  template.value?.locations.forEach((location) => {
    if (!promo.value[language.value]?.locations[getLocationKey(location)]) {
      promo.value[language.value].locations[getLocationKey(location)] = {
        description: "",
      };
    }
  });
});

const setActiveWidgetPlace = (value = "service") => {
  activeWidgetPlace.value = value;
};
const deleteIcon = (id) => {
  currentLocation.value.icons = currentLocation.value.icons.filter(
    (i) => i.id !== id
  );
};
const addIcon = () => {
  currentLocation.value.icons.push({
    id: currentLocation.value.icons.length,
    src: newIcon.value.file,
  });
  closeAddIcon();
};
const closeAddIcon = () => {
  newIcon.value = { file: null };
  addIconDialog.value = false;
};
const save = () => {
  isSaveLoading.value = true;

  const data = JSON.parse(JSON.stringify(template.value));
  const defaultKeys = ["languages", "main"];

  promo.value[language.value] = currentLocation.value;
  Object.keys(promo.value).forEach((key) => {
    if (!defaultKeys.includes(key) && !promo.value.languages?.includes(key)) {
      promo.value[key] = undefined;
    }
  });

  data.promo = promo.value;
  api.showcases
    .update(data)
    .then(() => {
      store.commit("snackbar/showSnackbarSuccess", {
        message: "Promo edited successfully",
      });
    })
    .catch((err) => {
      store.commit("snackbar/showSnackbarError", { message: err });
    })
    .finally(() => {
      isSaveLoading.value = false;
    });
};
const getLocationKey = (location) => {
  return location.id;
};
</script>

<style lang="scss" scoped>
.widget__template {
  display: flex;
  flex-direction: column;
  height: 275px;
  width: 350px;
  margin-bottom: 20px;
  .main {
    display: flex;
    flex-direction: row;
    height: 70%;
    .map {
      padding: 2px;
      width: 70%;
    }
    .right-menus {
      width: 30%;
      display: flex;
      flex-direction: column;
      .service {
        background-color: var(--v-background-dark-base);
        margin: 2px;
        border-radius: 5px;
        height: 50%;
        cursor: pointer;
      }
      .location {
        background-color: var(--v-background-dark-base);
        margin: 2px;
        border-radius: 5px;
        height: 50%;
        cursor: pointer;
      }
    }
  }
  .footer {
    display: flex;
    height: 30%;
    min-height: 75px;
    .rewards {
      background-color: var(--v-background-dark-base);
      margin: 2px;
      border-radius: 5px;
      width: 70%;
      cursor: pointer;
    }
    .offer {
      background-color: var(--v-background-dark-base);
      margin: 2px 2px 2px 6px;
      border-radius: 5px;
      width: 30%;
      cursor: pointer;
    }
  }
}
.active {
  background-color: var(--v-primary-base) !important;
}
</style>
