import useRate from "@/hooks/useRate";
import { ref } from "vue";

const usePlnRate = () => {
  const currency = ref("PLN");
  const { rate } = useRate(currency);

  return rate;
};

export default usePlnRate;
