import { ref, watch } from "vue";
import { useStore } from "@/store";

const useInstanceAddons = (instance, setValue) => {
  const store = useStore();

  const tarrifAddons = ref([]);
  const fetchedAddons = ref({});
  const isAddonsLoading = ref({});

  const setTariffAddons = () => {
    const addons = [];
    if (instance.value.product) {
      instance.value.billing_plan.addons.forEach((key) => addons.push(key));
      instance.value.billing_plan.products[
        instance.value.product
      ].addons.forEach((key) => addons.push(key));
    }
    tarrifAddons.value = addons;
  };

  watch(tarrifAddons, (value) => {
    value.forEach(async (uuid) => {
      try {
        if (!fetchedAddons.value[uuid]) {
          fetchedAddons.value[uuid] = store.getters["addons/addonsClient"].get({
            uuid,
          });
          fetchedAddons.value[uuid] = await fetchedAddons.value[uuid];
        }
      } catch {
        fetchedAddons.value[uuid] = undefined;
      } finally {
        setTimeout(() => {
          isAddonsLoading.value = Object.values(fetchedAddons.value).some(
            (acc) => acc instanceof Promise
          );
        }, 0);
      }
    });

    setValue(
      "addons",
      instance.value.addons.filter((addon) =>
        tarrifAddons.value.includes(addon)
      )
    );
  });

  const getAvailableAddons = () => {
    return tarrifAddons.value
      .map((uuid) => fetchedAddons.value[uuid])
      .filter((addon) => !(addon instanceof Promise));
  };

  return {
    tarrifAddons,
    setTariffAddons,
    getAvailableAddons,
    isAddonsLoading,
  };
};

export default useInstanceAddons;
