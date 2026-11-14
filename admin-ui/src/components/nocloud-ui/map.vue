<template>
  <div class="map" ref="map">
    <div class="map__buttons">
      <button class="map__button" @click="(e) => zoom(e, 1)">
        +
      </button>
      <button class="map__button" @click="(e) => zoom(e, -1)">
        –
      </button>
    </div>

    <slot name="actions"></slot>

    <svg
      ref="svgwrapper"
      :viewBox="`0 0 ${widthMap} ${heightMap}`"
      @click="mapClickHandler"
      @mousemove="drag"
      @mousedown="beginDrag"
      @mousewheel="zoom"
      @touchmove="drag"
      @touchstart="beginDrag"
    >
      <defs>
        <g id="marker">
          <slot name="marker">
            <path
              d="M14,0 C21.732,0 28,5.641 28,12.6 C28,23.963 14,36 14,36 C14,36 0,24.064 0,12.6 C0,5.641 6.268,0 14,0 Z"
              id="Shape"
              fill="#FF6E6E"
              style="fill: inherit"
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

        <g v-for="(icon, id) in markerIcons" :key="id" v-html="icon" :id="id"></g>
      </defs>
      <g
        class="map__viewport"
				ref="viewport"
				transform="matrix(1 0 0 1 0 0)"
      >
        <g v-for="country in mapData.countries" :key="country.id + country.title">
          <title>{{ country.title }}</title>
          <path
            :style="{
              fill: (isSelected(country.id)) ? (mapActiveColor ?? '#427cf7') : null,
              cursor: (highlightHoveredCountry) ? 'pointer': null
            }"
            class="map__part"
            :id="country.id"
            :title="country.title"
            :d="country.d"
            @mouseenter="hoverCountry(country.id, $event)"
            @mouseleave="unhoverCountry(country.id, $event)"
          />
        </g>

        <g
          class="map_ui"
          ref="notscale"
          transform="matrix(1 0 0 1 0 0)"
        >
          <g v-for="marker in markerOrder" :key="marker.id + '_1'">
            <use
              x="0"
              y="0"
              class="map__marker"
              :style="{
                fill: (selected == marker.id)
                  ? (marker.extra.activeColor ?? markerActiveColor ?? '#0fd058')
                  : (marker.extra.color ?? markerColor ?? '#f9383b')
              }"
              :href="`#${marker.svgId || 'marker'}`"
              :data-id="marker.id"
              :transform="`matrix(${markerScaleDivider / scale} 0 0 ${markerScaleDivider / scale} ${marker.x} ${marker.y})`"
              transform-origin="14 36"
              @mouseenter="(e) => mouseEnterHandler(marker.id, e)"
              @mouseleave="(e) => mouseLeaveHandler(marker.id, e)"
            />
          </g>
          <g
            v-for="marker in markerOrder"
            :key="marker.id + '_2'"
            class="map__popup"
            :class="{ 'map__popup--hovered': hovered == marker.id }"
          >
            <!-- popup -->
            <slot name="popup" :marker="marker">
              <rect
                x="0"
                y="0"
                :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(marker.x + 14 - popupWidth[marker.id] / 2, 1)} ${(marker.y - 45)})`"
                :transform-origin="`${popupWidth[marker.id]/2} 80`"
                :width="popupWidth[marker.id]"
                height="40"
                fill="#fff"
                stroke-width="1"
                stroke="#000"
                rx="8"
                @mouseenter="(e) => mouseEnterHandler(marker.id, e)"
                @mouseleave="(e) => mouseLeaveHandler(marker.id, e)"
              ></rect>

              <!-- text -->
              <foreignObject
                v-if="marker.title"
                x="0"
                y="0"
                :transform="`matrix(${1 / scale} 0 0 ${1 / scale} ${Math.max(marker.x + 14 - popupWidth[marker.id] / 2, 1)} ${(marker.y - 45)})`"
                :transform-origin="`${popupWidth[marker.id]/2} 80`"
                :width="popupWidth[marker.id]"
                height="40"
                @mouseenter="(e) => mouseEnterHandler(marker.id, e)"
                @mouseleave="(e) => mouseLeaveHandler(marker.id, e)"
              >
                <div class="map__popup-content">
                  <slot name="popup" :marker="marker">
                    <div class="map__popup-content--default">
                      {{ marker.title }}
                    </div>
                  </slot>
                </div>
              </foreignObject>
            </slot>
          </g>
        </g>
      </g>
    </svg>
  </div>
</template>

<script>
import mapData from "./map.json";

export default {
  name: "nc-map",
  emits: ["update:modelValue"],
  props: {
    modelValue: { type: String, required: true },
    markers: { type: Array, required: true },
    mapActiveColor: { type: String, default: null },
    markerActiveColor: { type: String, default: null },
    markerColor: { type: String, default: null },
    mapClick: { type: Function, default: null },
    notScale: { type: Boolean, default: false },
    markerScaleDivider: { type: Number, default: 1 },
    highlightHoveredCountry: { type: Boolean, default: false },
    markerUrl: { type: String, default: null }
  },
  data: () => ({
    selected: "",
    hovered: "",
    popupWidth: {},
    markerIcons: {},

    leaveDelay: 300,
    leaveDelayInterval: -1,
    scale: 1,
    maxScale: 6,
		minScale: 1,

		selectedDrag: null,
    svg: null,
    mapData,

    widthMap: 1010,
    heightMap: 666,
  }),
  methods: {
    mapClickHandler(event) {
      if (this.mapClick) {
        this.mapClick(event);
        return;
      }

      if (event.target.dataset.id) {
        this.selected = event.target.getAttribute("data-id");
        this.$emit("update:modelValue", this.selected);
      }
    },
    mouseEnterHandler(id, e, select) {
      this.hovered = id;
      if (select) this.selected = id;
      clearInterval(this.leaveDelayInterval);
    },
    mouseLeaveHandler() {
      this.leaveDelayInterval = setInterval(() => {
        this.hovered = "";
      }, this.leaveDelay);
    },
    beginDrag(e) {
			if (e.target.closest('.map_ui')) return;
			if (e.target.classList.contains('draggable')) {
				this.selectedDrag = e.target;
			} else {
				this.selectedDrag = this.$refs.viewport;
			}

			this.selectedDrag.dataset.startMouseX = (e.clientX)
        ? e.clientX
        : e.changedTouches[0].clientX;
			this.selectedDrag.dataset.startMouseY = (e.clientY)
        ? e.clientY
        : e.changedTouches[0].clientY;
			e.stopPropagation();

      if (!e.clientX) {
        document.querySelector('body').style.overflow = 'hidden';
        document.querySelector('.newCloud_wrapper').style.overflow = 'hidden';
      }
		},
		drag(e) {
			if (!this.selectedDrag) return;
			e.stopPropagation();
      e.preventDefault();

			const startX = parseFloat(this.selectedDrag.dataset.startMouseX);
			const startY = parseFloat(this.selectedDrag.dataset.startMouseY);
			let dx = ((e.clientX) ? e.clientX : e.changedTouches[0].clientX) - startX;
			let dy = ((e.clientY) ? e.clientY : e.changedTouches[0].clientY) - startY;

			if (this.selectedDrag.classList.contains('draggable')) {
        const selectedBox = this.selectedDrag.getBoundingClientRect();
        const boundaryBox = this.selectedDrag.parentElement.getBoundingClientRect();

        if (selectedBox.right + dx > boundaryBox.right) {
          dx = (boundaryBox.right - selectedBox.right);
        } else if (selectedBox.left + dx < boundaryBox.left) {
          dx = (boundaryBox.left - selectedBox.left);
        }

        if (selectedBox.bottom + dy > boundaryBox.bottom) {
          dy = (boundaryBox.bottom - selectedBox.bottom);
        } else if (selectedBox.top + dy < boundaryBox.top) {
          dy = (boundaryBox.top - selectedBox.top);
        }
			}

			const currentMatrix = this.selectedDrag.transform.baseVal.consolidate().matrix;
			const newMatrix = currentMatrix.translate(dx / this.scale, dy / this.scale);
			const transform = this.svg.createSVGTransformFromMatrix(newMatrix);

			this.selectedDrag.transform.baseVal.initialize(transform);
			this.selectedDrag.dataset.startMouseX = (dx / (e.clientX ? 1 : 3)) + startX;
			this.selectedDrag.dataset.startMouseY = (dy / (e.clientY ? 1 : 3)) + startY;
		},
		endDrag(e) {
			e.stopPropagation();
			if (this.selectedDrag) {
				this.selectedDrag = undefined;
			}

      if (!e.clientX) {
        document.querySelector('body').style.overflow = '';
        document.querySelector('.newCloud_wrapper').style.overflow = '';
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
			point.x = (delta) ? box.x / 2 + box.left : e.clientX - box.left;
			point.y = (delta) ? box.y / 2 + box.top : e.clientY - box.top;

			const currentZoomMatrix = container.transform.baseVal[0].matrix;
			point = point.matrixTransform(currentZoomMatrix.inverse());
			const matrix = this.svg.createSVGMatrix()
					.translate(point.x, point.y)
					.scale(scaleStep)
					.translate(-point.x, -point.y);
			const newZoomMatrix = currentZoomMatrix.multiply(matrix);

			container.transform.baseVal.initialize(this.svg.createSVGTransformFromMatrix(newZoomMatrix));
		},
    hoverCountry(id, { target }) {
      if (!this.highlightHoveredCountry) return;
      const isCurrentCountry = !this.selected.split('-').at(0).includes(id);

      if (id === target.id && isCurrentCountry) {
        target.style.fill = this.mapActiveColor ?? '#427cf7';
      }
    },
    unhoverCountry(id, { target }) {
      if (!this.highlightHoveredCountry) return;
      const isCurrentCountry = !this.selected.split('-').at(0).includes(id);

      if (id === target.id && isCurrentCountry) {
        target.style.fill = "";
      }
    },
    isSelected(id) {
      const { extra } = this.markers.find((marker) => marker.id === this.selected) ?? {};

      if (!extra) return false;
      if (extra.country) return extra.country === id;
      if (extra.region) return this.selected?.split('-').at(-2) === id;
      else return this.selected?.split('-').at(-1) === id;
    },
    changeMarkers() {
      const links = [];

      if (this.markerUrl) links.push({ link: this.markerUrl });

      this.markers.forEach(({ id, extra }) => {
        if (extra.link) links.push({ id, link: extra.link });
      });

      links.forEach(async ({ id, link }) => {
        const response = await fetch(link);
        const blob = await response.blob();
        const text = await blob.text();
        const i = text.indexOf('<svg');

        if (id) {
          this.markerIcons[id] = text.slice(i);
        } else {
          this.markerIcons['default'] = text.slice(i);
        }
      });
    },
    setScale() {
      const container = this.$refs.viewport;
      const min = { x: Infinity, y: Infinity };
      const max = { x: -Infinity, y: -Infinity };
      let x, y, right = 0, down = 0;

      this.markers.forEach(({ x, y }) => {
        if (min.x > x) min.x = x;
        if (min.y > y) min.y = y;
        if (max.x < x) max.x = x;
        if (max.y < y) max.y = y;
      });

      const xScale = mapData.meta.width / (max.x - min.x + 70);
      const yScale = mapData.meta.height / (max.y - min.y + 70);

      this.scale = Math.min(xScale, yScale);
      if (this.scale > this.maxScale) this.scale = this.maxScale;
      if (xScale > yScale) right = 20 * xScale;
      if (yScale > xScale) down = 20 * yScale;

      x = (min.x - 20) * this.scale - right;
      y = (min.y - 10) * this.scale - down;

      if (this.markers.length < 2) {
        const { width, height } = mapData.meta;

        this.scale = this.maxScale;
        x = (this.markers[0].x * this.scale - width / 2) + 70;
        y = (this.markers[0].y * this.scale - height / 2) + 140;
      }

		  container.setAttribute('transform', `matrix(${this.scale} 0 0 ${this.scale} ${-x} ${-y})`);
    }
  },
  computed: {
    markerOrder() {
      const tempMarkers = [...this.markers];

      tempMarkers.sort((a, b) => {
        if (a.id == this.hovered) {
          return 1;
        }
        if (b.id == this.hovered) {
          return -1;
        }
        return 0;
      });

      return tempMarkers.map((marker) => {
        const icon = this.markerIcons[marker.id];
        const defaultIcon = this.markerIcons.default;
        let svgId = null;

        if (icon) svgId = marker.id;
        else if (defaultIcon) svgId = 'default';
        return { ...marker, svgId };
      });
    },
  },
  created() {
    this.selected = this.modelValue ?? this.selected;

    if (this.markers.length === 1) {
      this.$emit('update:modelValue', this.markers[0].id);
    }

    this.markers.forEach(({ id, title }) => {
      this.popupWidth[id] = title.length * 10 + 15;
    });
  },
  mounted() {
    this.changeMarkers();
    this.widthMap = mapData.meta.width;
    this.heightMap = mapData.meta.height;

    const container = this.$refs.viewport;

    if (this.notScale) {
      this.widthMap = +this.$refs.map.getBoundingClientRect().width;
      this.heightMap = +this.$refs.map.getBoundingClientRect().height;

      let x = parseInt(this.widthMap - mapData.meta.width) / 2;
      let y = parseInt(this.heightMap - mapData.meta.height) / 2;

      container.setAttribute("transform", `matrix(1 0 0 1 ${x} ${y})`);
    } else if (this.markers.length > 0) {
      this.setScale();
    }

		this.svg = this.$refs.svgwrapper;
		window.addEventListener('mouseup', this.endDrag);
    window.addEventListener('touchend', this.endDrag);
	},
	beforeUnmount() {
		window.removeEventListener('mouseup', this.endDrag);
    window.removeEventListener('touchend', this.endDrag);
	},
  watch: {
    modelValue(newVal) {
      this.selected = newVal;
    },
    markers(value) {
      value.forEach(({ id, title }) => {
        this.popupWidth[id] = title.length * 10 + 15;
      });
      this.setScale()
    }
  },
};
</script>

<style scoped>
.map {
  /* width: 800px;
  height: 600px; */
  /* display: flex; */
  position: relative;
  align-items: center;
  justify-content: center;
}
/* .map .vue-pan-zoom-item {
  width: 720px;
} */
/* .map .vue-pan-zoom-scene {
  width: 100%;
} */
.map__marker {
  cursor: pointer;
}
.map__popup {
  visibility: hidden;
  opacity: 0;
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
  text-align: center;
}
.map__popup-content--default {
  font-size: 16px;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}
.map svg {
  width: 100%;
  height: 100%;
}
.map__part {
  fill: #c9c9c9;
  transition: fill 0.2s ease;
  stroke: #fff;
  stroke-opacity: 1;
  stroke-width: 1;
}

.map__buttons {
  position: absolute;
  right: 25px;
  top: 13px;
  display: flex;
}

.map__button {
  padding: 3px 13px 6px;
  font-size: 22px;
  line-height: 1;
  border: 1px solid rgba(0, 0, 0, 0.15);
  border-radius: 4px;
  background: #fff;
  color: #000;
  cursor: pointer;
}

.map__button:not(:last-child) {
  margin-right: 5px;
}

.map__button:last-child {
  padding: 2px 13px 7px;
}
</style>
