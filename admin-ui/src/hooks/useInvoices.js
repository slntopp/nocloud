import { BillingStatus } from "nocloud-proto/proto/es/billing/billing_pb";

const useInvoices = () => {
  const getInvoiceStatusColor = (status) => {
    switch (BillingStatus[status]) {
      case BillingStatus.CANCELED: {
        return "warning";
      }
      case BillingStatus.RETURNED: {
        return "blue";
      }
      case BillingStatus.DRAFT: {
        return "brown darked";
      }
      case BillingStatus.PAID: {
        return "green";
      }
      case BillingStatus.UNPAID: {
        return "grey-lighten-2";
      }
      case BillingStatus.BILLING_STATUS_UNKNOWN:
      case BillingStatus.TERMINATED:
      default: {
        return "red";
      }
    }
  };

  const getTotalColor = (item) => {
    if (
      BillingStatus[item.status] === BillingStatus.UNPAID &&
      item.deadline < Date.now() / 1000
    ) {
      return "red";
    }

    return "grey-lighten-2";
  };

  return { getInvoiceStatusColor, getTotalColor };
};

export default useInvoices;
