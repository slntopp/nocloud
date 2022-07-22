<template>
  <v-textarea
    label="JSON"
    v-model="tree"
    :rows="rows"
    :disabled="disabled"
    :rules="typeRule"
    @keyup="formatting"
    @change="$emit('getTree', tree)"
  />
</template>

<script>
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
      const tree = JSON.stringify(this.json)
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
              return ': '
            case ',':
              if (this.amountQuotes(i, tree)) {
                return simbol
              }
              return `,\n${'\t'.repeat(count)}`
            default:
              return simbol
          }
        })
        .join('')
    },
    amountQuotes(num, tree) {
      let quotes = 0

      tree
        .slice(0, num)
        .split('')
        .forEach((simbol) => {
          if (simbol === '"' && quotes) {
            quotes--
          } else if (simbol === '"') {
            quotes++
          }
        })
      
      return quotes
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
          return !!JSON.parse(v)
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
