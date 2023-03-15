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
            label="title"
            :value="instance.title"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.first_name', newVal)"
            label="first name"
            :value="instance.resources.user.first_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.last_name', newVal)"
            label="last name"
            :value="instance.resources.user.last_name"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.org_name', newVal)"
            label="organization name"
            :value="instance.resources.user.org_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.address1', newVal)"
            label="address1"
            :value="instance.resources.user.address1"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.address2', newVal)"
            label="address2"
            :value="instance.resources.user.address2"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.city', newVal)"
            label="city"
            :value="instance.resources.user.city"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.country', newVal)"
            label="country"
            :value="instance.resources.user.country"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.state', newVal)"
            label="state"
            :value="instance.resources.user.state"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.postal_code', newVal)"
            label="postal_code"
            :value="instance.resources.user.postal_code"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.phone', newVal)"
            label="phone"
            :value="instance.resources.user.phone"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.user.email', newVal)"
            label="email"
            :value="instance.resources.user.email"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.reg_username', newVal)"
            label="reg_username"
            :value="instance.resources.reg_username"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.reg_password', newVal)"
            label="reg_password"
            :value="instance.resources.reg_password"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="4">
          <v-switch
            @change="(newVal) => setValue('resources.auto_renew', newVal)"
            :value="instance.resources.auto_renew"
            label="auto_renew"
          />
        </v-col>
        <v-col cols="4">
          <v-switch
            @change="(newVal) => setValue('resources.who_is_privacy', newVal)"
            :value="instance.resources.who_is_privacy"
            label="who_is_privacy"
          />
        </v-col>
        <v-col cols="4">
          <v-switch
            @change="(newVal) => setValue('resources.lock_domain', newVal)"
            :value="instance.resources.lock_domain"
            label="lock_domain"
          />
        </v-col>
      </v-row>
      <domains-table
        :sp-uuid="spUuid"
        @input:period="setValue('resources.period', $event)"
        @input:domain="setValue('resources.domain', $event)"
      />
    </v-card>
  </div>
</template>

<script>
import DomainsTable from "@/components/domains_table.vue";
const getDefaultInstance = () => ({
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
});
export default {
  name: "instance-opensrs-create",
  components: { DomainsTable },
  props: ["plans", "instance", "planRules", "sp-uuid"],
  mounted() {
    this.$emit("set-instance", getDefaultInstance());
  },
  methods: {
    setValue(key, value) {
      this.$emit("set-value", { key, value });
    },
  },
};
</script>

<style scoped></style>
