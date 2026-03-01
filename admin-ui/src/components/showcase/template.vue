<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <template-json-editor-new
      :value="template"
      title="Template JSON"
      @save="editShowcase"
    />
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditorNew from "@/components/TemplateJsonEditorNew.vue";

export default {
  name: "showcase-template",
  components: { TemplateJsonEditorNew },
  mixins: [snackbar],
  props: {
    template: {
      type: Object,
      required: true,
    },
  },
  methods: {
    async editShowcase(parsedValue) {
      try {
        await api.showcases.update(this.template.uuid, parsedValue);
        this.showSnackbarSuccess({
          message: "Showcase edited successfully",
        });
        this.$router.go();
      } catch (err) {
        this.showSnackbarError({ message: err });
      }
    },
  },
};
</script>

<style scoped lang="scss">
</style>
