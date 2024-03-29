<template>
  <v-textarea
    label="YAML"
    v-model="tree"
    :rows="rows"
    :disabled="disabled"
    :rules="typeRule"
    @keyup="formatting"
    @change="$emit('getTree', tree)"
  />
</template>

<script>
import yaml from 'yaml';

export default {
  props: {
    json: { type: Object, required: true },
    disabled: { type: Boolean }
  },
  data: () => ({
    tree: ''
  }),
  methods: {
    changeTree () {
      const tree = yaml.stringify(this.json)
      let count = 0

      this.tree = tree
        .split('')
        .map((simbol, i, arr) => {
          switch (simbol) {
            case '{':
              count++
              if (arr[i + 1] === '}') return simbol
              return `{\n${'\t'.repeat(count)}`
            case '}':
              count--
              if (arr[i - 1] === '{') return simbol
              return `\n${'\t'.repeat(count)}}`
            case ':':
              for (let j = i; arr[j] !== '\n'; j--) {
                if (arr[j] === ':') return simbol
              }
              return ': '
            default:
              return simbol
          }
        })
        .join('')
    },
    formatting ({ target, key }) {
      const start = target.selectionStart
      const endString = target.value
        .slice(start)
      const count = () => endString
        .split('')
        .filter((simbol) => simbol === '}')
        .length
      let string = ''

      switch (key) {
        case '{':
          string = '}'
          break
        case '"':
          string = '"'
          break
        case ':':
          string = ' '
          break
        case 'Enter':
          if (endString[0] === '}') {
            string = `${'\t'.repeat(count())}\n` +
              '\t'.repeat(count() - 1)
          } else {
            string = '\t'.repeat(count())
          }
          break
      }

      this.tree = target.value
        .slice(0, start)
        .concat(string, endString)

      setTimeout(() => {
        const num = start + count()

        switch (key) {
          case ':':
            target.setSelectionRange(start + 1, start + 1)
            break
          case 'Enter':
            target.setSelectionRange(num, num)
            break
          default:
            target.setSelectionRange(start, start)
        }
      })
    }
  },
  beforeMount () {
    this.changeTree()
  },
  watch: {
    disabled () {
      setTimeout(this.changeTree)
    }
  },
  computed: {
    typeRule () {
      return [v => {
        try {
          return !!yaml.parse(v)
        } catch (e) {
          return e.message
        }
      }]
    },
    rows() {
      let rows = 0

      for (let i = 0; i < this.tree.length; i++) {
        if (this.tree[i] === '\n') rows++
      }
      return rows;
    }
  }
}
</script>
