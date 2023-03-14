<template>
  <div class="module">
    <instance-create-card
      v-for="(instance, index) in instances"
      :key="index"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
      :instance="instance"
      @remove="remove(index)"
      @set-value="setValue(`${index}.${$event.key}`, $event.value)"
      :sp-uuid="spUuid"
    />
    <v-row>
      <v-col class="d-flex justify-center">
        <add-instance :disabled="isOpenSRS" @add="addInstance" />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import InstanceCreateCard from "@/components/modules/opensrs/instanceCreateCard.vue";
import AddInstance from "@/components/ui/addInstance.vue";

export default {
  name: "ione-create-service-module",
  props: ["instances-group", "plans", "planRules"],
  components: {AddInstance, InstanceCreateCard },
  mixins: [snackbar],
  data: () => ({
    defaultItem: {
      title: "instance",
      resources: {
        user: {
          first_name: "",
          last_name: "",
          org_name: "",
          address1: "",
          address2: "",
          city: "",
          country: "",
          state: "",
          postal_code: "",
          phone: "",
          email: "",
        },
        reg_username: "",
        reg_password: "",
        domain: "",
        period: 1,
        auto_renew: true,
        who_is_privacy: false,
        lock_domain: true,
      },
    },
  }),
  methods: {
    addInstance() {
      const item = JSON.parse(JSON.stringify(this.defaultItem));
      const data = JSON.parse(this.instancesGroup);
      item.title += "#" + (data.body.instances.length + 1);

      data.body.instances.push(item);
      this.change(data);
    },
    remove(index) {
      const data = JSON.parse(this.instancesGroup);

      data.body.instances.splice(index, 1);
      this.change(data);
    },
    setValue(path, val) {
      if (val === undefined) return;

      const data = JSON.parse(this.instancesGroup);

      setToValue(data.body.instances, val, path);
      this.change(data);
    },
    change(data) {
      this.$emit("update:instances-group", JSON.stringify(data));
    },
  },
  computed: {
    instances() {
      const data = JSON.parse(this.instancesGroup);
      return data.body.instances;
    },
    isOpenSRS() {
      const isOpenSrsSp =
        JSON.parse(this.instancesGroup).body.type === "opensrs";
      const isSpEmpty = JSON.parse(this.instancesGroup).sp;
      return isOpenSrsSp && !isSpEmpty;
    },
    spUuid() {
      return JSON.parse(this.instancesGroup).sp;
    },
  },
  created() {
    const data = JSON.parse(this.instancesGroup);
    if (!data.body.instances) {
      data.body.instances = [];
    }

    this.change(data);
  },
};

function setToValue(obj, value, path) {
  path = path.split(".");
  let i;
  for (i = 0; i < path.length - 1; i++) {
    if (path[i] === "__proto__" || path[i] === "constructor")
      throw new Error("Can't use that path because of: " + path[i]);
    obj = obj[path[i]];
  }

  obj[path[i]] = value;
}
</script>

<style></style>
