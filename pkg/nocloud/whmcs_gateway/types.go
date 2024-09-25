package whmcs_gateway

import pb "github.com/slntopp/nocloud-proto/billing"

func statusToWhmcs(status pb.BillingStatus) string {
	switch status {
	case pb.BillingStatus_DRAFT:
		return "Draft"
	case pb.BillingStatus_UNPAID:
		return "Unpaid"
	case pb.BillingStatus_PAID:
		return "Paid"
	case pb.BillingStatus_CANCELED:
		return "Cancelled"
	case pb.BillingStatus_RETURNED:
		return "Returned"
	case pb.BillingStatus_TERMINATED:
		return "Terminated"
	default:
		return "Unknown"
	}
}

type CreateInvoiceQuery struct {
	Action          string `url:"action"`
	Username        string `url:"username"`
	Password        string `url:"password"` // md5 hash
	UserId          string `url:"userid"`
	Status          string `url:"status"`
	SendInvoice     string `url:"sendinvoice"`
	PaymentMethod   string `url:"paymentmethod"`
	TaxRate         string `url:"taxrate"`
	Date            string `url:"date"`
	DueDate         string `url:"duedate"`
	AutoApplyCredit string `url:"autoapplycredit"`
	ResponseType    string `url:"responsetype"`
}

type InvoiceResponse struct {
	Status    string `json:"status"`
	InvoiceId int    `json:"invoiceid"`
	Result    string `json:"result"`
}
