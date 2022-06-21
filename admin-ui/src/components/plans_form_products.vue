<template>
  <div class="pa-4">
    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Title</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-text-field
          label="Title"
          v-model="title"
          :rules="generalRule"
          @change="$emit('change:product', { key: 'title', value: title })"
        />
      </v-col>
    </v-row>

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
          @change="$emit('change:product', { key: 'price', value: price })"
        />
      </v-col>
    </v-row>

    <date-field
      :period="fullDate"
      @changeDate="(date) => $emit('change:product', {
        key: 'date',
        value: date
      })"
    />

    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Kind</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-radio-group
          row
          mandatory
          v-model="kind"
          @change="$emit('change:product', { key: 'kind', value: kind })"
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

    <v-row>
      <v-col cols="3">
        <v-subheader>Resources</v-subheader>
      </v-col>
      <v-col cols="9">
        <json-editor
          :json="resources"
          @changeValue="(data) => resources = data"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import dateField from '@/components/date.vue';
import JsonEditor from '@/components/JsonEditor.vue';

export default {
  props: {
    product: Object,
    preset: { type: Object, default: () => ({}) }
  },
  components: { dateField, JsonEditor },
  data: () => ({
    title: '',
    price: '',
    kind: '',
    kinds: ['POSTPAID', 'PREPAID'],
    resources: {},
    fullDate: null,

    generalRule: [v => !!v || 'This field is required!'],
  }),
  created() {
    if (!this.$route.params?.planId) {
      this.resources = this.preset;
      this.$emit('change:product', { key: 'resources', value: this.resources });
    }
    if (!this.product) {
      this.$emit('change:product', { key: 'kind', value: this.kind });

      return;
    }

    Object.entries(this.product)
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
          this.$emit('change:product', {
            key: 'date',
            value: this.fullDate
          });
        } else {
          this[key] = value;
        }
      });
  },
  watch: {
    resources() {
      this.$emit('change:product', { key: 'resources', value: this.resources });
    }
  }
}
</script>
