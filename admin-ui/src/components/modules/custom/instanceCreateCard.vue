<template>
  <v-card
    :id="instance.uuid"
    class="mb-4 pa-2"
    elevation="0"
    color="background"
  >
    <v-row>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) => $emit('set-value', { key: 'title', value: newVal })
          "
          label="title"
          :value="instance.title"
        >
        </v-text-field>
      </v-col>
      <v-col v-if="showRemove" class="d-flex justify-end">
        <v-btn @click="() => $emit('remove')"> remove </v-btn>
      </v-col>
    </v-row>

    <v-row>
      <v-col><h3>Config:</h3></v-col>
      <v-col cols="12">
        <json-editor
          label="config"
          :json="instance.config"
          @changeValue="
            (newVal) => $emit('set-value', { key: 'config', value: newVal })
          "
        />
      </v-col>

      <v-col><h3>Resources:</h3></v-col>
      <v-col cols="12">
        <json-editor
          label="resources"
          :json="instance.resources"
          @changeValue="
            (newVal) => $emit('set-value', { key: 'resources', value: newVal })
          "
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col lg="6" cols="12">
        <v-select
          label="price model"
          item-text="title"
          item-value="uuid"
          :value="instance.plan"
          :items="plans.list"
          :rules="planRules"
          @change="
            (newVal) =>
              $emit('set-value', { key: 'billing_plan', value: newVal })
          "
        />
      </v-col>
    </v-row>
  </v-card>
</template>

<script>
import JsonEditor from "@/components/JsonEditor.vue";

export default {
  name: "instanceCreateCard",
  components: { JsonEditor },
  props: {
    instance: {},
    "plan-rules": {},
    plans: {},
    "show-remove": { type: Boolean, default: true },
  },
  emits: ["remove", "set-value"],
};
</script>

<style scoped></style>
