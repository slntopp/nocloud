<template>
  <v-row density="compact">
    <v-col class="d-flex align-center">
      <v-text-field
        density="compact"
        required
        label="New key"
        v-if="add"
        :model-value="newKey"
        :rules="generalRule"
        @change="changeValue('newKey', $event.target.value)"
      />
      <v-text-field
        density="compact"
        required
        label="Key"
        v-else
        :model-value="newKey"
        :rules="generalRule"
        @change="changeValue('newKey', $event.target.value)"
      />
      <v-icon
        
        size="small"
        class="ml-2 mb-2"
        color="yellow"
        v-if="add"
        @click="() => $emit('changeAdd', false)"
      >
        mdi-minus
      </v-icon>
      <v-icon
        
        size="small"
        class="ml-2 mb-2"
        color="green"
        v-else-if="typeValue === 'object'"
        @click="() => $emit('changeAdd', true)"
      >
        mdi-plus
      </v-icon>
    </v-col>
    <v-col>
      <v-select
        density="compact"
        required
        label="Choose type"
        :model-value="typeValue"
        :items="typesValue"
        :rules="generalRule"
        @update:model-value="(value) => changeValue('typeValue', value)"
      />
    </v-col>
    <v-col>
      <v-select
        density="compact"
        required
        v-if="typeValue === 'boolean'"
        :label="(add) ? 'New value' : 'Value'"
        :model-value="newValue"
        :items="['true', 'false']"
        :rules="generalRule"
        @update:model-value="(value) => changeValue('newValue', value)"
      />
      <v-text-field
        density="compact"
        disabled
        :label="(add) ? 'New value' : 'Value'"
        model-value="{}"
        v-else-if="typeValue === 'object'"
      />
      <v-text-field
        density="compact"
        :label="(add) ? 'New value' : 'Value'"
        v-else
        :model-value="newValue"
        :rules="typeRule"
        @change="changeValue('newValue', $event.target.value)"
      />
    </v-col>
  </v-row>
</template>

<script>
export default {
  emits: ["changeAdd", "changeValue"],
  props: {
    fieldKey: { type: String, required: true },
    newKey: { type: String, required: true },
    newValue: { type: String, required: true },
    typeValue: { type: String, required: true },
    add: { type: Boolean, required: true }
  },
  data: () => ({
    typesValue: ['string', 'number', 'boolean', 'object'],
    generalRule: [v => !!v || 'This field is required!']
  }),
  methods: {
    changeValue (key, value) {
      this.$emit('changeValue', { key, value })
    }
  },
  watch: {
    typeValue () {
      if (this.typeValue === '') {
        this.$emit('changeAdd', false)
      } else if (this.typeValue === 'object') {
        this.changeValue('newValue', '{}')
      }
    }
  },
  computed: {
    typeRule () {
      if (this.typeValue === '') return [];
      return [v => {
        try {
          const value = (this.typeValue !== 'string') ? JSON.parse(v) : v

          switch (this.typeValue) {
            case typeof value:
              return !!v
            default:
              throw TypeError('Invalid value type!')
          }
        } catch (e) {
          return e.message
        }
      }]
    }
  }
}
</script>
