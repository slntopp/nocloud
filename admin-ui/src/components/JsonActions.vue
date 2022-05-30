<template>
  <v-row dense>
    <v-col class="column" style="justify-content:flex-end">
      <v-btn
        dark
        small
        color="success"
        :disabled="!isValid"
        @click="changeField"
      >
        Save
      </v-btn>
    </v-col>
    <v-col>
      <v-select
        dense
        required
        label="Choose key"
        v-model="key"
        :items="keys"
        :rules="[v => !!v || 'This field is required!']"
        @change="changeKeys"
      />
      <v-icon
        dark
        small
        color="red"
        v-if="key !== null"
        @click="cancel"
      >
        mdi-cancel
      </v-icon>
    </v-col>
    <v-col class="column">
      <v-btn
        dark
        small
        color="error"
        :disabled="disabled"
        @click="deleteField"
      >
        Delete
      </v-btn>
    </v-col>
  </v-row>
</template>

<script>
export default {
  props: {
    json: { type: Object, required: true },
    isValid: { type: Boolean, required: true },
    disabled: { type: Boolean, required: true },
    changeField: { type: Function, required: true },
    deleteField: { type: Function, required: true },
    cancel: { type: Function, required: true }
  },
  data: () => ({
    key: null,
    keys: ['/']
  }),
  methods: {
    addKeys (tree, prevKey = '') {
      Object.entries(tree).forEach(([key, value]) => {
        if (typeof value === 'object') {
          this.keys.push(`${prevKey}${key}`)

          this.addKeys(value, `${key}/`)
        } else {
          this.keys.push(`${prevKey}${key}`)
        }
      })
    },
    changeKeys () {
      if (this.key === null) {
        this.$emit('changeKey', '')
      } else if (this.key.includes('/', 1)) {
        this.$emit('changeKey', this.key.split('/')[1])
      } else {
        this.$emit('changeKey', this.key)
      }
      this.$emit('changeDisable')

      this.keys = ['/']
      this.addKeys(this.json)
    }
  },
  mounted () {
    this.addKeys(this.json)
  },
  watch: {
    isValid () {
      this.addKeys(this.json)
    }
  }
}
</script>

<style scoped>
.column {
  display: flex;
  align-items: center;
}
</style>
