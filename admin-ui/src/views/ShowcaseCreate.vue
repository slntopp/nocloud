<template>
  <div color="background-light" :loading="isLoading" class="pa-10 h-100 w-100">
    <h1 class="page__title" v-if="!isEdit">Create showcase</h1>
    <v-form ref="showcaseForm" align="center">
      <v-row>
        <v-col cols="6">
          <v-text-field
            :rules="[requiredRule]"
            v-model="showcase.title"
            label="title"
          />
        </v-col>
        <v-col cols="6">
          <icons-autocomplete
            label="preview icon"
            :value="showcase.icon"
            @input:value="showcase.icon = $event"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-autocomplete
            item-text="title"
            v-model="showcase.servicesProviders"
            item-value="uuid"
            multiple
            label="Service providers"
            :items="serviceProviders"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            item-text="title"
            item-value="uuid"
            multiple
            label="Plans"
            v-model="showcase.plans"
            :items="plans"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-autocomplete
            item-value="id"
            item-text="title"
            multiple
            label="Locations"
            return-object
            :items="locations"
            v-model="showcase.locations"
          >
            <template v-slot:item="{ item }">
              <span>{{ item.title }}({{ item.sp }})</span>
            </template>
          </v-autocomplete>
        </v-col>
      </v-row>
      <v-row justify="end">
        <v-btn @click="save" :loading="isSaveLoading">{{
          isEdit ? "Save" : "Create"
        }}</v-btn>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import { onMounted, ref, computed, toRefs, watch } from "vue";
import IconsAutocomplete from "@/components/ui/iconsAutocomplete.vue";
import { useStore } from "@/store";
import api from "@/api";
import { useRouter } from "vue-router/composables";

const props = defineProps({"real-showcase":{}, isEdit:{type:Boolean,default:false}});
// const emits=defineEmits(['input'])

const { realShowcase, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const showcase = ref({
  title: "",
  icon: "",
  servicesProviders: [],
  plans: [],
  promo: {},
  locations: [],
});
const isLoading = ref(false);
const isSaveLoading = ref(false);

const requiredRule = ref((val) => !!val || "Required field");
const serviceProviders = computed(() => store.getters["servicesProviders/all"]);
const plans = computed(() => store.getters["plans/all"]);
const locations = computed(() => {
  const sps = serviceProviders.value.filter((sp) =>
    showcase.value.servicesProviders?.includes(sp.uuid)
  );
  const locations = [];
  sps.forEach((sp) => {
    locations.push(...sp.locations.map((l) => ({ ...l, sp: sp.title })));
  });
  return locations;
});

onMounted(async () => {
  try {
    isLoading.value = true;
    await Promise.all([
      store.dispatch("servicesProviders/fetch"),
      store.dispatch("plans/fetch"),
    ]);
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch info",
    });
  } finally {
    isLoading.value = false;
  }
});

const save = async () => {
  try {
    isSaveLoading.value = true;
    if (isEdit.value) {
      const data = await api.showcases.update({
        ...showcase.value,
        locations: showcase.value.locations.map((l) => ({
          ...l,
          sp: undefined,
        })),
      });
      console.log(data);
    } else {
      await api.showcases.create({
        ...showcase.value,
        locations: showcase.value.locations.map((l) => ({
          ...l,
          sp: undefined,
        })),
      });
    }
    store.commit("snackbar/showSnackbarSuccess", {
      message: `Showcase successfully ${isEdit.value ? "saved" : "created"}`,
    });
    router.push({ name: "Showcases" });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during save showcase",
    });
  } finally {
    isSaveLoading.value = false;
  }
};

watch(realShowcase, () => {
  showcase.value = realShowcase.value;
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
</style>
