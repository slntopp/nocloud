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
      <v-col>
        <v-select
          label="service provider"
          item-value="uuid"
          item-text="title"
          v-model="serviceProvider"
          :items="servicesProviders"
          :rules="rules.req"
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
      <v-col cols="4" md="4" lg="3">
        <v-list dence color="background-light">
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
          <v-btn class="ml-2" :loading="isTestsLoading" @click="testService">
            test
          </v-btn>
          <v-btn
            class="ml-2"
            :disabled="!testsPassed"
            :loading="isLoading"
            @click="createService"
          >
            create
          </v-btn>
        </div>
      </v-col>
      <v-col cols="8" md="8" lg="9">
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
                apply plan
              </v-btn>
            </template>
            <v-card>
              <v-card-title>Apply plan to group</v-card-title>
              <v-card-actions class="d-flex flex-column align-end">
                <v-select
                  dense
                  label="plan"
                  style="width: 200px"
                  v-model="currentInstancesGroups.plan"
                  :items="plans.titles"
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
            :rules="rules.req"
            v-model="instances[currentInstancesGroupsIndex].title"
            @change="(newVal) => (currentInstancesGroups.title = newVal)"
          />

          <v-select
            :items="types"
            v-model="currentInstancesGroups.body.type"
            label="type"
          />

          <v-select
            label="service provider"
            item-value="uuid"
            item-text="title"
            v-model="instances[currentInstancesGroupsIndex].sp"
            :items="servicesProviders"
            :rules="rules.req"
            @change="(newVal) => (currentInstancesGroups.sp = newVal)"
          />

          <component
            :is="templates[currentInstancesGroups.body.type]"
            :instances-group="JSON.stringify(currentInstancesGroups)"
            :planRules="planRules"
            :plans="plans"
            @update:instances-group="receiveObject"
          >
          </component>
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
    serviceProvider: "",
    instances: [],
    currentInstancesGroups: {},
    currentInstancesGroupsIndex: -1,
    types: ["ione", "custom"],
    templates: {},
    plans: {
      titles: [],
      list: [],
      products: []
    },

    isVisible: false,
    isLoading: false,
    isTestsLoading: false,
    plansVisible: false,
    testsPassed: false,
  }),
  mixins: [snackbar],
  methods: {
    addInstancesGroup(title = "") {
      if (this.instances.some((inst) => inst.title == "")) return;
      this.instances = [...this.instances, this.defaultInstance(title)];
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
    defaultInstance(title = "") {
      return {
        title,
        body: {
          type: "ione",
          resources: {
            ips_public: 0,
          },
        },
        plan: ''
      };
    },
    selectInstance(index = -1) {
      if (this.instances.length > 0) {
        this.currentInstancesGroups = JSON.parse(
          JSON.stringify(this.instances[index])
        );
      } else {
        this.currentInstancesGroups = {};
      }
      this.currentInstancesGroupsIndex = index;
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
        current.plan.includes(plan.title)
      );
      const [product] = Object.entries(plan.products)
        .find(([, prod]) =>
          prod.title === current.product
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
      const deploy_policies = { 0: this.serviceProvider }

      instances.forEach((inst) => {
        inst.body.resources.ips_public = inst.body.instances?.length || 0;
        data.instances_groups.push({
          ...inst.body,
          title: inst.title,
          sp: inst.sp
        });
        // console.log(data.instances_groups[inst.title])
        // console.log(data.instances_groups)
        // let ips = 0;
        // Object.keys(data.instances_groups[inst.title].instances).forEach(key => {
        // 	const item = data.instances_groups[inst.title].instances[key];
        // 	ips += item.resources.ips_public;
        // })
        // data.instances_groups[inst.title].resources.ips_public = ips;
      });
      return { namespace: this.namespace, service: data, deploy_policies };
    },
    createService() {
      const data = this.getService();

      if (!this.formValid) {
        this.$refs.form.validate();
        return;
      }

      this.isLoading = true;
      api.services
        ._create(data)
        .then(() => {
          this.showSnackbar({ message: `Service created successfully` });
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
    testService() {
      const data = this.getService();

      if (!this.formValid) {
        this.$refs.form.validate();
        return;
      }

      this.isTestsLoading = true;
      api.services
        .testConfig(data)
        .then((res) => {
          if (res.result) {
            this.showSnackbar({ message: `Service passed tests successfully` });
            this.testsPassed = true;
          } else {
            throw res;
          }
        })
        .catch((err) => {
          this.testsPassed = false;
          const opts = {
            message: err.errors.map((error) => error),
          };
          this.showSnackbarError(opts);
        })
        .finally(() => {
          this.isTestsLoading = false;
        });
    },
    setProducts() {
      const { plan } = this.currentInstancesGroups;
      const uuid = plan.split('(')[1]?.slice(0, 8);
      const products = this.plans.list.find((el) =>
        el.uuid.includes(uuid)
      )?.products || {};

      this.plans.products = [];
      delete this.currentInstancesGroups.product;
      delete this.currentInstancesGroups.products;
      Object.values(products).forEach(({ title }) => {
        this.plans.products.push(title);
      });
    }
  },
  computed: {
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    namespacesLoading() {
      return this.$store.getters["namespaces/isLoading"];
    },
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
    planRules() {
      return (this.plansVisible)
        ? this.rules.req
        : []
    }
  },
  created() {
    this.$store.dispatch("namespaces/fetch");
    this.$store.dispatch("servicesProviders/fetch");
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
    
    api.plans.list()
      .then((res) => res.pool
        .forEach((plan) => {
          const title = `${plan.title} (${plan.uuid.slice(0, 8)}...)`;

          this.plans.titles.push(title)
          this.plans.list.push(plan)
        })
      )

    api.settings
      .get(['instance-billing-plan-settings'])
      .then((res) => {
        const key = res['instance-billing-plan-settings']

        if (key) {
          this.plansVisible = JSON.parse(key).required
        } else {
          this.plansVisible = false
        }
      })
  },
  watch: {
    service: {
      handler() {
        this.testsPassed = false;
      },
      deep: true,
    },
  },
};
</script>

<style></style>
