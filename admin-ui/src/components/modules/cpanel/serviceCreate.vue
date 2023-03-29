<template>
  <div class="module">
    <v-card
      class="mb-4 pa-2"
      color="background"
      elevation="0"
      v-for="(instance, index) of instances"
      :key="index"
      :id="instance.uuid"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="title"
            v-model="instance.title"
            :rules="rules.req"
            @change="setValue(index + '.title', $event)"
          />
        </v-col>
        <v-col class="d-flex justify-end">
          <v-btn @click="() => remove(index)">Remove</v-btn>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="domain"
            :rules="rules.req"
            v-model="instance.config.domain"
            @change="setValue(index + '.config.domain', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="mail"
            :rules="rules.req"
            v-model="instance.config.mail"
            @change="setValue(index + '.config.mail', $event)"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="password"
            :rules="rules.req"
            v-model="instance.config.password"
            @change="setValue(index + '.config.password', $event)"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            label="plan"
            :rules="rules.req"
            v-model="instance.config.plan"
            @change="setValue(index + '.config.plan', $event)"
          />
        </v-col>
      </v-row>
    </v-card>
    <v-row>
      <v-col class="d-flex justify-center">
        <add-instance-btn @click="addInstance" :disabled="isDisabled" />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import AddInstanceBtn from "@/components/ui/addInstanceBtn.vue";

export default {
  name: "ovh-create-service-module",
  components: { AddInstanceBtn },
  props: ["instances-group", "plans", "planRules", "meta"],
  data: () => ({
    defaultItem: {
      title: "instance",
      config: {
        domain: "",
        mail: "",
        plan: "",
        password: "",
      },
      billing_plan: {},
    },
    rules: {
      req: [(v) => !!v || "required field"],
    },
  }),
  methods: {
    addInstance() {
      const item = JSON.parse(JSON.stringify(this.defaultItem));
      const data = JSON.parse(this.instancesGroup);
      const i = data.body.instances.length;

      item.title += "#" + (i + 1);
      data.body.instances.push(item);
      this.change(data);
    },
    remove(index) {
      const data = JSON.parse(this.instancesGroup);

      data.body.instances.splice(index, 1);
      this.change(data);
    },
    setValue(path, val) {
      const data = JSON.parse(this.instancesGroup);
      const index = +path.slice(0, path.indexOf("."));
      if (path.includes("config")) {
        path.split(".").forEach((key) => {
          path = key;
        });
        data.body.instances[index].config[path] = val;
      }

      this.change(data);
    },
    change(data) {
      this.$emit("update:instances-group", JSON.stringify(data));
    },
  },
  computed: {
    instances() {
      return JSON.parse(this.instancesGroup).body.instances;
    },
    isDisabled() {
      const group = JSON.parse(this.instancesGroup);

      return group.body.type === "cpanel" && !group.sp;
    },
  },
};
</script>
