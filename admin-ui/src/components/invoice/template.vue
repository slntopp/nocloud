<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <template-json-editor-new
      :value="invoice"
      title="Template JSON"
      @save="editInvoice"
    />
  </v-card>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditorNew from "@/components/TemplateJsonEditorNew.vue";
import { Invoice } from "nocloud-proto/proto/es/billing/billing_pb";

export default {
  name: "invoice-template",
  components: { TemplateJsonEditorNew },
  mixins: [snackbar],
  props: {
    invoice: {
      type: Object,
      required: true,
    },
  },
  methods: {
    async editInvoice(parsedValue) {
      try {
        await this.$store.getters["invoices/invoicesClient"].updateInvoice(
          Invoice.fromJson({ ...parsedValue, uuid: this.invoice.uuid })
        );
        this.showSnackbarSuccess({
          message: "Invoice edited successfully",
        });
        this.$router.go();
      } catch (err) {
        this.showSnackbarError({ message: err });
      }
    },
  },
};
</script>

<style scoped lang="scss">
</style>
