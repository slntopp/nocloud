<template>
  <div class="pa-10">
    <v-card-title class="text-center">Service settings</v-card-title>
    <v-text-field v-model.trim="promo.service.title" outlined label="Title" />
    <v-text-field v-model.trim="promo.service.btn" outlined label="Btn title" />
    <v-textarea
      v-model.trim="promo.service.description"
      outlined
      label="Description"
    />
    <v-card-title class="text-center">Location default settings</v-card-title>
    <v-text-field label="Title" outlined v-model="promo.location.title" />
    <v-textarea
      v-model.trim="promo.location.description"
      outlined
      label="Description"
    />
    <v-card-title>Individual location setting</v-card-title>
    <v-expansion-panels>
      <v-expansion-panel
        v-for="location in template?.locations"
        :key="getLocationKey(location)"
      >
        <v-expansion-panel-header color="background-light">{{
          location.title
        }}</v-expansion-panel-header>
        <v-expansion-panel-content color="background-light">
          <v-textarea
            outlined
            label="Description"
            v-model="promo.locations[getLocationKey(location)].description"
          ></v-textarea>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
    <v-card-title class="text-center">Offer settings</v-card-title>
    <v-textarea
      label="Text"
      outlined
      v-model.trim="promo.offer.text"
    ></v-textarea>
    <v-text-field
      outlined
      label="Media src"
      v-model="promo.offer.src"
    ></v-text-field>
    <v-text-field outlined label="Media src link" v-model="promo.offer.link" />
    <v-card-title class="text-center">Rewards settings</v-card-title>
    <v-text-field label="Title" outlined v-model.trim="promo.rewards.title" />
    <v-textarea
      outlined
      label="Description"
      v-model="promo.rewards.description"
    />
    <v-card-title>Icons:</v-card-title>
    <v-row>
      <v-col
        heigh="100%"
        xs="4"
        md="3"
        lg="2"
        xl="2"
        v-for="icon in promo.icons"
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
          <v-card color="background-light" class="pa-5 ma-auto" max-width="600">
            <v-card-title class="text-h5"> Add new icon: </v-card-title>
            <v-text-field label="icon link" v-model="newIcon.file" />
            <v-card-actions class="flex-row-reverse">
              <v-btn class="mx-5" color="red" @click="closeAddIcon">
                close
              </v-btn>
              <v-btn class="mx-5" color="primary" @click="addIcon"> add </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>
    <v-row class="mt-3 justify-end">
      <v-btn :loading="isSaveLoading" @click="save" color="primary">Save</v-btn>
    </v-row>
  </div>
</template>

<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";

export default {
  name: "promo-tab",
  props: { template: { type: Object, required: true } },
  mixins: [snackbar],
  data: () => ({
    addIconDialog: false,
    newIcon: { file: null },
    promo: {
      icons: [],
      service: {},
      location: {},
      locations: {},
      offer: { text: "", src: "", link: "" },
      rewards: { description: "", title: "" },
    },
    isSaveLoading: false,
  }),
  mounted() {
    if (this.template.meta.promo) {
      this.promo = { ...this.promo, ...this.template.meta.promo };
    }
    this.template?.locations.forEach((location) => {
      if (!this.promo.locations[this.getLocationKey(location)]) {
        this.promo.locations[this.getLocationKey(location)] = {
          description: "",
        };
      }
    });
  },
  methods: {
    deleteIcon(id) {
      this.promo.icons = this.promo.icons.filter((i) => i.id !== id);
    },
    addIcon() {
      this.promo.icons.push({
        id: this.promo.icons.length,
        src: this.newIcon.file,
      });
      this.closeAddIcon();
    },
    closeAddIcon() {
      this.newIcon = { file: null };
      this.addIconDialog = false;
    },
    save() {
      this.isSaveLoading = true;

      const data = JSON.parse(JSON.stringify(this.template));
      data.meta.promo = this.promo;

      api.servicesProviders
        .update(this.template.uuid, data)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Promo edited successfully",
          });
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
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
