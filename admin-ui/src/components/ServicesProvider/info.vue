<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="template uuid"
          style="display: inline-block; width: 330px"
          :value="provider.uuid"
          :append-icon="copyed == 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(provider.uuid, 'rootUUID')"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="template type"
          style="display: inline-block; width: 150px"
          :value="provider.type"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="proxy"
          style="display: inline-block; width: 250px"
          :value="provider.proxy?.socket"
        />
      </v-col>
      <v-col>
        <v-switch readonly label="public" v-model="provider.public" />
      </v-col>
      <v-col cols="12">
        <div class="d-flex align-start">
          <v-btn
            :to="{
              name: 'ServicesProvider edit',
              params: { uuid: provider.uuid },
            }"
          >
            Edit
          </v-btn>
          <download-template-button
            :name="downloadedFileName"
            :template="template"
            class="mx-2"
            :type="isJson ? 'JSON' : 'YAML'"
          />
          <v-switch
            class="mr-2"
            style="margin-top: 5px; padding-top: 0"
            v-model="isJson"
            :label="!isJson ? 'YAML' : 'JSON'"
          />
        </div>
      </v-col>
    </v-row>

    <component :is="spTypes" :template="provider">
      <!-- Date -->
      <v-row>
        <v-col cols="12" lg="6" class="mt-5 mb-5">
          <v-alert dark type="info" color="background ">
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
          <v-dialog max-width="60%" v-model="isDialogVisible">
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
            <v-card
              max-width="100%"
              class="ma-auto pa-5"
              color="background-light"
            >
              <nocloud-table
                table-name="sp-binded-plans"
                :items="plans"
                :headers="headers"
                :loading="isPlanLoading"
                :footer-error="fetchError"
                v-model="selectedNewPlans"
              />
              <v-card-actions class="d-flex justify-end">
                <v-btn class="mr-5" @click="isDialogVisible = false">
                  Cancel
                </v-btn>
                <v-btn :loading="isLoading" @click="bindPlans">Add</v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>
          <confirm-dialog
            :disabled="selected.length < 1"
            @confirm="unbindPlans"
          >
            <v-btn :disabled="selected.length < 1" :loading="isDeleteLoading">
              Remove
            </v-btn>
          </confirm-dialog>
        </v-col>
        <v-col>
          <nocloud-table
            table-name="service-providers"
            :items="relatedPlans"
            :headers="headers"
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

    <v-row> </v-row>
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
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";

export default {
  name: "services-provider-info",
  components: {
    DownloadTemplateButton,
    JsonEditor,
    nocloudTable,
    ConfirmDialog,
  },
  props: { template: { type: Object, required: true } },
  mixins: [snackbar],
  data: () => ({
    format,
    copyed: null,
    opened: [],
    extentionsMap,

    provider: {},
    isJson: true,
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
    selectedNewPlans: [],
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
        if (
          (!errorMessage && vlanStart === undefined) ||
          vlanSize === undefined
        ) {
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
      if (this.selectedNewPlans.length < 1) return;
      this.isLoading = true;

      const plans = this.selectedNewPlans.map((el) => el.uuid);

      api.servicesProviders
        .bindPlan(this.template.uuid, plans)
        .then(() => {
          const ending = plans.length === 1 ? "" : "s";
          this.relatedPlans.push(...this.selectedNewPlans);
          this.selectedNewPlans = [];
          this.isDialogVisible = false;
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

      const plans = this.selected.map((el) => el.uuid);

      api.servicesProviders
        .unbindPlan(this.template.uuid, plans)
        .then(() => {
          const ending = plans.length === 1 ? "" : "s";
          this.relatedPlans = this.relatedPlans.filter(
            (rp) => this.selected.findIndex((s) => s.uuid === rp.uuid) === -1
          );
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
      .dispatch("plans/fetch", {
        params: {
          sp_uuid: this.template.uuid,
        },
      })
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

      return this.$store.getters["plans/all"].filter(
        (plan) =>
          plan.type.includes(this.provider.type) && !plans.includes(plan.uuid)
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
    downloadedFileName() {
      return this.template.title
        ? this.template.title.replaceAll(" ", "_")
        : "unknown_sp";
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
