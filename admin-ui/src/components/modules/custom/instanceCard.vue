<template>
  <instance-card
    :chipColor="chipColor"
    :type="group.type"
    :editing="editing"
    :instance="instance"
    :service="service"
  ></instance-card>
</template>

<script>
import instanceCard from "@/components/modules/ione/instanceCard.vue";

export default {
  name: "instance-card",
  components: { instanceCard },
  props: {
    instance: { type: Object },
    service: { type: Object },
    editing: { type: Boolean },
    type: { type: String },
    chipColor: { type: String },
  },
};
</script>
