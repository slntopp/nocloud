<template>
  <div>
    <v-row v-if="!editing">
      <v-col>
        <service-control
          :service="service"
          :instance_uuid="localInstance.uuid"
          :chip-color="chipColor"
          @closePanel="openedlocalInstances = {}"
        />
      </v-col>
    </v-row>
    <v-row v-if="!editing">
      <template>
        <v-col md="2">
          <v-text-field
            readonly
            :value="localInstance.state && localInstance.state.meta.state"
            label="state"
            style="display: inline-block; width: 100px"
          />
        </v-col>
        <v-col md="2">
          <v-text-field
            readonly
            :value="localInstance.state && localInstance.state.state"
            label="lcm state"
            style="display: inline-block; width: 100px"
          />
        </v-col>
      </template>
      <v-col md="2">
        <v-text-field
          readonly
          :value="localInstance.billingPlan.title"
          label="price model"
          style="display: inline-block; width: 100px"
        />
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col>
        <v-text-field
          v-if="editing"
          v-model="localInstance.title"
          label="title"
          style="display: inline-block; width: 160px"
        />
      </v-col>
    </v-row>
    <v-row>
      <json-editor
        v-if="editing"
        :json="localInstance.config"
        @changeValue="(data) => (localInstance.config = data)"
      />
      <json-textarea v-else :json="localInstance.config" :readonly="true" />
    </v-row>
  </div>
</template>

<script>
import JsonTextarea from "@/components/JsonTextarea.vue";
import jsonEditor from "@/components/JsonEditor.vue";
import serviceControl from "@/components/modules/ovh/serviceControls.vue";

import snackbar from "@/mixins/snackbar.js";

export default {
  name: "instance-card",
  components: { JsonTextarea, jsonEditor, serviceControl },
  mixins: [snackbar],
  props: {
    instance: { type: Object },
    service: { type: Object },
    editing: { type: Boolean },
    type: { type: String },
    chipColor: { type: String },
  },
  data() {
    return { localInstance: {}, isVisible: false };
  },
  created() {
    this.localInstance = this.instance;
  },
  methods: {},
};
</script>
