<template>
  <div class="pa-10">
    <v-card-title class="text-center">{{ sp.title }}</v-card-title>
    <v-textarea outlined label="Description:" />
    <v-card-title>Icons:</v-card-title>
    <v-row>
      <v-col
        heigh="100%"
        xs="4"
        md="3"
        lg="2"
        xl="2"
        v-for="icon in icons"
        :key="icon.id"
      >
        <v-card height="100%" color="background-light">
          <v-img cover :src="icon.img" />
          <v-divider />
          <div class="d-flex flex-row-reverse">
            <v-btn color="primary" @click="deleteIcon(icon.id)" icon>
              <v-icon>mdi-delete</v-icon>
            </v-btn>
          </div>
        </v-card>
      </v-col>
      <v-col class="d-flex justify-center align-center" cols="1">
        <v-dialog v-model="addIconDialog" persistent>
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
            <v-file-input
              accept="image/*"
              v-model="newIcon.file"
              clearable
              label="File input"
              underlined
            />
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
      <v-btn color="primary">Save</v-btn>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "promo-tab",
  props: { template: { type: Object, required: true } },
  data: () => ({
    sp: null,
    addIconDialog: false,
    newIcon: { file: null },

    icons: [
      {
        id: "1",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      {
        id: "2",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      {
        id: "3",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      {
        id: "4",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      {
        id: "5",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      {
        id: "6",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      {
        id: "7",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      { id: "8", img: "https://cdn.vuetifyjs.com/images/cards/cooking.png" },
      {
        id: "912",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
      {
        id: "212",
        img: "https://cdn.vuetifyjs.com/images/cards/cooking.png",
      },
    ],
  }),
  methods: {
    deleteIcon(id) {
      console.log(id);
    },
    addIcon() {
      console.log(this.newIcon);
    },
    closeAddIcon() {
      this.newIcon = { file: null };
      this.addIconDialog = false;
    },
  },
  mounted() {
    console.log(this.template);
    this.sp = this.template;
  },
};
</script>
