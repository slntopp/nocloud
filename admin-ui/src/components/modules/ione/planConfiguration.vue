<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <v-text-field
          label="Minimum disk size"
          type="number"
          suffix="GB"
          v-model="meta.minDisk"
        />
      </v-col>
      <v-col>
        <v-text-field
          label="Maximum disk size"
          type="number"
          suffix="GB"
          v-model="meta.maxDisk"
        />
      </v-col>
    </v-row>
    <v-btn :loading="isSaveLoading" @click="save">Save</v-btn>
  </v-card>
</template>

<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar";

export default {
  name: "ione-plan-configuration",
  props: {
    template: {
      type: Object,
      required: true,
    },
  },
  mixins: [snackbar],
  data: () => ({
    meta: {},
    isSaveLoading: false,
  }),
  methods: {
    async save() {
      this.isSaveLoading = true;
      try {
        const data = { ...this.template, meta: this.meta };
        await api.plans.update(this.template.uuid, data);
        this.showSnackbarSuccess({
          message: "Configuration edited successfully",
        });
      } catch (err) {
        this.showSnackbarError({ message: err });
      } finally {
        this.isSaveLoading = false;
      }
    },
  },
  mounted() {
    this.meta = this.template.meta;
  },
  watch: {
    "template.meta"(newVal) {
      this.meta = newVal;
    },
  },
};
</script>

<style scoped lang="scss"></style>
