<template>
  <div class="pa-4">
    <div class="d-flex">
      <h1 class="page__title">Showcases</h1>
    </div>

    <v-menu
      offset-y
      :close-on-content-click="false"
      :close-on-click="!isOpen"
      @input="clearShowcase"
    >
      <template v-slot:activator="{ on, attrs }">
        <v-btn class="mr-2" ref="create" v-bind="attrs" v-on="on">Create</v-btn>
      </template>

      <v-card class="pa-2">
        <v-select
          label="Icon"
          :items="icons"
          :rules="generalRule"
          :value="toPascalCase(newShowcase.icon)"
          @change="(icon) => newShowcase.icon = toKebabCase(icon)"
        >
          <template v-slot:item="{ item }">
            <icon-title-preview :is-mdi="false" :title="item" :icon="item" />
          </template>
        </v-select>

        <v-select
          dense
          chips
          multiple
          return-object
          item-text="title"
          label="Services providers"
          v-model="newShowcase.sp"
          :items="sp"
          @focus="isOpen = true"
          @blur="changeOpen"
        >
          <template v-slot:selection="{ item, index }">
            <v-chip small v-if="index === 0">
              <span>{{ item.title }}</span>
            </v-chip>
            <span v-if="index === 1" class="grey--text text-caption">
              (+{{ newShowcase.sp.length - 1 }} others)
            </span>
          </template>
        </v-select>

        <v-btn :disabled="isDisabled" @click="addShowcase">Add</v-btn>
      </v-card>
    </v-menu>

    <v-row v-if="Object.keys(showcases).length > 0">
      <v-col lg="6" cols="12" v-for="(showcase, key) in showcases" :key="key">
        <v-card color="background">
          <v-card-title>
            {{ showcase.title }}:
            <confirm-dialog @confirm="updateShowcase(key, 'sp', [])">
              <v-icon class="ml-2" color="error" style="cursor: pointer">mdi-close-circle</v-icon>
            </confirm-dialog>
          </v-card-title>

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
              :items="getPlans(showcase.sp)"
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
              persistent-hint
              item-text="title"
              label="Services providers"
              hint="If you clear the list of providers, the showcase will automatically be deleted."
              :items="sp"
              :value="showcase.sp"
              @change="(value) => updateShowcase(key, 'sp', value)"
            >
              <template v-slot:selection="{ item, index }">
                <v-chip small v-if="index === 0">
                  <span>{{ item.title }}</span>
                </v-chip>
                <span v-if="index === 1" class="grey--text text-caption">
                  (+{{ showcase.sp.length - 1 }} others)
                </span>
              </template>
            </v-select>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-btn
      :class="(Object.keys(showcases).length > 0) ? 'mt-4' : null"
      :isLoading="isLoading"
      @click="tryToSend"
    >
      Save
    </v-btn>
  </div>
</template>

<script>
import api from '@/api.js';
import snackbar from '@/mixins/snackbar.js';
import { toKebabCase, toPascalCase } from '@/functions.js';
import IconTitlePreview from '@/components/ui/iconTitlePreview.vue';
import ConfirmDialog from '@/components/confirmDialog.vue';

export default {
  name: 'showcases-view',
  mixins: [snackbar],
  components: { IconTitlePreview, ConfirmDialog },
  data: () => ({
    newShowcase: {
      title: 'Title',
      icon: '',
      billing_plans: [],
      sp: []
    },
    updated: [],
    isOpen: false,
    isLoading: false,
    generalRule: [(v) => !!v || "This field is required!"]
  }),
  methods: {
    updateShowcase(id, key, value) {
      if (key === 'sp') {
        let provider = null;

        if (value.length < this.showcases[id].sp.length) {
          provider = JSON.parse(JSON.stringify(
            this.showcases[id].sp.find(
              ({ uuid }) => !value.find((el) => uuid === el.uuid)
            )
          ));

          delete provider.meta.showcase[id];
        } else {
          const showcase = JSON.parse(JSON.stringify(this.showcases[id]));

          delete showcase.sp;
          provider = JSON.parse(JSON.stringify(
            value.find(({ uuid }) =>
              !this.showcases[id].sp.find((el) => uuid === el.uuid)
            )
          ));

          if (!provider.meta.showcase) provider.meta.showcase = {};
          provider.meta.showcase[id] = showcase;
        }

        const index = this.updated.findIndex(({ uuid }) => uuid === provider.uuid);
        this.$store.commit('servicesProviders/updateService', provider);

        if (index === -1) {
          this.updated.push(provider);
        } else {
          this.updated.splice(index, 1, provider);
        }
        return;
      }

      this.showcases[id].sp.forEach((el) => {
        const provider = JSON.parse(JSON.stringify(el));
        const index = this.updated.findIndex(({ uuid }) => uuid === provider.uuid);

        provider.meta.showcase[id][key] = value;
        this.$store.commit('servicesProviders/updateService', provider);

        if (index === -1) {
          this.updated.push(provider);
        } else {
          this.updated.splice(index, 1, provider);
        }
      });
    },
    addShowcase() {
      const id = `${this.newShowcase.icon}-${Date.now()}`;

      this.newShowcase.sp.forEach((el) => {
        const provider = JSON.parse(JSON.stringify(el));
        const index = this.updated.findIndex(({ uuid }) => uuid === provider.uuid);

        if (!provider.meta.showcase) provider.meta.showcase = {};
        provider.meta.showcase[id] = {
          title: this.newShowcase.title,
          icon: this.newShowcase.icon,
          billing_plans: []
        };
        this.$store.commit('servicesProviders/updateService', provider);
        this.$refs.create.$el.click();

        if (index === -1) {
          this.updated.push(provider);
        } else {
          this.updated.splice(index, 1, provider);
        }
      });
    },
    clearShowcase(isVisible) {
      if (isVisible) return;
      this.newShowcase = { title: 'Title', icon: '', billing_plans: [], sp: [] };
    },
    changeOpen() {
      setTimeout(() => { this.isOpen = false }, 100);
    },
    getPlans(sp) {
      return this.plans.filter(({ uuid }) =>
        sp.find(({ meta }) => meta.plans?.includes(uuid))
      );
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
    const promises = [];

    if (this.sp.length < 1) {
      promises.push(this.$store.dispatch("servicesProviders/fetch", false));
    }
    if (this.plans.length < 1) {
      promises.push(this.$store.dispatch("plans/fetch"));
    }

    Promise.all(promises)
      .catch((err) => {
        this.showSnackbarError({ message: err });
        console.error(err);
      });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "servicesProviders/fetch",
      params: false
    });
  },
  computed: {
    sp() {
      return this.$store.getters['servicesProviders/all'].filter(
        ({ locations, type }) => locations.length > 0 || !['ione', 'ovh'].includes(type)
      );
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
    },
    isDisabled() {
      return this.newShowcase.icon === '' || this.newShowcase.sp.length < 1;
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
