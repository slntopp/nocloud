import { onUnmounted, onMounted } from "vue";
import { useStore } from "@/store";

const useSearch = ({ name, defaultLayout, noSearch }) => {
  const store = useStore();

  onMounted(() => {
    if (!noSearch) {
      store.commit("appSearch/setSearchName", name);
      store.commit("appSearch/setDefaultLayout", defaultLayout);
    }
  });

  onUnmounted(() => {
    store.commit("appSearch/setSearchName", "");
    store.commit("appSearch/setFields", []);
    store.commit("appSearch/setDefaultLayout", null);
  });
};

export default useSearch;
