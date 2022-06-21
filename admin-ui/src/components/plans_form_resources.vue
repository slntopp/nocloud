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
          @change="$emit('change:resource', { key: 'price', value: price })"
        />
      </v-col>
    </v-row>

    <date-field
      :period="fullDate"
      @changeDate="(date) => $emit('change:resource', {
        key: 'date',
        value: date
      })"
    />

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
          @change="$emit('change:resource', { key: 'on', value: state })"
        />
      </v-col>
      <v-col cols="3" class="d-flex justify-end">
        <v-switch
          label="Except"
          v-model="except"
          @change="$emit('change:resource', { key: 'except', value: except })"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Kind</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-radio-group
          row
          mandatory
          v-model="kind"
          @change="$emit('change:resource', { key: 'kind', value: kind })"
        >
          <v-radio
            v-for="item of kinds"
            :key="item"
            :value="item"
            :label="item.toLowerCase()"
          />
        </v-radio-group>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import dateField from '@/components/date.vue';

export default {
  props: {
    resource: Object,
    keyForm: String
  },
  components: { dateField },
  data: () => ({
    price: '',
    state: '',
    kind: '',
    except: false,
    fullDate: null,

    generalRule: [v => !!v || 'This field is required!'],
    kinds: ['POSTPAID', 'PREPAID'],
    states: [
      'INIT',
      'UNKNOWN',
      'STOPPED',
      'RUNNING',
      'FAILURE' ,
      'DELETED',
      'SUSPENDED',
      'OPERATION'
    ]
  }),
  created() {
    if (!this.resource) {
      this.$emit('change:resource', { key: 'key', value: this.keyForm });
      this.$emit('change:resource', { key: 'kind', value: this.kind });
      this.$emit('change:resource', { key: 'except', value: this.except });

      return;
    }

    Object.entries(this.resource)
      .forEach(([key, value]) => {
        if (key === 'period') {
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
          this.$emit('change:resource', {
            key: 'date',
            value: this.fullDate
          });
        } else if (key === 'on') {
          this.state = value;
        } else {
          this[key] = value;
        }
      });
  }
}
</script>
