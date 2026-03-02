<template>
  <v-card elevation="0" color="background-light" class="pa-4 template-card">
    <template-json-editor
      :value="template"
      title="Template JSON"
      @save="editPlan"
    />
  </v-card>
</template>

<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar";
import TemplateJsonEditor from "@/components/TemplateJsonEditor.vue";

export default {
  name: "plan-template",
  components: { TemplateJsonEditor },
  mixins: [snackbar],
  props: {
    template: { type: Object, required: true },
  },
  methods: {
    async editPlan(parsedValue) {
      try {
        await api.plans.update(this.template.uuid, parsedValue);
        this.showSnackbarSuccess({ message: "Price model edited successfully" });
        this.$router.go();
      } catch (err) {
        this.showSnackbarError({ message: err?.message || "Save failed" });
      }
    },
  },
};
</script>

<style scoped lang="scss">
.template-card {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
}
</style>
