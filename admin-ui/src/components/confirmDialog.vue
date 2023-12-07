<template>
  <div class="empty">
    <div
      :class="{ empty: true, flex: !!flex }"
      @click.stop="() => !disabled && open()"
    >
      <slot></slot>
    </div>
    <v-dialog persistent v-model="dialog" :max-width="width">
      <v-card>
        <div class="confirm-card">
          <v-card-title class="text-h6">{{ title }}</v-card-title>
          <v-card-subtitle v-if="subtitle">{{ subtitle }}</v-card-subtitle>
          <v-card-text style="white-space: break-spaces" v-if="text">
            <span v-html="text"></span>
          </v-card-text>
          <slot name="actions" />
          <v-card-actions class="buttons">
            <v-btn color="red darken-1" @click="onCancel"> Cancel </v-btn>
            <v-btn
              color="primary darken-1"
              :disabled="successDisabled"
              @click="onConfirm"
            >
              Confirm
            </v-btn>
          </v-card-actions>
        </div>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
export default {
  props: {
    successDisabled: { type: Boolean, default: false },
    disabled: { type: Boolean, default: false },
    flex: { type: Boolean, default: false },
    text: { type: String, default: null },
    title: { type: String, default: "Are you sure you want to do this?" },
    subtitle: { type: String, default: null },
    width: { type: Number, default: 400 },
  },
  name: "confirm-dialog",
  data: () => ({ dialog: false }),
  methods: {
    open() {
      this.dialog = true;
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
  background-color: var(--v-background-base);
  padding: 30px;

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
