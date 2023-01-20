<template>
  <div>
    <h3 v-if="dense">Resources:</h3>
    <v-card-title v-else class="px-0">Resources:</v-card-title>
    <v-row align="center">
      <v-col v-for="(item, key) in resources" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key.replace('_', ' ')"
          :value="item"
        />
      </v-col>
    </v-row>

    <h3 v-if="dense">CSR:</h3>
    <v-card-title v-else class="px-0 pb-0">CSR:</v-card-title>
    <v-row align="center">
      <v-col lg="6" cols="12">
        <v-textarea readonly :value="template.resources.csr" />
      </v-col>
    </v-row>

    <h3 v-if="dense">User:</h3>
    <v-card-title v-else class="px-0">User:</v-card-title>
    <v-row align="center">
      <v-col v-for="(item, key) in user" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key.replace('_', ' ')"
          :value="item"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "instance-card",
  props: {
    template: { type: Object, required: true },
    dense: { type: Boolean },
  },
  data: () => ({
    dictionary: {
      phonenumber: "phone",
      companyname: "company",
    }
  }),
  computed: {
    resources() {
      const res = JSON.parse(JSON.stringify(this.template.resources));

      delete res.user;
      delete res.csr;
      return res;
    },
    user() {
      const res = JSON.parse(JSON.stringify(this.template.resources.user));

      delete res.domain;
      delete res.order;
      return res;
    }
  }
}
</script>
