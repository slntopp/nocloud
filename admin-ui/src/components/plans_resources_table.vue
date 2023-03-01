<template>
  <div class="pa-4">
    <nocloud-table
      item-key="id"
      v-model="selected"
      :show-expand="true"
      :items="resources"
      :headers="headers"
      :expanded.sync="expanded"
    >
      <template v-slot:top>
        <v-toolbar flat color="background">
          <v-toolbar-title>Actions</v-toolbar-title>
          <v-divider inset vertical class="mx-4" />
          <v-spacer />

          <v-btn class="mr-2" color="background-light" @click="addConfig">
            Create
          </v-btn>
          <confirm-dialog @confirm="removeConfig">
            <v-btn color="background-light" :disabled="selected.length < 1">Delete</v-btn>
          </confirm-dialog>
        </v-toolbar>
      </template>
      <template v-slot:[`item.key`]="{ item }">
        <v-text-field
          dense
          :value="item.key"
          :rules="generalRule"
          @change="(value) => changeResource('key', value, item.id)"
        />
      </template>
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field
          dense
          type="number"
          :value="item.price"
          :rules="generalRule"
          @input="(value) => changeResource('price', value, item.id)"
        />
      </template>
      <template v-slot:[`item.period`]="{ item }">
        <date-field
          :period="fullDate[item.id]"
          @changeDate="(value) => changeDate(value, item.id)"
        />
      </template>
      <template v-slot:[`item.kind`]="{ item }">
        <v-radio-group
          row mandatory
          :value="item.kind"
          @change="(value) => changeResource('kind', value, item.id)"
        >
          <v-radio
            v-for="(kind, i) of kinds"
            :style="{ marginRight: (i === kinds.length - 1) ? 0 : 16 }"
            :key="kind"
            :value="kind"
            :label="kind.toLowerCase()"
          />
        </v-radio-group>
      </template>
      <template v-slot:expanded-item="{ headers, item }">
        <td />
        <td :colspan="headers.length - 1">
          <v-select
            dense
            multiple
            label="State"
            class="d-inline-block mr-4"
            :value="item.on"
            :items="states"
            :rules="generalRule"
            @change="(value) => changeResource('on', value, item.id)"
          />
          <v-switch
            label="Except"
            class="d-inline-block"
            :value="item.except"
            @change="(value) => changeResource('except', value, item.id)"
          />
        </td>
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import { ref, toRefs } from 'vue';
import nocloudTable from '@/components/table.vue';
import dateField from '@/components/date.vue';
import confirmDialog from '@/components/confirmDialog.vue';

const props = defineProps({
  resources: { type: Array, required: true }
});
const emits = defineEmits(['change:resource']);
const { resources } = toRefs(props);

const fullDate = ref({});
const selected = ref([]);
const expanded = ref([]);
const generalRule = [v => !!v || 'This field is required!'];
const kinds = ['POSTPAID', 'PREPAID'];

const states = [
  'INIT',
  'UNKNOWN',
  'STOPPED',
  'RUNNING',
  'FAILURE' ,
  'DELETED',
  'SUSPENDED',
  'OPERATION'
];
const headers = [
  { text: 'Key', value: 'key' },
  { text: 'Price', value: 'price' },
  { text: 'Period', value: 'period' },
  { text: 'Kind', value: 'kind', width: 228 }
];

function changeDate({ value }, id) {
  fullDate.value[id] = value;
  emits("change:resource", { key: 'date', value, id });
}

function changeResource(key, value, id) {
  console.log(value);
  emits('change:resource', { key, value, id });
}

function addConfig() {
  const value = [...resources.value];

  value.push({
    key: '',
    kind: 'POSTPAID',
    price: 0,
    period: 0,
    except: false,
    on: [],
    id: Math.random().toString(16).slice(2)
  });
  changeResource('resources', value);
}

function removeConfig() {
  const value = resources.value.filter(({ id }) =>
    !selected.value.find((el) => el.id === id)
  );
  changeResource('resources', value);
}

resources.value.forEach(({ period, id }) => {
  const date = new Date(period * 1000);
  const time = date.toUTCString().split(' ');

  fullDate.value[id] = {
    day: `${date.getUTCDate() - 1}`,
    month: `${date.getUTCMonth()}`,
    year: `${date.getUTCFullYear() - 1970}`,
    quarter: '0',
    week: '0',
    time: time.at(-2)
  };
});
</script>
