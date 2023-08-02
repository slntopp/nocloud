const useAccountConverter = () => {
  const toAccountPrice = (price, rate) => {
    return rate ? (price / rate).toFixed(2) : 0;
  };
  const fromAccountPrice = (price, rate) => {
    return rate ? (price * rate).toFixed(2) : 0;
  };

  return { toAccountPrice, fromAccountPrice };
};

export default useAccountConverter;
