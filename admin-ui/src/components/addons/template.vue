<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <template-json-editor-new
      :value="addon"
      title="Template JSON"
      @save="editAddon"
    />
  </v-card>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditorNew from "@/components/TemplateJsonEditorNew.vue";

export default {
  name: "addon-template",
  components: { TemplateJsonEditorNew },
  mixins: [snackbar],
  props: {
    addon: {
      type: Object,
      required: true,
    },
  },
  methods: {
    async editAddon(parsedValue) {
      try {
        await this.$store.getters["addons/addonsClient"].update(parsedValue);
        this.showSnackbarSuccess({
          message: "Addon edited successfully",
        });
        this.$router.push({ name: "Addons" });
      } catch (err) {
        this.showSnackbarError({ message: err });
      }
    },
  },
};
</script>

<style scoped lang="scss">
</style>
