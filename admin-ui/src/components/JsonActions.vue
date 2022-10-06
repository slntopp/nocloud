<template>
  <v-row dense>
    <v-col class="column justify-end" cols="4">
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
    <v-col class="d-flex" cols="4">
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
        class="ml-2 mb-2"
        v-if="key !== null"
        @click="reset"
      >
        mdi-cancel
      </v-icon>
    </v-col>
    <v-col class="column" cols="4">
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
    addKeys (tree, prevKey = '/') {
      Object.entries(tree).forEach(([key, value]) => {
        if (typeof value === 'object') {
          this.keys.push(`${prevKey}${key}`)

          this.addKeys(value, `${prevKey}${key}/`)
        } else {
          this.keys.push(`${prevKey}${key}`)
        }
      })
    },
    changeKeys () {
      if (this.key === '/') {
        this.$emit('changeKey', '/')
      } else if (this.key.includes('/', 1)) {
        this.$emit('changeKey', this.key.replace('/', ''))
      } else {
        this.$emit('changeKey', this.key.replace('/', ''))
      }
      this.$emit('changeDisable')

      this.keys = ['/']
      this.addKeys(this.json)
    },
    reset () {
      this.key = null
      this.cancel()
    }
  },
  mounted () {
    this.addKeys(this.json)
  },
  watch: {
    json: {
      handler() {
        this.key = null
        this.keys = ['/']
        this.addKeys(this.json)
      },
      deep: true
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
