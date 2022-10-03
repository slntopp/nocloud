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
        <p class="d-flex justify-space-between">
          <v-chip label> CPU: {{ cpu }} core(s)</v-chip>
          <v-chip label> RAM: {{ ram }} Mb</v-chip>
        </p>
        <p class="d-flex justify-space-between">
          <v-chip label>
            Size:
            {{ driveSize(drive_size) }} GiB
          </v-chip>
          <v-chip label>
            Type:
            {{ drive_type }}
          </v-chip>
        </p>
        <p class="d-flex  mb-0">
          <v-chip class="hash" label :color="copyed == index ? 'green' : ''">
            Hash:
            <v-btn icon @click="addToClipboard(hash, index)">
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
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((res) => {
            console.error(res);
          });
      } else {
        alert('Clipboard is not supported!');
      }
    },
  },
};
</script>

<style scoped>
.v-card__text .v-chip {
  width: 48%;
  display: flex;
  justify-content: center;
  align-items: center;
}
.v-card__text .hash {
  width: 100%;
}
</style>
