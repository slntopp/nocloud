<template>
  <v-card color="background-light" class="ma-5">
    <v-card-title>DNS's</v-card-title>
    <v-data-table
      class="mt-2 normal-color"
      :headers="headers"
      :items="dns"
      sort-by="calories"
    >
      <template v-slot:top>
        <v-toolbar color="background-light" flat class="d-flex justify-end">
          <v-dialog
            @click:outside="closeModal"
            v-model="dialog"
            max-width="500px"
          >
            <template v-slot:activator="{}">
              <v-btn class="mr-5" color="background-light" @click="addNew">
                Add
              </v-btn>
              <v-btn color="background-light" @click="saveDNS"> Save </v-btn>
            </template>
            <v-card color="background-light">
              <add-DNS
                @close="closeModal"
                @editDNS="editDNS"
                @addDNS="addDNS"
                :isOpen="dialog"
                :item="editedItem"
                :isEdit="isEdit"
              />
            </v-card>
          </v-dialog>
        </v-toolbar>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-icon class="mr-2" @click="editDnsItem(item)"> mdi-pencil </v-icon>
        <v-icon class="mr-6" @click="deleteDnsItem(item.id)">
          mdi-delete
        </v-icon>
      </template>
    </v-data-table>
  </v-card>
</template>

<script>
import api from "@/api";
import addDNS from "@/components/dns/addDNS.vue";
import snackbar from "@/mixins/snackbar.js";

export default {
  name: "instance-dns",
  components: { addDNS },
  mixins: [snackbar],
  data() {
    return {
      dialog: false,
      editedItem: {},
      isEdit: false,
      editedDNSId: null,
      dns: [],
      selected: [],
      headers: [
        {
          text: "Type",
          value: "type",
        },
        { text: "Subdomain", value: "subdomain" },
        {
          text: "Actions",
          value: "actions",
          align: "end",
          width: "200px",
        },
      ],
    };
  },
  created() {
    api.instances
      .action({ action: "get_dns", uuid: this.instanceId })
      .then((res) => {
        const keys = Object.keys(res.meta.records);

        keys.forEach((key) => {
          res.meta.records[key].forEach((dns) => {
            this.dns.push({
              type: key,
              ...dns,
              id: "id" + Math.random().toString(16).slice(2),
            });
          });
        });
      })
      .catch((e) => {
        console.log(e);
        this.showSnackbarError(e);
        this.$router.go(-1);
      });
  },
  computed: {
    instanceId() {
      return this.$route.params.instanceId;
    },
  },
  methods: {
    getDnsFields(field) {
      const notShowedKeys = ["id", "type"];

      return Object.keys(field).filter((key) => !notShowedKeys.includes(key));
    },
    editDnsItem(item) {
      this.dialog = true;
      this.isEdit = true;
      this.editedDNSId = item.id;
      this.editedItem = item;
    },
    deleteDnsItem(id) {
      this.dns = this.dns.filter((dns) => dns.id !== id);
    },
    addNew() {
      this.dialog = true;
    },
    closeModal() {
      this.isEdit = false;
      this.dialog = false;
      this.editedItem = {};
    },
    saveDNS() {
      const allTypes = new Set([...this.dns.map((d) => d.type)]);

      const dto = { dns: {} };

      allTypes.forEach((type) => {
        dto.dns[type] = this.dns
          .filter((dns) => dns.type === type)
          .map((dns) => {
            return dns;
          });
      });

      api.instances
        .action({
          action: "set_dns",
          uuid: this.instanceId,
          params: dto,
        })
        .catch((e) => {
          this.showSnackbarError({ message: e.message });
        });
    },
    editDNS(oldDNS) {
      this.dns = this.dns.map((dns) => {
        if (dns.id === oldDNS.id) {
          return oldDNS;
        }
        return dns;
      });
      this.closeModal();
    },
    addDNS(newDNS) {
      this.dns.push(newDNS);
      this.closeModal();
    },
  },
};
</script>

<style scoped>
.normal-color {
  background-color: rgba(12, 12, 60, 0.9);
}
</style>
