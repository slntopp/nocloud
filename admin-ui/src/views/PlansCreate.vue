<template>
  <div class="pa-4">
    <h1 class="page__title">{{ title || 'Create' }} plan</h1>
    <v-form v-model="isValid" ref="form">
      <v-row>
        <v-col lg="6" cols="12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Plan type</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-select
                label="Type"
                v-model="plan.type"
                :items="types"
                :rules="generalRule"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Plan title</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                label="Title"
                v-model="plan.title"
                :rules="generalRule"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Public</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-switch v-model="plan.public" />
            </v-col>
          </v-row>

          <v-divider />

          <v-tabs v-model="form.title" background-color="background">
            <v-tab
              v-for="title of form.titles"
              :key="title"
            >
              {{ title }}
              <v-icon
                small
                right
                color="error"
                v-if="plan.type === 'custom'"
                @click="removeConfig(title)"
              >
                mdi-close
              </v-icon>
            </v-tab>
            <v-text-field
              dense
              outlined
              label="New config"
              class="ml-2 mt-1 mw-20"
              v-if="plan.type === 'custom' && isVisible"
              @change="addConfig"
            />
            <v-icon
              class="ml-2"
              v-else-if="plan.type === 'custom'"
              @click="isVisible = true"
            >
              mdi-plus
            </v-icon>
          </v-tabs>

          <v-divider />

          <v-tabs-items v-model="form.title">
            <v-tab-item
              v-for="(title, i) of form.titles"
              :key="title"
            >
              <plans-form
                :keyForm="title"
                :data="(plan.uuid)
                  ? plan.resources[i]
                  : null
                "
                @changeValue="(data) => changeValue(i, data)"
              />
            </v-tab-item>
          </v-tabs-items>
        </v-col>
      </v-row>
      
      <v-row>
        <v-col>
          <v-btn
            class="mr-2"
            color="background-light"
            :loading="isLoading"
            :disabled="!isTestSuccess"
            @click="tryToSend"
          >
            {{ title || 'Create' }}
          </v-btn>
          <v-btn
            class="mr-2"
            :color="testButtonColor"
            @click="testConfig"
          >
            Test
          </v-btn>
        </v-col>
      </v-row>
    </v-form>

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
import snackbar from '@/mixins/snackbar.js';
import PlansForm from '@/components/plans_form.vue';

export default {
  name: 'plansCreate-view',
  components: { PlansForm },
  mixins: [snackbar],
  props: ['title'],
  data: () => ({
    types: [],
    plan: {
      title: '',
      type: 'custom',
      public: false,
      resources: []
    },
    form: {
      title: '',
      titles: []
    },
    generalRule: [v => !!v || 'This field is required!'],

    isVisible: true,
    isValid: false,
    isLoading: false,
    isTestSuccess: false,
    testButtonColor: 'background-light',
  }),
  methods: {
    changeValue(num, { key, value }) {
      try {
        value = JSON.parse(value);
      } catch {
        value;
      }

      if (this.plan.resources[num]) {
        this.plan.resources[num][key] = value;
      } else {
        this.plan.resources.push({ [key]: value });
      }
    },
    addConfig(title) {
      this.form.titles.push(title);
      // this.form.title = title;
      this.isVisible = false;
    },
    removeConfig(title) {
      this.form.titles = this.form.titles
        .filter((el) => el !== title);

      if (this.form.titles.length <= 0) {
        this.isVisible = true;
      }
    },
    tryToSend() {
      if (!this.isValid) {
        this.$refs.form.validate();
        this.testButtonColor = 'background-light';
        this.isTestSuccess = false;

        return;
      }

      this.isLoading = true;

      const id = this.$route.params?.planId;
      const request = (this.title === 'Edit')
        ? api.plans.update(id, this.plan)
        : api.plans.create(this.plan);

      request.then(() => {
          this.showSnackbarSuccess({
            message: (this.title === 'Edit')
              ? 'Plan edited successfully'
              : 'Plan created successfully'
          });

          setTimeout(() => {
            this.$router.push({ name: 'Plans' });
          }, 1500);
        })
        .catch((err) => {
          this.showSnackbarError({
              message: err,
          });
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    testConfig() {
      if (!this.isValid) {
        this.$refs.form.validate();
        this.testButtonColor = 'background-light';
        this.isTestSuccess = false;

        this.showSnackbarError({
          message: 'Validation failed!',
        });

        return;
      }

      console.log(this.plan.resources);

      this.plan.resources.forEach((form, i, arr) => {
        arr[i].period = this.getTimestamp(form.date);
      });

      this.testButtonColor = 'success';
      this.isTestSuccess = true;
    },
    getTimestamp({ day, month, year, quarter, week, time }) {
      year = +year + 1970;
      month = +month + quarter * 3 + 1;
      day = +day + week * 7 + 1;

      if (`${day}`.length < 2) {
        day = '0' + day;
      }
      if (`${month}`.length < 2) {
        month = '0' + month;
      }

      return Date.parse(
        `${year}-${month}-${day}T${time}Z`
      ) / 1000;
    }
  },
  created() {
    const id = this.$route.params?.planId;
    const types = require.context(
      "@/components/modules/",
      true,
      /serviceProviders\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(
        /\.\/([A-Za-z0-9-_,\s]*)\/serviceProviders\.vue/i
      );
      if (matched && matched.length > 1) {
        const type = matched[1];
        this.types.push(type);
      }
    });

    if (id) this.$store.dispatch('plans/fetchItem', id);
  },
  watch: {
    'plan.type'() {
      switch (this.plan.type) {
        case 'ione':
          this.form.titles = ['CPU', 'RAM', 'IP public'];
          break;
        default:
          this.form.titles = [];
          break;
      }
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

.theme--dark.v-tabs-items {
  background: var(--v-background-base);
}

.mw-20 {
  max-width: 150px;
}
</style>