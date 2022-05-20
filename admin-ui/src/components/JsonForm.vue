<template>
  <v-row dense>
    <v-col>
      <v-text-field
        dense
        required
        label="New key"
        v-if="add"
        :value="newKey"
        :rules="generalRule"
        @change="(value) => changeValue('newKey', value)"
      />
      <v-text-field
        dense
        required
        label="Key"
        v-else
        :value="newKey"
        :rules="generalRule"
        @change="(value) => changeValue('newKey', value)"
      />
      <v-icon
        dark
        dense
        color="yellow"
        v-if="add"
        @click="() => $emit('changeAdd', false)"
      >
        mdi-minus
      </v-icon>
      <v-icon
        dark
        dense
        color="green"
        v-else-if="typeValue === 'object'"
        @click="() => $emit('changeAdd', true)"
      >
        mdi-plus
      </v-icon>
    </v-col>
    <v-col>
      <v-select
        dense
        required
        label="Choose type"
        :value="typeValue"
        :items="typesValue"
        :rules="generalRule"
        @change="(value) => changeValue('typeValue', value)"
      />
    </v-col>
    <v-col>
      <v-select
        dense
        required
        label="Value"
        v-if="typeValue === 'boolean'"
        :value="newValue"
        :items="['true', 'false']"
        :rules="generalRule"
        @change="(value) => changeValue('newValue', value)"
      />
      <v-text-field
        dense
        disabled
        label="Value"
        value="{}"
        v-else-if="typeValue === 'object'"
      />
      <v-text-field
        dense
        label="Value"
        v-else
        :value="newValue"
        :rules="typeRule"
        @change="(value) => changeValue('newValue', value)"
      />
    </v-col>
  </v-row>
</template>

<script>
export default {
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
      }
    }
  },
  computed: {
    typeRule () {
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
