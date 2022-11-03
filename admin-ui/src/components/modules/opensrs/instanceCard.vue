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
      <v-col cols="4">
        <v-text-field
          readonly
          :value="localInstance.billingPlan.title"
          label="billing plan"
        />
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col>
        <v-text-field
          v-if="editing"
          v-model="localInstance.title"
          label="title"
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
import serviceControl from "@/components/modules/opensrs/serviceControls.vue";

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
