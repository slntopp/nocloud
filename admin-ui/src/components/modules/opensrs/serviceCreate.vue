<template>
  <div class="module">
    <v-card
      v-for="(instance, index) in instances"
      :key="index"
      :id="instance.uuid"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue(index + '.title', newVal)"
            label="title"
            v-model="instance.title"
          />
        </v-col>
        <v-col class="d-flex justify-end">
          <v-btn @click="() => remove(index)"> remove </v-btn>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.first_name', newVal)
            "
            label="first name"
            v-model="instance.resources.user.first_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.last_name', newVal)
            "
            label="last name"
            v-model="instance.resources.user.last_name"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.org_name', newVal)
            "
            label="organization name"
            v-model="instance.resources.user.org_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.address1', newVal)
            "
            label="address1"
            v-model="instance.resources.user.address1"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.address2', newVal)
            "
            label="address2"
            v-model="instance.resources.user.address2"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.city', newVal)
            "
            label="city"
            v-model="instance.resources.user.city"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.country', newVal)
            "
            label="country"
            v-model="instance.resources.user.country"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.state', newVal)
            "
            label="state"
            v-model="instance.resources.user.state"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) =>
                setValue(index + '.resources.user.postal_code', newVal)
            "
            label="postal_code"
            v-model="instance.resources.user.postal_code"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.phone', newVal)
            "
            label="phone"
            v-model="instance.resources.user.phone"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.email', newVal)
            "
            label="email"
            v-model="instance.resources.user.email"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.reg_username', newVal)
            "
            label="reg_username"
            v-model="instance.resources.reg_username"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.reg_password', newVal)
            "
            label="reg_password"
            v-model="instance.resources.reg_password"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="4">
          <v-switch
            @change="
              (newVal) => setValue(index + '.resources.auto_renew', newVal)
            "
            v-model="instance.resources.auto_renew"
            label="auto_renew"
          />
        </v-col>
        <v-col cols="4">
          <v-switch
            @change="
              (newVal) => setValue(index + '.resources.who_is_privacy', newVal)
            "
            v-model="instance.resources.who_is_privacy"
            label="who_is_privacy"
          />
        </v-col>
        <v-col cols="4">
          <v-switch
            @change="
              (newVal) => setValue(index + '.resources.lock_domain', newVal)
            "
            v-model="instance.resources.lock_domain"
            label="lock_domain"
          />
        </v-col>
      </v-row>
      <domains-table
        :sp-uuid="spUuid"
        @input:period="setValue(index + '.resources.period', $event)"
        @input:domain="setValue(index + '.resources.domain', $event)"
      />
    </v-card>
    <v-row>
      <v-col class="d-flex justify-center">
        <add-instance-btn @click="addInstance" :disabled="isOpenSRS" />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import domainsTable from "@/components/domains_table.vue";
import AddInstanceBtn from "@/components/ui/addInstanceBtn.vue";

export default {
  name: "ione-create-service-module",
  props: ["instances-group", "plans", "planRules"],
  components: { AddInstanceBtn, domainsTable },
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
