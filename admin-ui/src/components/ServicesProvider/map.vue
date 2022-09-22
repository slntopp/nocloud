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
    <div v-if="selectedC" style="position: absolute; right: 25px; bottom: 13px">
      <v-btn
        @click="saveCountry"
        style="margin-right: 5px; background-color: #4caf50"
        class="ant-btn-primary"
      >
        Save
      </v-btn>
      <v-btn @click="CancelSelectedCountry" style="background-color: #272727">
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
              fill="#FF6E6E"
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
              'map__part--selected': selected == country.id,
              'is-settings': isSettings,
              'is-settings-selected': selectedC === country.id,
              'to-del': toDel === country.id,
              'settings-selected': checkSettingsSelected(country.id),
            }"
            class="map__part"
            :id="country.id"
            :title="country.title"
            :d="country.d"
            @click="(e) => selectedCountry(country.id, country.title)"
          />
        </g>

        <g
          @click.stop
          class="map_ui"
          ref="notscale"
          transform="matrix(1 0 0 1 0 0)"
        >
          <g
            v-for="marker in markerOrder"
            :key="marker.id + '_' + marker.x + '_' + marker.y + '_1'"
          >
            <use
              :href="`#${marker.svgId || 'marker'}`"
              class="map__marker"
              :data-id="marker.id + '_' + marker.x + '_' + marker.y"
              x="0"
              y="0"
              :transform="`matrix(${0.8 / scale} 0 0 ${0.8 / scale} ${
                marker.x
              } ${marker.y})`"
              transform-origin="14 36"
              @mouseenter="
                (e) =>
                  mouseEnterHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
              @mouseleave="
                (e) =>
                  mouseLeaveHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
            />
          </g>
          <g
            v-for="marker in markerOrder"
            :key="marker.id + '_' + marker.x + '_' + marker.y + '_2'"
            class="map__popup"
            :class="{
              'map__popup--active':
                selected == marker.id + '_' + marker.x + '_' + marker.y,
              'map__popup--hovered':
                hovered == marker.id + '_' + marker.x + '_' + marker.y,
            }"
          >
            <!-- popup -->
            <rect
              x="0"
              y="0"
              :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(
                marker.x + 14 - popupWidth / 2,
                1
              )} ${marker.y - 45})`"
              :transform-origin="`${popupWidth / 2} 80`"
              :width="popupWidth"
              height="40"
              fill="#fff"
              stroke-width="1"
              stroke="#000"
              rx="8"
              @mouseenter="
                (e) =>
                  mouseEnterHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
              @mouseleave="
                (e) =>
                  mouseLeaveHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
            ></rect>
            <!-- text -->
            <foreignObject
              v-if="marker.title"
              x="0"
              y="0"
              :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(
                marker.x + 14 - popupWidth / 2,
                1
              )} ${marker.y - 45})`"
              :transform-origin="`${popupWidth / 2} 80`"
              :width="popupWidth"
              height="40"
              @mouseenter="
                (e) =>
                  mouseEnterHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
              @mouseleave="
                (e) =>
                  mouseLeaveHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
            >
              <div class="map__popup-content">
                <slot name="popup" :marker="marker">
                  <div class="map__popup-content--default">
                    <!-- {{ marker.title }} -->
                    <v-form>
                      <v-text-field
                        :ref="
                          'textFiel_' +
                          marker.id +
                          '_' +
                          marker.x +
                          '_' +
                          marker.y
                        "
                        @keyup.enter="
                          (e) =>
                            onEnterHandler(
                              marker.id + '_' + marker.x + '_' + marker.y,
                              e
                            )
                        "
                        @input="
                          (e) =>
                            inputHandler(
                              marker.id + '_' + marker.x + '_' + marker.y,
                              e,
                              marker
                            )
                        "
                        v-model="marker.title"
                        label=""
                        solo
                        dense
                      >
                      </v-text-field>
                    </v-form>
                  </div>
                </slot>
              </div>
            </foreignObject>

            <foreignObject
              v-if="marker.title"
              x="65"
              y="40"
              :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(
                marker.x + 14 - popupWidth / 2,
                1
              )} ${marker.y - 45})`"
              :transform-origin="`${popupWidth / 2} 80`"
              width="40"
              height="40"
              @mouseenter="
                (e) =>
                  mouseEnterHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
              @mouseleave="
                (e) =>
                  mouseLeaveHandler(
                    marker.id + '_' + marker.x + '_' + marker.y,
                    e
                  )
              "
            >
              <div class="map__popup-content">
                <slot name="popup" :marker="marker">
                  <div
                    @click="(e) => delMarker(e, marker.id, marker.x, marker.y)"
                    class="map__popup-content--default del"
                  >
                    X
                  </div>
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
import mapData from "../../map.json";
import snackbar from "@/mixins/snackbar.js";
import api from "@/api.js";

export default {
  mixins: [snackbar],
  name: "support-map",
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
    svg: null,
    mapData,
    // markersInComponent: [],

    widthMap: 1010,
    heightMap: 666,

    markersSave: [
      {
        id: "PL",
        title: "Test 1",
        x: 517,
        y: 255,
      },
      {
        id: "BY",
        title: "Test 1",
        x: 540,
        y: 250,
      },
    ],
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
    onEnterHandler(id, e) {
      e.stopPropagation();
      const ref = this.$refs["textFiel_" + id][0];
      this.mouseLeaveHandler(id);
      ref.blur();
    },
    inputHandler(id, e, marker) {
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

      this.selectedC = id;
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
            this.$refs["textFiel_" + el.id + "_" + el.x + "_" + el.y][0];
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
      this.isLoading = true;
      this.item.locations = JSON.parse(JSON.stringify(this.markers));

      console.log("this.item = ", this.item);
      console.log("this.uuid = ", this.uuid);

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

      this.mouseLeaveHandler();

      // console.log("this.markerOrder = ", this.markerOrder);
    },
    CancelSelectedCountry() {
      this.selectedC = "";
      this.markers = JSON.parse(JSON.stringify(this.markersSave));
    },
    checkSettingsSelected(countryId) {
      let f = false;
      this.markerOrder.forEach((el) => {
        if (el.id === countryId) {
          f = true;
        }
      });
      return f;
    },
    // ---------------------------
    mapClickHandler({ target, offsetX, offsetY }) {
      if (!target.id) {
        return false;
      }
      // -------------------------
      const kx = this.widthMap / (this.widthMap * this.scale);
      const ky = this.heightMap / (this.heightMap * this.scale);
      const w = this.$refs.viewport.getAttribute("transform").split(" ")[4];
      const h = this.$refs.viewport.getAttribute("transform").split(" ")[5];

      this.x = parseInt(offsetX * kx - parseInt(w) / (this.scale * 1.03));
      this.y = parseInt(offsetY * ky - parseInt(h) / (this.scale * 1.161));

      const obg = {
        id: JSON.parse(JSON.stringify(this.selectedC)),
        // title: JSON.parse(JSON.stringify(this.titleMarker)),
        title: " ",
        x: JSON.parse(JSON.stringify(this.x)),
        y: JSON.parse(JSON.stringify(this.y)),
      };

      this.markers.push(obg);

      this.mouseEnterHandler(obg.id + "_" + obg.x + "_" + obg.y);

      setTimeout(() => {
        const ref =
          this.$refs["textFiel_" + obg.id + "_" + obg.x + "_" + obg.y][0];
        ref.focus();
      }, 200);
    },
    mouseEnterHandler(id) {
      this.hovered = id;
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
    },
  },

  mounted() {
    this.markers = JSON.parse(JSON.stringify(this.markersSave));
    this.widthMap = +this.$refs.map.getBoundingClientRect().width;
    this.heightMap = +this.$refs.map.getBoundingClientRect().height;

    this.uuid = this.$route.params.uuid;
    api.servicesProviders
      .list()
      .then((response) => {
        this.item = response.pool.find((el) => el.uuid == this.uuid);
      })
      .catch((error) => {
        console.log("error = ", error);
      })
      .finally(() => {});

    //---------------------------
    const container = this.$refs.viewport;
    const min = { x: Infinity, y: Infinity };
    const max = { x: 0, y: 0 };

    let x, y;
    this.markers.forEach(({ x, y }) => {
      if (min.x > x) min.x = x;
      if (min.y > y) min.y = y;
      if (max.x < x) max.x = x;
      if (max.y < y) max.y = y;
    });
    this.scale = this.widthMap / (max.x - min.x);
    if (this.scale > this.maxScale) {
      this.scale = this.maxScale;
    } else {
      if (this.scale < this.minScale) {
        this.scale = this.minScale;
      }
    }
    x =
      (min.x + (max.x - min.x) / 2 - this.widthMap / (2 * this.scale * 1.028)) *
      this.scale *
      1.028;
    y =
      (min.y +
        (max.y - min.y) / 2 -
        this.heightMap / (2 * this.scale * 1.143)) *
      this.scale *
      1.143;
    if (!x) {
      x = 0;
    }
    if (!y) {
      y = 0;
    }

    this.svg = this.$refs.svgwrapper;
    container.setAttribute(
      "transform",
      `matrix(${this.scale} 0 0 ${this.scale} ${-x} ${-y})`
    );
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
#mapMain .theme--dark.v-input input,
.theme--dark.v-input textarea {
  color: #000;
}
#mapMain .v-application--is-ltr .v-messages {
  z-index: -10;
}
</style>
