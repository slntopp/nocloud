<template>
  <div class="pa-4">
    <div class="d-flex mb-5" v-if="!isEdit">
      <h1 class="page__title">Create Account Group</h1>
    </div>
    <div class="d-flex mb-5" v-else>
      <h1 class="page__title">Edit Account Group</h1>
    </div>
    <v-form v-model="isValid" ref="accountGroupForm">
      <!-- Basic Info -->
      <v-row>
        <v-col cols="2" class="align-center d-flex">
          <v-subheader>Title</v-subheader>
        </v-col>
        <v-col cols="6" class="align-center d-flex">
          <v-text-field
            label="Title"
            v-model="newGroup.title"
            :rules="[rules.required]"
          />
        </v-col>
        <v-col cols="1"></v-col>
        <v-col cols="2" class="align-center d-flex">
          <color-picker label="Color" v-model="newGroup.color" />
        </v-col>
        <v-col cols="1">
          <div
            :style="{ backgroundColor: newGroup.color }"
            class="color-box"
          ></div>
        </v-col>
      </v-row>

      <v-divider class="my-4" />

      <!-- Invoice Order Settings -->
      <div class="d-flex align-center mb-5">
        <v-expansion-panels>
          <v-expansion-panel style="background: var(--v-background-light-base)">
            <v-expansion-panel-header
              style="padding: 5px 20px"
              disable-icon-rotate
            >
              Invoice Order Settings

              <v-switch
                class="ml-3"
                v-model="newGroup.has_own_invoice_order"
                @click.stop
              />
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Template</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-text-field
                    dense
                    hide-details
                    label="Template"
                    v-model="newGroup.invoice_order_settings.template"
                    :disabled="!newGroup.has_own_invoice_order"
                  />
                </v-col>
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>New Template</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-text-field
                    dense
                    hide-details
                    label="New Template"
                    v-model="newGroup.invoice_order_settings.new_template"
                    :disabled="!newGroup.has_own_invoice_order"
                  />
                </v-col>
              </v-row>

              <v-row class="px-5 pt-0">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Start With Number</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-text-field
                    label="Start With Number"
                    v-model.number="
                      newGroup.invoice_order_settings.start_with_number
                    "
                    type="number"
                    min="0"
                    dense
                    hide-details
                    :disabled="!newGroup.has_own_invoice_order"
                  />
                </v-col>
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Reset Counter Mode</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-select
                    label="Reset Counter Mode"
                    v-model="newGroup.invoice_order_settings.reset_counter_mode"
                    :items="resetCounterModes"
                    dense
                    hide-details
                    :disabled="!newGroup.has_own_invoice_order"
                  />
                </v-col>
              </v-row>
            </v-expansion-panel-content> </v-expansion-panel
        ></v-expansion-panels>
      </div>

      <v-divider class="my-4" />

      <!-- Invoice Parameters Custom -->
      <div class="d-flex align-center mb-5">
        <v-expansion-panels>
          <v-expansion-panel style="background: var(--v-background-light-base)">
            <v-expansion-panel-header
              style="padding: 5px 20px"
              disable-icon-rotate
            >
              Invoice Parameters Custom

              <v-switch
                class="ml-3"
                v-model="newGroup.has_own_invoice_base"
                @click.stop
              />
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Invoice From Name</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-textarea
                    label="Invoice From Name"
                    v-model="
                      newGroup.invoice_parameters_custom.invoiceFromFields.name
                    "
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    rows="1"
                    hide-details
                  />
                </v-col>
              </v-row>

              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Invoice From Company domain</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-textarea
                    rows="1"
                    label="Invoice From Company domain"
                    v-model="
                      newGroup.invoice_parameters_custom.invoiceFromFields
                        .companyDomain
                    "
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    hide-details
                  />
                </v-col>
              </v-row>

              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Invoice From Address</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-textarea
                    rows="1"
                    label="Invoice From Address"
                    v-model="
                      newGroup.invoice_parameters_custom.invoiceFromFields
                        .address
                    "
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    hide-details
                  />
                </v-col>
              </v-row>

              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Invoice From City</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-textarea
                    rows="1"
                    label="Invoice From City"
                    v-model="
                      newGroup.invoice_parameters_custom.invoiceFromFields.city
                    "
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    hide-details
                  />
                </v-col>
              </v-row>

              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Invoice From Postal Code</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-textarea
                    rows="1"
                    label="Invoice From Postal Code"
                    v-model="
                      newGroup.invoice_parameters_custom.invoiceFromFields
                        .postalCode
                    "
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    hide-details
                  />
                </v-col>
              </v-row>

              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Invoice From Country</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-textarea
                    rows="1"
                    label="Invoice From  Country"
                    v-model="
                      newGroup.invoice_parameters_custom.invoiceFromFields
                        .country
                    "
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    hide-details
                  />
                </v-col>
              </v-row>

              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Invoice From Tax ID</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-textarea
                    rows="1"
                    label="Invoice From  Tax ID"
                    v-model="
                      newGroup.invoice_parameters_custom.invoiceFromFields.taxId
                    "
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    hide-details
                  />
                </v-col>
              </v-row>

              <v-row class="px-5">
                <v-col cols="3" class="align-center d-flex">
                  <v-subheader>Logo URL</v-subheader>
                </v-col>
                <v-col cols="9" class="align-center d-flex">
                  <v-text-field
                    label="Logo URL"
                    v-model="newGroup.invoice_parameters_custom.logo_url"
                    :disabled="!newGroup.has_own_invoice_base"
                    dense
                    hide-details
                  />
                </v-col>
              </v-row>
            </v-expansion-panel-content> </v-expansion-panel
        ></v-expansion-panels>
      </div>

      <v-divider class="my-4" />

      <!-- Actions -->
      <v-row class="mt-3" justify="end">
        <v-col cols="3">
          <v-btn
            :loading="isSaveLoading"
            @click="saveAccountGroup"
            class="mr-2"
          >
            Save
          </v-btn>
          <v-btn @click="$router.back()" text> Cancel </v-btn>
        </v-col>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import { onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { useRouter } from "vue-router/composables";
import ColorPicker from "@/components/ui/colorPicker.vue";

const props = defineProps({
  accountGroup: {},
  isEdit: { type: Boolean, default: false },
});
const { accountGroup, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const newGroup = ref({
  title: "",
  color: "#ffffff",
  uuid: "",
  has_own_invoice_order: false,
  invoice_order_settings: {
    template: "",
    new_template: "",
    start_with_number: 0,
    reset_counter_mode: "",
  },
  has_own_invoice_base: false,
  invoice_parameters_custom: {
    invoiceFromFields: {},
    logo_url: "",
  },
});

const isSaveLoading = ref(false);
const isValid = ref(false);
const accountGroupForm = ref(null);

onMounted(() => {
  store.dispatch("accountGroups/fetch");
});

const resetCounterModes = ref(["NONE", "YEARLY", "MONTHLY", "DAILY"]);

const rules = ref({
  required: (v) => !!v || "This field is required!",
});

onMounted(() => {
  if (isEdit.value && accountGroup.value) {
    setAccountGroup(accountGroup.value);
  }
});

const setAccountGroup = (val) => {
  newGroup.value = {
    uuid: val.uuid || "",
    title: val.title || "",
    color: val.color || "",
    has_own_invoice_order: val.hasOwnInvoiceOrder || false,
    invoice_order_settings: {
      template: val.invoiceOrderSettings?.template || "",
      new_template: val.invoiceOrderSettings?.newTemplate || "",
      start_with_number: val.invoiceOrderSettings?.startWithNumber || 0,
      reset_counter_mode: val.invoiceOrderSettings?.resetCounterMode || "",
    },
    has_own_invoice_base: val.hasOwnInvoiceBase || false,
    invoice_parameters_custom: {
      invoiceFromFields: val.invoiceParametersCustom?.invoiceFromFields || {},
      logo_url: val.invoiceParametersCustom?.logoUrl || "",
    },
  };
};

const saveAccountGroup = async () => {
  if (!(await accountGroupForm.value.validate())) {
    return;
  }
  isSaveLoading.value = true;

  try {
    const payload = {
      title: newGroup.value.title,
      color: newGroup.value.color,
      has_own_invoice_order: newGroup.value.has_own_invoice_order,
      invoice_order_settings: newGroup.value.has_own_invoice_order
        ? newGroup.value.invoice_order_settings
        : null,
      has_own_invoice_base: newGroup.value.has_own_invoice_base,
      invoice_parameters_custom: newGroup.value.has_own_invoice_base
        ? newGroup.value.invoice_parameters_custom
        : null,
    };

    if (!isEdit.value) {
      await store.dispatch("accountGroups/create", payload);
      store.commit("snackbar/showSnackbarSuccess", {
        message: "Account Group created successfully",
      });
      router.push({ name: "AccountGroups" });
    } else {
      payload.uuid = newGroup.value.uuid;
      await store.dispatch("accountGroups/update", payload);
      store.commit("snackbar/showSnackbarSuccess", {
        message: "Account Group updated successfully",
      });
    }
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSaveLoading.value = false;
  }
};

watch(accountGroup, (val) => {
  if (val) {
    setAccountGroup(val);
  }
});
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}

.color-box {
  width: 24px;
  height: 24px;
  border: 1px solid #ccc;
  border-radius: 4px;
  margin: 20px 0px 0px 0px;
}
</style>
