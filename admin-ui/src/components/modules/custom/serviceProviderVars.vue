<template>
  <div class="pa-10">
    <v-row v-if="!editing">
      <v-col v-for="(variable, varTitle) in template.vars" :key="varTitle">
        {{ varTitle.replaceAll("_", " ") }}
        <v-row>
          <v-col :cols="12" v-for="(value, key) in variable.value" :key="key">
            <v-text-field
              readonly
              :value="JSON.stringify(value)"
              :label="key"
              style="display: inline-block; width: 200px"
            >
            </v-text-field>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
    <slot v-else></slot>
  </div>
</template>

<script>
export default {
  props: { template: { type: Object, required: true } },
  data: () => ({ editing: false })
}
</script>
