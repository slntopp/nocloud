<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="template uuid"
          style="display: inline-block; width: 330px"
          v-if="!editing"
          :value="template.uuid"
          :append-icon="copyed == 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(template.uuid, 'rootUUID')"
        />
        <v-text-field
          v-else
          label="title"
          style="display: inline-block; width: 330px"
          v-model="provider.title"
        />
      </v-col>
      <v-col>
        <v-select
          label="template type"
          style="display: inline-block; width: 150px"
          v-model="provider.type"
          :items="types"
          :readonly="!editing"
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

    <!-- Secrets -->
    <v-card-title class="px-0 mb-3"> Secrets:</v-card-title>
    <v-row v-if="!editing">
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.group"
          label="group"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.host"
          label="host"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.user"
          label="user"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.pass"
          label="password"
          style="display: inline-block; width: 330px"
          :type="showPassword ? 'text' : 'password'"
          :append-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append="() => (showPassword = !showPassword)"
        >
        </v-text-field>
      </v-col>
    </v-row>
    <v-row v-else>
      <json-editor
        :json="template.secrets"
        @changeValue="(data) => provider.secrets = data"
      />
    </v-row>

    <!-- Variables -->
    <v-card-title class="px-0 mb-3">Variables:</v-card-title>
    <v-row v-if="!editing">
      <v-col v-for="(variable, varTitle) in template.vars" :key="varTitle">
        {{ varTitle.replaceAll("_", " ") }}
        <v-row>
          <v-col :cols="12" v-for="(value, key) in variable.value" :key="key">
            <v-text-field
              readonly
              :value="JSON.stringify(value)"
              :label="key"
              style="display: inline-block; width: 200px"
            >
            </v-text-field>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
    <v-row v-else>
      <json-editor
        :json="template.vars"
        @changeValue="(data) => provider.vars = data"
      />
    </v-row>

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
          {{
            template.state.meta.ts &&
            format(new Date(template.state.meta.ts * 1000), "dd MMMM yyy  H:mm")
          }}
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
              <v-btn class="ml-2" @click="isDialogVisible = false">Cancel</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-btn :loading="isDeleteLoading" @click="unbindPlans">Remove</v-btn>
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

    <!-- Hosts -->
    <v-card-title class="px-0 mb-3">Hosts:</v-card-title>
    <v-row class="mb-7">
      <v-col v-if="template.state.meta.hosts.error" cols="12" lg="6">
        <v-alert type="error"> {{ template.state.meta.hosts.error }}</v-alert>
      </v-col>
      <v-col
        v-for="(host, idx) in template.state.meta.hosts"
        :key="idx"
        cols="12"
        v-else
      >
        <v-row>
          <v-col v-if="host.error" cols="12" lg="6">
            <v-alert type="error"> {{ host.error }}</v-alert>
          </v-col>
          <v-col class="order-2 order-lg-1" cols="12" lg="6" v-else>
            <v-row>
              <v-col cols="12">
                <div class="title_progress">
                  <span>CPU</span>
                  <div>
                    <span>{{ host.total_cpu - host.free_cpu }}</span> /
                    <span>{{ host.total_cpu }}</span>
                  </div>
                </div>
                <v-progress-linear
                  :value="
                    Math.round(
                      ((host.total_cpu - host.free_cpu) / host.total_cpu) * 100
                    )
                  "
                  color="green"
                  height="20"
                >
                  <template v-slot:default="{ value }">
                    <strong>{{ value }}%</strong>
                  </template>
                </v-progress-linear>
              </v-col>
              <v-col cols="12">
                <div class="title_progress">
                  <span>Memory</span>
                  <div>
                    <span
                      >{{
                        ((host.total_ram - host.free_ram) / 1048576).toFixed(2)
                      }}
                      GiB</span
                    >
                    /
                    <span>{{ (host.total_ram / 1048576).toFixed(2) }} GiB</span>
                  </div>
                </div>
                <v-progress-linear
                  :value="
                    Math.round(
                      ((host.total_ram - host.free_ram) / host.total_ram) * 100
                    )
                  "
                  color="green"
                  height="20"
                >
                  <template v-slot:default="{ value }">
                    <strong>{{ value }}%</strong>
                  </template>
                </v-progress-linear>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="12" lg="6" class="order-1 order-lg-2">
            <p>name: {{ host.name }}</p>
            <p>state: {{ host.state }}</p>
            <p>vm_mad: {{ host.vm_mad }}</p>
            <p>im_mad: {{ host.im_mad }}</p>
          </v-col>
        </v-row>
      </v-col>
    </v-row>

    <!-- Networking -->
    <v-card-title class="px-0 mb-3">Networking:</v-card-title>
    <v-row class="mb-7">
      <v-col>
        <v-card-subtitle class="px-0">Public</v-card-subtitle>
        <v-row>
          <v-col
            v-if="template.state.meta.networking.public_vnet.error"
            cols="12"
            lg="6"
          >
            <v-alert type="error">
              {{ template.state.meta.networking.public_vnet.error }}</v-alert
            >
          </v-col>

          <v-col v-else cols="12" lg="6">
            <div class="title_progress">
              <span>
                {{ template.state.meta.networking.public_vnet.name }}
              </span>
              <div>
                <span>{{
                  template.state.meta.networking.public_vnet.used
                }}</span>
                /
                <span>{{
                  template.state.meta.networking.public_vnet.total
                }}</span>
              </div>
            </div>
            <v-progress-linear
              :value="
                Math.round(
                  (template.state.meta.networking.public_vnet.used /
                    template.state.meta.networking.public_vnet.total) *
                    100
                )
              "
              color="green"
              height="20"
              class="mb-10"
            >
              <template v-slot:default="{ value }">
                <strong>{{ value }}%</strong>
              </template>
            </v-progress-linear>
          </v-col>
        </v-row>
        <v-card-subtitle class="px-0">Private</v-card-subtitle>
        <v-row ref="private">
          <v-col
            v-if="template.state.meta.networking.private_vnet.error"
            cols="12"
            lg="6"
          >
            <v-alert type="error">
              {{ template.state.meta.networking.private_vnet.error }}</v-alert
            >
          </v-col>
          <v-col v-else cols="12" lg="6">
            <div class="title_progress">
              <span>
                {{ template.state.meta.networking.private_vnet.name }}
              </span>
              <div>
                <span>{{
                  template.state.meta.networking.private_vnet.used
                }}</span>
                /
                <span>{{
                  template.state.meta.networking.private_vnet.total
                }}</span>
              </div>
            </div>
            <v-progress-linear
              :value="
                Math.round(
                  (template.state.meta.networking.private_vnet.used /
                    template.state.meta.networking.private_vnet.total) *
                    100
                )
              "
              color="green"
              height="20"
              class="mb-10"
            >
              <template v-slot:default="{ value }">
                <strong>{{ value }}%</strong>
              </template>
            </v-progress-linear>
          </v-col>
          <v-col cols="12">
            <p>Vlans:</p>
            <v-tooltip
              bottom
              color="info"
              v-for="(vlan, i) of vlans"
              :key="i"
            >
              <template v-slot:activator="{ on, attrs }">
                <span
                  class="ceil"
                  v-bind="attrs"
                  v-on="on"
                  :class="(vlan === 0) ? 'occupied' : 'free'"
                />
              </template>
              <span>{{ i }}</span>
            </v-tooltip>
            <div class="mt-2">
              <v-btn
                class="mr-2"
                v-if="counter > 1"
                @click="counter--"
              >
                less
              </v-btn>
              <v-btn
                v-if="counter < 8"
                @click="counter++"
              >
                more
              </v-btn>
            </div>
          </v-col>
        </v-row>
      </v-col>
    </v-row>

    <!-- Datastores -->
    <v-card-title class="px-0 mb-3">Datastores:</v-card-title>
    <v-row class="mb-7">
      <v-col v-if="template.state.meta.datastores.error" cols="12" lg="6">
        <v-alert type="error">
          {{ template.state.meta.datastores.error }}</v-alert
        >
      </v-col>
      <v-col
        v-for="(datastor, idx) in template.state.meta.datastores"
        :key="idx"
        v-else
        cols="12"
      >
        <v-row>
          <v-col class="order-2 order-lg-1" cols="12" lg="6" order-2>
            <div class="title_progress">
              <span>{{ datastor.drive_type }}</span>
              <div>
                <span>{{ (datastor.used / 1024).toFixed(2) }}</span> /
                <span>{{ (datastor.total / 1024).toFixed(2) }} GiB</span>
              </div>
            </div>
            <v-progress-linear
              :value="Math.round((datastor.used / datastor.total) * 100)"
              color="green"
              height="20"
            >
              <template v-slot:default="{ value }">
                <strong>{{ value }}%</strong>
              </template>
            </v-progress-linear>
          </v-col>
          <v-col class="order-1 order-lg-2" cols="12" lg="6" order-1>
            <p>name: {{ datastor.name }}</p>
            <p>tm_mad: {{ datastor.tm_mad }}</p>
            <p>ds_mad: {{ datastor.tm_mad }}</p>
          </v-col>
        </v-row>
      </v-col>
    </v-row>

    <template
      v-if="template.extentions && Object.keys(template.extentions).length > 0"
    >
      <v-card-title class="px-0">Extentions:</v-card-title>
      <component
        v-for="(extention, extName) in template.extentions"
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
import api from '@/api.js';
import snackbar from '@/mixins/snackbar.js';
import JsonEditor from '@/components/JsonEditor.vue';
import extentionsMap from "@/components/extentions/map.js";
import nocloudTable from '@/components/table.vue';
import { format } from "date-fns";

export default {
  name: "services-provider-info",
  components: { JsonEditor, nocloudTable },
  mixins: [snackbar],
  data: () => ({
    format,
    copyed: null,
    opened: [],
    showPassword: false,
    extentionsMap,
    counter: 1,

    types: [],
    provider: {},
    editing: false,
    isLoading: false,
    isTestLoading: false,
    isTestSuccess: false,

    headers: [
      { text: 'Title ', value: 'title' },
      { text: 'UUID ', value: 'uuid' },
      { text: 'Public ', value: 'public' },
      { text: 'Type ', value: 'type' }
    ],
    isDeleteLoading: false,
    isDialogVisible: false,
    relatedPlans: [],
    selected: [],
    fetchError: ''
  }),
  props: {
    template: {
      type: Object,
      required: true,
    },
  },
  //   cumputed:{

  //   },
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
        alert('Clipboard is not supported!');
      }
    },
    changeWidth() {
      const { clientWidth } = this.$refs.private;
      let cols = 64;

      for (let i = 4; i > 0; i--) {
        if (clientWidth / 30 >= cols) {
          this.$refs.private.style.width = `${cols * 30 + 24}px`;
          break;
        } else {
          cols /= 2;
        }
      }
    },
    editServiceProvider() {
      if (!this.isTestSuccess) {
        this.showSnackbarError({
          message: 'Error: Test must be passed before creation.',
        });
        return;
      }
      this.isLoading = true;

      api.servicesProviders
        .update(this.template.uuid, this.template)
        .then(() => {
          this.showSnackbarSuccess({
            message: 'Service edited successfully'
          });
        })
        .catch((err) => {
          this.showSnackbarError({
            message: err
          });
        })
        .finnaly(() => {
          this.isLoading = false;
        });
    },
    testConfig() {
      this.isTestLoading = true;
      api.servicesProviders
        .testConfig(this.template)
        .then(() => {
          this.showSnackbarSuccess({
            message: 'Tests passed'
          });
          this.isTestSuccess = true;
        })
        .catch((err) => {
          this.showSnackbarError({
            message: err
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
        api.servicesProviders
          .bindPlan(this.template.uuid, el.uuid)
      );

      Promise.all(bindPromises)
        .then(() => {
          const ending = bindPromises.length === 1 ? '' : 's';

          this.showSnackbarSuccess({
            message: `Plan${ending} added successfully.`,
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
      if (this.selected.length < 1) return;
      this.isDeleteLoading = true;
      
      const unbindPromises = this.selected.map((el) =>
        api.servicesProviders
          .unbindPlan(this.template.uuid, el.uuid)
      );

      Promise.all(unbindPromises)
        .then(() => {
          const ending = unbindPromises.length === 1 ? '' : 's';

          this.showSnackbarSuccess({
            message: `Plan${ending} deleted successfully.`,
          });
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isDeleteLoading = false;
        });
    }
  },
  mounted() {
    this.provider = this.template;
    this.changeWidth();
  },
  created() {
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
      }
    });
    
    this.$store.dispatch('plans/fetch', {
      sp_uuid: this.template.uuid,
      anonymously: false
    })
      .then(() => {
        this.relatedPlans = this.$store.getters['plans/all'];
        this.fetchError = '';
      })
      .catch((err) => {
        console.error(err);

        this.fetchError = 'Can\'t reach the server';
        if (err.response) {
          this.fetchError += `: [ERROR]: ${err.response.data.message}`;
        } else {
          this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
        }
      });
  },
  computed: {
    vlans() {
      const { free_vlans } = this.template?.state.meta.networking.private_vnet;
      let vlans = 0;

      Object.values(free_vlans || {}).forEach((value) => {
        vlans += +value;
      });

      const res = Array.from({ length: 512 * this.counter })
        .fill(1, 0, vlans)
        .fill(0, vlans);

      return res;
    },
    plans() {
      return this.$store.getters['plans/all'];
    },
    isPlanLoading() {
      return this.$store.getters['plans/isLoading'];
    }
  },
  watch: {
    counter() { this.changeWidth() }
  }
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
  width: 20px;
  height: 20px;
  margin: 5px;
  vertical-align: middle;
  border-radius: 5px;
}
.occupied {
  background: var(--v-success-base);
}
.free {
  background: var(--v-error-base);
}
</style>
