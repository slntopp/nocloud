<template>
  <div class="pa-4">
    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Price</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-text-field
          type="number"
          label="Price"
          v-model="price"
          :rules="generalRule"
          @change="$emit('changeValue', { key: 'price', value: price })"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Period</v-subheader>
      </v-col>
      <v-col :cols="(this.date === 'Custom') ? 8 : 7">
        <v-select
          label="Period"
          v-model="date"
          :items="typesDate"
          :rules="generalRule"
          @change="$emit('changeValue', { key: 'date', value: fullDate })"
        />
      </v-col>
      <v-col cols="2" v-if="this.date !== 'Custom'">
        <v-text-field
          v-model="amountDate"
          v-if="date === 'Time'"
          :rules="timeRules"
          @change="$emit('changeValue', { key: 'date', value: fullDate })"
        />
        <v-text-field
          v-else
          type="number"
          v-model="amountDate"
          :rules="numRules"
          @change="$emit('changeValue', { key: 'date', value: fullDate })"
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
                      ? timeRules
                      : numRules
                    "
                    @change="$emit('changeValue', { key: 'date', value: fullDate })"
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

    <v-row align="center">
      <v-col cols="3">
        <v-subheader>State</v-subheader>
      </v-col>
      <v-col cols="6">
        <v-select
          multiple
          label="State"
          v-model="state"
          :items="states"
          :rules="generalRule"
          @change="$emit('changeValue', { key: 'on', value: state })"
        />
      </v-col>
      <v-col cols="3" class="d-flex justify-end">
        <v-switch
          label="Except"
          v-model="except"
          @change="$emit('changeValue', { key: 'except', value: except })"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Kind</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-radio-group row v-model="kind">
          <v-radio
            v-for="item of kinds"
            :key="item"
            :value="item"
            :label="item.toLowerCase()"
            @change="$emit('changeValue', { key: 'kind', value: kind })"
          />
        </v-radio-group>
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  props: {
    data: { type: Object },
    keyForm: { type: String }
  },
  data: () => ({
    price: '',
    date: '',
    state: '',
    kind: 'POSTPAID',
    except: false,

    fullDate: {
      day: '0',
      month: '0',
      year: '0',
      quarter: '0',
      week: '0',
      time: '00:00:00'
    },
    amountDate: 0,
    menuVisible: false,

    generalRule: [v => !!v || 'This field is required!'],
    numRules: [
      value => !!value || 'Is required!',
      value =>
        /^[1-9][0-9]{0,1}|0$/.test(value) || 'Invalid!'
    ],
    timeRules: [
      value => !!value || 'Is required!',
      value =>
        /^([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$/.test(value) || 'Invalid!'
    ],

    kinds: ['POSTPAID', 'PREPAID'],
    typesDate: [
      'Day',
      'Week',
      'Month',
      'Quarter',
      'Year',
      'Time',
      'Custom'
    ],
    states: [
      'INIT',
      'UNKNOWN',
      'STOPPED',
      'RUNNING',
      'FAILURE' ,
      'DELETED',
      'SUSPENDED',
      'OPERATION'
    ],
    items: [
      {
        title: 'Day',
        model: 'day'
      },
      {
        title: 'Week',
        model: 'week'
      },
      {
        title: 'Month',
        model: 'month'
      },
      {
        title: 'Quarter',
        model: 'quarter'
      },
      {
        title: 'Year',
        model: 'year'
      },
      {
        title: 'Time',
        model: 'time'
      }
    ]
  }),
  methods: {
    resetDate(date) {
      Object.keys(date).forEach((key) => {
        date[key] = (key === 'time')
          ? '00:00:00'
          : '0';
      });
    }
  },
  mounted() {
    this.$emit('changeValue', { key: 'key', value: this.keyForm });
    this.$emit('changeValue', { key: 'kind', value: this.kind });
    this.$emit('changeValue', { key: 'except', value: this.except });
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

      const key = this.date.toLowerCase();
      const value = this.amountDate;

      this.resetDate(this.fullDate);

      this.fullDate[key] = value;
    },
    data () {
      if (!this.data) return;

      Object.entries(this.data)
        .forEach(([key, value]) => {
          if (key === 'on') {
            this.state = value;
          } else if (key === 'period') {
            const date = new Date(value * 1000);
            const time = date.toUTCString().split(' ');

            this.fullDate = {
              day: `${date.getUTCDate() - 1}`,
              month: `${date.getUTCMonth()}`,
              year: `${date.getUTCFullYear() - 1970}`,
              quarter: '0',
              week: '0',
              time: time[time.length - 2]
            };
            this.date = 'Custom';
            this.$emit('changeValue', {
              key: 'date',
              value: this.fullDate
            });
          } else {
            this[key] = `${value}`;
          }
        })
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
