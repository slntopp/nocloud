<template>
  <v-form class="pa-10" ref="dnsForm" v-model="isValid">
    <v-select :items="dnsTypes" v-model="newDNS.type" v-if="!isEdit" />
    <v-text-field
      v-for="(value, key) in dnsKeys"
      :key="key"
      :label="key"
      v-model="newDNS[key]"
      :rules="mainRule"
    />
    <div class="d-flex justify-end">
      <v-btn class="mr-2" @click="close">Close</v-btn>
      <v-btn class="mr-4" @click="add">{{ isEdit ? "Edit" : "Add" }}</v-btn>
    </div>
  </v-form>
</template>

<script>
export default {
  name: "add-dns",
  props: {
    item: { type: Object },
    isEdit: { type: Boolean, default: false },
    isOpen: { type: Boolean, default: false },
  },
  data() {
    return {
      dnsTypes: ["AAAA", "A", "CNAME", "MX", "SRV", "TXT"],
      newDNS: {},
      isValid: false,
      mainRule: [(d) => !!d || "Field is required"],
    };
  },
  created() {
    if (this.isEdit) {
      this.newDNS = this.item;
    } else {
      this.setDefaultDNS();
    }
  },
  methods: {
    setDefaultDNS() {
      this.newDNS = {
        type: "A",
        ip_adress: "",
        subdomain: "",
      };
    },
    close() {
      this.$emit("close");
    },
    add() {
      if (!this.isValid) {
        this.$refs.dnsForm.validate();
        return;
      }

      if (this.isEdit) {
        this.$emit("editDNS", this.newDNS);
      } else {
        this.$emit("addDNS", this.newDNS);
      }

      this.setDefaultDNS();
    },
  },
  computed: {
    dnsKeys() {
      const withoutType = { ...this.newDNS };
      delete withoutType.type;
      delete withoutType.id;
      return withoutType;
    },
  },
  watch: {
    isOpen() {
      this.setDefaultDNS();
    },
    item() {
      if (this.isEdit) {
        this.newDNS = this.item;
      }
    },
    "newDNS.type"(newValue) {
      if (this.isEdit) {
        return;
      }

      switch (newValue) {
        case "A": {
          this.setDefaultDNS();
          break;
        }
        case "AAAA": {
          this.newDNS = {
            type: "AAAA",
            ipv6_adress: "",
            subdomain: "",
          };
          break;
        }
        case "CNAME": {
          this.newDNS = {
            type: "CNAME",
            hostname: "",
            subdomain: "",
          };
          break;
        }
        case "MX": {
          this.newDNS = {
            type: "MX",
            hostname: "",
            subdomain: "",
            priority: "",
          };
          break;
        }
        case "SRV": {
          this.newDNS = {
            type: "SRV",
            hostname: "",
            subdomain: "",
            priority: "",
            weight: "",
            port: "",
          };
          break;
        }
        case "TXT": {
          this.newDNS = {
            type: "TXT",
            text: "",
            subdomain: "",
          };
          break;
        }
      }
    },
  },
};
</script>
