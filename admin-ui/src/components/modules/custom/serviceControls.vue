<template>
  <service-control
    :service="service"
    :instance_uuid="localInstance.uuid"
    :chip-color="chipColor"
    @closePanel="openedlocalInstances = {}"
  />
</template>

<script>
import serviceControl from "@/components/modules/ovh/serviceControls.vue";

export default {
  name: "service-state",
  components: { serviceControl },
  props: {
    service: { type: Object, required: true },
    instance_uuid: { type: String, required: true },
    "chip-color": { type: String, required: true },
  },
};
</script>
