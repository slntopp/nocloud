<template>
  <nocloud-table
    show-expand
    class="mt-4"
    sort-by="proc"
    sort-desc
    :items="transactions"
    :headers="headers"
    :loading="isLoading"
    :expanded.sync="expanded"
    :footer-error="fetchError"
    @input="selectTransaction"
  >
    <template v-slot:[`item.account`]="{ item }">
      {{ account(item.account) }}
    </template>

    <template v-slot:[`item.service`]="{ item, index }">
      <template v-if="item.service">
        <router-link
          :to="{ name: 'Service', params: { serviceId: item.service } }"
        >
          {{ service(item.service) }}
        </router-link>

        <v-icon
          class="ml-2"
          v-if="!visibleItems.includes(index)"
          @click="visibleItems.push(index)"
        >
          mdi-eye-outline
        </v-icon>
        <template v-else>
          ({{ hashTrim(item.service) }})
          <v-btn icon @click="addToClipboard(item.service, index)">
            <v-icon v-if="copyed === index"> mdi-check </v-icon>
            <v-icon v-else> mdi-content-copy </v-icon>
          </v-btn>
        </template>
      </template>
      <template v-else>-</template>
    </template>

    <template v-slot:[`item.total`]="{ item }">
      <balance :value="-item.total" />
    </template>
    <template v-slot:[`item.proc`]="{ item }">
      {{ date(item.proc) }}
    </template>

    <template v-slot:expanded-item="{ headers, item }">
      <td :colspan="headers.length" style="padding: 0">
        <nocloud-table
          class="mx-8"
          style="background: var(--v-background-base) !important"
          :show-select="false"
          :loading="isRecordsLoading"
          :items="records[item.uuid] || []"
          :headers="recordHeaders"
        >
          <template v-slot:[`item.instance`]="{ item, index }">
            <v-icon
              class="ml-2"
              v-if="!visibleRecords.includes(`${index}.${item.uuid}`)"
              @click="visibleRecords.push(`${index}.${item.uuid}`)"
            >
              mdi-eye-outline
            </v-icon>
            <template v-else>
              {{ hashTrim(item.instance) }}
            </template>
            <v-btn icon @click="addToClipboard(item.instance, index)">
              <v-icon v-if="copyed === index"> mdi-check </v-icon>
              <v-icon v-else> mdi-content-copy </v-icon>
            </v-btn>
          </template>
          <template v-slot:[`header.product`]="{ header }">
            {{
              records[item.uuid] && records[item.uuid][0].product
                ? header.text
                : "Resource"
            }}
          </template>
          <template v-slot:[`item.product`]="{ item }">
            {{
              item.product
                ? item.product.replaceAll("_", " ").toUpperCase()
                : item.resource.toUpperCase()
            }}
          </template>
          <template v-slot:[`item.total`]="{ item }">
            <balance :value="-item.total" />
          </template>
          <template v-slot:[`item.exec`]="{ item }">
            {{ date(item.exec) }}
          </template>
        </nocloud-table>
      </td>
    </template>
  </nocloud-table>
</template>

<script>
import api from "@/api.js";
import nocloudTable from "@/components/table.vue";
import balance from "@/components/balance.vue";

export default {
  name: "transactions-table",
  components: { nocloudTable, balance },
  props: {
    selectTransaction: { type: Function, required: true },
    transactions: { type: Array, required: true },
  },
  data: () => ({
    headers: [
      { text: "Account ", value: "account" },
      { text: "Service ", value: "service" },
      { text: "Amount ", value: "total" },
      { text: "Date ", value: "proc" },
    ],
    recordHeaders: [
      { text: "Instance", value: "instance" },
      { text: "Product", value: "product" },
      { text: "Amount ", value: "total" },
      { text: "Date ", value: "exec" },
    ],
    isRecordsLoading: false,
    records: {},
    visibleItems: [],
    visibleRecords: [],
    selected: [],
    expanded: [],
    copyed: -1,
    fetchError: "",
  }),
  methods: {
    account(uuid) {
      return this.accounts.find((acc) => acc.uuid === uuid)?.title;
    },
    service(uuid) {
      const service = this.$store.getters["services/all"].find(
        (serv) => serv.uuid === uuid
      );

      return service?.title;
    },
    date(timestamp) {
      const date = new Date(timestamp * 1000);
      const time = date.toUTCString().split(" ")[4];

      const year = date.toUTCString().split(" ")[3];
      let month = date.getUTCMonth() + 1;
      let day = date.getUTCDate();

      if (`${month}`.length < 2) month = `0${month}`;
      if (`${day}`.length < 2) day = `0${day}`;

      return `${day}.${month}.${year} ${time}`;
    },
    hashTrim(hash) {
      if (hash) return ` ${hash.slice(0, 12)}... `;
      else return " XXXXXXXX... ";
    },
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((err) => {
            this.showSnackbarError({
              message: err,
            });
          });
      } else {
        this.showSnackbarError({
          message: "Clipboard is not supported!",
        });
      }
    },
    getRecords(uuid) {
      if (this.records[uuid]) return;

      this.isRecordsLoading = true;
      api
        .get(`/billing/transactions/${uuid}`)
        .then(({ pool }) => {
          this.records[uuid] = pool;
        })
        .catch((err) => {
          this.showSnackbarError({
            message: err,
          });
        })
        .finally(() => {
          this.isRecordsLoading = false;
        });
    },
  },
  computed: {
    accounts() {
      return this.$store.getters["accounts/all"];
    },
    isLoading() {
      return this.$store.getters["transactions/isLoading"];
    },
  },
  watch: {
    transactions() {
      this.fetchError = "";
    },
    expanded() {
      this.expanded.forEach(({ uuid }) => {
        this.getRecords(uuid);
      });
    },
  },
};
</script>
