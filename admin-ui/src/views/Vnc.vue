<template>
  <div class="vnc__page">
    <template>
      <div id="noVNC_status"></div>

      <div id="noVNC_control_bar_anchor" class="noVNC_vcenter">
        <div id="noVNC_control_bar">
          <div id="noVNC_control_bar_handle" title="Hide/Show the control bar">
            <div></div>
          </div>

          <div class="noVNC_scroll">
            <div @click="removeCanvases">
              <!-- <router-link class="noVNC_button noVNC_button--goBack" :to="{path: `/cloud-${$route.params.pathMatch}`}" title="Go back">
							<v-icon type="left"></v-icon>
						</router-link> -->
              <a
                class="noVNC_button noVNC_button--goBack"
                @click="$router.go(-1)"
                title="Go back"
              >
                <v-icon type="left"></v-icon>
              </a>
            </div>

            <!-- Drag/Pan the viewport -->
            <input
              type="image"
              alt="Drag"
              src="img/images/drag.svg"
              id="noVNC_view_drag_button"
              class="noVNC_button noVNC_hidden"
              title="Move/Drag Viewport"
            />

            <!--noVNC Touch Device only buttons-->
            <div id="noVNC_mobile_buttons">
              <input
                type="image"
                alt="Keyboard"
                src="img/images/keyboard.svg"
                id="noVNC_keyboard_button"
                class="noVNC_button"
                title="Show Keyboard"
              />
            </div>

            <!-- Extra manual keys -->
            <input
              type="image"
              alt="Extra keys"
              src="img/images/toggleextrakeys.svg"
              id="noVNC_toggle_extra_keys_button"
              class="noVNC_button"
              title="Show Extra Keys"
            />
            <div class="noVNC_vcenter">
              <div id="noVNC_modifiers" class="noVNC_panel">
                <input
                  type="image"
                  alt="Ctrl"
                  src="img/images/ctrl.svg"
                  id="noVNC_toggle_ctrl_button"
                  class="noVNC_button"
                  title="Toggle Ctrl"
                />
                <input
                  type="image"
                  alt="Alt"
                  src="img/images/alt.svg"
                  id="noVNC_toggle_alt_button"
                  class="noVNC_button"
                  title="Toggle Alt"
                />
                <input
                  type="image"
                  alt="Windows"
                  src="img/images/windows.svg"
                  id="noVNC_toggle_windows_button"
                  class="noVNC_button"
                  title="Toggle Windows"
                />
                <input
                  type="image"
                  alt="Tab"
                  src="img/images/tab.svg"
                  id="noVNC_send_tab_button"
                  class="noVNC_button"
                  title="Send Tab"
                />
                <input
                  type="image"
                  alt="Esc"
                  src="img/images/esc.svg"
                  id="noVNC_send_esc_button"
                  class="noVNC_button"
                  title="Send Escape"
                />
                <input
                  type="image"
                  alt="Ctrl+Alt+Del"
                  src="img/images/ctrlaltdel.svg"
                  id="noVNC_send_ctrl_alt_del_button"
                  class="noVNC_button"
                  title="Send Ctrl-Alt-Del"
                />
              </div>
            </div>

            <!-- Shutdown/Reboot -->
            <input
              type="image"
              alt="Shutdown/Reboot"
              src="img/images/power.svg"
              id="noVNC_power_button"
              class="noVNC_button"
              title="Shutdown/Reboot..."
            />
            <div class="noVNC_vcenter">
              <div id="noVNC_power" class="noVNC_panel">
                <div class="noVNC_heading">
                  <img alt="" src="img/images/power.svg" /> Power
                </div>
                <input
                  type="button"
                  id="noVNC_shutdown_button"
                  value="Shutdown"
                />
                <input type="button" id="noVNC_reboot_button" value="Reboot" />
                <input type="button" id="noVNC_reset_button" value="Reset" />
              </div>
            </div>

            <!-- Clipboard -->
            <input
              type="image"
              alt="Clipboard"
              src="img/images/clipboard.svg"
              id="noVNC_clipboard_button"
              class="noVNC_button"
              title="Clipboard"
            />
            <div class="noVNC_vcenter">
              <div id="noVNC_clipboard" class="noVNC_panel">
                <div class="noVNC_heading">
                  <img alt="" src="img/images/clipboard.svg" /> Clipboard
                </div>
                <textarea id="noVNC_clipboard_text" rows="5"></textarea>
                <br />
                <input
                  id="noVNC_clipboard_clear_button"
                  type="button"
                  value="Clear"
                  class="noVNC_submit"
                />
              </div>
            </div>

            <!-- Toggle fullscreen -->
            <input
              type="image"
              alt="Fullscreen"
              src="img/images/fullscreen.svg"
              id="noVNC_fullscreen_button"
              class="noVNC_button noVNC_hidden"
              title="Fullscreen"
            />

            <!-- Settings -->
            <input
              type="image"
              alt="Settings"
              src="img/images/settings.svg"
              id="noVNC_settings_button"
              class="noVNC_button"
              title="Settings"
            />
            <div class="noVNC_vcenter">
              <div id="noVNC_settings" class="noVNC_panel">
                <ul>
                  <li class="noVNC_heading">
                    <img alt="" src="img/images/settings.svg" /> Settings
                  </li>
                  <template v-if="instance && !isLoading">
                    <li>password:</li>
                    <li>
                      <v-text-field
                        type="password"
                        v-model="instance.config.password"
                      >
                      </v-text-field>
                    </li>
                  </template>
                  <li>
                    <hr />
                  </li>
                  <li>
                    <label
                      ><input id="noVNC_setting_shared" type="checkbox" />
                      Shared Mode</label
                    >
                  </li>
                  <li>
                    <label
                      ><input id="noVNC_setting_view_only" type="checkbox" />
                      View Only</label
                    >
                  </li>
                  <li>
                    <hr />
                  </li>
                  <li>
                    <label
                      ><input id="noVNC_setting_view_clip" type="checkbox" />
                      Clip to Window</label
                    >
                  </li>
                  <li>
                    <label for="noVNC_setting_resize">Scaling Mode:</label>
                    <select id="noVNC_setting_resize" name="vncResize">
                      <option value="off">None</option>
                      <option value="scale">Local Scaling</option>
                      <option value="remote">Remote Resizing</option>
                    </select>
                  </li>
                  <li>
                    <hr />
                  </li>
                  <li>
                    <div class="noVNC_expander">Advanced</div>
                    <div>
                      <ul>
                        <li>
                          <label for="noVNC_setting_quality">Quality:</label>
                          <input
                            id="noVNC_setting_quality"
                            type="range"
                            min="0"
                            max="9"
                            value="6"
                          />
                        </li>
                        <li>
                          <label for="noVNC_setting_compression"
                            >Compression level:</label
                          >
                          <input
                            id="noVNC_setting_compression"
                            type="range"
                            min="0"
                            max="9"
                            value="2"
                          />
                        </li>
                        <!-- <li><hr></li> -->
                        <!-- <li>
										<label for="noVNC_setting_repeaterID">Repeater ID:</label>
										<input id="noVNC_setting_repeaterID" type="text" value="">
									</li> -->
                        <!-- <li>
										<div class="noVNC_expander">WebSocket</div>
										<div><ul>
											<li>
												<label><input id="noVNC_setting_encrypt" type="checkbox"> Encrypt</label>
											</li>
											<li>
												<label for="noVNC_setting_host">Host:</label>
												<input id="noVNC_setting_host">
											</li>
											<li>
												<label for="noVNC_setting_port">Port:</label>
												<input id="noVNC_setting_port" type="number">
											</li>
											<li>
												<label for="noVNC_setting_path">Path:</label>
												<input id="noVNC_setting_path" type="text" value="websockify">
											</li>
										</ul></div>
									</li> -->
                        <li>
                          <hr />
                        </li>
                        <li>
                          <label
                            ><input
                              id="noVNC_setting_reconnect"
                              type="checkbox"
                            />
                            Automatic Reconnect</label
                          >
                        </li>
                        <li>
                          <label for="noVNC_setting_reconnect_delay"
                            >Reconnect Delay (ms):</label
                          >
                          <input
                            id="noVNC_setting_reconnect_delay"
                            type="number"
                          />
                        </li>
                        <li>
                          <hr />
                        </li>
                        <li>
                          <label
                            ><input
                              id="noVNC_setting_show_dot"
                              type="checkbox"
                            />
                            Show Dot when No Cursor</label
                          >
                        </li>
                        <!-- <li><hr></li> -->
                        <!-- <li>
										<label>Logging:
											<select id="noVNC_setting_logging" name="vncLogging">
											</select>
										</label>
									</li> -->
                      </ul>
                    </div>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <div id="noVNC_control_bar_hint"></div>
      </div>
      <!-- End of noVNC_control_bar -->

      <!-- Transition Screens -->
      <div id="noVNC_transition">
        <div id="noVNC_transition_text"></div>
        <div>
          <input
            type="button"
            id="noVNC_cancel_reconnect_button"
            value="Cancel"
            class="noVNC_submit"
          />
        </div>
        <div class="noVNC_spinner"></div>
      </div>

      <div id="noVNC_container">
        <!-- Note that Google Chrome on Android doesn't respect any of these,
						html attributes which attempt to disable text suggestions on the
						on-screen keyboard. Let's hope Chrome implements the ime-mode
						style for example -->
        <textarea
          id="noVNC_keyboardinput"
          autocapitalize="off"
          autocomplete="off"
          spellcheck="false"
          tabindex="-1"
        ></textarea>
      </div>

      <main>
        <div id="vnc-screen" ref="vncscreen"></div>
      </main>
    </template>
  </div>
</template>

<script>
import UI from "vnc-ui-vue";

export default {
  name: "vnc-view",
  data: () => ({ desktopName: "", token: "", url: "", rfb: null }),
  created() {
    document.title = `Console ${this.instanceId} | NoCloud`;

    if (this.instance) return;

    this.$store.dispatch("instances/get", this.instanceId).catch((err) => {
      this.$router.go(-1);
      alert(err);
    });
  },
  mounted() {
    if (!this.instance) return;
    this.getToken();
  },
  methods: {
    getToken() {
      this.$store
        .dispatch("vnc/actionVMInvoke", {
          uuid: this.instanceId,
          action: "start_vnc",
        })
        .then((res) => {
          this.token = res.meta.token;
          this.desktopName = this.instance?.title ?? "Unknown";

          this.url = `wss://${
            this.instance.sp
          }.${window.location.hostname.replace("api", "proxy")}/socket?${
            res.meta.url
          }`;
          this.connect(this.$store.state.auth.token);
        })
        .catch((err) => console.error(err));
    },
    connect(token) {
      this.$refs.vncscreen.innerHTML = "";
      UI.connect(this.url, token);
      UI.prime();
      if (UI.connected) location.reload();
    },
    credentialsAreRequired() {
      const password = prompt("Password Required:");

      this.rfb.sendCredentials({ password });
    },
    updateDesktopName(e) {
      this.desktopName = e.detail.name;
    },
    removeCanvases() {
      const conv = document.getElementsByTagName("canvas");

      Array.from(conv).forEach((el) => el.remove());
    },
  },
  computed: {
    instanceId() {
      return this.$route.params.instanceId;
    },
    instance() {
      if (!this.$store.getters["instances/one"]) {
        return;
      }

      return {
        ...this.$store.getters["instances/one"].instance,
        ...this.$store.getters["instances/one"],
      };
    },
    isLoading() {
      return (
        this.$store.getters["vnc/isLoading"] ||
        this.$store.getters["instances/isLoading"]
      );
    },
  },
  watch: {
    instance() {
      this.getToken();
    },
  },
};
</script>

<style scoped>
.vnc__page {
  width: 100vw;
  height: 100vh;
}
.vnc__header {
  display: flex;
  justify-content: space-around;
}

.container {
  min-width: 1200px;
}

.header__back {
  font-size: 1.2rem;
}

.noVNC_button--goBack {
  color: white;
  text-align: center;
  font-size: 1.1rem;
}

/*
 * noVNC base CSS
 * Copyright (C) 2019 The noVNC Authors
 * noVNC is licensed under the MPL 2.0 (see LICENSE.txt)
 * This file is licensed under the 2-Clause BSD license (see LICENSE.txt).
 */

/*
 * Z index layers:
 *
 * 0: Main screen
 * 10: Control bar
 * 50: Transition blocker
 * 60: Connection popups
 * 100: Status bar
 * ...
 * 1000: Javascript crash
 * ...
 * 10000: Max (used for polyfills)
 */

body {
  margin: 0;
  padding: 0;
  font-family: Helvetica;
  /*Background image with light grey curve.*/
  background-color: #494949;
  background-repeat: no-repeat;
  background-position: right bottom;
  height: 100%;
  touch-action: none;
}

html {
  height: 100%;
}

.noVNC_only_touch.noVNC_hidden {
  display: none;
}

.noVNC_disabled {
  color: rgb(128, 128, 128);
}

/* ----------------------------------------
 * Spinner
 * ----------------------------------------
 */

.noVNC_spinner {
  position: relative;
}

.noVNC_spinner,
.noVNC_spinner::before,
.noVNC_spinner::after {
  width: 10px;
  height: 10px;
  border-radius: 2px;
  box-shadow: -60px 10px 0 rgba(255, 255, 255, 0);
  animation: noVNC_spinner 1s linear infinite;
}

.noVNC_spinner::before {
  content: "";
  position: absolute;
  left: 0px;
  top: 0px;
  animation-delay: -0.1s;
}

.noVNC_spinner::after {
  content: "";
  position: absolute;
  top: 0px;
  left: 0px;
  animation-delay: 0.1s;
}

@keyframes noVNC_spinner {
  0% {
    box-shadow: -60px 10px 0 rgba(255, 255, 255, 0);
    width: 20px;
  }

  25% {
    box-shadow: 20px 10px 0 rgba(255, 255, 255, 1);
    width: 10px;
  }

  50% {
    box-shadow: 60px 10px 0 rgba(255, 255, 255, 0);
    width: 10px;
  }
}

/* ----------------------------------------
 * Input Elements
 * ----------------------------------------
 */

input:not([type]),
input[type="date"],
input[type="datetime-local"],
input[type="email"],
input[type="month"],
input[type="number"],
input[type="password"],
input[type="search"],
input[type="tel"],
input[type="text"],
input[type="time"],
input[type="url"],
input[type="week"],
textarea {
  /* Disable default rendering */
  -webkit-appearance: none;
  -moz-appearance: none;
  background: none;

  margin: 2px;
  padding: 2px;
  border: 1px solid rgb(192, 192, 192);
  border-radius: 5px;
  color: black;
  background: linear-gradient(
    to top,
    rgb(255, 255, 255) 80%,
    rgb(240, 240, 240)
  );
}

input[type="button"],
input[type="color"],
input[type="reset"],
input[type="submit"],
select {
  /* Disable default rendering */
  -webkit-appearance: none;
  -moz-appearance: none;
  background: none;

  margin: 2px;
  padding: 2px;
  border: 1px solid rgb(192, 192, 192);
  border-bottom-width: 2px;
  border-radius: 5px;
  color: black;
  background: linear-gradient(to top, rgb(255, 255, 255), rgb(240, 240, 240));

  /* This avoids it jumping around when :active */
  vertical-align: middle;
}

input[type="button"],
input[type="color"],
input[type="reset"],
input[type="submit"] {
  padding-left: 20px;
  padding-right: 20px;
}

option {
  color: black;
  background: white;
}

input:not([type]):focus,
input[type="button"]:focus,
input[type="color"]:focus,
input[type="date"]:focus,
input[type="datetime-local"]:focus,
input[type="email"]:focus,
input[type="month"]:focus,
input[type="number"]:focus,
input[type="password"]:focus,
input[type="reset"]:focus,
input[type="search"]:focus,
input[type="submit"]:focus,
input[type="tel"]:focus,
input[type="text"]:focus,
input[type="time"]:focus,
input[type="url"]:focus,
input[type="week"]:focus,
select:focus,
textarea:focus {
  box-shadow: 0px 0px 3px rgba(74, 144, 217, 0.5);
  border-color: rgb(74, 144, 217);
  outline: none;
}

input[type="button"]::-moz-focus-inner,
input[type="color"]::-moz-focus-inner,
input[type="reset"]::-moz-focus-inner,
input[type="submit"]::-moz-focus-inner {
  border: none;
}

input:not([type]):disabled,
input[type="button"]:disabled,
input[type="color"]:disabled,
input[type="date"]:disabled,
input[type="datetime-local"]:disabled,
input[type="email"]:disabled,
input[type="month"]:disabled,
input[type="number"]:disabled,
input[type="password"]:disabled,
input[type="reset"]:disabled,
input[type="search"]:disabled,
input[type="submit"]:disabled,
input[type="tel"]:disabled,
input[type="text"]:disabled,
input[type="time"]:disabled,
input[type="url"]:disabled,
input[type="week"]:disabled,
select:disabled,
textarea:disabled {
  color: rgb(128, 128, 128);
  background: rgb(240, 240, 240);
}

input[type="button"]:active,
input[type="color"]:active,
input[type="reset"]:active,
input[type="submit"]:active,
select:active {
  border-bottom-width: 1px;
  margin-top: 3px;
}

:root:not(.noVNC_touch) input[type="button"]:hover:not(:disabled),
:root:not(.noVNC_touch) input[type="color"]:hover:not(:disabled),
:root:not(.noVNC_touch) input[type="reset"]:hover:not(:disabled),
:root:not(.noVNC_touch) input[type="submit"]:hover:not(:disabled),
:root:not(.noVNC_touch) select:hover:not(:disabled) {
  background: linear-gradient(to top, rgb(255, 255, 255), rgb(250, 250, 250));
}

/* ----------------------------------------
 * WebKit centering hacks
 * ----------------------------------------
 */

.noVNC_center {
  /*
   * This is a workaround because webkit misrenders transforms and
   * uses non-integer coordinates, resulting in blurry content.
   * Ideally we'd use "top: 50%; transform: translateY(-50%);" on
   * the objects instead.
   */
  display: flex;
  align-items: center;
  justify-content: center;
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.noVNC_center > * {
  pointer-events: auto;
}

.noVNC_vcenter {
  display: flex;
  flex-direction: column;
  justify-content: center;
  position: fixed;
  top: 0;
  left: 0;
  height: 100%;
  pointer-events: none;
}

.noVNC_vcenter > * {
  pointer-events: auto;
}

/* ----------------------------------------
 * Layering
 * ----------------------------------------
 */

.noVNC_connect_layer {
  z-index: 60;
}

/* ----------------------------------------
 * Fallback error
 * ----------------------------------------
 */

#noVNC_fallback_error {
  z-index: 1000;
  visibility: hidden;
}

#noVNC_fallback_error.noVNC_open {
  visibility: visible;
}

#noVNC_fallback_error > div {
  max-width: 90%;
  padding: 15px;

  transition: 0.5s ease-in-out;

  transform: translateY(-50px);
  opacity: 0;

  text-align: center;
  font-weight: bold;
  color: #fff;

  border-radius: 10px;
  box-shadow: 6px 6px 0px rgba(0, 0, 0, 0.5);
  background: rgba(200, 55, 55, 0.8);
}

#noVNC_fallback_error.noVNC_open > div {
  transform: translateY(0);
  opacity: 1;
}

#noVNC_fallback_errormsg {
  font-weight: normal;
}

#noVNC_fallback_errormsg .noVNC_message {
  display: inline-block;
  text-align: left;
  font-family: monospace;
  white-space: pre-wrap;
}

#noVNC_fallback_error .noVNC_location {
  font-style: italic;
  font-size: 0.8em;
  color: rgba(255, 255, 255, 0.8);
}

#noVNC_fallback_error .noVNC_stack {
  max-height: 50vh;
  padding: 10px;
  margin: 10px;
  font-size: 0.8em;
  text-align: left;
  font-family: monospace;
  white-space: pre;
  border: 1px solid rgba(0, 0, 0, 0.5);
  background: rgba(0, 0, 0, 0.2);
  overflow: auto;
}

/* ----------------------------------------
 * Control Bar
 * ----------------------------------------
 */

#noVNC_control_bar_anchor {
  /* The anchor is needed to get z-stacking to work */
  position: fixed;
  z-index: 10;

  transition: 0.5s ease-in-out;

  /* Edge misrenders animations wihthout this */
  transform: translateX(0);
}

:root.noVNC_connected #noVNC_control_bar_anchor.noVNC_idle {
  opacity: 0.8;
}

#noVNC_control_bar_anchor.noVNC_right {
  left: auto;
  right: 0;
}

#noVNC_control_bar {
  position: relative;
  left: -100%;

  transition: 0.5s ease-in-out;

  background-color: rgb(110, 132, 163);
  border-radius: 0 10px 10px 0;
}

#noVNC_control_bar.noVNC_open {
  box-shadow: 6px 6px 0px rgba(0, 0, 0, 0.5);
  left: 0;
}

#noVNC_control_bar::before {
  /* This extra element is to get a proper shadow */
  content: "";
  position: absolute;
  z-index: -1;
  height: 100%;
  width: 30px;
  left: -30px;
  transition: box-shadow 0.5s ease-in-out;
}

#noVNC_control_bar.noVNC_open::before {
  box-shadow: 6px 6px 0px rgba(0, 0, 0, 0.5);
}

.noVNC_right #noVNC_control_bar {
  left: 100%;
  border-radius: 10px 0 0 10px;
}

.noVNC_right #noVNC_control_bar.noVNC_open {
  left: 0;
}

.noVNC_right #noVNC_control_bar::before {
  visibility: hidden;
}

#noVNC_control_bar_handle {
  position: absolute;
  left: -15px;
  top: 0;
  transform: translateY(35px);
  width: calc(100% + 30px);
  height: 50px;
  z-index: -1;
  cursor: pointer;
  border-radius: 5px;
  background-color: rgb(83, 99, 122);
  background-image: url("/public/img/images/handle_bg.svg");
  background-repeat: no-repeat;
  background-position: right;
  box-shadow: 3px 3px 0px rgba(0, 0, 0, 0.5);
}

#noVNC_control_bar_handle:after {
  content: "";
  transition: transform 0.5s ease-in-out;
  background: url("/public/img/images/handle.svg");
  position: absolute;
  top: 22px;
  /* (50px-6px)/2 */
  right: 5px;
  width: 5px;
  height: 6px;
}

#noVNC_control_bar.noVNC_open #noVNC_control_bar_handle:after {
  transform: translateX(1px) rotate(180deg);
}

:root:not(.noVNC_connected) #noVNC_control_bar_handle {
  display: none;
}

.noVNC_right #noVNC_control_bar_handle {
  background-position: left;
}

.noVNC_right #noVNC_control_bar_handle:after {
  left: 5px;
  right: 0;
  transform: translateX(1px) rotate(180deg);
}

.noVNC_right #noVNC_control_bar.noVNC_open #noVNC_control_bar_handle:after {
  transform: none;
}

#noVNC_control_bar_handle div {
  position: absolute;
  right: -35px;
  top: 0;
  width: 50px;
  height: 50px;
}

:root:not(.noVNC_touch) #noVNC_control_bar_handle div {
  display: none;
}

.noVNC_right #noVNC_control_bar_handle div {
  left: -35px;
  right: auto;
}

#noVNC_control_bar .noVNC_scroll {
  max-height: 100vh;
  /* Chrome is buggy with 100% */
  overflow-x: hidden;
  overflow-y: auto;
  padding: 0 10px 0 5px;
}

.noVNC_right #noVNC_control_bar .noVNC_scroll {
  padding: 0 5px 0 10px;
}

/* Control bar hint */
#noVNC_control_bar_hint {
  position: fixed;
  left: calc(100vw - 50px);
  right: auto;
  top: 50%;
  transform: translateY(-50%) scale(0);
  width: 100px;
  height: 50%;
  max-height: 600px;

  visibility: hidden;
  opacity: 0;
  transition: 0.2s ease-in-out;
  background: transparent;
  box-shadow: 0 0 10px black, inset 0 0 10px 10px rgba(110, 132, 163, 0.8);
  border-radius: 10px;
  transition-delay: 0s;
}

#noVNC_control_bar_anchor.noVNC_right #noVNC_control_bar_hint {
  left: auto;
  right: calc(100vw - 50px);
}

#noVNC_control_bar_hint.noVNC_active {
  visibility: visible;
  opacity: 1;
  transition-delay: 0.2s;
  transform: translateY(-50%) scale(1);
}

/* General button style */
.noVNC_button {
  display: block;
  padding: 4px 4px;
  margin: 10px 0;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
}

.noVNC_button.noVNC_selected {
  border-color: rgba(0, 0, 0, 0.8);
  background: rgba(0, 0, 0, 0.5);
}

.noVNC_button:disabled {
  opacity: 0.4;
}

.noVNC_button:focus {
  outline: none;
}

.noVNC_button:active {
  padding-top: 5px;
  padding-bottom: 3px;
}

/* Android browsers don't properly update hover state if touch events
 * are intercepted, but focus should be safe to display */
:root:not(.noVNC_touch) .noVNC_button.noVNC_selected:hover,
.noVNC_button.noVNC_selected:focus {
  border-color: rgba(0, 0, 0, 0.4);
  background: rgba(0, 0, 0, 0.2);
}

:root:not(.noVNC_touch) .noVNC_button:hover,
.noVNC_button:focus {
  background: rgba(255, 255, 255, 0.2);
}

.noVNC_button.noVNC_hidden {
  display: none;
}

/* Panels */
.noVNC_panel {
  transform: translateX(25px);

  transition: 0.5s ease-in-out;

  max-height: 100vh;
  /* Chrome is buggy with 100% */
  overflow-x: hidden;
  overflow-y: auto;

  visibility: hidden;
  opacity: 0;

  padding: 15px;

  background: #fff;
  border-radius: 10px;
  color: #000;
  border: 2px solid #e0e0e0;
  box-shadow: 6px 6px 0px rgba(0, 0, 0, 0.5);
}

.noVNC_panel.noVNC_open {
  visibility: visible;
  opacity: 1;
  transform: translateX(75px);
}

.noVNC_right .noVNC_vcenter {
  left: auto;
  right: 0;
}

.noVNC_right .noVNC_panel {
  transform: translateX(-25px);
}

.noVNC_right .noVNC_panel.noVNC_open {
  transform: translateX(-75px);
}

.noVNC_panel hr {
  border: none;
  border-top: 1px solid rgb(192, 192, 192);
}

.noVNC_panel label {
  display: block;
  white-space: nowrap;
}

.noVNC_panel .noVNC_heading {
  background-color: rgb(110, 132, 163);
  border-radius: 5px;
  padding: 5px;
  /* Compensate for padding in image */
  padding-right: 8px;
  color: white;
  font-size: 20px;
  margin-bottom: 10px;
  white-space: nowrap;
}

.noVNC_panel .noVNC_heading img {
  vertical-align: bottom;
}

.noVNC_submit {
  float: right;
}

/* Expanders */
.noVNC_expander {
  cursor: pointer;
}

.noVNC_expander::before {
  content: url("/public/img/images/expander.svg");
  display: inline-block;
  margin-right: 5px;
  transition: 0.2s ease-in-out;
}

.noVNC_expander.noVNC_open::before {
  transform: rotateZ(90deg);
}

.noVNC_expander ~ * {
  margin: 5px;
  margin-left: 10px;
  padding: 5px;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 5px;
}

.noVNC_expander:not(.noVNC_open) ~ * {
  display: none;
}

/* Control bar content */

:root:not(.noVNC_connected) #noVNC_view_drag_button {
  display: none;
}

/* noVNC Touch Device only buttons */
:root:not(.noVNC_connected) #noVNC_mobile_buttons {
  display: none;
}

:root:not(.noVNC_touch) #noVNC_mobile_buttons {
  display: none;
}

/* Extra manual keys */
:root:not(.noVNC_connected) #noVNC_toggle_extra_keys_button {
  display: none;
}

#noVNC_modifiers {
  background-color: rgb(92, 92, 92);
  border: none;
  padding: 0 10px;
}

/* Shutdown/Reboot */
:root:not(.noVNC_connected) #noVNC_power_button {
  display: none;
}

#noVNC_power_buttons {
  display: none;
}

#noVNC_power input[type="button"] {
  width: 100%;
}

/* Clipboard */
:root:not(.noVNC_connected) #noVNC_clipboard_button {
  display: none;
}

#noVNC_clipboard {
  /* Full screen, minus padding and left and right margins */
  max-width: calc(100vw - 2 * 15px - 75px - 25px);
}

#noVNC_clipboard_text {
  width: 500px;
  max-width: 100%;
}

/* Settings */
#noVNC_settings {
  min-width: 250px;
}

#noVNC_settings ul {
  list-style: none;
  margin: 0px;
  padding: 0px;
}

#noVNC_setting_port {
  width: 80px;
}

#noVNC_setting_path {
  width: 100px;
}

/* Version */

.noVNC_version_wrapper {
  font-size: small;
}

.noVNC_version {
  margin-left: 1rem;
}

/* Connection Controls */
:root:not(.noVNC_connected) #noVNC_disconnect_button {
  display: none;
}

/* ----------------------------------------
 * Status Dialog
 * ----------------------------------------
 */

#noVNC_status {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  z-index: 100;
  transform: translateY(-100%);

  cursor: pointer;

  transition: 0.5s ease-in-out;

  visibility: hidden;
  opacity: 0;

  padding: 5px;

  display: flex;
  flex-direction: row;
  justify-content: center;
  align-content: center;

  line-height: 25px;
  word-wrap: break-word;
  color: #fff;

  border-bottom: 1px solid rgba(0, 0, 0, 0.9);
}

#noVNC_status.noVNC_open {
  transform: translateY(0);
  visibility: visible;
  opacity: 1;
}

#noVNC_status::before {
  content: "";
  display: inline-block;
  width: 25px;
  height: 25px;
  margin-right: 5px;
}

#noVNC_status.noVNC_status_normal {
  background: rgba(128, 128, 128, 0.9);
}

#noVNC_status.noVNC_status_normal::before {
  content: url("/public/img/images/info.svg") " ";
}

#noVNC_status.noVNC_status_error {
  background: rgba(200, 55, 55, 0.9);
}

#noVNC_status.noVNC_status_error::before {
  content: url("/public/img/images/error.svg") " ";
}

#noVNC_status.noVNC_status_warn {
  background: rgba(180, 180, 30, 0.9);
}

#noVNC_status.noVNC_status_warn::before {
  content: url("/public/img/images/warning.svg") " ";
}

/* ----------------------------------------
 * Connect Dialog
 * ----------------------------------------
 */

#noVNC_connect_dlg {
  transition: 0.5s ease-in-out;

  transform: scale(0, 0);
  visibility: hidden;
  opacity: 0;
}

#noVNC_connect_dlg.noVNC_open {
  transform: scale(1, 1);
  visibility: visible;
  opacity: 1;
}

#noVNC_connect_dlg .noVNC_logo {
  transition: 0.5s ease-in-out;
  padding: 10px;
  margin-bottom: 10px;

  font-size: 80px;
  text-align: center;

  border-radius: 5px;
}

@media (max-width: 440px) {
  #noVNC_connect_dlg {
    max-width: calc(100vw - 100px);
  }

  #noVNC_connect_dlg .noVNC_logo {
    font-size: calc(25vw - 30px);
  }
}

#noVNC_connect_button {
  cursor: pointer;

  padding: 10px;

  color: white;
  background-color: rgb(110, 132, 163);
  border-radius: 12px;

  text-align: center;
  font-size: 20px;

  box-shadow: 6px 6px 0px rgba(0, 0, 0, 0.5);
}

#noVNC_connect_button div {
  margin: 2px;
  padding: 5px 30px;
  border: 1px solid rgb(83, 99, 122);
  border-bottom-width: 2px;
  border-radius: 5px;
  background: linear-gradient(to top, rgb(110, 132, 163), rgb(99, 119, 147));

  /* This avoids it jumping around when :active */
  vertical-align: middle;
}

#noVNC_connect_button div:active {
  border-bottom-width: 1px;
  margin-top: 3px;
}

:root:not(.noVNC_touch) #noVNC_connect_button div:hover {
  background: linear-gradient(to top, rgb(110, 132, 163), rgb(105, 125, 155));
}

#noVNC_connect_button img {
  vertical-align: bottom;
  height: 1.3em;
}

/* ----------------------------------------
 * Password Dialog
 * ----------------------------------------
 */

#noVNC_credentials_dlg {
  position: relative;

  transform: translateY(-50px);
}

#noVNC_credentials_dlg.noVNC_open {
  transform: translateY(0);
}

#noVNC_credentials_dlg ul {
  list-style: none;
  margin: 0px;
  padding: 0px;
}

.noVNC_hidden {
  display: none;
}

/* ----------------------------------------
 * Main Area
 * ----------------------------------------
 */

/* Transition screen */
#noVNC_transition {
  display: none;

  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  right: 0;

  color: white;
  background: rgba(0, 0, 0, 0.5);
  z-index: 50;

  /*display: flex;*/
  align-items: center;
  justify-content: center;
  flex-direction: column;
}

:root.noVNC_loading #noVNC_transition,
:root.noVNC_connecting #noVNC_transition,
:root.noVNC_disconnecting #noVNC_transition,
:root.noVNC_reconnecting #noVNC_transition {
  display: flex;
}

:root:not(.noVNC_reconnecting) #noVNC_cancel_reconnect_button {
  display: none;
}

#noVNC_transition_text {
  font-size: 1.5em;
}

/* Main container */
#noVNC_container {
  width: 100%;
  height: 100%;
  background-color: #313131;
  /* border-bottom-right-radius: 800px 600px; */
  /*border-top-left-radius: 800px 600px;*/
}

#noVNC_keyboardinput {
  width: 1px;
  height: 1px;
  background-color: #fff;
  color: #fff;
  border: 0;
  position: absolute;
  left: -40px;
  z-index: -1;
  ime-mode: disabled;
}

.noVNC_logo span {
  color: green;
}

#noVNC_bell {
  display: none;
}

/* ----------------------------------------
 * Media sizing
 * ----------------------------------------
 */

@media screen and (max-width: 640px) {
  #noVNC_logo {
    font-size: 150px;
  }
}

@media screen and (min-width: 321px) and (max-width: 480px) {
  #noVNC_logo {
    font-size: 110px;
  }
}

@media screen and (max-width: 320px) {
  #noVNC_logo {
    font-size: 90px;
  }
}
</style>
