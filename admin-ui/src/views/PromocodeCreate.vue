<template>
  <div class="pa-4">
    <h1 class="page__title" v-if="!isEdit">Create promocode</h1>

    <v-form v-model="isValid" ref="promocodeForm">
      <v-row>
        <v-col cols="3">
          <v-text-field
            v-model="newPromocode.title"
            :rules="[requiredRule]"
            label="Title"
          />
        </v-col>

        <v-col cols="2">
          <v-text-field
            v-model="newPromocode.code"
            :rules="[requiredRule, codeRule]"
            label="Code"
          >
            <template v-slot:append>
              <v-btn small @click="newPromocode.code = generateCode()" icon
                ><v-icon small>mdi-pound</v-icon></v-btn
              >
            </template>
          </v-text-field>
        </v-col>

        <v-col cols="2">
          <date-picker
            :min="formatSecondsToDateString(Date.now() / 1000)"
            label="Due date"
            v-model="newPromocode.dueDate"
          />
        </v-col>

        <v-col cols="2" v-if="isEdit">
          <date-picker
            disabled
            label="Created"
            :value="formatSecondsToDateString(promocode.created)"
          />
        </v-col>

        <v-col cols="3" v-if="isEdit">
          <div class="d-flex">
            <div class="mt-5">
              <span>Condition:</span>
              <promocode-condition-chip class="ml-2" :item="promocode" />
            </div>
            <div class="mt-5 ml-3">
              <span>Status:</span>
              <promocode-status-chip class="ml-2" :item="promocode" />
            </div>
          </div>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="2">
          <v-select
            v-model.number="newPromocode.type"
            :rules="[requiredRule]"
            label="Promocode type"
            :items="promocodeTypes"
            item-key="value"
            item-text="text"
          />
        </v-col>
        <v-col cols="1">
          <v-text-field
            v-model.number="newPromocode.value"
            :rules="[requiredNumberRule, promocodeValueRule]"
            type="number"
            label="Amount"
            :max="isDiscountPercent && 100"
            :suffix="isDiscountPercent ? '%' : defaultCurrency.code"
          />
        </v-col>

        <v-col cols="1">
          <v-text-field
            v-model.number="newPromocode.limit"
            :rules="[requiredNumberRule]"
            type="number"
            label="Limit"
          />
        </v-col>

        <v-col cols="1">
          <v-text-field
            v-model.number="newPromocode.usesPerUser"
            :rules="[requiredNumberRule]"
            type="number"
            label="Uses per user"
          />
        </v-col>

        <v-col cols="1">
          <v-switch v-model="isOneTime" label="One time" />
        </v-col>

        <v-col cols="2" v-if="!isOneTime">
          <div class="d-flex align-center">
            <span>Period:</span>
            <date-field
              class="mt-3 ml-2"
              label="Period"
              placeholder="Period"
              :period="newPromocode.activeTime"
              @changeDate="newPromocode.activeTime = $event"
            />
          </div>
        </v-col>
      </v-row>

      <v-textarea
        outlined
        no-resize
        label="Description"
        v-model="newPromocode.description"
      ></v-textarea>

      <v-tabs v-model="currentTab" background-color="background-light">
        <v-tab key="Plans"> Plans </v-tab>

        <v-tab key="Showcases"> Showcases </v-tab>
      </v-tabs>

      <v-tabs-items
        style="background: var(--v-background-light-base)"
        v-model="currentTab"
      >
        <v-tab-item key="Plans">
          <div>
            <div class="d-flex justify-end">
              <v-btn
                class="mr-2"
                @click="deletePlans"
                :disabled="!selectedBindedPlans.length"
                >Delete</v-btn
              >

              <v-dialog v-model="isAddPlansOpen" width="50%">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn v-bind="attrs" v-on="on" class="mr-2">Add</v-btn>
                </template>

                <v-card color="background-light" class="pa-4">
                  <v-text-field
                    label="Search..."
                    v-model="availablePlansSearchParam"
                  />

                  <plans-table
                    :custom-headers="customPlansHeaders"
                    no-search
                    show-select
                    :custom-params="{
                      showDeleted: false,
                      anonymously: false,
                      excludeUuids: [...newPromocode.plans],
                      filters: {
                        search_param: availablePlansSearchParam,
                      },
                    }"
                    table-name="plans-promocode-add-table"
                    @fetch:plans="fetchAvailablePlans"
                    :plans="availablePlans"
                    :isLoading="isAvailablePlansLoading"
                    :total="availablePlansTotal"
                    v-model="selectedAvailablePlans"
                  />

                  <div class="d-flex justify-end">
                    <v-btn
                      class="mr-2"
                      @click="addPlans"
                      :disabled="!selectedAvailablePlans.length"
                      >Add</v-btn
                    >
                  </div>
                </v-card>
              </v-dialog>
            </div>
            <nocloud-table
              :items="bindedPlans"
              :loading="isBindedPlansLoading"
              table-name="plans-promocode-table"
              v-model="selectedBindedPlans"
              :headers="customPlansHeaders"
            />
          </div>
        </v-tab-item>

        <v-tab-item key="Showcases">
          <div>
            <div class="d-flex justify-end">
              <v-btn
                class="mr-2"
                @click="deleteShowcases"
                :disabled="!selectedBindedShowcases.length"
                >Delete</v-btn
              >

              <v-dialog v-model="isAddShowcasesOpen" width="50%">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn v-bind="attrs" v-on="on" class="mr-2">Add</v-btn>
                </template>

                <v-card color="background-light" class="pa-4">
                  <v-text-field
                    label="Search..."
                    v-model="availableShowcasesSearchParam"
                  />

                  <showcases-table
                    table-name="showcases-promocode-add-table"
                    :items="availableShowcases"
                    :loading="isShowcasesLoading"
                    v-model="selectedAvailableShowcases"
                  />

                  <div class="d-flex justify-end">
                    <v-btn
                      class="mr-2"
                      @click="addShowcases"
                      :disabled="!selectedAvailableShowcases.length"
                      >Add</v-btn
                    >
                  </div>
                </v-card>
              </v-dialog>
            </div>
            <showcases-table
              :items="bindedShowcases"
              :loading="isShowcasesLoading"
              table-name="showcases-promocode-table"
              v-model="selectedBindedShowcases"
            />
          </div>
        </v-tab-item>
      </v-tabs-items>

      <v-row justify="space-between" class="mt-4 mb-4">
        <div class="mt-2 ml-1">
          <v-btn
            class="mx-3"
            color="background-light"
            :loading="isSaveLoading"
            @click="savePromocode"
          >
            {{ isEdit ? "Save" : "Create" }}
          </v-btn>
        </div>

        <promocode-status-btns
          v-if="isEdit"
          @click="() => emit('refresh')"
          :items="[promocode]"
        />
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import dateField from "@/components/date.vue";
import nocloudTable from "@/components/table.vue";
import showcasesTable from "@/components/showcases_table.vue";
import datePicker from "@/components/ui/datePicker.vue";
import { formatSecondsToDateString } from "../functions";
import useCurrency from "@/hooks/useCurrency";
import { useStore } from "@/store";
import plansTable from "@/components/plansTable.vue";
import { ListRequest } from "nocloud-proto/proto/es/billing/billing_pb";
import { Promocode } from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";
import { useRouter } from "vue-router/composables";
import promocodeConditionChip from "@/components/promocode/ui/promocodeConditionChip.vue";
import promocodeStatusChip from "@/components/promocode/ui/promocodeStatusChip.vue";
import promocodeStatusBtns from "@/components/promocode/ui/promocodeStatusBtns.vue";

const props = defineProps({
  isEdit: { type: Boolean, default: false },
  promocode: {},
});
const { isEdit, promocode } = toRefs(props);

const emit = defineEmits("refresh");

const { defaultCurrency } = useCurrency();
const store = useStore();
const router = useRouter();

const isValid = ref(false);
const requiredRule = ref((val) => !!val || "Field required");
const requiredNumberRule = ref((val) => +val >= 0 || "Field required");
const codeRule = ref((val) => val.length === 6 || "Code must be 6 symbols");
const promocodeValueRule = ref((val) =>
  isDiscountPercent.value
    ? (+val >= 0 && +val <= 100) || "Min 0% Max 100%"
    : true
);

const promocodeTypes = [{ text: "Discount percent", value: "percent" }];

const promocodeForm = ref(null);
const newPromocode = ref({
  title: "",
  code: "",
  limit: 0,
  usesPerUser: 1,
  dueDate: null,
  activeTime: 2592000,
  description: "",
  type: "percent",
  value: 20,
  plans: [],
  showcases: [],
});
const isOneTime = ref(true);
const isSaveLoading = ref(false);

const currentTab = ref("Showcases");

const customPlansHeaders = ref([
  { text: "Title ", value: "title" },
  { text: "UUID ", value: "uuid" },
  { text: "Type ", value: "type" },
]);

const isBindedPlansLoading = ref(false);
const bindedPlans = ref([]);
const selectedBindedPlans = ref([]);

const selectedBindedShowcases = ref([]);
const isAddShowcasesOpen = ref(false);
const selectedAvailableShowcases = ref([]);
const availableShowcasesSearchParam = ref("");

const isAddPlansOpen = ref(false);
const availablePlansSearchParam = ref("");
const selectedAvailablePlans = ref([]);
const isAvailablePlansLoading = ref(false);
const availablePlans = ref([]);
const availablePlansTotal = ref(0);

onMounted(() => {
  store.dispatch("showcases/fetch", { anonymously: false });

  setPromocde();
});

const isDiscountPercent = computed(() => newPromocode.value.type === "percent");
const promocodePlans = computed(() => newPromocode.value.plans);

const isShowcasesLoading = computed(() => store.getters["showcases/isLoading"]);
const bindedShowcases = computed(() =>
  store.getters["showcases/all"].filter(({ uuid }) =>
    newPromocode.value.showcases.includes(uuid)
  )
);
const availableShowcases = computed(() =>
  store.getters["showcases/all"].filter(
    ({ uuid, title }) =>
      !newPromocode.value.showcases.includes(uuid) &&
      title
        .toLowerCase()
        .includes(availableShowcasesSearchParam.value.toLowerCase())
  )
);

const fetchBindedPlans = async () => {
  if (!promocodePlans.value.length) {
    bindedPlans.value = [];
    return;
  }

  try {
    isBindedPlansLoading.value = true;

    const response = await store.getters["plans/plansClient"].listPlans(
      ListRequest.fromJson({
        filters: { uuid: [...promocodePlans.value] },
        page: "1",
        limit: "100",
        sort: "ASC",
        field: "title",
      })
    );
    bindedPlans.value = response.toJson().pool;
  } finally {
    isBindedPlansLoading.value = false;
  }
};

const fetchAvailablePlans = async (options) => {
  isAvailablePlansLoading.value = true;

  try {
    const response = await store.getters["plans/plansClient"].listPlans(
      ListRequest.fromJson(options)
    );

    const data = response.toJson();
    availablePlans.value = data.pool;
    availablePlansTotal.value = +data.total;
  } finally {
    isAvailablePlansLoading.value = false;
  }
};

const savePromocode = async () => {
  if (!(await promocodeForm.value.validate())) {
    return;
  }
  isSaveLoading.value = true;

  try {
    const data = {
      title: newPromocode.value.title,
      code: newPromocode.value.code,
      description: newPromocode.value.description || "",
      dueDate: new Date(newPromocode.value.dueDate).getTime() / 1000,
      limit: newPromocode.value.limit,
      usesPerUser: newPromocode.value.usesPerUser,
      promoItems: [],
    };

    if (!isOneTime.value) {
      data.oneTime = false;
      data.activeTime = newPromocode.value.activeTime;
    } else {
      data.oneTime = true;
      data.activeTime = 0;
    }

    const promoSchema = {};

    if (isDiscountPercent.value) {
      promoSchema.discountPercent = newPromocode.value.value / 100;
    }

    newPromocode.value.plans.forEach((uuid) => {
      const promoItem = {
        schema: promoSchema,
        planPromo: { billingPlan: uuid },
      };

      data.promoItems.push(promoItem);
    });

    newPromocode.value.showcases.forEach((uuid) => {
      const promoItem = {
        schema: promoSchema,
        showcasePromo: { showcase: uuid },
      };

      data.promoItems.push(promoItem);
    });

    if (isEdit.value) {
      data.uuid = promocode.value.uuid;
      data.status = promocode.value.status || 0;
      data.condition = promocode.value.condition || 0;
      data.uses = promocode.value.uses || [];
      data.created = promocode.value.created;
      data.meta = promocode.value.meta || {};
    }

    let message;
    if (isEdit.value) {
      await store.getters["promocodes/promocodesClient"].update(
        Promocode.fromJson(data)
      );
      message = "Promocode saved successfully";
    } else {
      await store.getters["promocodes/promocodesClient"].create(
        Promocode.fromJson(data)
      );
      message = "Promocode created successfully";
    }

    store.commit("snackbar/showSnackbarSuccess", {
      message,
    });

    if (isEdit.value) {
      emit("refresh");
    } else {
      router.push({ name: "Promocodes" });
    }
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSaveLoading.value = false;
  }
};

const setPromocde = () => {
  if (isEdit.value) {
    newPromocode.value.title = promocode.value.title;
    newPromocode.value.code = promocode.value.code;
    newPromocode.value.description = promocode.value.description;
    newPromocode.value.limit = +(promocode.value.limit || 0);
    newPromocode.value.usesPerUser = +(promocode.value.usesPerUser || 0);

    newPromocode.value.dueDate = promocode.value.dueDate
      ? formatSecondsToDateString(promocode.value.dueDate)
      : null;

    if (promocode.value.oneTime) {
      newPromocode.value.activeTime = 0;
      isOneTime.value = true;
    } else {
      newPromocode.value.activeTime = +(promocode.value.activeTime || 0);
      isOneTime.value = false;
    }

    const showcases = [];
    const plans = [];

    if (promocode.value.promoItems?.length) {
      const promoSchema = promocode.value.promoItems[0].schema;

      if (promoSchema.discountPercent) {
        newPromocode.value.type = "percent";
        newPromocode.value.value = +(promoSchema.discountPercent || 0) * 100;
      }

      promocode.value.promoItems.forEach((item) => {
        if (item.showcasePromo) {
          showcases.push(item.showcasePromo.showcase);
        } else if (item.planPromo) {
          plans.push(item.planPromo.billingPlan);
        }
      });
    }

    newPromocode.value.showcases = showcases;
    newPromocode.value.plans = plans;
  }
};

const addPlans = () => {
  newPromocode.value.plans.push(
    ...selectedAvailablePlans.value.map((p) => p.uuid)
  );
  selectedAvailablePlans.value = [];
  isAddPlansOpen.value = false;
  availablePlansSearchParam.value = "";
};

const deletePlans = () => {
  const uuids = selectedBindedPlans.value.map((p) => p.uuid);
  newPromocode.value.plans = newPromocode.value.plans.filter(
    (uuid) => !uuids.includes(uuid)
  );
  selectedBindedPlans.value = [];
};

const addShowcases = () => {
  newPromocode.value.showcases.push(
    ...selectedAvailableShowcases.value.map((p) => p.uuid)
  );
  selectedAvailableShowcases.value = [];
  isAddShowcasesOpen.value = false;
  availableShowcasesSearchParam.value = "";
};

const deleteShowcases = () => {
  const uuids = selectedBindedShowcases.value.map((p) => p.uuid);
  newPromocode.value.showcases = newPromocode.value.showcases.filter(
    (uuid) => !uuids.includes(uuid)
  );
  selectedBindedShowcases.value = [];
};

function generateCode() {
  const gen = () =>
    Math.random().toString(16).slice(2).slice(0, 6).toUpperCase();

  let val = gen();

  while (
    val.split("").reduce((c, e) => (Number.isInteger(+e) ? c + 1 : c), 0) > 3
  ) {
    val = gen();
  }

  return val;
}

watch(
  promocodePlans,
  () => {
    fetchBindedPlans();
  },
  { deep: true }
);
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

.promocode__container {
  display: flex;
  flex-wrap: wrap;

  .item {
    margin-left: 10px;
    margin-left: 10px;
    width: 200px;
    &.date {
      width: 140px;
    }
  }
}
</style>
