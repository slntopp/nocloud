<template>
  <div>
    <v-card-title class="px-0 mb-3"> Vlans:</v-card-title>
    <v-row v-if="template.secrets.vlans">
      <v-col>
        <v-text-field
          readonly
          :value="vlansKey"
          label="vlans key"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.vlans[vlansKey].start"
          label="start"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.vlans[vlansKey].size"
          label="size"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
    </v-row>
    <slot></slot>
    <!-- Hosts -->
    <v-card-title class="px-0 mb-3">Hosts:</v-card-title>
    <v-row class="mb-7">
      <v-col v-if="template.state.meta.hosts.error" cols="12" lg="6">
        <v-alert type="error"> {{ template.state.meta.hosts.error }}</v-alert>
      </v-col>
      <v-col
        v-for="(host, idx) in template.state.meta.hosts"
        :key="idx"
        cols="12"
        v-else
      >
        <v-row>
          <v-col v-if="host.error" cols="12" lg="6">
            <v-alert type="error"> {{ host.error }}</v-alert>
          </v-col>
          <v-col class="order-2 order-lg-1" cols="12" lg="6" v-else>
            <v-row>
              <v-col cols="12">
                <div class="title_progress">
                  <span>CPU</span>
                  <div>
                    <span>{{ host.total_cpu - host.free_cpu }}</span> /
                    <span>{{ host.total_cpu }}</span>
                  </div>
                </div>
                <v-progress-linear
                  :value="
                    Math.round(
                      ((host.total_cpu - host.free_cpu) / host.total_cpu) * 100
                    )
                  "
                  color="green"
                  height="20"
                >
                  <template v-slot:default="{ value }">
                    <strong>{{ value }}%</strong>
                  </template>
                </v-progress-linear>
              </v-col>
              <v-col cols="12">
                <div class="title_progress">
                  <span>Memory</span>
                  <div>
                    <span
                      >{{
                        ((host.total_ram - host.free_ram) / 1048576).toFixed(2)
                      }}
                      GiB</span
                    >
                    /
                    <span>{{ (host.total_ram / 1048576).toFixed(2) }} GiB</span>
                  </div>
                </div>
                <v-progress-linear
                  :value="
                    Math.round(
                      ((host.total_ram - host.free_ram) / host.total_ram) * 100
                    )
                  "
                  color="green"
                  height="20"
                >
                  <template v-slot:default="{ value }">
                    <strong>{{ value }}%</strong>
                  </template>
                </v-progress-linear>
              </v-col>
            </v-row>
          </v-col>
          <v-col cols="12" lg="6" class="order-1 order-lg-2">
            <p>name: {{ host.name }}</p>
            <p>state: {{ host.state }}</p>
            <p>vm_mad: {{ host.vm_mad }}</p>
            <p>im_mad: {{ host.im_mad }}</p>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
    <!-- Networking -->
    <v-card-title class="px-0 mb-3">Networking:</v-card-title>
    <v-row class="mb-7">
      <v-col>
        <v-card-subtitle class="px-0">Public</v-card-subtitle>
        <v-row>
          <v-col
            v-if="template.state.meta.networking.public_vnet.error"
            cols="12"
            lg="6"
          >
            <v-alert type="error">
              {{ template.state.meta.networking.public_vnet.error }}
            </v-alert>
          </v-col>

          <v-col v-else cols="12" lg="6">
            <div class="title_progress">
              <span>
                {{ template.state.meta.networking.public_vnet.name }}
              </span>
              <div>
                <span>{{
                  template.state.meta.networking.public_vnet.used
                }}</span>
                /
                <span>{{
                  template.state.meta.networking.public_vnet.total
                }}</span>
              </div>
            </div>
            <v-progress-linear
              :value="percentUsePublic"
              :color="colorUsePublic"
              height="20"
              class="mb-10"
            >
              <template v-slot:default="{ value }">
                <strong>{{ value }}%</strong>
              </template>
            </v-progress-linear>
          </v-col>
        </v-row>
        <v-card-subtitle class="px-0">Private</v-card-subtitle>
        <v-row>
          <v-col
            v-if="template.state.meta.networking.private_vnet.error"
            cols="12"
            lg="6"
          >
            <v-alert type="error">
              {{ template.state.meta.networking.private_vnet.error }}
            </v-alert>
          </v-col>
          <v-col v-else cols="12" lg="6">
            <div class="title_progress">
              <span>
                {{ template.state.meta.networking.private_vnet.name }}
              </span>
              <div>
                <span>{{
                  template.state.meta.networking.private_vnet.used
                }}</span>
                /
                <span>{{
                  template.state.meta.networking.private_vnet.total
                }}</span>
              </div>
            </div>
            <v-progress-linear
              :value="percentUsePrivate"
              :color="colorUsePrivate"
              height="20"
              class="mb-10"
            >
              <template v-slot:default="{ value }">
                <strong>{{ value }}%</strong>
              </template>
            </v-progress-linear>
          </v-col>
        </v-row>
      </v-col>
    </v-row>

    <!-- Datastores -->
    <v-card-title class="px-0 mb-3">Datastores:</v-card-title>
    <v-row class="mb-7">
      <v-col v-if="template.state.meta.datastores.error" cols="12" lg="6">
        <v-alert type="error">
          {{ template.state.meta.datastores.error }}
        </v-alert>
      </v-col>
      <v-col
        v-for="(datastor, idx) in template.state.meta.datastores"
        :key="idx"
        v-else
        cols="12"
      >
        <v-row>
          <v-col class="order-2 order-lg-1" cols="12" lg="6" order-2>
            <div class="title_progress">
              <span>{{ datastor.drive_type }}</span>
              <div>
                <span>{{ (datastor.used / 1024).toFixed(2) }}</span> /
                <span>{{ (datastor.total / 1024).toFixed(2) }} GiB</span>
              </div>
            </div>
            <v-progress-linear
              :value="Math.round((datastor.used / datastor.total) * 100)"
              color="green"
              height="20"
            >
              <template v-slot:default="{ value }">
                <strong>{{ value }}%</strong>
              </template>
            </v-progress-linear>
          </v-col>
          <v-col class="order-1 order-lg-2" cols="12" lg="6" order-1>
            <p>name: {{ datastor.name }}</p>
            <p>tm_mad: {{ datastor.tm_mad }}</p>
            <p>ds_mad: {{ datastor.tm_mad }}</p>
          </v-col>
        </v-row>
      </v-col>
      <v-col cols="12">
        <p>OS's:</p>
        <div class="newCloud__template">
          <div
            v-for="OS in getPublicTemplates"
            class="newCloud__template-item"
            :key="OS.name"
          >
            <div class="newCloud__template-image">
              <img :src="getOSImgLink(OS.name)" :alt="OS.name" />
            </div>
            <div class="newCloud__template-name">
              {{ OS.name }}
            </div>
          </div>
        </div>
      </v-col>
      <v-col ref="private">
        <p>Vlans:</p>
        <v-tooltip bottom color="info" v-for="(vlan, i) of vlans" :key="i">
          <template v-slot:activator="{ on, attrs }">
            <span
              class="ceil"
              v-bind="attrs"
              v-on="on"
              :class="vlan === 0 ? 'occupied' : 'free'"
            />
          </template>
          <span>{{ getVlanIndex(i) }}</span>
        </v-tooltip>
        <div class="mt-2">
          <v-btn class="mr-2" v-if="counter > 1" @click="counter--">
            less
          </v-btn>
          <v-btn v-if="counter < vlansRowCount" @click="counter++">
            more
          </v-btn>
        </div>
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "service-provider-ione",
  props: { template: { type: Object, required: true } },
  data: () => ({ counter: 1 }),
  methods: {
    changeWidth() {
      const { clientWidth } = this.$refs?.private;
      let cols = 64;

      for (let i = 4; i > 0; i--) {
        if (clientWidth / 20 >= cols) {
          this.$refs.private.style.maxWidth = `${cols * 20 + 24}px`;
          break;
        } else {
          cols /= 2;
        }
      }
    },
    getVlanIndex(index) {
      return this.vlansStart ? this.vlansStart + index : index;
    },
    getOSImgLink(name) {
      const os = name.replace(/[^a-zA-Z]+/g, "").toLowerCase();

      return "/admin/img/" + os + ".png";
    },
  },
  mounted() {
    this.changeWidth();
  },
  computed: {
    vlansKey() {
      return Object.keys(this.template.secrets.vlans ?? {})[0];
    },
    vlans() {
      const { free_vlans } =
        this.template?.state?.meta?.networking?.private_vnet;
      let vlans = 0;

      Object.values(free_vlans || {}).forEach((value) => {
        vlans += +value;
      });

      const res = Array.from({
        length: (this.vlansCount / this.vlansRowCount) * this.counter + 1,
      })
        .fill(1, 0, vlans)
        .fill(0, vlans);
      return res;
    },
    vlansCount() {
      if (!this.template.secrets.vlans) return 0;
      return this.template.secrets.vlans[this.vlansKey]?.size;
    },
    vlansStart() {
      if (!this.template.secrets.vlans) return 0;
      return this.template.secrets.vlans[this.vlansKey]?.start;
    },
    vlansRowCount() {
      if (this.vlansCount < 400) {
        return 1;
      }
      return this.vlansCount < 2000 ? 4 : 8;
    },
    getPublicTemplates() {
      if (!this.template.publicData.templates) {
        return [];
      }

      return Object.values(this.template.publicData.templates).filter(
        (t) => t.is_public
      );
    },
    percentUsePublic() {
      const { public_vnet } = this.template.state.meta.networking;

      return Math.round((public_vnet.used / public_vnet.total) * 100);
    },
    percentUsePrivate() {
      const { private_vnet } = this.template.state.meta.networking;

      return Math.round((private_vnet.used / private_vnet.total) * 100);
    },
    colorUsePublic() {
      if (this.percentUsePublic >= 95) return "red";
      if (this.percentUsePublic > 80) return "orange";
      return "green";
    },
    colorUsePrivate() {
      if (this.percentUsePrivate >= 95) return "red";
      if (this.percentUsePrivate > 80) return "orange";
      return "green";
    },
  },
  watch: {
    counter() {
      this.changeWidth();
    },
  },
};
</script>
<style lang="scss" scoped>
.newCloud__template {
  display: flex;
  flex-wrap: wrap;

  &.one-line {
    flex-wrap: nowrap;
    justify-content: space-between;
  }
}

.newCloud__template-item {
  width: 150px;
  margin-left: 20px;
  margin-bottom: 10px;
  box-shadow: 3px 2px 6px rgba(189, 188, 188, 0.08),
    0px 0px 8px rgba(65, 64, 64, 0.05);
  border-radius: 15px;
  border: solid 1px white;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
  cursor: pointer;
  text-align: center;
  overflow: hidden;
  display: grid;
  grid-template-columns: 1fr;
  grid-template-rows: max-content auto;

  &:not(:last-child) {
    margin-right: 10px;
  }

  &:first-child {
    margin-left: 0px;
  }

  &:hover {
    box-shadow: 5px 8px 10px rgba(0, 0, 0, 0.08),
      0px 0px 12px rgba(0, 0, 0, 0.05);
  }
}

.newCloud__template-image {
  padding: 10px;
}

.newCloud__template-image img {
  margin: auto;
  width: 90%;
  height: 100%;
}

.newCloud__template-name {
  padding: 10px;
}
</style>
