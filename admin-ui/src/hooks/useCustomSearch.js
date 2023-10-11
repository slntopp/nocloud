import { onUnmounted } from "vue";
import { useStore } from "@/store";

const useCustomSearch = () => {
  const store = useStore();

  onUnmounted(() => {
    store.commit("appSearch/resetSearchParams");
  });
};

export default useCustomSearch;
