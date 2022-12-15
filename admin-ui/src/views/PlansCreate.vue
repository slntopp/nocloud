<template>
  <div class="pa-4">
    <div class="d-flex">
      <h1 class="page__title" v-if="!item">Create price model</h1>
      <v-icon class="mx-3" large color="light" @click="openPlanWiki">
        mdi-information-outline
      </v-icon>
    </div>

    <v-form v-model="isValid" ref="form">
      <v-row>
        <v-col lg="6" cols="12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Price model type</v-subheader>
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
              <v-subheader>Price model title</v-subheader>
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
              <v-subheader>Price model kind</v-subheader>
            </v-col>
            <v-col cols="9">
              <confirm-dialog @cancel="changePlan(true)" @confirm="changePlan">
                <v-radio-group v-model="selectedKind" row mandatory>
                  <v-radio
                    v-for="item in kinds"
                    :key="item"
                    :value="item"
                    :label="item.toLowerCase()"
                  />
                </v-radio-group>
              </confirm-dialog>
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

          <template v-if="!['ovh', 'goget'].includes(plan.type)">
            <v-tabs v-model="form.title" background-color="background-light">
              <v-tab
                draggable="true"
                active-class="background"
                v-for="(title, i) of form.titles"
                :key="title"
                @drag="(e) => dragTab(e, i)"
                @dragstart="dragTabStart"
                @dragend="dragTabEnd"
                @dblclick="edit = { isVisible: true, title }"
              >
                {{ title }}
                <v-icon small right color="error" @click="removeConfig(title)">
                  mdi-close
                </v-icon>
              </v-tab>
              <v-text-field
                dense
                outlined
                :label="edit.isVisible ? `Edit ${edit.title}` : 'New config'"
                class="ml-2 mt-1 mw-20"
                v-if="isVisible || edit.isVisible"
                @change="addConfig"
              />
              <v-icon v-else class="ml-2" @click="isVisible = true">
                mdi-plus
              </v-icon>
            </v-tabs>

            <v-divider />

            <v-subheader v-if="form.titles.length > 0">
              To edit the title, double-click the LMB
            </v-subheader>

            <v-tabs-items v-model="form.title">
              <v-tab-item v-for="(title, i) of form.titles" :key="title">
                <component
                  :is="template"
                  :keyForm="title"
                  :resource="plan.resources[i]"
                  :product="getProduct(i)"
                  :preset="preset(i)"
                  @change:resource="(data) => changeResource(i, data)"
                  @change:product="(data) => changeProduct(title, data)"
                />
              </v-tab-item>
            </v-tabs-items>
          </template>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-btn
            class="mr-2"
            color="background-light"
            v-if="isEdit"
            :disabled="!isTestSuccess"
            @click="isDialogVisible = true"
          >
            Save
          </v-btn>
          <v-btn
            v-else
            class="mr-2"
            color="background-light"
            :loading="isLoading"
            @click="tryToSend('create')"
          >
            Create
          </v-btn>
          <v-btn :color="testButtonColor" @click="testConfig">Test</v-btn>
        </v-col>
        <v-col>
          <v-dialog :max-width="600" v-model="isDialogVisible">
            <v-card color="background-light">
              <v-card-title>Do you really want to change your current price model?</v-card-title>
              <v-card-subtitle>You can also create a new price model based on the current one.</v-card-subtitle>
              <v-card-actions>
                <v-btn class="mr-2" :loading="isLoading" @click="tryToSend('create')">
                  Create
                </v-btn>
                <v-btn v-if="item" :loading="isLoading" @click="tryToSend('edit')">
                  Edit
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>
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
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import ConfirmDialog from "@/components/confirmDialog.vue";
import PlanOpensrs from "@/components/plan/opensrs/planOpensrs.vue";

export default {
  name: "plansCreate-view",
  mixins: [snackbar],
  components: { ConfirmDialog, PlanOpensrs },
  props: { item: { type: Object }, isEdit: { type: Boolean, default: false } },
  data: () => ({
    types: [],
    kinds: ["DYNAMIC", "STATIC"],
    selectedKind: "",
    products: [],
    plan: {
      title: "",
      type: "custom",
      kind: "DYNAMIC",
      public: true,
      resources: [],
      products: {},
      fee: null,
    },
    form: {
      title: "",
      titles: [],
    },
    edit: {
      isVisible: false,
      title: "",
    },
    generalRule: [(v) => !!v || "This field is required!"],

    isDialogVisible: false,
    isVisible: true,
    isValid: false,
    isFeeValid: true,
    isLoading: false,
    isTestSuccess: false,
    testButtonColor: "background-light",
  }),
  methods: {
    changeResource(num, { key, value }) {
      try {
        value = JSON.parse(value);
      } catch {
        value;
      }

      if (key === "date") {
        this.setPeriod(value, num);
        return;
      }
      if (this.plan.resources[num]) {
        this.plan.resources[num][key] = value;
      } else {
        this.plan.resources.push({ [key]: value });
      }
    },
    changeProduct(obj, { key, value }) {
      try {
        value = JSON.parse(value);
      } catch {
        value;
      }

      if (key === "date") {
        this.setPeriod(value, obj);
        return;
      } else if (key === "resources") {
        this.plan.resources = value;
        return;
      } else if (key === "amount") {
        key = "resources";
      }

      if (this.plan.products[obj]) {
        this.plan.products[obj][key] = value;
      } else {
        this.plan.products[obj] = { [key]: value };
      }
    },
    changeFee({ key, value }) {
      try {
        value = JSON.parse(value);
      } catch {
        value;
      }

      this.plan.fee[key] = value;
    },
    preset(i) {
      const title = this.form.titles[i - 1];

      if (this.plan.products[title]) {
        return this.plan.products[title].resources;
      }
      if (this.plan.type === "custom") return;
      return {
        cpu: 1,
        ram: 1024,
        ip_public: 0,
      };
    },
    dragTabStart(e) {
      const el = document.createElement("div");

      e.dataTransfer.dropEffect = "move";
      e.dataTransfer.effectAllowed = "move";
      e.dataTransfer.setDragImage(el, 0, 0);
    },
    dragTab(e, i) {
      const width = parseInt(getComputedStyle(e.target).width);
      const all = Array.from(e.target.parentElement.children);
      const next = Math.round(e.layerX / width) + i;
      const prev = e.target.getAttribute("data-x");

      e.target.style.cssText = `transform: translateX(${e.layerX}px)`;
      e.target.setAttribute("data-x", `${e.layerX}`);
      all.shift();
      all.pop();

      if (!all[next] || next === i) return;

      all[next].style.transition = "0.3s";
      if (prev < e.layerX) {
        if (e.layerX > width / 2) {
          all[next].style.transform = `translateX(-${width}px)`;
        } else {
          all[next].style.transform = "";
        }
      } else if (prev > e.layerX) {
        if (e.layerX > width / 2) {
          all[next].style.transform = "";
        } else {
          all[next].style.transform = `translateX(${width}px)`;
        }
      }

      const titles = [...this.form.titles];
      const [newTitle] = titles.splice(i, 1);

      titles.splice(next, 0, newTitle);
      localStorage.setItem("titles", JSON.stringify(titles));
    },
    dragTabEnd(e) {
      const all = Array.from(e.target.parentElement.children);
      const titles = localStorage.getItem("titles");
      const wrapper = all.shift();

      all.forEach((el) => el.removeAttribute("style"));
      this.form.titles = JSON.parse(titles);
      localStorage.removeItem("titles");

      setTimeout(() => {
        const left = all.find((el) =>
          el.className.includes("tab--active")
        ).offsetLeft;

        wrapper.style.left = `${left}px`;
      });
    },
    addConfig(title) {
      if (this.edit.isVisible) {
        const i = this.form.titles.indexOf(this.edit.title);

        this.form.titles[i] = title;
        this.edit.isVisible = false;

        return;
      }

      this.form.titles.push(title);
      this.isVisible = false;
    },
    removeConfig(title) {
      this.form.titles = this.form.titles.filter((el) => el !== title);

      if (this.form.titles.length <= 0) {
        this.isVisible = true;
      }
    },
    tryToSend(action) {
      if (!this.isValid) {
        this.$refs.form.validate();
        this.testButtonColor = "background-light";
        this.isTestSuccess = false;

        return;
      }
      if (action === 'create') delete this.plan.uuid;

      function checkName(name, obj, num = 2) {
        const value = obj.find(({ title }) => title === name);

        if (value) return checkName(`${name} ${num}`, obj, num + 1);
        else return name;
      }

      this.isLoading = true;
      this.plan.title = checkName(this.plan.title, this.plans);
      Object.entries(this.plan.products).forEach(([key, form]) => {
        const num = this.form.titles.findIndex((el) => el === key);

        form.sorter = num;
      });

      const id = this.$route.params?.planId;
      const request = (action === 'edit')
        ? api.plans.update(id, this.plan)
        : api.plans.create(this.plan);

      request.then(() => {
        this.showSnackbarSuccess({
          message: (action === 'edit')
            ? "Price model edited successfully"
            : "Price model created successfully",
        });
        setTimeout(() => {
          this.$router.push({ name: "Plans" });
        }, 100);
      })
      .catch((err) => {
        this.showSnackbarError({ message: err });
      })
      .finally(() => {
        this.isLoading = false;
      });
    },
    checkPeriods(periods) {
      const wrongPeriod = periods.find((p) => p.period === 0);

      return (
        wrongPeriod &&
        `Period cannot be zero in an ${
          wrongPeriod.key || wrongPeriod.title
        } config`
      );
    },
    checkPlanPeriods(plan) {
      if (
        !Object.keys(plan.products).length &&
        !Array.isArray(plan.resources)
      ) {
        return;
      } else if (plan.products) {
        return this.checkPeriods(Object.values(plan.products));
      } else {
        return this.checkPeriods(plan.resources);
      }
    },
    testConfig() {
      let message = "";

      if (!this.isValid || !this.isFeeValid) {
        this.$refs.form.validate();
        message = "Validation failed!";
      }

      if (!message && (this.plan.fee?.ranges?.length === 0)) {
        if (this.plan.type === 'ovh' || this.plan.type === 'opensrs') {
          message = "Ranges cant be empty!";
        }
      }

      if (!message) {
        message = this.checkPlanPeriods(this.plan);
      }

      if (message) {
        this.testButtonColor = "background-light";
        this.isTestSuccess = false;
        this.showSnackbarError({ message });
        return;
      }

      this.$store.dispatch('plans/fetch')
        .then(() => {
          this.testButtonColor = "success";
          this.isTestSuccess = true;
        })
        .catch((err) => {
          const message = err.response?.data?.message ?? err.message ?? err;

          this.showSnackbarError({ message })
          console.error(err);
        });
    },
    setPeriod(date, res) {
      const period = this.getTimestamp(date);

      if (this.plan.kind === "DYNAMIC") {
        this.plan.resources[res].period = period;
        this.plan.products = {};
      } else if (this.plan.products[res]) {
        this.plan.products[res].period = period;
      }
    },
    getTimestamp({ day, month, year, quarter, week, time }) {
      year = +year + 1970;
      month = +month + quarter * 3 + 1;
      day = +day + week * 7 + 1;

      if (`${day}`.length < 2) day = "0" + day;
      if (`${month}`.length < 2) month = "0" + month;
      let seconds = Date.parse(`${year}-${month}-${day}T${time}Z`) / 1000;

      if (month > 1) {
        seconds -= 60 * 60 * 24 * (month - 1);
      }

      return seconds;
    },
    getItem() {
      this.form.titles = [];
      if (Object.keys(this.item).length > 0) {
        this.plan = this.item;
        this.isVisible = false;
        if (this.item.kind === "DYNAMIC") {
          this.item.resources.forEach((el) => {
            this.form.titles.push(el.key);
          });
        } else {
          this.products = this.item.products;
          Object.keys(this.item.products).forEach((key) => {
            this.form.titles.push(key);
          });
        }
      }
    },
    getProduct(index) {
      const product = Object.values(this.products)[index];
      if (!product) return {};
      return {
        ...product,
        amount: product.resources,
        resources: this.item.resources,
      };
    },
    changePlan(isReset) {
      if (isReset) {
        this.selectedKind = "";
        return;
      }
      this.plan.kind = this.selectedKind;
    },
    openPlanWiki() {
      window.open(
        "https://github.com/slntopp/nocloud/wiki/Billing-Plans",
        "_blank"
      );
    },
  },
  created() {
    if (this.isEdit) {
      this.plan.resources = this.item.resources;
    }
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

    if (this.item) this.getItem();
  },
  computed: {
    template() {
      let type;
      switch (this.plan.kind) {
        case "DYNAMIC":
          type = "resources";
          break;
        default:
          type = "products";
      }

      return () => import(`@/components/plans_form_${type}.vue`);
    },
    plans() {
      return this.$store.getters['plans/all'];
    }
  },
  watch: {
    "plan.type"() {
      this.plan.fee = {};
      switch (this.plan.type) {
        case "ione":
          if (this.plan.kind === "STATIC") return;
          if (this.isEdit) return;

          this.form.titles = ["cpu", "ram", "ips_public"];
          this.isVisible = false;
          break;
        default:
          this.form.titles = [];
          this.isVisible = true;
      }
    },
    "plan.kind"() {
      if (!this.isEdit) {
        this.form.titles = [];
        if (this.plan.kind === "STATIC") {
          this.plan.products = {};
        } else {
          this.plan.resources = [];
        }
      }
    },
  },
};
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
