<template>
  <v-container>
    <v-row>
      <v-col cols="3">
        <v-select v-model="selectedEventType" :items="eventsTypes" />
      </v-col>
      <v-col cols="3">
        <v-btn @click="deleteEvents" class="mt-3" color="primary">delete</v-btn>
        <v-dialog v-model="addEventDialog" width="50%">
          <template v-slot:activator="{ on, attrs }">
            <v-btn class="ml-7 mt-3" color="primary" v-on="on" v-bind="attrs"
              >add</v-btn
            >
          </template>
          <v-card class="pa-5">
            <v-card-title class="text-center">New event:</v-card-title>
            <v-autocomplete
              @input.native="userKey = $event.target.value"
              :items="keysWithNew"
              v-model="newEvent.key"
            />
            <v-autocomplete
              @input.native="userType = $event.target.value"
              :items="eventsTypesWithNew"
              v-model="newEvent.type"
            />
            <json-editor
              :json="newEvent.data"
              @changeValue="newEvent.data = $event"
            />
            <v-card-actions>
              <v-btn color="primary" @click="addNewEvent">Add</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>
    <nocloud-table
      table-name="events"
      item-key="id"
      v-model="selectedEvents"
      :headers="headers"
      :items="events"
    >
      <template v-slot:item.data="{ item }">
        <p>{{ JSON.stringify(item.data).slice(0, 40) + "..." }}</p>
      </template>
    </nocloud-table>
  </v-container>
</template>

<script>
import api from "@/api";
import nocloudTable from "@/components/table.vue";
import jsonEditor from "@/components/JsonEditor.vue";
import snackbar from "@/mixins/snackbar";
export default {
  name: "accounts-events",
  data: () => ({
    selectedEventType: "email",
    eventsTypes: ["email", "test"],
    keyItems: [
      "instance_suspended",
      "instance_unsuspended",
      "transaction_created",
      "account_created",
      "instance_created",
      "instance_deleted",
    ],
    eventsStorage: {
      email: { items: [], isLoaded: false },
      test: { items: [], isLoaded: false },
    },
    headers: [
      { text: "ID", value: "id" },
      { text: "Data", value: "data" },
      { text: "Key", value: "key" },
    ],
    newEvent: { type: "email", key: "", data: {} },
    userType: "",
    userKey: "",
    addEventDialog: false,
    selectedEvents: [],
  }),
  mixins: [snackbar],
  components: { nocloudTable, jsonEditor },
  props: ["account"],
  mounted() {
    this.fetchEvents();
  },
  methods: {
    async fetchEvents(check = true) {
      if (check && this.eventsStorage[this.selectedEventType].isLoaded) {
        return;
      }
      const res = await api.events.list(this.selectedEventType, this.uuid);
      this.eventsStorage[this.selectedEventType].items = res.events;
      this.eventsStorage[this.selectedEventType].isLoaded = true;
    },
    async addNewEvent() {
      try {
        await api.events.publish({ ...this.newEvent, uuid: this.uuid });

        if (!this.eventsStorage[this.newEvent.type]) {
          this.eventsStorage[this.newEvent.type] = {
            isLoaded: false,
            items: [],
          };
        } else {
          this.fetchEvents(false);
        }

        this.newEvent = { type: "email", key: "", data: {} };
        this.addEventDialog = false;
      } catch {
        this.showSnackbarError("Error during add event");
      }
    },
    async deleteEvents() {
      try {
        await Promise.all(
          this.selectedEvents.map((e) => {
            return api.events.cancel(e.type, e.id, e.uuid);
          })
        );

        this.eventsStorage[this.selectedEvents[0].type].items =
          this.eventsStorage[this.selectedEvents[0].type].items.filter((i) => {
            return this.selectedEvents.findIndex((el) => el.id === i.id) === -1;
          });

        this.selectedEvents = [];
      } catch {
        this.showSnackbarError("Error during delete events");
      }
    },
  },
  computed: {
    events() {
      return this.eventsStorage[this.selectedEventType].items;
    },
    eventsTypesWithNew() {
      if (!this.userType) {
        return this.eventsTypes;
      }

      return this.eventsTypes.concat([this.userType]);
    },
    keysWithNew() {
      if (!this.userKey) {
        return this.keyItems;
      }

      return this.keyItems.concat([this.userKey]);
    },
    uuid() {
      return this.account.uuid;
    },
  },
  watch: {
    selectedEventType() {
      this.selectedEvents = [];
      this.fetchEvents();
    },
  },
};
</script>
