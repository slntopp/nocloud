<template>
  <nc-map
    not-scale
    class="map"
    id="mapMain"
    ref="map"
    v-model="selected"
    :markers="markers"
    :mapClick="mapClickHandler"
    :markerScaleDivider="0.8"
    :highlightHoveredCountry="true"
  >
    <template #actions>
      <!-- byn  .ant-btn-primary -->
      <div
        v-if="selected || multiSelect"
        style="position: absolute; right: 25px; bottom: 13px"
      >
        <v-btn
          class="ant-btn-primary"
          style="margin-right: 5px; background-color: #4caf50"
          @click="saveCountry"
        >
          Save
        </v-btn>
        <v-btn style="background-color: #272727" @click="cancelSelectedCountry">
          Cancel
        </v-btn>
      </div>
      <!-- end byn -->
    </template>

    <template v-slot:popup="{ marker }">
      <!-- text -->
      <foreignObject
        x="65"
        y="55"
        width="40"
        height="40"
        v-if="marker.title"
        :transform-origin="`${120 / 2} 80`"
        :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(
          marker.x + 14 - 120 / 2,
          1
        )} ${marker.y - 45})`"
        @mouseenter="(e) => mouseEnterHandler(marker.id, e)"
        @mouseleave="(e) => mouseLeaveHandler(marker.id, e)"
      >
        <div class="map__popup-content">
          <slot name="popup" :marker="marker">
            <div class="map__popup-content--default">
              <!-- {{ marker.title }} -->
              <v-dialog :ref="`edit-dialog.${marker.id}`" width="800">
                <template v-slot:activator="{ on, attrs }">
                  <v-icon v-on="on" v-bind="attrs" color="secondary">
                    mdi-cog
                  </v-icon>
                </template>
                <v-card class="pa-4" color="background-light">
                  <v-icon
                    @click="formatText('b', marker.id)"
                    @mousedown.prevent
                  >
                    mdi-format-bold
                  </v-icon>

                  <v-icon
                    @click="formatText('i', marker.id)"
                    @mousedown.prevent
                  >
                    mdi-format-italic
                  </v-icon>

                  <v-icon
                    @click="formatText('u', marker.id)"
                    @mousedown.prevent
                  >
                    mdi-format-underline
                  </v-icon>

                  <v-icon
                    @click="formatText('s', marker.id)"
                    @mousedown.prevent
                  >
                    mdi-format-strikethrough
                  </v-icon>

                  <v-icon
                    class="mx-1"
                    @click="formatText('a', marker.id)"
                    @mousedown.prevent
                  >
                    mdi-link
                  </v-icon>

                  <v-icon
                    class="mx-1"
                    @click="formatText('img', marker.id)"
                    @mousedown.prevent
                  >
                    mdi-image-outline
                  </v-icon>

                  <v-dialog width="400" :ref="`color-dialog.${marker.id}`">
                    <template v-slot:activator="{ on, attrs }">
                      <v-icon v-on="on" v-bind="attrs">mdi-palette</v-icon>
                    </template>

                    <v-card class="pa-2">
                      <v-color-picker
                        hide-mode-switch
                        style="margin: 0 auto"
                        mode="hexa"
                        v-model="textColor"
                      />

                      <v-card-actions class="justify-end">
                        <v-btn @click="formatText('color', marker.id)">
                          Save
                        </v-btn>
                      </v-card-actions>
                    </v-card>
                  </v-dialog>

                  <v-textarea
                    label="Description"
                    style="
                      color: #fff;
                      background: var(--v-background-light-base);
                      transition: 0.3s;
                    "
                    :ref="`textarea_${marker.id}`"
                    v-model="marker.extra.description"
                  />

                  <v-text-field
                    dense
                    label="Title"
                    v-model="marker.title"
                    :ref="`textField_${marker.id}`"
                    @keyup.enter="(e) => onEnterHandler(marker.id, e)"
                    @input="(e) => inputHandler(e, marker)"
                  />

                  <v-card-actions class="justify-end">
                    <v-btn @click.stop="saveAndClose(marker.id)">
                      Save
                    </v-btn>
                  </v-card-actions>
                </v-card>
              </v-dialog>
            </div>
          </slot>
        </div>
      </foreignObject>

      <foreignObject
        x="65"
        y="30"
        width="40"
        height="40"
        v-if="marker.title"
        :transform-origin="`${120 / 2} 80`"
        :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(
          marker.x + 14 - 120 / 2,
          1
        )} ${marker.y - 45})`"
        @mouseenter="(e) => mouseEnterHandler(marker.id, e)"
        @mouseleave="(e) => mouseLeaveHandler(marker.id, e)"
      >
        <div class="map__popup-content">
          <slot name="popup" :marker="marker">
            <div
              class="map__popup-content--default del"
              @click="(e) => delMarker(e, marker.id, marker.x, marker.y)"
            >
              X
            </div>
          </slot>
        </div>
      </foreignObject>
    </template>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </nc-map>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import api from "@/api.js";
import { NcMap } from "nocloud-ui";

export default {
  components: { NcMap },
  mixins: [snackbar],
  name: "support-map",
  props: {
    template: { type: Object, required: true },
    region: { type: String, default: "" },
    multiSelect: { type: Boolean, default: false },
    activePinTitle: { type: String, default: "" },
    canAddPin: { type: Boolean, default: true },
    error: { type: String, default: "" },
  },

  data: () => ({
    selected: "",
    textColor: "#fff",

    markersSave: [],
    markers: [],
    item: {}
  }),
  methods: {
    formatText(tag, id) {
      const textarea = this.$refs[`textarea_${id}`][0].$el.querySelector("textarea");
      const { selectionStart, selectionEnd } = textarea;
      const text = textarea.value.slice(selectionStart, selectionEnd);

      switch (tag) {
        case "a": {
          const pos = selectionStart + 17;

          textarea.setRangeText(`<a href="https://">${text}</a>`);
          textarea.setSelectionRange(pos, pos);
          break;
        }

        case "img": {
          const pos = selectionStart + 10;

          textarea.setRangeText(`<img src="" alt="">${text}`);
          textarea.setSelectionRange(pos, pos);
          break;
        }

        case "color": {
          const color = `color: ${this.textColor.toLowerCase()}`;
          const pos = selectionStart + color.length + 13;

          this.$refs[`color-dialog.${id}`][0].isActive = false;
          setTimeout(() => { textarea.focus() });

          setTimeout(() => {
            textarea.setRangeText(`<span style="${color}">${text}</span>`);
            textarea.setSelectionRange(pos - 8, pos);
          }, 100);
          break;
        }

        default:
          textarea.setRangeText(`<${tag}>${text}</${tag}>`);
      }
    },
    onEnterHandler(id, e) {
      const ref = this.$refs["textField_" + id][0];

      this.mouseLeaveHandler(id);
      ref.blur();
      e.stopPropagation();
    },
    inputHandler(e, marker) {
      if (!e) {
        marker.title = " ";
      } else {
        marker.title = e.trim();
      }
    },
    delMarker(e, id, x, y) {
      e.stopPropagation();
      this.markers.forEach((m, i) => {
        if (m.id === id && m.x === x && m.y === y) {
          this.markers.splice(i, 1);
          this.selected = "";
        }
      });
    },
    saveCountry() {
      let error = 0;
      this.markers.forEach((el) => {
        if (el.title && !el.title.trim()) {
          const ref = this.$refs["textField_" + el.id][0];

          this.mouseEnterHandler(el.id);
          setTimeout(() => {
            ref.focus();
          }, 100);

          error = 1;
        }
      });
      if (error) {
        this.showSnackbarError({
          message: "Error: Enter location names.",
        });
        return;
      }
      this.$emit("save", this.item);
      this.isLoading = true;
      this.item.locations = JSON.parse(JSON.stringify(this.markers));

      if (this.item.locations.length < 1) {
        this.item.locations = [{ id: "_nocloud.remove" }];
      }
      api.servicesProviders
        .update(this.item.uuid, this.item)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Service edited successfully",
          });
        })
        .catch((err) => {
          this.showSnackbarError({
            message: err,
          });
          alert(err);
        })
        .finally(() => {
          this.isLoading = false;
        });

      this.markersSave = JSON.parse(JSON.stringify(this.markers));
    },
    saveAndClose(id) {
      this.$refs["edit-dialog." + id][0].isActive = false;
      this.saveCountry();
    },
    cancelSelectedCountry() {
      this.markers = JSON.parse(JSON.stringify(this.markersSave));
    },
    // ---------------------------
    mapClickHandler({ target, offsetX, offsetY }) {
      if (!this.canAddPin) {
        this.$emit("errorAddPin");
        return;
      }
      if (target.id) {
        this.selected = (this.region) ? `${target.id}-${this.region}` : target.id;
      } else {
        return false;
      }
      let stop = false;

      const kx = this.widthMap / (this.widthMap * this.scale);
      const ky = this.heightMap / (this.heightMap * this.scale);
      const w = this.$refs.map.$refs.viewport.getAttribute("transform").split(" ")[4];
      const h = this.$refs.map.$refs.viewport.getAttribute("transform").split(" ")[5];
      const x =
        parseInt(offsetX * kx - parseInt(w) / this.scale) -
        12 -
        this.scale * 0.12;
      const y =
        parseInt(offsetY * ky - parseInt(h) / this.scale) -
        35 -
        this.scale * 0.07;

      this.markers.forEach((el) => {
        if (el.x == x && el.y == y) {
          stop = true;
        }
      });

      if (stop) {
        return false;
      }

      setTimeout(() => {
        const marker = { id: this.selected, title: " ", extra: {}, x, y };

        if (this.multiSelect) {
          this.markers.push({ ...marker, extra: { region: this.region } });
        } else {
          this.markers = [marker];
        }

        this.mouseEnterHandler(marker.id);

        setTimeout(() => {
          const ref = this.$refs["textField_" + marker.id][0];

          ref.focus();
        }, 200);
      }, 10);
    },
    mouseEnterHandler(id) {
      this.selected = id;
      this.$emit("pinHover", id);
      this.$refs.map.mouseEnterHandler(id, null, true);
    },
    mouseLeaveHandler(id) {
      this.$refs.map.mouseLeaveHandler(id);
    }
  },
  mounted() {
    this.item = JSON.parse(JSON.stringify(this.template));
    this.markers = this.template.locations;
  },
  computed: {
    scale() {
      return this.$refs.map?.scale ?? 1;
    },
    widthMap() {
      return this.$refs.map?.widthMap ?? 1010;
    },
    heightMap() {
      return this.$refs.map?.heightMap ?? 666;
    }
  },
  watch: {
    error(message) {
      if (message === "") return;
      this.showSnackbarError({ message });
    },
    activePinTitle(value) {
      this.selected = this.markers?.find(({ title }) => title === value)?.id ?? "";
    }
  }
};
</script>

<style>
.map__popup-content--default.del {
  color: red;
  font-size: 24px;
  font-weight: bold;
  cursor: pointer;
}

#mapMain {
  height: calc(100vh - 220px);
  margin-top: 10px;
}
#mapMain .v-input--dense > .v-input__control > .v-input__slot {
  margin-bottom: -21px;
}

#mapMain .theme--dark.v-text-field--solo > .v-input__control > .v-input__slot {
  background: #ffffff;
  color: #000;
}
#mapMain .theme--dark.v-input input {
  color: #000;
}
/* #mapMain .v-application--is-ltr .v-messages {
  z-index: -10;
}

#mapMain .v-text-field.v-text-field--enclosed .v-text-field__details {
  z-index: -10;
} */
</style>
