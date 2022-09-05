<template>
  <div>
    <!-- Secrets -->
    <v-card-title class="px-0 mb-3"> Secrets:</v-card-title>
    <v-row v-if="!editing">
      <v-col>
        <json-textarea readonly :json="template" />
      </v-col>
    </v-row>

    <slot></slot>
  </div>
</template>

<script>
import JsonTextarea from "@/components/JsonTextarea.vue";

export default {
  name: "service-provider-custom",
  components: { JsonTextarea },
  props: {
    template: { type: Object, required: true },
    editing: { type: Boolean, required: true }
  }
};
</script>