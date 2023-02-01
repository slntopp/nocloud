<template>
  <v-form v-model="formValid" ref="form">
    <v-row justify="start">
      <v-col cols="6" md="4" lg="3">
        <v-text-field
          label="service title"
          :rules="rules.req"
          v-model="service.title"
        />
      </v-col>
      <v-col cols="6" md="4" lg="3">
        <v-select
          label="namespace"
          :rules="rules.req"
          :items="namespaces"
          :loading="namespacesLoading"
          item-text="title"
          item-value="uuid"
          v-model="namespace"
        />
      </v-col>
      <v-col cols="6" md="4" lg="3">
        <v-text-field
          label="version"
          :rules="rules.req"
          v-model="service.version"
          readonly
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="4" md="4" lg="4">
        <v-list color="background-light">
          <v-subheader>instances groups</v-subheader>

          <v-list-item
            v-for="(instance, index) in instances"
            :key="index"
            @click="() => selectInstance(index)"
            :class="{
              'v-list-item--active': index == currentInstancesGroupsIndex,
            }"
          >
            <v-list-item-icon>
              <v-icon>mdi-playlist-star</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>{{
                instance.title || "(Enter title)"
              }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>

          <v-list-item @click="() => addInstancesGroup()">
            <v-list-item-icon>
              <v-icon>mdi-plus-circle-outline</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>add instances group</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list>

        <div class="btns__wrapper d-flex justify-end mt-4">
          <router-link :to="{ name: 'Services' }" style="text-decoration: none">
            <v-btn>cancel</v-btn>
          </router-link>
          <v-btn class="ml-2" @click="downloadJSON"> download </v-btn>
          <v-btn
            class="ml-2"
            :loading="isLoading"
            @click="createService"
          >
            Save
          </v-btn>
        </div>
      </v-col>
      <v-col cols="8" md="8" lg="8">
        <v-card
          v-if="currentInstancesGroupsIndex != -1"
          color="background-light"
          elevation="0"
          class="pa-4"
          :key="currentInstancesGroups.title"
        >
          <v-btn
            class="mb-4"
            @click="() => removeInstance(currentInstancesGroupsIndex)"
            >Remove</v-btn
          >
          <v-menu
            bottom
            offset-y
            transition="slide-y-transition"
            v-model="isVisible"
            :close-on-content-click="false"
          >
            <template v-slot:activator="{ on, attrs }">
              <v-btn class="mx-4 mb-4" v-on="on" v-bind="attrs">
                apply price model
              </v-btn>
            </template>
            <v-card>
              <v-card-title>Apply price model to group</v-card-title>
              <v-card-actions class="d-flex flex-column align-end">
                <v-select
                  dense
                  label="price model"
                  style="width: 200px"
                  item-text="title"
                  item-value="uuid"
                  v-model="currentInstancesGroups.plan"
                  :items="filteredPlans"
                  @change="setProducts"
                />
                <v-select
                  dense
                  label="product"
                  style="width: 200px"
                  v-model="currentInstancesGroups.product"
                  v-if="plans.products.length > 0"
                  :items="plans.products"
                />
                <v-btn @click="applyGroup">apply</v-btn>
              </v-card-actions>
            </v-card>
          </v-menu>

          <v-text-field
            label="instances group title"
            v-model="instances[currentInstancesGroupsIndex].title"
            :rules="rules.req"
            @change="(value) => (currentInstancesGroups.title = value)"
          />

          <v-select
            label="type"
            v-model="currentInstancesGroups.body.type"
            :items="types"
          />
          <v-text-field
            label="Type name"
            v-if="currentInstancesGroups.body.type === 'custom'"
            v-model="customTitles[currentInstancesGroupsIndex]"
            :rules="rules.req"
          />

          <v-select
            label="service provider"
            item-value="uuid"
            item-text="title"
            v-model="currentInstancesGroups.sp"
            :items="servicesProviders"
            :rules="rules.req"
          />

          <component
            :is="templates[currentInstancesGroups.body.type] ?? templates.custom"
            :instances-group="JSON.stringify(currentInstancesGroups)"
            :plans="{ list: filteredPlans, products: plans.products }"
            :planRules="planRules"
            :meta="meta"
            @update:instances-group="receiveObject"
            @changeMeta="(value) => meta = value"
          />
        </v-card>

        <v-card v-else color="background-light" elevation="0" class="pa-4">
          no instances group selected
        </v-card>
      </v-col>
    </v-row>

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
  </v-form>
</template>

<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";

import { downloadJSONFile } from "@/functions.js";

export default {
  name: "service-create",
  data: () => ({
    formValid: false,
    rules: {
      req: [(v) => !!v || "required field"],
      password: [
        (v) => !!v || "password required",
        (v) => v.length > 6 || "password must be at least 6 characters length",
      ],
      nubmer: [(v) => Number(v) == v || "must be a correct number"],
    },
    service: {
      version: "1",
      title: "",
      context: {},
      instances_groups: [],
    },
    namespace: "",
    instances: [],
    customTitles: {},
    currentInstancesGroups: {},
    currentInstancesGroupsIndex: -1,
    prevInstancesGroupsIndex: -1,
    types: ["ione", "custom"],
    templates: {},
    plans: {
      list: [],
      products: [],
    },
    meta: {},

    isVisible: false,
    isLoading: false,
    plansVisible: false,
  }),
  mixins: [snackbar],
  methods: {
    addInstancesGroup(title = "", type) {
      if (this.instances.some((inst) => inst.title == "")) return;
      this.instances = [...this.instances, this.defaultInstance(title, type)];
      if (this.instances.length == 1) this.selectInstance(0);
    },
    removeInstance(index) {
      this.instances.splice(index, 1);

      if (this.instances.length > 0) {
        this.selectInstance(Math.max(index - 1, 0));
      } else {
        this.selectInstance(-1);
      }
    },
    defaultInstance(title = "", type = "ione") {
      return {
        title,
        body: {
          type,
          resources: {
            ips_public: 0,
          },
        },
        plan: "",
      };
    },
    selectInstance(index = -1) {
      this.prevInstancesGroupsIndex = this.currentInstancesGroupsIndex;
      this.currentInstancesGroupsIndex = index;

      if (this.instances.length > 0) {
        this.currentInstancesGroups = JSON.parse(
          JSON.stringify(this.instances[index])
        );
      } else {
        this.currentInstancesGroups = {};
      }
    },
    receiveObject(newVal) {
      this.instances[this.currentInstancesGroupsIndex] = JSON.parse(newVal);
      this.selectInstance(this.currentInstancesGroupsIndex);
      this.testsPassed = false;
    },
    applyGroup() {
      const current = this.currentInstancesGroups;
      const instances = current.body.instances;
      const plan = this.plans.list.find((plan) =>
        current.plan.includes(plan.uuid)
      );
      const [product] =
        Object.entries(plan.products).find(
          ([, prod]) => prod.title === current.product
        ) || [];

      instances.forEach((inst) => {
        inst.billing_plan = plan;
        inst.plan = current.plan;
        inst.product = product;
        inst.productTitle = current.product;
        inst.products = this.plans.products;
      });

      this.instances[this.currentInstancesGroupsIndex] = current;
      this.isVisible = false;
    },
    getService() {
      const data = JSON.parse(JSON.stringify(this.service));
      const instances = JSON.parse(JSON.stringify(this.instances));

      instances.forEach((inst, i) => {
        if (inst.type === 'custom') {
          inst.body.type = this.customTitles[i];
        }

        inst.body.resources.ips_public = inst.body.instances?.length || 0;
        data.instances_groups.push({ ...inst.body, title: inst.title, sp: inst.sp });
        // console.log(data.instances_groups[inst.title])
        // console.log(data.instances_groups)
        // let ips = 0;
        // Object.keys(data.instances_groups[inst.title].instances).forEach(key => {
        // 	const item = data.instances_groups[inst.title].instances[key];
        // 	ips += item.resources.ips_public;
        // })
        // data.instances_groups[inst.title].resources.ips_public = ips;
      });
      return { namespace: this.namespace, service: data };
    },
    createService() {
      const action = (this.$route.params.serviceId) ? 'edit' : 'create';
      const data = this.getService();

      if (!this.formValid) {
        this.$refs.form.validate();
        return;
      }

      this.isLoading = true;
      api.services.testConfig(data)
        .then((res) => {
          if (res.result) return (action === 'create')
            ? api.services._create(data)
            : api.services._update(data.service);
          else throw res;
        })
        .then(() => {
          this.showSnackbarSuccess({ message: (action === 'create')
            ? "Service created successfully"
            : "Service updated successfully"
          });
          this.$router.push({ name: "Services" });
        })
        .catch((err) => {
          const opts = {
            message: err.errors.map((error) => error),
          };
          this.showSnackbarError(opts);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    setProducts() {
      const { plan } = this.currentInstancesGroups;
      const products =
        this.plans.list.find((el) => el.uuid.includes(plan.uuid))?.products ||
        {};

      this.plans.products = [];
      delete this.currentInstancesGroups.product;
      delete this.currentInstancesGroups.products;
      Object.values(products).forEach(({ title }) => {
        this.plans.products.push(title);
      });
    },
    downloadJSON() {
      const data = this.getService();
      const name = data.service.title
        ? (data.service.title + " service").replaceAll(" ", "_")
        : "unknown_service";

      downloadJSONFile(data, name);
    },
  },
  computed: {
    currentType() {
      const { type } = this.currentInstancesGroups.body;

      if (type === 'custom') {
        return this.customTitles[this.currentInstancesGroupsIndex];
      }
      return type;
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    namespacesLoading() {
      return this.$store.getters["namespaces/isLoading"];
    },
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"].filter(
        (el) => el.type === this.currentType
      );
    },
    filteredPlans() {
      return this.plans.list.filter((plan) => plan.type === this.currentType);
    },
    planRules() {
      return this.plansVisible ? this.rules.req : [];
    },
  },
  created() {
    this.$store.dispatch("services/fetch")
      .then(({ pool }) => {
        const service = pool.find((el) => el.uuid === this.$route.params.serviceId);
        const group = this.$route.params.type;
        const i = service?.instancesGroups.findIndex(({ type }) => type === group);
        const { instance } = this.$route.params;

        if (service) {
          this.service = service;
          this.namespace = service.access.namespace;

          this.instances = service.instancesGroups.map((group, i) => {
            if (!this.types.includes(group.type)) {
              this.customTitles[i] = group.type;
              group.type = 'custom';
            }
            return { title: group.title, sp: group.sp, body: group };
          });
        }

        if (instance) {
          this.selectInstance(i);
          setTimeout(() => {
            const top = -document.getElementsByTagName('header')[0].offsetHeight;

            document.getElementById(instance).scrollIntoView();
            window.scrollBy({ top });
          }, 300);
        } else if (group) {
          if (i !== -1) this.selectInstance(i);
          else {
            const type = (this.types.includes(group)) ? group : "custom";

            this.addInstancesGroup("", type);
            this.selectInstance(this.instances.length - 1);

            if (!this.types.includes(group)) {
              this.customTitles[this.currentInstancesGroupsIndex] = group;
            }
          }

          setTimeout(() => {
            const button = document.getElementById('button');

            button.click();
            button.scrollIntoView(true);
          }, 300);
        }
      });

    this.$store.dispatch("namespaces/fetch");
    this.$store.dispatch("servicesProviders/fetch", false);

    const types = require.context(
      "@/components/modules/",
      true,
      /serviceCreate\.vue$/
    );
    types.keys().forEach((key) => {
      const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/serviceCreate\.vue/i);
      if (matched && matched.length > 1) {
        const type = matched[1];
        this.types.push(type);
        this.templates[type] = () =>
          import(`@/components/modules/${type}/serviceCreate.vue`);
      }
    });

    api.plans.list().then((res) =>
      res.pool.forEach((plan) => {
        const end = (plan.uuid.length > 8) ? '...' : '';
        const title = `${plan.title} (${plan.uuid.slice(0, 8)}${end})`;

        this.plans.list.push({ ...plan, title });
      })
    );

    api.settings.get(["instance-billing-plan-settings"]).then((res) => {
      const key = res["instance-billing-plan-settings"];

      if (key) {
        this.plansVisible = JSON.parse(key).required;
      } else {
        this.plansVisible = false;
      }
    });
  },
  watch: {
    service: {
      handler() {
        this.testsPassed = false;
      },
      deep: true,
    },
    "currentInstancesGroups.body.type"() {
      const i = this.currentInstancesGroupsIndex;
      const j = this.prevInstancesGroupsIndex;
      if (i !== j) {
        this.prevInstancesGroupsIndex = this.currentInstancesGroupsIndex;
        return;
      }

      this.currentInstancesGroups.body.instances = [];
      this.currentInstancesGroups.sp = "";
    },
  },
};
</script>

<style>
</style>
