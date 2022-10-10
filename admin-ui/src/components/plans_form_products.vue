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
        <v-subheader>Amount of resources</v-subheader>
      </v-col>
      <v-col cols="9">
        <json-editor
          :json="amountResources"
          @changeValue="(data) => amountResources = data"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Resources</v-subheader>
      </v-col>
      <v-col cols="7" v-if="!isDialogVisible">
        <span v-for="(title, i) of form.titles" :key="title">
          "{{ title }}"
          <template v-if="i !== form.titles.length - 1">, </template>
        </span>
      </v-col>
      <v-col cols="2">
        <v-dialog width="fit-content" class="pa-4" v-model="isDialogVisible">
          <template v-slot:activator="{ on, attrs }">
            <v-btn v-on="on" v-bind="attrs">
              Add
            </v-btn>
          </template>

          <v-tabs v-model="form.title" background-color="background-light">
            <v-tab
              active-class="background"
              v-for="title of form.titles"
              :key="title"
              @dblclick="edit = {
                isVisible: true,
                title
              }"
            >
              {{ title }}
              <v-icon
                small
                right
                color="error"
                @click="removeConfig(title)"
              >
                mdi-close
              </v-icon>
            </v-tab>
            <v-text-field
              dense
              outlined
              class="mx-2 mt-1 mw-20"
              v-if="isVisible || edit.isVisible"
              :label="(edit.isVisible) ? `Edit ${edit.title}` : 'New config'"
              @change="addConfig"
            />
            <v-icon
              v-else
              class="mx-2"
              @click="isVisible = true"
            >
              mdi-plus
            </v-icon>
          </v-tabs>

          <v-divider />

          <v-subheader v-if="form.titles.length > 0" style="background: var(--v-background-base)">
            To edit the title, double-click the LMB
          </v-subheader>

          <v-tabs-items v-model="form.title">
            <v-tab-item
              active-class="background"
              v-for="(title, i) of form.titles"
              :key="title"
            >
              <v-row class="flex-column">
                <v-col>
                  <plans-form-resources
                    :keyForm="title"
                    :resource="resources[i]"
                    @change:resource="(data) => changeResource(i, data)"
                  />
                </v-col>
                <v-col>
                  <v-btn class="mb-4 ml-4" @click="setResources">
                    Add
                  </v-btn>
                </v-col>
              </v-row>
            </v-tab-item>
          </v-tabs-items>
        </v-dialog>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import dateField from '@/components/date.vue';
import JsonEditor from '@/components/JsonEditor.vue';
import plansFormResources from './plans_form_resources.vue';

export default {
  props: {
    product: Object,
    preset: { type: Object, default: () => ({}) }
  },
  components: { dateField, JsonEditor, plansFormResources },
  data: () => ({
    title: '',
    price: '',
    kind: '',
    kinds: ['POSTPAID', 'PREPAID'],
    resources: [],
    amountResources: {},
    fullDate: null,
    form: { title: '', titles: [] },
    edit: { isVisible: false, title: '' },
    isVisible: false,
    isDialogVisible: false,
    generalRule: [v => !!v || 'This field is required!'],
  }),
  methods: {
    changeResource(num, { key, value }) {
      try {
        value = JSON.parse(value, num);
      } catch {
        value;
      }

      if (key === 'date') {
        this.setPeriod(value, num);
        return;
      }
      if (this.resources[num]) {
        this.resources[num][key] = value;
      } else {
        this.resources.push({ [key]: value });
      }
    },
    setResources() {
      this.$emit('change:product', { key: 'resources', value: this.resources });
      this.isDialogVisible = false;
    },
    setPeriod(date, res) {
      const period = this.getTimestamp(date);

      this.resources[res].period = period;
    },
    getTimestamp({ day, month, year, quarter, week, time }) {
      year = +year + 1970;
      month = +month + quarter * 3 + 1;
      day = +day + week * 7 + 1;

      if (`${day}`.length < 2) {
        day = '0' + day;
      }
      if (`${month}`.length < 2) {
        month = '0' + month;
      }

      return Date.parse(`${year}-${month}-${day}T${time}Z`) / 1000;
    },
    addConfig(title) {
      if (this.edit.isVisible) {
        const i = this.form.titles.indexOf(this.edit.title);
        const j = this.resources.findIndex(({ key }) =>
          key === this.edit.title
        );

        this.resources[j].key = title;
        this.form.titles[i] = title;
        this.edit.isVisible = false;

        return;
      }

      this.form.titles.push(title);
      this.isVisible = false;
    },
    removeConfig(title) {
      this.form.titles = this.form.titles
        .filter((el) => el !== title);

      if (this.form.titles.length <= 0) {
        this.isVisible = true;
      }
    }
  },
  created() {
    if (!this.$route.params?.planId) {
      this.amountResources = this.preset;
      this.$emit('change:product', { key: 'amount', value: this.amountResources });
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
          this.$emit('change:product', { key: 'date', value: this.fullDate });
        } else if (key === 'amount') {
          this.amountResources = value;
        } else if (key === 'resources') {
          this.resources = value;
          value.forEach(({ key }) => { this.form.titles.push(key) });
        } else {
          this[key] = value;
        }
      });
  },
  watch: {
    amountResources() {
      this.$emit('change:product', { key: 'amount', value: this.amountResources });
    }
  }
}
</script>

<style scoped>
.mw-20 {
  max-width: 150px;
}
</style>
