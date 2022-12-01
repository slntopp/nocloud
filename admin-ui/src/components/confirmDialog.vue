<template>
  <div class="empty">
    <div class="empty" @click.stop="open"><slot></slot></div>
    <v-dialog v-model="dialog" max-width="400">
      <v-card>
        <div class="confirm-card">
          <v-card-title class="text-h6">
            Are you sure you want to do this?
          </v-card-title>
          <v-card-text v-if="text">{{ text }}</v-card-text>
          <v-card-actions class="buttons">
            <v-btn color="red darken-1" @click="onCancel"> Cancel </v-btn>
            <v-btn color="primary darken-1" @click="onConfirm"> Confirm </v-btn>
          </v-card-actions>
        </div>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
export default {
  props: {
    disabled: { type: Boolean, default: false },
    text: { type: String, default: null }
  },
  name: "confirm-dialog",
  data: () => ({ dialog: false }),
  methods: {
    open() {
      if (!this.disabled) {
        this.dialog = true;
      }
    },
    onConfirm() {
      this.$emit("confirm");
      this.dialog = false;
    },
    onCancel() {
      this.$emit("cancel");
      this.dialog = false;
    },
  },
};
</script>

<style lang="scss">
.empty {
  display: inline;
}
.confirm-card {
  background-color: rgba(12, 12, 60, 0.9);
  padding: 30px;
  div,
  p {
    background-color: rgba(12, 12, 60, 0.9);
  }
  .buttons {
    margin-top: 10px;
    display: flex;
    justify-content: center;
    button {
      margin: 0 20px;
    }
  }
}
</style>
