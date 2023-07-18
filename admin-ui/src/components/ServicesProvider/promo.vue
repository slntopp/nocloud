<template>
  <div class="pa-10">
    <v-row align="center" justify="space-between">
      <v-col cols="6">
        <v-autocomplete multiple label="acceptable languages" v-model="promo.languages" :items="languages"/>
      </v-col>
      <v-col cols="6">
        <v-select label="current language" v-model="language" :items="promo.languages"/>
      </v-col>
    </v-row>
    <template v-if="language && Object.keys(currentLocation).length">
      <v-card-title class="text-center">Service settings</v-card-title>
      <v-text-field v-model.trim="currentLocation.service.title" outlined label="Title"/>
      <v-text-field v-model.trim="currentLocation.service.btn" outlined label="Btn title"/>
      <v-card-subtitle>Description:</v-card-subtitle>
      <rich-editor v-model="currentLocation.service.description"/>
      <v-card-title class="text-center">Location default settings</v-card-title>
      <v-text-field label="Title" outlined v-model="currentLocation.location.title"/>
      <v-card-subtitle>Description:</v-card-subtitle>
      <rich-editor v-model="currentLocation.location.description"/>
      <v-card-title>Individual location setting</v-card-title>
      <v-expansion-panels>
        <v-expansion-panel
            v-for="location in template?.locations"
            :key="getLocationKey(location)"
        >
          <v-expansion-panel-header color="background-light">{{
              location.title
            }}
          </v-expansion-panel-header>
          <v-expansion-panel-content color="background-light">
            <v-card-subtitle>Description:</v-card-subtitle>
            <rich-editor
                v-model="currentLocation.locations[getLocationKey(location)].description"
            />
          </v-expansion-panel-content>
        </v-expansion-panel>
      </v-expansion-panels>
      <v-card-title class="text-center">Offer settings</v-card-title>
      <v-card-subtitle class="mt-3">Description:</v-card-subtitle>
      <rich-editor v-model="currentLocation.offer.text"/>
      <v-text-field
          class="mt-5"
          outlined
          label="Media src"
          v-model="currentLocation.offer.src"
      ></v-text-field>
      <v-text-field outlined label="Media src link" v-model="currentLocation.offer.link"/>
      <v-card-title class="text-center">Rewards settings</v-card-title>
      <v-text-field label="Title" outlined v-model.trim="currentLocation.rewards.title"/>
      <v-card-subtitle>Description:</v-card-subtitle>
      <rich-editor v-model="currentLocation.rewards.description"/>
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
            <v-img :src="icon.src"/>
            <v-divider/>
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
            <v-card color="background-light" class="pa-5 ma-auto" max-width="600">
              <v-card-title class="text-h5"> Add new icon:</v-card-title>
              <v-text-field label="icon link" v-model="newIcon.file"/>
              <v-card-actions class="flex-row-reverse">
                <v-btn class="mx-5" color="red" @click="closeAddIcon">
                  close
                </v-btn>
                <v-btn class="mx-5" color="primary" @click="addIcon"> add</v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>
        </v-col>
      </v-row>
      <v-row class="mt-3 justify-end">
        <v-btn :loading="isSaveLoading" @click="save" color="primary">Save</v-btn>
      </v-row>
    </template>
  </div>
</template>

<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";
import RichEditor from "@/components/ui/richEditor.vue";

export default {
  name: "promo-tab",
  components: {RichEditor},
  props: {template: {type: Object, required: true}},
  mixins: [snackbar],
  data: () => ({
    language: null,
    languages: ['en', 'ru', 'pl'],
    addIconDialog: false,
    newIcon: {file: null},
    isSaveLoading: false,
    promo:{
      languages:['en']
    },
    currentLocation:{}
  }),
  created() {
    if(this.template.meta.promo?.languages?.length){
      this.promo.languages=this.template.meta.promo.languages
      this.language=this.promo.languages[0]
    }else {
      this.language = 'en'
    }
  },
  watch: {
    language(newValue,prevValue) {
      if(!newValue){
        return
      }
      this.promo[prevValue]=this.currentLocation

      if(!this.promo[this.language]){
        this.promo[this.language] = {
          icons: [],
          service: {},
          location: {},
          locations: {},
          offer: {text: "", src: "", link: ""},
          rewards: {description: "", title: ""}, ...this.template.meta.promo?.[this.language]
        };
      }

      this.currentLocation=this.promo[newValue]

      this.template?.locations.forEach((location) => {
        if (!this.promo[this.language]?.locations[this.getLocationKey(location)]) {
          this.promo[this.language].locations[this.getLocationKey(location)] = {
            description: "",
          };
        }
      });
    }
  },
  methods: {
    deleteIcon(id) {
      this.currentLocation.icons = this.currentLocation.icons.filter((i) => i.id !== id);
    },
    addIcon() {
      this.currentLocation.icons.push({
        id: this.currentLocation.icons.length,
        src: this.newIcon.file,
      });
      this.closeAddIcon();
    },
    closeAddIcon() {
      this.newIcon = {file: null};
      this.addIconDialog = false;
    },
    save() {
      this.isSaveLoading = true;

      const data = JSON.parse(JSON.stringify(this.template));

      this.promo[this.language]=this.currentLocation
      Object.keys(this.promo).forEach(key=>{
        if(key!=='languages' && !this.promo.languages.includes(key)){
          this.promo[key]=undefined
        }
      })

      data.meta.promo = this.promo;
      api.servicesProviders
          .update(this.template.uuid, data)
          .then(() => {
            this.showSnackbarSuccess({
              message: "Promo edited successfully",
            });
          })
          .catch((err) => {
            this.showSnackbarError({message: err});
          })
          .finally(() => {
            this.isSaveLoading = false;
          });
    },
    getLocationKey(location) {
      return `${this.template.title} ${location.id}`;
    },
  },
};
</script>
