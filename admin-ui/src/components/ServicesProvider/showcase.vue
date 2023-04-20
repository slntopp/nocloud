<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row align="center" style="gap: 10px" class="mx-0" :class="{ 'mb-4': (isDisplay) ? false : true }">
      <v-col
        class="indigo darken-4"
        style="flex-grow: 0; flex-basis: 150px; border-radius: 10px; cursor: pointer"
        v-for="(item, key) in provider.meta.showcase"
        :key="key"
        :cols="viewport > 1600 ? 6 : 12"
        @click="editShowcase(key)"
      >
        <v-badge
          bordered
          icon="mdi-close"
          color="error"
          style="width: 100%"
          :ref="key"
        >
          <icon-title-preview
            class="align-center"
            :is-mdi="false"
            :title="item.title"
            :icon="toPascalCase(item.icon)"
          />
        </v-badge>
      </v-col>
      <v-col>
        <v-menu offset-y :close-on-content-click="false" @input="clearShowcase">
          <template v-slot:activator="{ on, attrs }">
            <v-icon ref="icon-plus" class="mr-2" v-bind="attrs" v-on="on">mdi-plus</v-icon>
          </template>

          <v-card class="pa-2">
            <v-text-field
              label="Title"
              v-model="showcase.title"
              :rules="generalRule"
            />
            <v-select
              label="Icon"
              :items="icons"
              :rules="generalRule"
              :value="toPascalCase(showcase.icon)"
              @change="(icon) => showcase.icon = toKebabCase(icon)"
            >
              <template v-slot:item="{ item }">
                <icon-title-preview :is-mdi="false" :title="item" :icon="item" />
              </template>
            </v-select>

            <v-select
              chips
              multiple
              item-text="title"
              item-value="uuid"
              label="Price models"
              v-model="showcase.billing_plans"
              :items="plans"
            >
              <template v-slot:selection="{ item, index }">
                <v-chip v-if="index === 0">
                  <span>{{ item.title }}</span>
                </v-chip>
                <span v-if="index === 1" class="grey--text text-caption">
                  (+{{ showcase.billing_plans.length - 1 }} others)
                </span>
              </template>
            </v-select>

            <v-btn :disabled="isDisabled" @click="addShowcase">
              {{ (beingEdited) ? 'Edit' : 'Add' }}
            </v-btn>
          </v-card>
        </v-menu>
      </v-col>
    </v-row>

    <v-btn v-if="!isDisplay" :loading="isLoading" @click="tryToSend">Save</v-btn>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import { toKebabCase, toPascalCase } from "@/functions.js";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";

export default {
  name: "services-provider-showcase",
  components: { IconTitlePreview },
  mixins: [snackbar],
  props: {
    template: { type: Object, required: true },
    isDisplay: { type: Boolean, default: false }
  },
  data: () => ({
    plans: [],
    provider: {},
    isLoading: false,
    beingEdited: null,
    showcase: { title: '', icon: '', billing_plans: [] },
    generalRule: [(v) => !!v || "This field is required!"],
  }),
  methods: {
    addShowcase() {
      const id = this.beingEdited ?? `${this.showcase.icon}-${Date.now()}`;

      if (!this.provider.meta.showcase) {
        this.provider.meta.showcase = {};
      }

      this.provider.meta.showcase[id] = JSON.parse(JSON.stringify(this.showcase));
      this.$emit('update:showcase', this.provider);
      this.clearShowcase();

      if (this.beingEdited) {
        this.beingEdited = null;
        return;
      }

      setTimeout(() => {
        const badge = this.$refs[id][0].$el.querySelector('.v-badge__badge');

        badge.style.cursor = 'pointer';
        badge.addEventListener('click', () => this.removeShowcase(id));
      }, 100);
    },
    editShowcase(key) {
      this.beingEdited = key;
      this.showcase = JSON.parse(JSON.stringify(this.provider.meta.showcase[key]));
      this.$refs['icon-plus'].$el.click();
    },
    removeShowcase(key) {
      this.$delete(this.provider.meta.showcase, key);
      this.provider.meta = Object.assign({}, this.provider.meta);
    },
    clearShowcase(isVisible) {
      if (isVisible && this.beingEdited) return;
      this.showcase = { title: '', icon: '', billing_plans: [] };
      this.beingEdited = null;
    },
    tryToSend() {
      const id = this.$route.params.uuid;

      this.isLoading = true;
      api.servicesProviders.update(id, this.provider)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Showcase changed successfully"
          });
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
          console.error(err);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    toKebabCase,
    toPascalCase,
  },
  created() {
    this.provider = JSON.parse(JSON.stringify(this.template));

    api.plans.list({ sp_uuid: this.provider.uuid })
      .then(({ pool }) => {
        this.plans = pool;
      })
      .catch((err) => {
        this.showSnackbarError({ message: err });
        console.error(err);
      });
  },
  mounted() {
    Object.keys(this.provider.meta.showcase ?? {}).forEach((id) => {
      const badge = this.$refs[id][0].$el.querySelector('.v-badge__badge');

      badge.style.cursor = 'pointer';
      badge.addEventListener('click', () => this.removeShowcase(id));
    });
  },
  computed: {
    viewport() {
      return document.documentElement.clientWidth;
    },
    icons() {
      const illustrations = require.context(
        "@ant-design/icons-vue/",
        true,
        /^.*\.js$/
      );
      const removedKeys = ["./", ".js", "Outlined"];

      return illustrations
        .keys()
        .map((icon) => {
          if (icon.includes("Filled") || icon.includes("TwoTone")) {
            return undefined;
          }

          removedKeys.forEach((key) => {
            icon = icon.replace(key, "");
          });

          if (icon.includes("/")) {
            return undefined;
          }

          return icon;
        })
        .filter((icon) => !!icon);
    },
    isDisabled() {
      return this.showcase.title === '' && this.showcase.icon === '';
    },
  }
}
</script>
