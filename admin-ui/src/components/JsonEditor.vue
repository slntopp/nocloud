<template>
  <v-form v-model="isValid" ref="form">
    <v-row>
      <v-col cols="12">
        <json-textarea
          :json="json"
          :disabled="disabled"
          @getTree="getTree"
        />
      </v-col>
    </v-row>
    <json-actions
      :json="json"
      :isValid="isValid"
      :disabled="disabledDelete"
      :changeField="changeField"
      :deleteField="deleteField"
      :cancel="cancel"
      @changeKey="(value) => fieldKey = value"
      @changeDisable="changeFields"
    />
    <json-form
      :fieldKey="fieldKey"
      :newKey="newKey"
      :newValue="newValue"
      :typeValue="typeValue"
      :add="add"
      @changeValue="changeValue"
      @changeAdd="(value) => add = value"
    />
  </v-form>
</template>

<script>
import JsonActions from '@/components/JsonActions'
import JsonForm from '@/components/JsonForm'
import JsonTextarea from '@/components/JsonTextarea'

export default {
  name: 'JsonEditor',
  components: { JsonForm, JsonTextarea, JsonActions },
  props: {
    json: { type: Object, required: true }
  },
  data: () => ({
    tree: '',
    isValid: false,
    fieldKey: '',
    newKey: '',
    newValue: '',
    typeValue: '',
    disabled: false,
    disabledDelete: true,
    add: false
  }),
  methods: {
    changeValue ({ key, value }) {
      this[key] = value
    },
    getTree (tree) {
      try {
        JSON.parse(tree)

        this.tree = tree
        this.isValid = true
      } catch {
        this.isValid = false
      }
    },
    deleteField () {
      if (this.fieldKey !== '/') {
        const tree = JSON.parse(JSON.stringify(this.json))

        const path = this.fieldKey.split('/')
        const key = path.pop()
        let node = tree

        path.forEach((el) => { node = node[el] })
        delete node[key]

        this.$emit('changeValue', tree)
      }

      this.cancel()
    },
    changeField () {
      if (this.tree !== '') {
        const tree = JSON.parse(this.tree)

        this.$emit('changeValue', tree)
        this.disabled = true
        this.isValid = false
        this.tree = ''

        setTimeout(this.cancel)
        return
      }

      if (this.typeValue === 'object' && this.add) {
        this.newValue = '{}'
      }

      const value = (this.typeValue === 'string')
        ? this.newValue
        : JSON.parse(this.newValue)
      const tree = JSON.parse(JSON.stringify(this.json))

      this.findNode(tree, value)
      this.$emit('changeValue', tree)
      this.cancel()
    },
    findNode (tree, newValue) {
      if (this.fieldKey === '/') {
        tree[this.newKey] = newValue
        return
      }
      const path = this.fieldKey.split('/')
      const key = path.pop()
      let node = tree

      path.forEach((el) => { node = node[el] })

      const isObject = typeof node[key] === 'object'
      const isUndefined = newValue === undefined

      if (isObject && !isUndefined && this.add) {
        node[key][this.newKey] = newValue
      } else if (!isUndefined) {
        delete node[key]

        node[this.newKey] = newValue
      } else {
        return { key, value: node[key], type: typeof node[key] }
      }
    },
    changeFields () {
      if (this.fieldKey === '/') {
        this.newKey = ''
        this.typeValue = ''
        this.newValue = ''
        this.disabledDelete = true
      } else {
        const tree = JSON.parse(JSON.stringify(this.json))
        const { key, value, type } = this.findNode(tree)

        this.newKey = key
        this.typeValue = type
        this.disabledDelete = false

        if (type === 'string') {
          this.newValue = value
        } else {
          this.newValue = JSON.stringify(value)
        }
      }

      this.disabled = true
    },
    cancel () {
      this.newKey = ''
      this.typeValue = ''
      this.newValue = ''

      this.$refs.form.resetValidation();
      this.disabledDelete = true
      this.disabled = false
    }
  }
}
</script>
