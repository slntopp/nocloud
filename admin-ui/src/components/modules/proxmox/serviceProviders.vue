<template>
  <div class="sp-proxmox">
    <!-- Connection -->
    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.host">Host</subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-text-field
          label="https://pve.example.com:8006"
          :value="secrets.host"
          :rules="[required, isHttpsURL]"
          :error-messages="errors.host"
          @change="(v) => changeSecret('host', v)"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.auth_mode">
          Authorization
        </subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-radio-group v-model="authMode" row @change="changeAuthMode">
          <v-radio label="User + password" value="password" />
          <v-radio label="API token" value="token" />
        </v-radio-group>
        <v-alert v-if="authMode === 'token'" dense text type="info">
          PVE vncwebsocket does not accept API tokens: the VNC console requires
          user + password. Everything else works with a token.
        </v-alert>
      </v-col>
    </v-row>

    <template v-if="authMode === 'token'">
      <v-row align="center">
        <v-col cols="4">
          <subheader-with-info infoText="SPInfo.proxmox.token_id">Token ID</subheader-with-info>
        </v-col>
        <v-col cols="8">
          <v-text-field
            label="nocloud@pve!driver"
            :value="secrets.token_id"
            :rules="[required, isTokenID]"
            @change="(v) => changeSecret('token_id', v)"
          />
        </v-col>
      </v-row>
      <v-row align="center">
        <v-col cols="4">
          <subheader-with-info infoText="SPInfo.proxmox.token_secret">Token secret</subheader-with-info>
        </v-col>
        <v-col cols="8">
          <v-text-field
            label="uuid"
            type="password"
            :value="secrets.token_secret"
            :rules="[required]"
            @change="(v) => changeSecret('token_secret', v)"
          />
        </v-col>
      </v-row>
    </template>
    <template v-else>
      <v-row align="center">
        <v-col cols="4">
          <subheader-with-info infoText="SPInfo.proxmox.user">User</subheader-with-info>
        </v-col>
        <v-col cols="8">
          <v-text-field
            label="nocloud@pve"
            :value="secrets.user"
            :rules="[required, isRealmUser]"
            @change="(v) => changeSecret('user', v)"
          />
        </v-col>
      </v-row>
      <v-row align="center">
        <v-col cols="4">
          <subheader-with-info infoText="SPInfo.proxmox.pass">Password</subheader-with-info>
        </v-col>
        <v-col cols="8">
          <v-text-field
            label="password"
            type="password"
            :value="secrets.pass"
            :rules="[required]"
            @change="(v) => changeSecret('pass', v)"
          />
        </v-col>
      </v-row>
    </template>

    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.insecure">Skip TLS verify</subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-switch
          label="insecure (self-signed certificate)"
          :input-value="!!secrets.insecure"
          @change="(v) => changeSecret('insecure', !!v)"
        />
      </v-col>
    </v-row>
    <v-row v-if="!secrets.insecure">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.ca_cert">CA certificate (PEM)</subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-textarea
          rows="3"
          label="-----BEGIN CERTIFICATE-----"
          :value="secrets.ca_cert"
          @change="(v) => changeSecret('ca_cert', v)"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.pool">Pool</subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-text-field
          label="nocloud (optional, VMs are added to this PVE pool)"
          :value="getDefault('pool')"
          @change="(v) => changeDefault('pool', v)"
        />
      </v-col>
    </v-row>

    <v-divider class="my-4" />

    <!-- Placement -->
    <v-row>
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.sched">
          Scheduler rules (sched)
        </subheader-with-info>
        <div class="text-caption">
          default / HCPU: {nodes, tags, strategy: least_used|round_robin, reserve_ram_mb}
        </div>
      </v-col>
      <v-col cols="8">
        <json-editor
          :json="getJSON('sched')"
          @changeValue="(v) => changeJSON('sched', v)"
        />
        <div v-if="errors.sched" class="error--text">{{ errors.sched }}</div>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.sched_ds">
          Storage by drive type (sched_ds)
        </subheader-with-info>
        <div class="text-caption">default: {"SSD": "local-lvm", "HDD": "ceph-hdd"}</div>
      </v-col>
      <v-col cols="8">
        <json-editor
          :json="getJSON('sched_ds')"
          @changeValue="(v) => changeJSON('sched_ds', v)"
        />
        <div v-if="errors.sched_ds" class="error--text">{{ errors.sched_ds }}</div>
      </v-col>
    </v-row>

    <!-- Networking -->
    <v-row>
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.public_ip_pool">
          Public IP pool
        </subheader-with-info>
        <div class="text-caption">
          default: {ranges: ["a.b.c.10-a.b.c.50", "x.y.z.0/24"], gateway, prefix|netmask, dns, bridge, vlan_tag}
        </div>
      </v-col>
      <v-col cols="8">
        <json-editor
          :json="getJSON('public_ip_pool')"
          @changeValue="(v) => changeJSON('public_ip_pool', v)"
        />
        <div v-if="errors.public_ip_pool" class="error--text">
          {{ errors.public_ip_pool }}
        </div>
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.private_vnet_ban">
          Private network functions
        </subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-switch
          label="disable private networks (private_vnet_ban)"
          :input-value="!!getDefault('private_vnet_ban')"
          @change="(v) => changeDefault('private_vnet_ban', !!v)"
        />
      </v-col>
    </v-row>
    <v-row v-if="!getDefault('private_vnet_ban')">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.private_vnet_tmpl">
          Private IP pool
        </subheader-with-info>
        <div class="text-caption">same format as the public pool, usually another bridge / vlan_tag</div>
      </v-col>
      <v-col cols="8">
        <json-editor
          :json="getJSON('private_vnet_tmpl')"
          @changeValue="(v) => changeJSON('private_vnet_tmpl', v)"
        />
      </v-col>
    </v-row>

    <v-divider class="my-4" />

    <!-- VM options -->
    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.full_clone">Full clone</subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-switch
          label="full clone (off = linked clone, template must stay on the same storage)"
          :input-value="getDefault('full_clone') !== false"
          @change="(v) => changeDefault('full_clone', !!v)"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.vmid_range">VMID range</subheader-with-info>
      </v-col>
      <v-col cols="4">
        <v-text-field
          type="number"
          label="min (>= 100)"
          :value="vmidRange[0]"
          :error-messages="errors.vmid_range"
          @change="(v) => changeVMIDRange(0, v)"
        />
      </v-col>
      <v-col cols="4">
        <v-text-field
          type="number"
          label="max"
          :value="vmidRange[1]"
          @change="(v) => changeVMIDRange(1, v)"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.suspend_mode">Suspend mode</subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-select
          :items="suspendModes"
          :value="getDefault('suspend_mode') || 'suspend_disk'"
          @change="(v) => changeDefault('suspend_mode', v)"
        />
      </v-col>
    </v-row>

    <v-row align="center">
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.qemu_agent_required">
          QEMU guest agent
        </subheader-with-info>
      </v-col>
      <v-col cols="8">
        <v-switch
          label="required (templates without agent are hidden, extra IPs and stats need it)"
          :input-value="!!getDefault('qemu_agent_required')"
          @change="(v) => changeDefault('qemu_agent_required', !!v)"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="4">
        <subheader-with-info infoText="SPInfo.proxmox.min_drive_size">
          Drive size limits, MB
        </subheader-with-info>
        <div class="text-caption">min_drive_size / max_drive_size by drive type</div>
      </v-col>
      <v-col cols="4">
        <json-editor
          :json="getJSON('min_drive_size')"
          @changeValue="(v) => changeJSON('min_drive_size', v)"
        />
      </v-col>
      <v-col cols="4">
        <json-editor
          :json="getJSON('max_drive_size')"
          @changeValue="(v) => changeJSON('max_drive_size', v)"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import JsonEditor from "@/components/JsonEditor.vue";
import subheaderWithInfo from "@/components/ui/subheaderWithInfo.vue";

const DEFAULT_VARS = {
  sched: { default: { strategy: "least_used", reserve_ram_mb: 2048 } },
  sched_ds: { default: { SSD: "local-lvm" } },
  public_ip_pool: {
    default: { ranges: [], gateway: "", prefix: 24, dns: ["1.1.1.1"], bridge: "vmbr0" },
  },
  private_vnet_tmpl: { default: { ranges: [], prefix: 24, bridge: "vmbr1" } },
  min_drive_size: { default: {} },
  max_drive_size: { default: {} },
};

export default {
  name: "servicesProviders-create-proxmox",
  components: { JsonEditor, subheaderWithInfo },
  props: {
    secrets: { type: Object, default: () => ({}) },
    vars: { type: Object, default: () => ({}) },
    passed: { type: Boolean, default: false },
  },
  data: () => ({
    authMode: "password",
    errors: {},
    suspendModes: [
      { text: "suspend to disk (hibernate)", value: "suspend_disk" },
      { text: "stop (power off, RAM freed)", value: "stop" },
      { text: "suspend in RAM (pause)", value: "suspend_ram" },
    ],
  }),
  computed: {
    vmidRange() {
      const r = this.getDefault("vmid_range");
      return Array.isArray(r) && r.length === 2 ? r : ["", ""];
    },
  },
  methods: {
    required(v) {
      return !!v || "Field is required";
    },
    isHttpsURL(v) {
      try {
        const u = new URL(v);
        return u.protocol === "https:" || "Must be https://";
      } catch {
        return "Is not a valid URL";
      }
    },
    isTokenID(v) {
      return /^[^@!\s]+@[^@!\s]+![^@!\s]+$/.test(v || "") || "Format: user@realm!tokenname";
    },
    isRealmUser(v) {
      return /^[^@\s]+@[^@\s]+$/.test(v || "") || "Format: user@realm (nocloud@pve)";
    },

    getDefault(key) {
      return this.vars?.[key]?.value?.default;
    },
    getJSON(key) {
      const v = this.vars?.[key]?.value;
      if (v && typeof v === "object") return v;
      return DEFAULT_VARS[key] || {};
    },

    changeSecret(key, value) {
      const secrets = { ...this.secrets, [key]: value };
      if (value === "" || value === null || value === undefined) delete secrets[key];
      this.$emit("change:secrets", secrets);
      this.validate(secrets, this.vars);
    },
    changeAuthMode(mode) {
      const secrets = { ...this.secrets };
      if (mode === "token") {
        delete secrets.user;
        delete secrets.pass;
      } else {
        delete secrets.token_id;
        delete secrets.token_secret;
      }
      this.$emit("change:secrets", secrets);
      this.validate(secrets, this.vars);
    },
    changeDefault(key, value) {
      const vars = { ...this.vars, [key]: { value: { default: value } } };
      this.$emit("change:vars", vars);
      this.validate(this.secrets, vars);
    },
    changeJSON(key, value) {
      const vars = { ...this.vars, [key]: { value } };
      this.$emit("change:vars", vars);
      this.validate(this.secrets, vars);
    },
    changeVMIDRange(idx, value) {
      const r = [...this.vmidRange];
      r[idx] = value === "" ? "" : +value;
      if (r[0] === "" && r[1] === "") {
        const vars = { ...this.vars };
        delete vars.vmid_range;
        this.$emit("change:vars", vars);
        this.validate(this.secrets, vars);
        return;
      }
      this.changeDefault("vmid_range", r);
    },

    validate(secrets, vars) {
      const errors = {};
      if (!secrets.host || this.isHttpsURL(secrets.host) !== true) {
        errors.host = "https host is required";
      }
      const hasToken = !!secrets.token_id && !!secrets.token_secret;
      const hasPass = !!secrets.user && !!secrets.pass;
      if (!hasToken && !hasPass) {
        errors.auth = "token_id+token_secret or user+pass";
      }
      const ds = vars.sched_ds?.value?.default;
      if (!ds || typeof ds !== "object" || Object.keys(ds).length === 0) {
        errors.sched_ds = "sched_ds.default must map drive types to storages";
      }
      const pool = vars.public_ip_pool?.value?.default;
      if (!pool || !Array.isArray(pool.ranges) || pool.ranges.length === 0) {
        errors.public_ip_pool = "public_ip_pool.default.ranges is required";
      } else if (!pool.gateway) {
        errors.public_ip_pool = "public_ip_pool.default.gateway is required";
      }
      const sched = vars.sched?.value?.default;
      if (sched && typeof sched !== "object") {
        errors.sched = "sched.default must be an object";
      }
      const r = vars.vmid_range?.value?.default;
      if (Array.isArray(r) && (r.length !== 2 || +r[0] < 100 || +r[1] < +r[0])) {
        errors.vmid_range = "expected [min >= 100, max >= min]";
      }
      this.errors = errors;
      this.$emit("passed", Object.keys(errors).length === 0);
    },
  },
  mounted() {
    if (this.secrets.token_id) this.authMode = "token";

    const vars = { ...this.vars };
    let changed = false;
    for (const key of ["sched", "sched_ds", "public_ip_pool"]) {
      if (!vars[key]?.value) {
        vars[key] = { value: DEFAULT_VARS[key] };
        changed = true;
      }
    }
    if (!vars.console?.value) {
      vars.console = { value: { default: "vnc" } };
      changed = true;
    }
    if (changed) this.$emit("change:vars", vars);
    this.validate(this.secrets, vars);
  },
};
</script>

<style scoped>
.text-caption {
  opacity: 0.7;
  word-break: break-word;
}
</style>
