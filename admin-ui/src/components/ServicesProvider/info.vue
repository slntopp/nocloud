<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="template uuid"
          style="display: inline-block; width: 330px"
          v-if="!editing"
          :value="provider.uuid"
          :append-icon="copyed == 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(provider.uuid, 'rootUUID')"
        />
        <v-text-field
          v-else
          label="title"
          style="display: inline-block; width: 330px"
          v-model="provider.title"
        />
      </v-col>
      <v-col v-if="!editing">
        <v-text-field
          readonly
          label="template type"
          style="display: inline-block; width: 150px"
          :value="provider.type"
        />
      </v-col>
      <v-col>
        <v-text-field
          v-if="!editing"
          readonly
          label="proxy"
          style="display: inline-block; width: 250px"
          :value="provider.proxy?.socket"
        />
        <v-text-field
          v-else
          label="proxy"
          style="display: inline-block; width: 250px"
          v-model="provider.proxy.socket"
        />
      </v-col>
      <v-col>
        <v-switch
          label="public"
          v-model="provider.public"
          :readonly="!editing"
        />
      </v-col>
    </v-row>

    <component :is="spTypes" :template="provider">
      <!-- Edit -->
      <v-row justify="end">
        <v-col col="6" v-if="editing">
          <v-tooltip bottom :disabled="isTestSuccess">
            <template v-slot:activator="{ on, attrs }">
              <div v-bind="attrs" v-on="on" class="d-inline-block">
                <v-btn
                  color="background-light"
                  class="mr-2"
                  :loading="isLoading"
                  :disabled="!isTestSuccess"
                  @click="editServiceProvider"
                >
                  Edit
                </v-btn>
              </div>
            </template>
            <span>Test must be passed before creation.</span>
          </v-tooltip>

          <v-btn
            color="background-light"
            class="mr-2"
            :loading="isTestLoading"
            @click="testConfig"
          >
            Test
          </v-btn>
        </v-col>
        <v-col>
          <v-switch v-model="editing" label="editing" />
        </v-col>
      </v-row>
      <!-- Date -->
      <v-row>
        <v-col cols="12" lg="6" class="mt-5 mb-5">
          <v-alert dark type="info" color="indigo ">
            <span class="mr-2 text-h6">Last Monitored:</span>
            <template v-if="provider.state && template.state.meta.ts">
              {{
                format(
                  new Date(provider.state.meta.ts * 1000),
                  "dd MMMM yyy  H:mm"
                )
              }}
            </template>
            <template v-else>unknown</template>
          </v-alert>
        </v-col>
      </v-row>

      <!-- Plans -->
      <v-card-title class="px-0 mb-3">Plans:</v-card-title>
      <v-row class="flex-column">
        <v-col>
          <v-dialog v-model="isDialogVisible">
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                class="mr-2"
                v-bind="attrs"
                v-on="on"
                @click="$store.dispatch('plans/fetch')"
              >
                Add
              </v-btn>
            </template>
            <v-card>
              <nocloud-table
                :items="plans"
                :headers="headers"
                :loading="isPlanLoading"
                :footer-error="fetchError"
                v-model="selected"
              />
              <v-card-actions style="background: var(--v-background-base)">
                <v-btn :loading="isLoading" @click="bindPlans">Add</v-btn>
                <v-btn class="ml-2" @click="isDialogVisible = false"
                  >Cancel</v-btn
                >
              </v-card-actions>
            </v-card>
          </v-dialog>
          <confirm-dialog
            :disabled="selected.length < 1"
            @confirm="unbindPlans"
          >
            <v-btn :disabled="selected.length < 1" :loading="isDeleteLoading"
              >Remove</v-btn
            >
          </confirm-dialog>
        </v-col>
        <v-col>
          <nocloud-table
            :items="relatedPlans"
            :headers="headers"
            :loading="isPlanLoading"
            :footer-error="fetchError"
            v-model="selected"
          />
        </v-col>
      </v-row>
    </component>

    <template
      v-if="provider.extentions && Object.keys(provider.extentions).length > 0"
    >
      <v-card-title class="px-0">Extentions:</v-card-title>
      <component
        v-for="(extention, extName) in provider.extentions"
        :is="extentionsMap[extName].pageComponent"
        :key="extName"
        :data="extention"
      >
      </component>
    </template>

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
import JsonEditor from "@/components/JsonEditor.vue";
import extentionsMap from "@/components/extentions/map.js";
import nocloudTable from "@/components/table.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import { format } from "date-fns";

export default {
  name: "services-provider-info",
  components: { JsonEditor, nocloudTable, ConfirmDialog },
  props: { template: { type: Object, required: true } },
  mixins: [snackbar],
  data: () => ({
    format,
    copyed: null,
    opened: [],
    extentionsMap,

    provider: {},
    editing: false,
    isLoading: false,
    isTestLoading: false,
    isTestSuccess: false,

    headers: [
      { text: "Title ", value: "title" },
      { text: "UUID ", value: "uuid" },
      { text: "Public ", value: "public" },
      { text: "Type ", value: "type" },
    ],
    isDeleteLoading: false,
    isDialogVisible: false,
    relatedPlans: [],
    selected: [],
    fetchError: "",
  }),
  methods: {
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((res) => {
            console.error(res);
          });
      } else {
        alert("Clipboard is not supported!");
      }
    },
    editServiceProvider() {
      if (!this.isTestSuccess) {
        this.showSnackbarError({
          message: "Error: Test must be passed before creation.",
        });
        return;
      }
      this.isLoading = true;
      api.servicesProviders
        .update(this.template.uuid, this.provider)
        .then(() => {
          this.isLoading = false;
          this.showSnackbarSuccess({
            message: "Service edited successfully",
          });
        })
        .catch((err) => {
          this.isLoading = false;
          this.showSnackbarError({
            message: err,
          });
        });
    },
    testConfig() {
      this.isTestLoading = true;
      if (this.template.type === "ione") {
        const maxVlans = 4096;
        let errorMessage = "";

        const vlansKeys = Object.keys(this.template.secrets.vlans);
        if (vlansKeys.length > 1) {
          errorMessage = "Can be only one vlan key!";
        }

        const vlanStart = this.template.secrets.vlans[vlansKeys[0]].start;
        const vlanSize = this.template.secrets.vlans[vlansKeys[0]].size;
        if (!errorMessage && vlanStart===undefined || vlanSize===undefined) {
          errorMessage = `Vlans need size and start keys!`;
        }

        if (!errorMessage && vlanSize + vlanStart > maxVlans) {
          errorMessage = `Vlans cant be more then ${maxVlans}!`;
        }

        if (errorMessage) {
          this.isTestLoading = false;
          this.showSnackbarError({
            message: errorMessage,
          });
          return;
        }
      }
      api.servicesProviders
        .testConfig(this.template)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Tests passed",
          });
          this.isTestSuccess = true;
        })
        .catch((err) => {
          this.showSnackbarError({
            message: err,
          });
        })
        .finally(() => {
          this.isTestLoading = false;
        });
    },
    bindPlans() {
      if (this.selected.length < 1) return;
      this.isLoading = true;

      const bindPromises = this.selected.map((el) =>
        api.servicesProviders.bindPlan(this.template.uuid, el.uuid)
      );

      Promise.all(bindPromises)
        .then(() => {
          const ending = bindPromises.length === 1 ? "" : "s";

          this.showSnackbarSuccess({
            message: `Price model${ending} added successfully.`,
          });
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    unbindPlans() {
      this.isDeleteLoading = true;

      const unbindPromises = this.selected.map((el) =>
        api.servicesProviders.unbindPlan(this.template.uuid, el.uuid)
      );

      Promise.all(unbindPromises)
        .then(() => {
          const ending = unbindPromises.length === 1 ? "" : "s";

          this.showSnackbarSuccess({
            message: `Price model${ending} deleted successfully.`,
          });
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isDeleteLoading = false;
        });
    },
  },
  mounted() {
    this.provider = this.template;
    if (!this.provider.proxy) {
      this.provider.proxy = { socket: "" };
    }
  },
  created() {
    this.$store
      .dispatch("plans/fetch", { params: {
        sp_uuid: this.template.uuid,
        anonymously: false,
      } })
      .then(() => {
        this.relatedPlans = this.$store.getters["plans/all"];
        this.fetchError = "";
      })
      .catch((err) => {
        console.error(err);

        this.fetchError = "Can't reach the server";
        if (err.response) {
          this.fetchError += `: [ERROR]: ${err.response.data.message}`;
        } else {
          this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
        }
      });
  },
  computed: {
    plans() {
      const plans = this.relatedPlans.map(({ uuid }) => uuid);

      return this.$store.getters["plans/all"].filter((plan) =>
        plan.type === this.provider.type && !plans.includes(plan.uuid)
      );
    },
    isPlanLoading() {
      return this.$store.getters["plans/isLoading"];
    },
    spTypes() {
      switch (this.provider.type) {
        case "ione":
          return () =>
            import("@/components/modules/ione/serviceProviderInfo.vue");
        case "ovh":
          return () =>
            import("@/components/modules/ovh/serviceProviderInfo.vue");
        default:
          return () =>
            import("@/components/modules/custom/serviceProviderInfo.vue");
      }
    },
  },
};
</script>

<style>
.title_progress {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.v-alert__icon.v-icon {
  margin-top: 5px;
}
.apexcharts-svg {
  background: none !important;
}
.ceil {
  display: inline-block;
  width: 15px;
  height: 15px;
  margin: 5px;
  vertical-align: middle;
  border-radius: 2px;
}
.occupied {
  background: var(--v-success-base);
}
.free {
  background: var(--v-error-base);
}
</style>
