import {computed, onMounted, ref} from "vue";
import api from "@/api";
import {useStore} from "@/store";

const useRate=()=>{
    const rate = ref(0);

    const store=useStore()

    const defaultCurrency = computed(() => {
        return store.getters["currencies/default"];
    });

    onMounted(()=>{
        api
            .get(`/billing/currencies/rates/PLN/${defaultCurrency.value}`)
            .then((res) => {
                rate.value = res.rate;
            })
            .catch(() =>
                api.get(`/billing/currencies/rates/${defaultCurrency.value}/PLN`)
            )
            .then((res) => {
                if (res) rate.value = 1 / res.rate;
            })
            .catch((err) => console.error(err));
    })

    return rate
}

export default useRate