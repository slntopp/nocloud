<template>
  <div class="settings pa-4">
    <div class="buttons__inline pb-4">
      <v-menu
        offset-y
        transition="slide-y-transition"
        bottom
        :close-on-content-click="false"
        v-model="newSetting.visible"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn color="background-light" class="mr-2" v-bind="attrs" v-on="on">
            create
          </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row>
            <v-col>
              <v-text-field
                dense
                v-model="newSetting.data.key"
                label="key"
                :rules="newSetting.rules"
              >
              </v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                dense
                v-model="newSetting.data.data.description"
                label="description"
                :rules="newSetting.rules"
              >
              </v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                dense
                v-model="newSetting.data.data.value"
                label="value"
                :rules="newSetting.rules"
              >
              </v-text-field>
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-btn :loading="newSetting.loading" @click="createKey">
                send
              </v-btn>
            </v-col>
          </v-row>
        </v-card>
      </v-menu>

      <v-btn
        color="background-light"
        class="mr-8"
        :disabled="selected.length < 1"
        @click="deleteSelectedKeys"
      >
        delete
      </v-btn>
    </div>

    <nocloud-table
      item-key="key"
      :loading="loading"
      :headers="headers"
      :items="settings"
      show-select
      v-model="selected"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.description`]="{ item }">
        <div
          class="d-flex align-center"
          v-if="edit.key == 'description' && edit.data == item"
        >
          <div class="control">
            <v-icon @click="saveEditDescription()" class="edit-btn mr-2">
              mdi-content-save-outline
            </v-icon>
            <v-icon @click="stopEdit()" class="edit-btn mr-3">
              mdi-close-circle-outline
            </v-icon>
          </div>
          <v-text-field v-model="edit.data.description"></v-text-field>
        </div>
        <template v-else>
          <v-icon
            @click="setEditSettingValue('description', item)"
            class="edit-btn"
          >
            mdi-border-color
          </v-icon>
          {{ item.description }}
        </template>
      </template>
      <template v-slot:[`item.value`]="{ item }">
        <v-icon @click="startEditValue(item)" class="edit-btn">
          mdi-border-color
        </v-icon>
        {{ item.value }}
      </template>
    </nocloud-table>

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
    <template v-if="isModalEditOpen">
      <v-dialog
        max-width="800px"
        v-model="isModalEditOpen"
      >
        <div class="edit-modal">
          <json-editor
            @changeValue="(data) => changeSettingValue(data)"
            :json="edit.data ? JSON.parse(edit.data) : {}"
          />
          <div class="buttons">
            <v-btn depressed color="error" @click="isModalEditOpen=false"> Close </v-btn>
            <v-btn depressed color="success" @click="saveModalEdit"> Save </v-btn>
          </div>
        </div>
      </v-dialog>
    </template>
  </div>
</template>

<script>
import { mapGetters } from "vuex";
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import noCloudTable from "@/components/table.vue";
import JsonEditor from "@/components/JsonEditor.vue";

const headers = [
  { text: "key", value: "key" },
  { text: "description", value: "description" },
  { text: "value", value: "value" },
];

const defaultData = {
  key: "",
  data: {
    description: "",
    // visible: false,
    value: "",
  },
};

export default {
  name: "settings-view",
  created() {
    this.$store
      .dispatch("settings/fetch")
      .then(() => {
        this.fetchError = "";
      })
      .catch((err) => {
        console.log(`err`, err);
        this.fetchError = "Can't reach the server";
        if (err.response) {
          this.fetchError += `: [ERROR]: ${err.response.data.message}`;
        } else {
          this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
        }
      });
    // this.$store.dispatch('settings/fetchKeys')
  },
  components: {
    "nocloud-table": noCloudTable,
    JsonEditor,
  },
  mixins: [snackbar],
  data: () => ({
    headers,
    selected: [],
    newSetting: {
      rules: [(value) => !!value || "Required."],
      isLoading: false,
      visible: false,
      data: {
        ...JSON.parse(JSON.stringify(defaultData)),
      },
    },
    edit: {
      key: "",
      data: {},
    },
    isModalEditOpen: false,
    fetchError: "",
  }),
  computed: {
    ...mapGetters("settings", {
      settings: "all",
      loading: "isLoading",
    }),
  },
  methods: {
    deleteSelectedKeys() {
      if (this.selected.length > 0) {
        const deletePromices = this.selected.map((el) =>
          api.settings.delete(el.key)
        );
        Promise.all(deletePromices)
          .then((res) => {
            if (res.every((el) => el.key)) {
              this.$store.dispatch("settings/fetch");

              const ending = deletePromices.length == 1 ? "" : "s";
              this.showSnackbar({
                message: `Setting${ending} deleted successfully.`,
              });
            } else {
              this.showSnackbar({ message: `Some error executed` });
            }
          })
          .catch((err) => {
            if (err.response.status == 501 || err.response.status == 502) {
              const opts = {
                message: `Service Unavailable: ${err.response.data.message}.`,
                timeout: 0,
              };
              this.showSnackbarError(opts);
            }
          });
      }
    },
    createKey() {
      if (
        Object.keys(this.newSetting.data).every((dataKey) => {
          const dataValue = this.newSetting.data[dataKey];
          return this.newSetting.rules.every((rule) => {
            const res = typeof rule(dataValue) == "boolean";
            return res;
          });
        })
      ) {
        this.newSetting.loading = true;
        this.sendKey()
          .then(() => {
            this.showSnackbar({ message: "Setting created successfully" });

            this.newSetting.data = {
              ...JSON.parse(JSON.stringify(defaultData)),
            };
            this.newSetting.visible = false;
          })
          .finally(() => {
            this.newSetting.loading = false;
          });
      } else {
        this.showSnackbarError({ message: "All fields are required" });
      }
    },
    sendKey(key, data) {
      let reqestKey = key;
      let reqestData = data;

      if (reqestKey == undefined || reqestData == undefined) {
        reqestKey = this.newSetting.data.key;
        reqestData = this.newSetting.data.data;
      }

      return new Promise((resolve, reject) => {
        api.settings
          .addKey(reqestKey, reqestData)
          .then((res) => {
            if (res.key != reqestKey) throw res;
            this.$store.dispatch("settings/fetch");
            resolve(res);
          })
          .catch((err) => {
            this.showSnackbarError({ message: err.response.data.message });
            reject(err);
          });
      });
    },
    startEditValue(data) {
      this.setEditSettingValue(data.key, data.value);
      this.isModalEditOpen = true;
    },
    setEditSettingValue(key, data) {
      this.edit.key = key;
      this.edit.data = data;
    },
    stopEdit() {
      this.edit.key = "";
      this.edit.data = {};
      this.isModalEditOpen = false;
    },
    changeSettingValue(data) {
      this.edit.data = JSON.stringify(data);
    },
    saveEditDescription() {
      const data = JSON.parse(JSON.stringify(this.edit.data));
      const key = this.edit.data.key;
      delete data.key;
      this.saveSettingVallue(key, data);
    },
    saveSettingVallue(key, data) {
      this.sendKey(key, data)
        .then(() => {
          this.$store.dispatch("settings/fetch");
          this.showSnackbar({ message: "Setting created successfully" });
          this.stopEdit();
        })
        .catch((err) => {
          this.showSnackbarError({ message: err.response.data.message });
        });
    },
    saveModalEdit() {
      const data = JSON.parse(JSON.stringify(this.edit.data));
      const key = this.edit.key;
      const el = { ...this.settings.find((d) => d.key === key) };
      el.value = data;
      delete el.key;
      this.saveSettingVallue(key, el);
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", { type: "settings/fetch" });
  },
  watch: {
    settings() {
      this.fetchError = "";
    },
  },
};
</script>

<style scoped lang="scss">
.edit-btn {
  opacity: 0.4;
  margin-right: 4px;

  &:hover {
    opacity: 1;
  }
}

.edit-modal {
  background-color: rgba(12, 12, 60, 0.9);
  overflow: hidden;
  border: 1px solid white;
  border-radius: 15px;
  padding: 30px;
}
.buttons{
  margin-top: 15px;
  display: flex;
  justify-content: right;
  button{
    margin: 0 10px
  }
}
</style>
