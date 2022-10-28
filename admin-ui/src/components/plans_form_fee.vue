<template>
  <v-row align="center" class="mt-2">
    <v-col cols="3">
      <v-subheader>Precision:</v-subheader>
    </v-col>
    <v-col cols="9">
      <v-text-field
        type="number"
        label="precision"
        v-model="fee.precision"
        :rules="rules.num"
        @change="(value) => changeValue('precision', value)"
      />
    </v-col>
    <v-col cols="3">
      <v-subheader>Rounding:</v-subheader>
    </v-col>
    <v-col cols="9">
      <v-radio-group
        row
        mandatory
        v-model="fee.round"
        @change="(value) => changeValue('round', value)"
      >
        <v-radio
          v-for="item of rounds"
          :key="item"
          :value="item"
          :label="item.toLowerCase()"
        />
      </v-radio-group>
    </v-col>
    <v-col cols="3">
      <v-subheader>Default:</v-subheader>
    </v-col>
    <v-col cols="9">
      <v-text-field
        type="number"
        label="default"
        v-model="fee.default"
        :rules="rules.num"
        @change="(value) => changeValue('default', value)"
      />
    </v-col>
    <v-col cols="3" align-self="start">
      <v-subheader>Ranges:</v-subheader>
    </v-col>
    <v-col cols="9">
      <v-text-field
        type="number"
        label="from"
        v-model="fee.ranges.from"
        :rules="rules.num"
        @change="(value) => changeValue('from', value)"
      />
      <v-text-field
        type="number"
        label="to"
        v-model="fee.ranges.to"
        :rules="rules.num"
        @change="(value) => changeValue('to', value)"
      />
      <v-text-field
        type="number"
        label="factor"
        v-model="fee.ranges.factor"
        :rules="rules.num"
        @change="(value) => changeValue('factor', value)"
      />
    </v-col>
  </v-row>
</template>

<script>
export default {
  data: () => ({
    fee: {
      precision: 0, round: 'NONE', default: 0.0,
      ranges: { from: 0.0, to: 0.0, factor: 0.0 }
    },
    rounds: ['NONE', 'FLOOR', 'ROUND', 'CEIL'],
    rules: { num: [(v) => !!v && isFinite(+v) || 'This field is number!'] }
  }),
  methods: {
    changeValue(key, value) {
      if (value[0] === '0') value = parseFloat(value);
      this.$emit('change', { key, value });
    }
  }
}
</script>
