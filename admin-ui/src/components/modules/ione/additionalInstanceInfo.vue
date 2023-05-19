<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="State"
          :value="template.state.meta?.state_str || template.state.state"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Lcm state"
          :value="template.state?.meta.lcm_state_str"
        />
      </v-col>
      <v-col>
        <instance-ip-menu :item="template" />
      </v-col>
      <v-col>
        <instance-ip-menu type="private" :item="template" />
      </v-col>
      <v-col>
        <v-text-field readonly :value="os" label="OS login" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Pass"
          :type="isVisible ? 'text' : 'password'"
          :value="template.config.password"
          :append-icon="isVisible ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append="isVisible = !isVisible"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field readonly :value="template.resources.cpu" label="CPU" />
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.resources.ram" label="RAM" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.resources.drive_type"
          label="Disk type"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.resources.drive_size"
          label="Disk size"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.resources.cpu"
          label="Software"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { toRefs, defineProps, ref, computed } from "vue";
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";

const props = defineProps(["template", "sp"]);

const { template } = toRefs(props);

const isVisible = ref(false);

const os = computed(() => {
  const id = props.template.config.template_id;

  return props.sp?.publicData.templates[id]?.name;
});
</script>

<style scoped></style>
