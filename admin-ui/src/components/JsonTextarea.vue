<template>
  <v-textarea
    label="JSON"
    v-model="tree"
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
      let count = 0

      this.tree = JSON.stringify(this.json)
        .split('')
        .map((simbol) => {
          switch (simbol) {
            case '{':
              count++
              return `{\n${'\t'.repeat(count)}`
            case '}':
              count--
              return `\n${'\t'.repeat(count)}}`
            case ':':
              return ': '
            case ',':
              return `,\n${'\t'.repeat(count)}`
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
          return !!JSON.parse(v)
        } catch (e) {
          return e.message
        }
      }]
    }
  }
}
</script>
