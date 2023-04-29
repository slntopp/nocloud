<template>
  <div class="pa-4">
    <div class="d-flex">
      <h1 class="page__title">Showcases</h1>
    </div>

    <v-row>
      <v-col lg="6" cols="12" v-for="(showcase, key) in showcases" :key="key">
        <v-card color="background">
          <v-card-title>{{ showcase.title }}:</v-card-title>
          <v-card-text>
            <v-text-field
              label="Title"
              :rules="generalRule"
              :value="showcase.title"
              @change="(value) => updateShowcase(key, 'title', value)"
            />
            <v-select
              label="Icon"
              :items="icons"
              :rules="generalRule"
              :value="toPascalCase(showcase.icon)"
              @change="(icon) => updateShowcase(key, 'icon', toKebabCase(icon))"
            >
              <template v-slot:item="{ item }">
                <icon-title-preview :is-mdi="false" :title="item" :icon="item" />
              </template>
            </v-select>

            <v-select
              dense
              chips
              multiple
              item-text="title"
              item-value="uuid"
              label="Price models"
              :items="plans"
              :value="showcase.billing_plans"
              @change="(value) => updateShowcase(key, 'billing_plans', value)"
            >
              <template v-slot:selection="{ item, index }">
                <v-chip small v-if="index === 0">
                  <span>{{ item.title }}</span>
                </v-chip>
                <span v-if="index === 1" class="grey--text text-caption">
                  (+{{ showcase.billing_plans.length - 1 }} others)
                </span>
              </template>
            </v-select>

            <v-select
              dense
              chips
              multiple
              return-object
              item-text="title"
              label="Services providers"
              :items="sp"
              :value="showcase.sp"
              @change="(value) => updateShowcase(key, 'sp', value)"
            >
              <template v-slot:selection="{ item, index }">
                <v-chip small v-if="index === 0">
                  <span>{{ item.title }}</span>
                </v-chip>
                <span v-if="index === 1" class="grey--text text-caption">
                  (+{{ showcase.billing_plans.length - 1 }} others)
                </span>
              </template>
            </v-select>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-btn class="mt-4" :isLoading="isLoading" @click="tryToSend">Save</v-btn>

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
  </div>
</template>

<script>
import api from '@/api.js';
import snackbar from "@/mixins/snackbar.js";
import { toKebabCase, toPascalCase } from "@/functions.js";
import IconTitlePreview from "@/components/ui/iconTitlePreview.vue";

export default {
  name: 'showcases-view',
  mixins: [snackbar],
  components: { IconTitlePreview },
  data: () => ({
    updated: [],
    isLoading: false,
    generalRule: [(v) => !!v || "This field is required!"]
  }),
  methods: {
    updateShowcase(showcase, key, value) {
      const sp = (key === 'sp') ? value : this.showcases[showcase].sp;

      sp.forEach((provider) => {
        if (key === 'sp') {
          const value = JSON.parse(JSON.stringify(this.showcases[showcase]));

          delete value.sp;
          provider.meta.showcase[showcase] = value;
        } else {
          provider.meta.showcase[showcase][key] = value;
        }
        this.$store.commit('servicesProviders/updateService', provider);

        if (this.updated.find(({ uuid }) => uuid === provider.uuid)) return;
        this.updated.push(provider);
      });
    },
    tryToSend() {
      const promises = this.updated.map((provider) =>
        api.servicesProviders.update(provider.uuid, provider)
      );

      this.isLoading = true;
      Promise.all(promises).then(() => {
        this.showSnackbarSuccess({
          message: "Showcases changed successfully"
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
    toPascalCase
  },
  created() {
    if (this.sp.length > 0) return;
    Promise.all([
      this.$store.dispatch("servicesProviders/fetch"),
      this.$store.dispatch("plans/fetch")
    ])
      .catch((err) => {
        this.showSnackbarError({ message: err });
        console.error(err);
      });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "servicesProviders/fetch",
    });
  },
  computed: {
    sp() {
      return this.$store.getters['servicesProviders/all'];
    },
    plans() {
      return this.$store.getters['plans/all'];
    },
    showcases() {
      const result = {};

      this.sp.forEach((provider) => {
        const showcase = JSON.parse(JSON.stringify(provider.meta.showcase ?? {}));

        Object.keys(showcase).forEach((key) => {
          showcase[key].sp = [provider, ...(result[key]?.sp ?? [])];
        });
        Object.assign(result, showcase);
      });

      return result;
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
    }
  }
}
</script>

<style scoped>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
