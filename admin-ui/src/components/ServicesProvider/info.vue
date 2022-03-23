<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <v-text-field
          readonly
          :value="template.uuid"
          label="template uuid"
          style="display: inline-block; width: 330px"
          :append-icon="copyed == 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(template.uuid, 'rootUUID')"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.type"
          label="template type"
          style="display: inline-block; width: 150px"
        >
        </v-text-field>
      </v-col>
    </v-row>

    <!-- Secrets -->
    <v-card-title class="px-0 mb-3"> Secrets:</v-card-title>
    <v-row>
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

    <!-- Variables -->
    <v-card-title class="px-0 mb-3">Variables:</v-card-title>
    <v-row>
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
        <v-row>
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
  </v-card>
</template>

<script>
import extentionsMap from "@/components/extentions/map.js";
import { format } from "date-fns";
export default {
  name: "services-provider-info",
  data: () => ({
    format,
    copyed: null,
    opened: [],
    showPassword: false,
    extentionsMap,
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
      navigator.clipboard
        .writeText(text)
        .then(() => {
          this.copyed = index;
        })
        .catch((res) => {
          console.error(res);
        });
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
</style>
