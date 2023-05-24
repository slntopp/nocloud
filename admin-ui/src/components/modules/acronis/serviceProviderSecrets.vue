<template>
  <v-container>
    <v-row>
      <v-col>
        <v-card-title>Client id</v-card-title>
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.secrets.clientId" />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-card-title>Client secret</v-card-title>
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.secrets.clientSecret" />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-card-title>Host</v-card-title>
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.secrets.datacenterUrl" />
      </v-col>
    </v-row>
    <v-card-title>Offering items</v-card-title>
    <v-row>
      <nocloud-table
        style="width: 100%"
        table-name="acronisPrices"
        :no-hide-uuid="false"
        :show-select="false"
        :loading="isLoading"
        :headers="headers"
        :items="offeringItems"
      >
        <template v-slot:[`item.price`]="{ item }">
          <v-text-field
            type="number"
            v-model.number="item.price"
          ></v-text-field>
        </template>
      </nocloud-table>
    </v-row>
    <v-row justify="end">
      <v-btn :loading="isSaveLoading" @click="saveOffering">Save</v-btn>
    </v-row>
  </v-container>
</template>

<script setup>
import { onMounted, ref, defineProps } from "vue";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";

const props = defineProps(["template"]);

const store = useStore();

const offeringItems = ref([]);
const isLoading = ref(false);
const isSaveLoading = ref(false);

const acronisResourcesMap = {
  servers: "Physical Servers",
  vms: "Virtual Machines",
  storage: "Storage",
  seats: "Seats",
  cloud_servers: "Cloud Servers",
  base: "Base",
  workstations: "Workstations",
  web_hosting_servers: "Hosting Servers",
  websites: "Websites",
  mobiles: "Mobile Devices",
  gsuite_seats: "Google Workspace Seats",
  o365_teams: "Microsoft 365 Seats (for SP/Local/Azure/Google Hosted Storage)",
  m366: "Microsoft 365 Seats (unlimited Acronis Hosted Cloud Storage included)",
  hsot: "Hosting Servers",
  notarizations: "Notarizations (notarized files)",
  esignatures: "eSignature Templates",
  local_storage: "Customer Local Storage",
  mailboxes: "Advanced Email Security Mailboxes",
  backup_workstations: "Backup Workstations",
  backup_servers: "Backup Physical Servers",
  backup_web_hosting_servers: "Backup Hosting Servers",
  backup_vms: "Backup Virtual Machines",
  management:
    "Management Workloads (Servers, VMs, Workstations, Hosting Servers)",
  security: "Security Workloads (Servers, VMs, Workstations, Hosting Servers)",
  email_security_mailboxes: "Email Security Mailboxes",
  security_edr_d180: "Security + EDR",
  dlp: "Data Loss Prevention Workloads (Servers, VMs, Workstations, Hosting Servers)",
  backup_m365_seats:
    "Backup Microsoft 365 Seats (unlimited Acronis Hosted Cloud Storage included)",
  backup_gworkspace_seats: "Backup Google Workspace Seats",
  nas: "Notary Cloud Storage - Acronis Hosted ",
  compute_points: "Monthly Compute Point Consumption ",
  public_ips: "Advanced Disaster Recovery Public IPs",
  notary_storage: "Notary",
  o365_mailboxes: "Office 365 Mailboxes",
  o365_onedrive: "Office 365 Onedrive",
  o365_sharepoint_sites: "Office 365 Sharepoint Sites",
  google_mail: "Google Mail",
  google_drive: "Google Drive",
  google_team_drive: "Google Team Drive",
  drives_shipped_to_cloud: "Drives Shipped To Cloud",
  hosted_exchange: "Hosted Exchange",
};
const acronisEditionMap = {
  per_workload: "Per Workload",
  per_gigabyte: "Per GB",
  per_user: "Per User",
  without_edition: "Without",
  pck: "Cyber protection ",
  fss: "File sync and share ",
};

const headers = ref([
  { text: "Usage name", value: "usage_name" },
  { text: "Type", value: "type" },
  { text: "Application id", value: "application_id" },
  { text: "Edition", value: "edition" },
  { text: "Required", value: "mandatory" },
  { text: "Price", value: "price" },
]);

onMounted(async () => {
  isLoading.value = true;
  try {
    const data = (
      await api.servicesProviders.action({
        action: "get_offering_items",
        uuid: props.template.uuid,
      })
    ).meta.offeringItems;

    for (const mandatoryKey of Object.keys(data)) {
      for (const key of Object.keys(data[mandatoryKey])) {
        offeringItems.value.push(
          ...data[mandatoryKey][key].map((of) => ({
            ...of,
            edition: getAcronisEdition(of),
            usage_name: getAcronisName(of),
            mandatory: mandatoryKey === "mandatory",
            price: props.template.secrets.offeringItems[of.name]?.price,
          }))
        );
      }
    }
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch offering items",
    });
  } finally {
    isLoading.value = false;
  }
});

const saveOffering = async () => {
  const offerings = offeringItems.value.filter((of) => !!of.price);
  if (offerings.length === 0) {
    return;
  }
  const sp = JSON.parse(JSON.stringify(props.template));
  sp.secrets.offeringItems = {};
  offerings.forEach((of) => {
    sp.secrets.offeringItems[of.name] = of;
  });
  isSaveLoading.value = true;
  try {
    await api.servicesProviders.update(props.template.uuid, sp);
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save offering items",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const getAcronisName = (item) => {
  let name = item.usage_name;
  const deletedKeys = ["dr", "fc"];
  if (deletedKeys.includes(name.split("_")[0])) {
    name = name.replace(name.split("_")[0] + "_", "");
  }
  const subKey = name.includes("pack_adv_") ? "Advanced " : undefined;
  name = name.replace("pack_adv_", "");
  return (
    (subKey ? subKey + acronisResourcesMap[name] : acronisResourcesMap[name]) ||
    item.usage_name
  );
};

const getAcronisEdition = (item) => {
  let edition = "";
  const keys = item.edition.split("_");
  if (acronisEditionMap[keys[0]]) {
    edition += acronisEditionMap[keys[0]];
    keys.shift();
  }
  edition += acronisEditionMap[keys.join("_")];
  return edition || item.edition;
};
</script>

<style scoped></style>
