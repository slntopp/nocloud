<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    Template
    <span
      class="template__display-trigger"
      @click="
        () => (ObjectDisplay = ObjectDisplay === 'YAML' ? 'JSON' : 'YAML')
      "
    >
      {{ ObjectDisplay }}
    </span>
    <v-switch
      style="display: inline-flex"
      v-model="ObjectDisplay"
      true-value="JSON"
      false-value="YAML"
    />
    :
    <v-spacer />
    <template v-if="editing">
      <v-btn
        class="mr-2"
        color="success"
        :loading="isLoading"
        :disabled="!isValid"
        @click="editInvoice"
      >
        Save
      </v-btn>
      <v-btn @click="cancel"> Cancel </v-btn>
      <json-textarea
        class="mt-4"
        v-if="ObjectDisplay === 'JSON'"
        :json="invoice"
        @getTree="changeTree"
      />
      <yaml-editor class="mt-4" v-else :json="invoice" @getTree="changeTree" />
    </template>
    <template v-else>
      <v-btn @click="editing = true">Edit</v-btn>
      <pre v-if="ObjectDisplay === 'YAML'" v-html="templateObjectYAML" />
      <pre v-else-if="ObjectDisplay === 'JSON'" v-html="templateObjectJSON" />
    </template>
  </v-card>
</template>

<script>
import yaml from "yaml";
import snackbar from "@/mixins/snackbar.js";
import JsonTextarea from "@/components/JsonTextarea.vue";
import YamlEditor from "@/components/YamlEditor.vue";
import { Invoice } from "nocloud-proto/proto/es/billing/billing_pb";

export default {
  name: "invoice-template",
  components: { JsonTextarea, YamlEditor },
  mixins: [snackbar],
  props: {
    invoice: {
      type: Object,
      required: true,
    },
  },
  data: () => ({
    ObjectDisplay: "YAML",
    tree: "",
    isValid: false,
    isLoading: false,
    editing: false,
  }),
  methods: {
    changeTree(value) {
      try {
        if (this.ObjectDisplay === "JSON") JSON.parse(value);
        else yaml.parse(value);

        this.tree = value;
        this.isValid = true;
      } catch {
        this.isValid = false;
      }
    },
    editInvoice() {
      const request =
        this.ObjectDisplay === "JSON"
          ? JSON.parse(this.tree)
          : yaml.parse(this.tree);

      this.isLoading = true;
      this.$store.getters["invoices/invoicesClient"]
        .updateInvoice(
          Invoice.fromJson({ ...request, uuid: this.invoice.uuid })
        )
        .then(() => {
          this.showSnackbarSuccess({
            message: "Invoice edited successfully",
          });

          this.$router.push({ name: "Invoices" });
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    cancel() {
      this.editing = false;
      this.isValid = false;
    },
  },
  computed: {
    templateObjectJSON() {
      let json = JSON.stringify(this.invoice, null, 2);
      json = json
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");
      return json.replace(
        /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g,
        function (match) {
          let cls = "number";
          if (/^"/.test(match)) {
            if (/:$/.test(match)) {
              cls = "key";
            } else {
              cls = "string";
            }
          } else if (/true|false/.test(match)) {
            cls = "boolean";
          } else if (/null/.test(match)) {
            cls = "null";
          }
          return '<span class="' + cls + '">' + match + "</span>";
        }
      );
    },
    templateObjectYAML() {
      const doc = new yaml.Document();
      doc.contents = this.invoice;

      return doc.toString();
    },
  },
};
</script>

<style scoped lang="scss">
.template__display-trigger {
  cursor: pointer;
  color: var(--v-primary-base);
}
</style>

<style lang="scss">
pre {
  padding: 5px;
  margin: 5px;
  background-color: var(--v-background-light-base);
  border-radius: 4px;
  white-space: pre-wrap;
}
.string {
  color: var(--v-success-base);
}
.number {
  color: var(--v-warning-base);
}
.boolean {
  color: var(--v-info-base);
}
.null {
  color: var(--v-accent-base);
}
.key {
  color: var(--v-error-base);
}
</style>
