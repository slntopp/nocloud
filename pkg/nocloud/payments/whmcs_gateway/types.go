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

type GetInvoiceQuery struct {
	Action       string `url:"action"`
	InvoiceId    int    `url:"invoiceid"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
}

type Item struct {
	Id          int    `json:"id"`
	Type        string `json:"type"`
	RelId       int    `json:"relid"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Taxed       int    `json:"taxed"`
}

type ItemHolder struct {
	Items []Item `json:"item"`
}

type Invoice struct {
	Result        string     `json:"result"`
	InvoiceId     int        `json:"invoiceid"`
	InvoiceNum    string     `json:"invoicenum"`
	UserId        int        `json:"userid"`
	Date          string     `json:"date"`     // YYYY-MM-DD
	DueDate       string     `json:"duedate"`  // YYYY-MM-DD
	DatePaid      string     `json:"datepaid"` // YYYY-MM-DD HH:ii:ss
	Subtotal      string     `json:"subtotal"` // type: float
	Credit        string     `json:"credit"`   // credit assigned
	Tax           string     `json:"tax"`      // first level tax charged
	Tax2          string     `json:"tax2"`     // second level tax charged
	Total         string     `json:"total"`
	Balance       string     `json:"balance"` // amount left to pay
	TaxRate       string     `json:"taxrate"`
	TaxRate2      string     `json:"taxrate2"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"paymentmethod"`
	Notes         string     `json:"notes"`
	CcGateway     bool       `json:"ccgateway"` // Whether the payment method is a credit card gateway that can be submitted to attempt capture.
	Items         ItemHolder `json:"items"`
	Transactions  string     `json:"transactions"`
}

type InvoiceResponse struct {
	Status    string `json:"status"`
	InvoiceId int    `json:"invoiceid"`
	Result    string `json:"result"`
	Message   string `json:"message"`
}

type InvoicePaid struct {
	InvoiceId int `json:"invoiceid"`
}
