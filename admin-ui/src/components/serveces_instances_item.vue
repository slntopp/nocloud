<template>
  <v-col cols="12" lg="4">
    <v-card style="background: var(--v-background-light-base)">
      <div class="d-flex justify-space-between align-center pr-4">
        <v-card-title>
          {{ title }}
        </v-card-title>
        <v-chip midle :color="chipColor(state)">
          {{ state }}
        </v-chip>
      </div>
      <v-divider class="mx-4"></v-divider>
      <v-card-text>
        <p class="d-flex justify-space-between align-center">
          CPU:
          <v-chip class="" color="indigo" label>
            {{ cpu }}
          </v-chip>
        </p>
        <p class="d-flex justify-space-between align-center">
          Drive size:
          <v-chip class="" color="indigo" label>
            {{ driveSize(drive_size) }} GiB
          </v-chip>
        </p>
        <p class="d-flex justify-space-between align-center">
          Drive type:
          <v-chip class="" color="indigo" label>
            {{ drive_type }}
          </v-chip>
        </p>
        <p class="d-flex justify-space-between align-center">
          RAM:
          <v-chip class="" color="indigo" label>
            {{ ram }}
          </v-chip>
        </p>
        <p class="d-flex justify-space-between align-center">
          <span> Hash: </span>
          <v-chip class="" :color="copyed == index ? 'green' : ''" label>
            <v-btn icon @click="addToClipboardItem(hash, index)">
              <v-icon v-if="copyed == index"> mdi-check </v-icon>
              <v-icon v-else> mdi-content-copy </v-icon>
            </v-btn>
            {{ hashTrim(hash) }}
          </v-chip>
        </p>
      </v-card-text>
    </v-card>
  </v-col>
</template>

<script>
export default {
  name: "servecesInstancesItem",
  props: {
    title: {
      type: String,
      default: "",
    },
    state: {
      type: String,
      default: "",
    },
    cpu: {
      type: Number,
      default: 0,
    },
    drive_type: {
      type: String,
      default: "",
    },
    drive_size: {
      type: Number,
      default: 0,
    },
    ram: {
      type: Number,
      default: 0,
    },
    hash: {
      type: String,
      default: "",
    },
    index: {
      type: Number,
      default: 0,
    },
    chipColor: {
      type: Function,
      default: () => {},
    },
    hashTrim: {
      type: Function,
      default: () => {},
    },
  },
  data() {
    return {
      copyed: -1,
      color: "teal",
    };
  },
  methods: {
    driveSize(data) {
      return (data / 1024).toFixed(2);
    },
    addToClipboardItem(hash, index) {
      navigator.clipboard
        .writeText(hash)
        .then(() => {
          console.log(index);
          console.log(hash);
          console.log(this.copyed);
          this.copyed = index;
        })
        .catch((res) => {
          console.error(res);
        });
    },
  },
};
</script>

<style></style>
