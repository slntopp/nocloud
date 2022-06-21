<template>
  <v-row align="center">
    <v-col cols="3">
      <v-subheader>Period</v-subheader>
    </v-col>
    <v-col :cols="(this.date === 'Custom') ? 8 : 7">
      <v-select
        label="Period"
        v-model="date"
        :items="typesDate"
        :rules="rules.general"
      />
    </v-col>
    <v-col cols="2" v-if="this.date !== 'Custom'">
      <v-text-field
        v-model="amountDate"
        v-if="date === 'Time'"
        :rules="rules.time"
      />
      <v-text-field
        v-else
        type="number"
        v-model="amountDate"
        :rules="rules.number"
      />
    </v-col>
    <v-col cols="1" v-else>
      <v-menu
        left
        v-model="menuVisible"
        :close-on-content-click="false"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-icon v-bind="attrs" v-on="on">
            mdi-playlist-edit
          </v-icon>
        </template>
  
        <v-card>
          <v-list class="columns-2">
            <v-list-item v-for="item of items" :key="item.title">
              <v-list-item-title >{{ item.title }}</v-list-item-title>
              <v-list-item-action>
                <v-text-field
                  v-model="fullDate[item.model]"
                  :type="(item.model === 'time')
                    ? 'text'
                    : 'number'
                  "
                  :rules="(item.model === 'time')
                    ? rules.time
                    : rules.number
                  "
                />
              </v-list-item-action>
            </v-list-item>
          </v-list>
  
          <v-card-actions>
            <v-spacer />
            <v-btn text @click="() => resetDate(this.fullDate)">
              Reset
            </v-btn>
            <v-btn text color="primary" @click="menuVisible = false">
              Close
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-menu>
    </v-col>
  </v-row>
</template>

<script>
export default {
  name: 'date-field',
  props: { period: Object },
  data: () => ({
    date: '',
    amountDate: '0',

    fullDate: {
      day: '0',
      month: '0',
      year: '0',
      quarter: '0',
      week: '0',
      time: '00:00:00'
    },
    typesDate: [
      'Day',
      'Week',
      'Month',
      'Quarter',
      'Year',
      'Time',
      'Hour',
      'Minute',
      'Custom'
    ],
    items: [
      { title: 'Day', model: 'day' },
      { title: 'Week', model: 'week' },
      { title: 'Month', model: 'month' },
      { title: 'Quarter', model: 'quarter' },
      { title: 'Year', model: 'year' },
      { title: 'Time', model: 'time' },
      { title: 'Hour' },
      { title: 'Minute' }
    ],

    rules: {
      general: [v => !!v || 'This field is required!'],
      number: [
        value => !!value || 'Is required!',
        value =>
          /^[1-9][0-9]{0,1}|0$/.test(value) || 'Invalid!'
      ],
      time: [
        value => !!value || 'Is required!',
        value =>
          /^([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/.test(value) || 'Invalid!'
      ]
    },

    menuVisible: false,
  }),
  methods: {
    resetDate(date) {
      Object.keys(date).forEach((key) => {
        date[key] = (key === 'time') ? '00:00:00' : '0';
      });
    }
  },
  mounted() {
    if (this.period) {
      this.date = 'Custom';
      this.fullDate = this.period;
    }
  },
  watch: {
    date () {
      if (this.date === 'Custom') return;

      const key = this.date.toLowerCase();
      const value = (key === 'time') ? '00:00:00' : '1';

      this.resetDate(this.fullDate);

      this.fullDate[key] = value;
      this.amountDate = value;
    },
    amountDate () {
      if (this.date === '') return;

      let key = this.date.toLowerCase();
      let value = this.amountDate;
      const newValue = (value.length < 2)
        ? `0${value}`
        : value;

      switch (key) {
        case 'hour':
          key = 'time';
          value = `${newValue}:00:00`;
          break;
        case 'minute':
          key = 'time';
          value = `00:${newValue}:00`;
      }

      this.resetDate(this.fullDate);

      this.fullDate[key] = value;
    },
    fullDate: {
      handler() {
        this.$emit('changeDate', this.fullDate);
      },
      deep: true
    },
    period() {
      if (this.period) {
        this.date = 'Custom';
        this.fullDate = this.period;
      }
    }
  }
}
</script>

<style scoped lang="scss">
.columns-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
}
</style>
