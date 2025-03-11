<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <v-select
          label="Preferred disk type"
          :items="['HDD', 'SSD']"
          :value="meta.preferedDiskType"
          @change="meta.preferedDiskType = $event || undefined"
          clearable
        />
      </v-col>
      <v-col>
        <v-text-field
          label="Minimum HDD size"
          type="number"
          suffix="GB"
          v-model="meta.minDiskSize.HDD"
        />
      </v-col>
      <v-col>
        <v-text-field
          label="Maximum HDD size"
          type="number"
          suffix="GB"
          v-model="meta.maxDiskSize.HDD"
        />
      </v-col>
      <v-col>
        <v-text-field
          label="Minimum SSD size"
          type="number"
          suffix="GB"
          v-model="meta.minDiskSize.SSD"
        />
      </v-col>
      <v-col>
        <v-text-field
          label="Maximum SSD size"
          type="number"
          suffix="GB"
          v-model="meta.maxDiskSize.SSD"
        />
      </v-col>
    </v-row>
    <div class="os-tab__card mb-5" v-if="templateOs.length">
      <v-card
        outlined
        class="pt-4 pl-4 d-flex"
        style="gap: 10px"
        color="background"
        v-for="item of templateOs"
        :key="item.id"
      >
        <v-chip
          close
          :color="!hidedOs?.includes(item.id) ? 'info' : 'error'"
          :close-icon="
            hidedOs?.includes(item.id) ? 'mdi-close-circle' : 'mdi-plus-circle'
          "
          @click:close="changeOsState(item)"
        >
          <span>
            {{ item.name }}
          </span>
        </v-chip>
      </v-card>
    </div>
    <v-card
      v-else
      outlined
      class="pt-4 pl-4 d-flex mb-5"
      style="gap: 10px"
      color="background"
    >
      <v-card-title>No os in binded service provider</v-card-title>
    </v-card>

    <v-btn :loading="isSaveLoading" @click="save">Save</v-btn>
  </v-card>
</template>

<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar";

export default {
  name: "ione-plan-configuration",
  props: {
    template: {
      type: Object,
      required: true,
    },
  },
  mixins: [snackbar],
  data: () => ({
    meta: { minDiskSize: {}, maxDiskSize: {} },
    hidedOs: [],
    isSaveLoading: false,
  }),
  methods: {
    async save() {
      this.isSaveLoading = true;
      try {
        const data = {
          ...this.template,
          meta: {
            ...this.meta,
            hidedOs: this.hidedOs.filter((osId) => {
              const indexOfExistedOs = this.templateOs.findIndex(
                (os) => os.id === osId
              );
              return indexOfExistedOs !== -1;
            }),
          },
        };
        await api.plans.update(this.template.uuid, data);
        this.showSnackbarSuccess({
          message: "Configuration edited successfully",
        });
      } catch (err) {
        this.showSnackbarError({ message: err });
      } finally {
        this.isSaveLoading = false;
      }
    },
    changeOsState(item) {
      if (this.hidedOs.includes(item.id)) {
        const index = this.hidedOs.findIndex((i) => item.id === i);
        this.hidedOs.splice(index, 1);
      } else {
        this.hidedOs.push(item.id);
      }
    },
  },
  computed: {
    sp() {
      return this.$store.getters["servicesProviders/all"].find((sp) =>
        sp.meta.plans?.includes(this.template.uuid)
      );
    },
    templateOs() {
      return Object.keys(this.sp?.publicData?.templates || {}).map((key) => ({
        ...this.sp?.publicData?.templates[key],
        id: key,
      }));
    },
  },
  mounted() {
    this.meta = this.template.meta || {};
    if (!this.meta.minDiskSize) {
      this.meta.minDiskSize = {};
    }
    if (!this.meta.maxDiskSize) {
      this.meta.maxDiskSize = {};
    }
    this.hidedOs = this.template.meta?.hidedOs || [];
  },
  watch: {
    "template.meta"(newVal) {
      this.meta = newVal;
      if (!this.meta.minDiskSize) {
        this.meta.minDiskSize = {};
      }
      if (!this.meta.maxDiskSize) {
        this.meta.maxDiskSize = {};
      }
    },
  },
};
</script>

<style scoped lang="scss">
.os-tab__card {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  background: var(--v-background-base);
  padding-bottom: 16px;
}
</style>
