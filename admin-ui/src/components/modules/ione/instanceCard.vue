<template>
  <div>
    <v-row align="center">
      <v-col>
        <v-text-field
          readonly
          label="id"
          style="display: inline-block; width: 200px"
          :value="template.data.vmid"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="next payment date"
          style="display: inline-block; width: 200px"
          :value="date"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="OS"
          style="display: inline-block; width: 200px"
          :value="osName"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="password"
          style="display: inline-block; width: 200px"
          :type="(isVisible) ? 'text' : 'password'"
          :value="template.config.password"
          :append-icon="(isVisible) ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append="isVisible = !isVisible"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col v-for="(item, key) in template.resources" :key="key">
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
  },
  data: () => ({
    isVisible: false,
    dictionary: {
      cpu: "CPU",
      ram: "RAM",
    }
  }),
  created() {
    this.$store.dispatch('servicesProviders/fetch');
  },
  computed: {
    sp() {
      return this.$store.getters['servicesProviders/all']
        .find(({ uuid }) => uuid === this.template.sp);
    },
    osName() {
      const id = this.template.config.template_id;

      return this.sp?.publicData.templates[id].name;
    },
    date() {
      if (!this.template.data.last_monitoring) return '-';
      const date = new Date(this.template.data.last_monitoring * 1000);

      const year = date.toUTCString().split(" ")[3];
      let month = date.getUTCMonth() + 1;
      let day = date.getUTCDate();

      if (`${month}`.length < 2) month = `0${month}`;
      if (`${day}`.length < 2) day = `0${day}`;

      return `${day}.${month}.${year}`;
    }
  }
}
</script>
