<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="OpenNebula State"
          :value="template.state.meta?.state_str || template.state.state"
        />
      </v-col>
      <v-col>
        <instance-ip-menu :item="template" />
      </v-col>
      <v-col>
        <instance-ip-menu type="private" :item="template" />
      </v-col>
      <v-col>
        <v-text-field readonly :value="os" label="OS"/>
      </v-col>
      <v-col>
        <password-text-field
          :value="template.config.password"
          copy
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
import { toRefs, defineProps, computed } from "vue";
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";
import PasswordTextField from "@/components/ui/passwordTextField.vue";

const props = defineProps(["template", "sp"]);

const { template } = toRefs(props);

const os = computed(() => {
  const id = props.template.config.template_id;

  return props.sp?.publicData.templates[id]?.name;
});
</script>

<style scoped></style>
