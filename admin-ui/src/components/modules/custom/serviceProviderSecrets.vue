<template>
  <div class="pa-4">
    <v-row v-if="!editing">
      <v-col :cols="12" :md="6">
        <json-textarea readonly :json="template.secrets" />
      </v-col>
    </v-row>

    <slot v-else></slot>
  </div>
</template>

<script>
import JsonTextarea from "@/components/JsonTextarea.vue";

export default {
  components: { JsonTextarea },
  props: { template: { type: Object, required: true } },
  data: () => ({ editing: false })
};
</script>