<template>
  <div ref="map" class="map" id="mapMain">
    <div style="position: absolute; right: 25px; top: 13px">
      <v-btn
        style="margin-right: 5px; font-size: 20px"
        @click="(e) => zoom(e, 1)"
      >
        +
      </v-btn>
      <v-btn style="font-size: 20px" @click="(e) => zoom(e, -1)"> - </v-btn>
    </div>

    <!-- byn  .ant-btn-primary -->
    <div v-if="selectedC || multiSelect" style="position: absolute; right: 25px; bottom: 13px">
      <v-btn
        class="ant-btn-primary"
        style="margin-right: 5px; background-color: #4caf50"
        @click="saveCountry"
      >
        Save
      </v-btn>
      <v-btn style="background-color: #272727" @click="CancelSelectedCountry">
        Cancel
      </v-btn>
    </div>
    <!-- end byn -->
    <svg
      ref="svgwrapper"
      :viewBox="`0 0 ${widthMap} ${heightMap}`"
      @click="mapClickHandler"
      @mousemove="drag"
      @mousedown="beginDrag"
      @mousewheel="zoom"
    >
      <defs>
        <g id="marker">
          <slot name="marker">
            <path
              d="M14,0 C21.732,0 28,5.641 28,12.6 C28,23.963 14,36 14,36 C14,36 0,24.064 0,12.6 C0,5.641 6.268,0 14,0 Z"
              id="Shape"
            ></path>
            <circle
              id="elips"
              fill="#FFFFFF"
              fill-rule="nonzero"
              cx="14"
              cy="14"
              r="7"
            ></circle>
          </slot>
        </g>
      </defs>
      <g class="map__viewport" ref="viewport" transform="matrix(1 0 0 1 0 0)">
        <g v-for="country in mapData.countries" :key="country.id">
          <title>{{ country.title }}</title>
          <path
            :key="country.id + country.title"
            :class="{
              'map__part--selected': selected === country.id,
              'is-settings': isSettings,
              'is-settings-selected': selectedC === country.id,
              'to-del': toDel === country.id,
              'settings-selected': checkSettingsSelected(country.id),
            }"
            class="map__part"
            :id="country.id"
            :title="country.title"
            :d="country.d"
            @click="selectedCountry(country.id, country.title)"
          />
        </g>

        <g class="map_ui" ref="notscale" transform="matrix(1 0 0 1 0 0)" @click.stop>
          <g v-for="marker in markerOrder" :key="`${marker.id}_${marker.x}_${marker.y}_1`">
            <use
              x="0"
              y="0"
              transform-origin="14 36"
              :href="`#${marker.svgId || 'marker'}`"
              :data-id="`${marker.id}_${marker.x}_${marker.y}`"
              :class="{
                map__marker: true,
                active: activePinTitle && activePinTitle === marker.title,
              }"
              :transform="`matrix(${0.8 / scale} 0 0 ${0.8 / scale} ${marker.x} ${marker.y})`"
              @mouseenter="(e) => mouseEnterHandler(`${marker.id}_${marker.x}_${marker.y}`, e)"
              @mouseleave="(e) => mouseLeaveHandler(`${marker.id}_${marker.x}_${marker.y}`, e)"
            />
          </g>
          <g
            class="map__popup"
            v-for="marker in markerOrder"
            :key="`${marker.id}_${marker.x}_${marker.y}_2`"
            :class="{
              'map__popup--active': selected === `${marker.id}_${marker.x}_${marker.y}`,
              'map__popup--hovered': hovered === `${marker.id}_${marker.x}_${marker.y}`,
            }"
          >
            <!-- text -->
            <foreignObject
              x="65"
              y="55"
              width="40"
              height="40"
              v-if="marker.title"
              :transform-origin="`${popupWidth / 2} 80`"
              :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(marker.x + 14 - popupWidth / 2, 1)} ${marker.y - 45})`"
              @mouseenter="(e) => mouseEnterHandler(`${marker.id}_${marker.x}_${marker.y}`, e)"
              @mouseleave="(e) => mouseLeaveHandler(`${marker.id}_${marker.x}_${marker.y}`, e)"
            >
              <div class="map__popup-content">
                <slot name="popup" :marker="marker">
                  <div class="map__popup-content--default">
                    <!-- {{ marker.title }} -->
                    <v-dialog width="800">
                      <template v-slot:activator="{ on, attrs }">
                        <v-icon v-on="on" v-bind="attrs" color="secondary">mdi-cog</v-icon>
                      </template>
                      <v-card class="pa-4" color="background-light">
                        <v-icon @click="formatText('b', `${marker.id}_${marker.x}_${marker.y}`)" @mousedown.prevent>
                          mdi-format-bold
                        </v-icon>

                        <v-icon @click="formatText('i', `${marker.id}_${marker.x}_${marker.y}`)" @mousedown.prevent>
                          mdi-format-italic
                        </v-icon>

                        <v-icon @click="formatText('u', `${marker.id}_${marker.x}_${marker.y}`)" @mousedown.prevent>
                          mdi-format-underline
                        </v-icon>

                        <v-icon @click="formatText('s', `${marker.id}_${marker.x}_${marker.y}`)" @mousedown.prevent>
                          mdi-format-strikethrough
                        </v-icon>

                        <v-icon class="mx-1" @click="formatText('a', `${marker.id}_${marker.x}_${marker.y}`)" @mousedown.prevent>
                          mdi-link
                        </v-icon>

                        <v-icon @click="formatText('color', `${marker.id}_${marker.x}_${marker.y}`)" @mousedown.prevent>
                          mdi-palette
                        </v-icon>

                        <v-textarea
                          label="Description"
                          style="
                            color: #fff;
                            background: var(--v-background-light-base);
                            transition: 0.3s;
                          "
                          :ref="`textarea_${marker.id}_${marker.x}_${marker.y}`"
                          v-model="marker.extra.description"
                        />

                        <v-text-field
                          dense
                          label="Title"
                          v-model="marker.title"
                          :ref="`textField_${marker.id}_${marker.x}_${marker.y}`"
                          @keyup.enter="(e) => onEnterHandler(`${marker.id}_${marker.x}_${marker.y}`, e)"
                          @input="(e) => inputHandler(e, marker)"
                        />
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
              :transform-origin="`${popupWidth / 2} 80`"
              :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(marker.x + 14 - popupWidth / 2, 1)} ${marker.y - 45})`"
              @mouseenter="(e) => mouseEnterHandler(`${marker.id}_${marker.x}_${marker.y}`, e)"
              @mouseleave="(e) => mouseLeaveHandler(`${marker.id}_${marker.x}_${marker.y}`, e)"
            >
              <div class="map__popup-content">
                <slot name="popup" :marker="marker">
                  <div class="map__popup-content--default del" @click="(e) => delMarker(e, marker.id, marker.x, marker.y)">X</div>
                </slot>
              </div>
            </foreignObject>
          </g>
        </g>
      </g>
    </svg>

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
  </div>
</template>

<script>
import mapData from "@/map.json";
import snackbar from "@/mixins/snackbar.js";
import api from "@/api.js";

export default {
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
    hovered: "",
    popupWidth: 120,
    leaveDelay: 300,
    leaveDelayInterval: -1,
    scale: 1,
    maxScale: 10,
    minScale: 1,
    selectedDrag: null,
    dragF: false,
    svg: null,
    mapData,
    // markersInComponent: [],
    widthMap: 1010,
    heightMap: 666,

    markersSave: [],
    markers: [],
    //----------------------
    isSettings: true,
    selectedC: "",
    titleMarker: "",
    toDel: "",
    x: "",
    y: "",
  }),
  methods: {
    formatText(tag, id) {
      const textarea = this.$refs[`textarea_${id}`][0].$el.querySelector('textarea');
      const { selectionStart, selectionEnd } = textarea;
      const text = textarea.value.slice(selectionStart, selectionEnd);

      if (tag === 'a') {
        const pos = selectionStart + 17;

        textarea.setRangeText(`<a href="https://">${text}</a>`);
        textarea.setSelectionRange(pos, pos);
      } else if (tag === 'color') {
        const pos = selectionStart + 20;

        textarea.setRangeText(`<span style="color: ">${text}</span>`);
        textarea.setSelectionRange(pos, pos);
      } else {
        textarea.setRangeText(`<${tag}>${text}</${tag}>`);
      }
    },
    onEnterHandler(id, e) {
      e.stopPropagation();
      const ref = this.$refs["textField_" + id][0];
      this.mouseLeaveHandler(id);
      ref.blur();
    },
    inputHandler(e, marker) {
      this.selectedC = "inputHandler";
      if (!e) {
        marker.title = " ";
      } else {
        marker.title = e.trim();
      }
    },
    selectedCountry(id, country) {
      if (!this.isSettings) {
        return false;
      }

      if (this.dragF) {
        return false;
      }

      if (this.multiSelect) this.selectedC = `${id}-${this.region}`;
      else this.selectedC = id;
      this.titleMarker = country;
    },
    delMarker(e, id, x, y) {
      e.stopPropagation();
      this.markers.forEach((m, i) => {
        if (m.id === id && m.x === x && m.y === y) {
          this.markers.splice(i, 1);
          this.selectedC = id + "_del";
        }
      });
    },
    saveCountry() {
      let error = 0;
      this.markers.forEach((el) => {
        if (el.title && !el.title.trim()) {
          const ref =
            this.$refs["textField_" + el.id + "_" + el.x + "_" + el.y][0];
          this.mouseEnterHandler(el.id + "_" + el.x + "_" + el.y);
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
        this.item.locations = [{ id: '_nocloud.remove' }];
      }
      api.servicesProviders
        .update(this.uuid, this.item)
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
      this.selectedC = "";

      // console.log("this.markerOrder = ", this.markerOrder);
    },
    CancelSelectedCountry() {
      this.selectedC = "";
      this.markers = JSON.parse(JSON.stringify(this.markersSave));
    },
    checkSettingsSelected(countryId) {
      let f = false;
      this.markerOrder.forEach((el) => {
        if (el.id.includes(countryId)) {
          f = true;
        }
      });
      return f;
    },
    // ---------------------------
    mapClickHandler({ target, offsetX, offsetY }) {
      if (!this.canAddPin) {
        this.selectedC = "";
        this.$emit("errorAddPin");
        return;
      }
      if (!target.id) {
        parseInt;
        return false;
      }
      let stop = false;

      const kx = this.widthMap / (this.widthMap * this.scale);
      const ky = this.heightMap / (this.heightMap * this.scale);
      const w = this.$refs.viewport.getAttribute("transform").split(" ")[4];
      const h = this.$refs.viewport.getAttribute("transform").split(" ")[5];
      this.x =
        parseInt(offsetX * kx - parseInt(w) / this.scale) -
        12 -
        this.scale * 0.12;
      this.y =
        parseInt(offsetY * ky - parseInt(h) / this.scale) -
        35 -
        this.scale * 0.07;

      this.markers.forEach((el) => {
        if (el.x == this.x && el.y == this.y) {
          stop = true;
        }
      });

      if (stop) {
        return false;
      }

      setTimeout(() => {
        if (this.dragF) {
          this.dragF = false;
          return false;
        }
        const obg = {
          id: JSON.parse(JSON.stringify(this.selectedC)),
          title: " ",
          x: JSON.parse(JSON.stringify(this.x)),
          y: JSON.parse(JSON.stringify(this.y)),
        };
        if (this.multiSelect) {
          this.markers.push({ ...obg, extra: { region: this.region } });
        } else {
          this.markers = [obg];
        }

        this.mouseEnterHandler(obg.id + "_" + obg.x + "_" + obg.y);

        setTimeout(() => {
          const ref =
            this.$refs["textField_" + obg.id + "_" + obg.x + "_" + obg.y][0];
          ref.focus();
        }, 200);
      }, 10);
    },
    mouseEnterHandler(id) {
      this.hovered = id;
      this.$emit("pinHover", id.substring(0, id.indexOf("_")));
      clearInterval(this.leaveDelayInterval);
    },
    mouseLeaveHandler() {
      this.leaveDelayInterval = setInterval(() => {
        this.hovered = "";
      }, this.leaveDelay);
    },
    beginDrag(e) {
      e.stopPropagation();
      if (e.target.closest(".map_ui")) return;
      let target = e.target;
      if (target.classList.contains("draggable")) {
        this.selectedDrag = target;
      } else {
        this.selectedDrag = this.$refs.viewport;
      }
      this.selectedDrag.dataset.startMouseX = e.clientX;
      this.selectedDrag.dataset.startMouseY = e.clientY;
    },
    drag(e) {
      if (!this.selectedDrag) return;
      e.stopPropagation();
      let startX = parseFloat(this.selectedDrag.dataset.startMouseX),
        startY = parseFloat(this.selectedDrag.dataset.startMouseY),
        dx = e.clientX - startX,
        dy = e.clientY - startY;
      if (this.selectedDrag.classList.contains("draggable")) {
        let selectedBox = this.selectedDrag.getBoundingClientRect(),
          boundaryBox = this.selectedDrag.parentElement.getBoundingClientRect();
        if (selectedBox.right + dx > boundaryBox.right) {
          dx = boundaryBox.right - selectedBox.right;
        } else if (selectedBox.left + dx < boundaryBox.left) {
          dx = boundaryBox.left - selectedBox.left;
        }
        if (selectedBox.bottom + dy > boundaryBox.bottom) {
          dy = boundaryBox.bottom - selectedBox.bottom;
        } else if (selectedBox.top + dy < boundaryBox.top) {
          dy = boundaryBox.top - selectedBox.top;
        }
      }
      let currentMatrix =
          this.selectedDrag.transform.baseVal.consolidate().matrix,
        newMatrix = currentMatrix.translate(dx / this.scale, dy / this.scale),
        transform = this.svg.createSVGTransformFromMatrix(newMatrix);
      this.selectedDrag.transform.baseVal.initialize(transform);

      if (
        this.selectedDrag.dataset.startMouseX != dx + startX ||
        this.selectedDrag.dataset.startMouseY != dy + startY
      ) {
        this.dragF = true;
      } else {
        this.dragF = false;
      }
      this.selectedDrag.dataset.startMouseX = dx + startX;
      this.selectedDrag.dataset.startMouseY = dy + startY;
    },
    endDrag(e) {
      e.stopPropagation();
      if (this.selectedDrag) {
        this.selectedDrag = undefined;
      }
    },
    zoom(e, delta) {
      e.stopPropagation();
      e.preventDefault();
      const container = this.$refs.viewport;
      let scaleStep = (delta || e.wheelDelta) > 0 ? 1.25 : 0.8;
      if (this.scale * scaleStep > this.maxScale) {
        scaleStep = this.maxScale / this.scale;
      }
      if (this.scale * scaleStep < this.minScale) {
        scaleStep = this.minScale / this.scale;
      }
      const box = this.svg.getBoundingClientRect();
      let point = this.svg.createSVGPoint();
      this.scale *= scaleStep;
      point.x = delta ? box.x / 2 + box.left : e.clientX - box.left;
      point.y = delta ? box.y / 2 + box.top : e.clientY - box.top;
      const currentZoomMatrix = container.transform.baseVal[0].matrix;
      point = point.matrixTransform(currentZoomMatrix.inverse());
      const matrix = this.svg
        .createSVGMatrix()
        .translate(point.x, point.y)
        .scale(scaleStep)
        .translate(-point.x, -point.y);
      const newZoomMatrix = currentZoomMatrix.multiply(matrix);
      container.transform.baseVal.initialize(
        this.svg.createSVGTransformFromMatrix(newZoomMatrix)
      );
    },
  },
  watch: {
    error(newVallue) {
      this.showSnackbarError({
        message: newVallue,
      });
    },
  },
  computed: {
    markerOrder() {
      const tempMarkers = [...this.markers];
      return tempMarkers.sort((a, b) => {
        if (a.id == this.hovered) {
          return 1;
        }
        if (b.id == this.hovered) {
          return -1;
        }
        return 0;
      });
    }
  },

  mounted() {
    // this.markers = JSON.parse(JSON.stringify(this.markersSave));
    this.widthMap = +this.$refs.map.getBoundingClientRect().width;
    this.heightMap = +this.$refs.map.getBoundingClientRect().height;
    this.uuid = this.template.uuid;
    this.item = this.template;
    this.markers = this.template.locations;
    const container = this.$refs.viewport;
    let x = parseInt(this.widthMap - 1010) / 2;
    let y = parseInt(this.heightMap - 666) / 2;
    this.svg = this.$refs.svgwrapper;
    container.setAttribute("transform", `matrix(1 0 0 1 ${x} ${y})`);
    window.addEventListener("mouseup", this.endDrag);
  },
  beforeUnmount() {
    window.removeEventListener("mouseup", this.endDrag);
  },
};
</script>
<style>
.map {
  position: relative;
  align-items: center;
  justify-content: center;
}

.map__marker {
  cursor: pointer;
  fill: red;
}

.map__marker.active {
  fill: rgb(76, 175, 80);
}

.map__popup {
  visibility: hidden;
  opacity: 0;
  color: black;
  transition: visibility 0s linear 300ms, opacity 300ms;
}
.map__popup--hovered,
.map__popup--active {
  visibility: visible;
  opacity: 1;
  transition: visibility 0s linear 0s, opacity 300ms;
}
.map__popup-content {
  height: 100%;
  width: 100%;
}
.map__popup-content--default {
  font-size: 16px;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}
.map__popup-content--default.del {
  color: red;
  font-size: 24px;
  font-weight: bold;
  cursor: pointer;
}
.map svg {
  width: 100%;
  height: 100%;
  border: solid 1px #fff;
}
.map__part {
  fill: #c9c9c9;
  transition: fill 0.2s ease;
  stroke: white;
  stroke-opacity: 1;
  stroke-width: 0.5;
}
.map__part--selected {
  fill: #6755b1;
  stroke: gray;
  stroke-width: 0.8;
}

/* ---------------------- */

.map__part.is-settings {
  cursor: pointer;
}

.map__part.is-settings.is-settings-selected {
  fill: #6755b1;
  stroke: #c9c9c9;
  stroke-width: 0.3;
}

.map__part.is-settings.settings-selected {
  fill: #6755b1;
  stroke: #c9c9c9;
  stroke-width: 0.3;
}

.map__part.is-settings:hover {
  /* fill: #8e8e8e; */
  fill: #6755b1;
  stroke: #c9c9c9;
  stroke-width: 0.3;
}

.map__part.is-settings.to-del {
  fill: #ff4545;
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
#mapMain .v-application--is-ltr .v-messages {
  z-index: -10;
}

#mapMain .v-text-field.v-text-field--enclosed .v-text-field__details {
  z-index: -10;
}
</style>
