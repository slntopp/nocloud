<template>
  <div class="module">
    <v-card
      v-if="Object.keys(instance).length > 0"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('title', newVal)"
            label="Name"
            :value="instance.title"
          >
          </v-text-field>
        </v-col>
      </v-row>

      <v-row>
        <v-col><h3>Config:</h3></v-col>
        <v-col cols="12">
          <json-editor
            label="Config"
            :json="instance.config"
            @changeValue="(newVal) => setValue('config', newVal)"
          />
        </v-col>

        <v-col><h3>Resources:</h3></v-col>
        <v-col cols="12">
          <json-editor
            label="Resources"
            :json="instance.resources"
            @changeValue="(newVal) => setValue('resources', newVal)"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col lg="6" cols="12">
          <v-autocomplete
            label="Price model"
            item-text="title"
            item-value="uuid"
            :value="instance.plan"
            :items="plans"
            :rules="planRules"
            @change="(newVal) => setValue('billing_plan', newVal)"
          />
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script>
import JsonEditor from "@/components/JsonEditor.vue";

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    template_id: "",
    password: "",
  },
  resources: {
    cpu: 1,
    ram: 1024,
    drive_type: "SSD",
    drive_size: 10000,
    ips_public: 0,
    ips_private: 0,
  },
  billing_plan: {},
});

export default {
  name: "instance-custom-create",
  components: { JsonEditor },
  props: ["plans", "instance", "planRules", "sp-uuid", "is-edit"],
  mounted() {
    if (!this.isEdit) {
      this.$emit("set-instance", getDefaultInstance());
    }
  },
  methods: {
    setValue(key, value) {
      this.$emit("set-value", { key, value });
    },
  },
};
</script>

<style scoped></style>
