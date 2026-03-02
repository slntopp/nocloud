<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <template-json-editor
      :value="template"
      title="Template JSON"
      @save="editTemplate"
    />
  </v-card>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditor from "@/components/TemplateJsonEditor.vue";
import { UpdateRequest } from "nocloud-proto/proto/es/instances/instances_pb";

export default {
  name: "instance-template",
  components: { TemplateJsonEditor },
  mixins: [snackbar],
  props: {
    template: {
      type: Object,
      required: true,
    },
  },
  methods: {
    async editTemplate(parsedValue) {
      try {
        const instance = parsedValue;
        console.log(instance);

        // const service = this.$store.getters["services/all"].find(
        //   ({ uuid }) => uuid === this.template.service,
        // );

        // const igIndex = service.instancesGroups.findIndex((ig) =>
        //   ig.instances.find((i) => i.uuid === this.template.uuid),
        // );
        // const instanceIndex = service.instancesGroups[
        //   igIndex
        // ].instances.findIndex((i) => i.uuid === this.template.uuid);

        // service.instancesGroups[igIndex].instances[instanceIndex] = instance;

        // this.isLoading = true;
        // api.services._update(service);

        await this.$store.getters["instances/instancesClient"].update(
          UpdateRequest.fromJson(
            { instance: parsedValue, uuid: this.template.uuid },
            { ignoreUnknownFields: true },
          ),
        );

        // await api.instances.update(this.template.uuid, parsedValue);
        this.showSnackbarSuccess({
          message: "Instance template edited successfully",
        });
        // this.$router.go();
      } catch (err) {
        this.showSnackbarError({ message: err });
      }
    },
  },
};
</script>

<style scoped lang="scss"></style>
