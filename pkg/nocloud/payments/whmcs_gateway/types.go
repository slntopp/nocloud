package whmcs_gateway

import (
	pb "github.com/slntopp/nocloud-proto/billing"
	"strconv"
	"strings"
)

type floatAsString float64

func (foe *floatAsString) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		if foe != nil {
			*foe = 0
		}
		return nil
	}
	num := strings.ReplaceAll(string(data), `"`, "")
	n, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return err
	}
	*foe = floatAsString(n)
	return nil
}

func (foe *floatAsString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatFloat(float64(*foe), 'f', -1, 64) + `"`), nil
}

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
		return "Refunded"
	case pb.BillingStatus_TERMINATED:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

func statusToNoCloud(status string) pb.BillingStatus {
	switch strings.ToLower(status) {
	case "draft":
		return pb.BillingStatus_DRAFT
	case "unpaid":
		return pb.BillingStatus_UNPAID
	case "paid":
		return pb.BillingStatus_PAID
	case "cancelled":
		return pb.BillingStatus_CANCELED
	case "refunded":
		return pb.BillingStatus_RETURNED
	default:
		return pb.BillingStatus_DRAFT
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
	Id          int           `json:"id"`
	Type        string        `json:"type"`
	RelId       int           `json:"relid"`
	Description string        `json:"description"`
	Amount      floatAsString `json:"amount"`
	Taxed       int           `json:"taxed"`
}

type ItemHolder struct {
	Items []Item `json:"item"`
}

type Invoice struct {
	Result        string        `json:"result"`
	InvoiceId     int           `json:"invoiceid"`
	InvoiceNum    string        `json:"invoicenum"`
	UserId        int           `json:"userid"`
	Date          string        `json:"date"`     // YYYY-MM-DD
	DueDate       string        `json:"duedate"`  // YYYY-MM-DD
	DatePaid      string        `json:"datepaid"` // YYYY-MM-DD HH:ii:ss
	Subtotal      floatAsString `json:"subtotal"`
	Credit        floatAsString `json:"credit"` // credit assigned
	Tax           floatAsString `json:"tax"`    // first level tax charged
	Tax2          floatAsString `json:"tax2"`   // second level tax charged
	Total         floatAsString `json:"total"`
	Balance       floatAsString `json:"balance"` // amount left to pay
	TaxRate       floatAsString `json:"taxrate"`
	TaxRate2      floatAsString `json:"taxrate2"`
	Status        string        `json:"status"`
	PaymentMethod string        `json:"paymentmethod"`
	Notes         string        `json:"notes"`
	CcGateway     bool          `json:"ccgateway"` // Whether the payment method is a credit card gateway that can be submitted to attempt capture.
	Items         ItemHolder    `json:"items"`
	Transactions  string        `json:"transactions"`
}

type InvoiceResponse struct {
	Status    string `json:"status"`
	InvoiceId int    `json:"invoiceid"`
	Result    string `json:"result"`
	Message   string `json:"message"`
}

type UpdateInvoiceQuery struct {
	Action       string `url:"action"`
	InvoiceId    int    `url:"invoiceid"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash

	Status              *string               `url:"status,omitempty"`
	PaymentMethod       *string               `url:"paymentmethod,omitempty"`
	TaxRate             *floatAsString        `url:"taxrate,omitempty"`
	TaxRate2            *floatAsString        `url:"taxrate2,omitempty"`
	Credit              *floatAsString        `url:"credit,omitempty"`
	Date                *string               `url:"date,omitempty"`
	DueDate             *string               `url:"duedate,omitempty"`
	DatePaid            *string               `url:"datepaid,omitempty"`
	Notes               *string               `url:"notes,omitempty"`
	ItemDescription     map[int]string        `url:"-,omitempty"`
	ItemAmount          map[int]floatAsString `url:"-,omitempty"`
	ItemTaxed           map[int]bool          `url:"-,omitempty"`
	NewItemDescription  map[int]string        `url:"-,omitempty"`
	NewItemAmount       map[int]floatAsString `url:"-,omitempty"`
	NewItemTaxed        map[int]bool          `url:"-,omitempty"`
	DeleteLineIds       []int                 `url:"deletelineids[],omitempty"`
	Publish             *bool                 `url:"publish,omitempty"`
	PublishAndSendEmail *bool                 `url:"publishandsendemail,omitempty"`
}

type PaymentURIQuery struct {
	ClientID  int `url:"client_id"`
	InvoiceID int `url:"invoice_id"`
}

type InvoicePaid struct {
	InvoiceId int `json:"invoiceid"`
}

type InvoiceCancelled struct {
	InvoiceId int `json:"invoiceid"`
}

type InvoiceRefunded struct {
	InvoiceId int `json:"invoiceid"`
}

type InvoiceModified struct {
	InvoiceId int `json:"invoiceid"`
}

type UpdateInvoiceTotal struct {
	InvoiceId int `json:"invoiceid"`
}

type InvoiceUnpaid struct {
	InvoiceId int `json:"invoiceid"`
}

type InvoiceDeleted struct {
	InvoiceId int `json:"invoiceid"`
}

// InvoiceCreated Missing user field
type InvoiceCreated struct {
	InvoiceId int    `json:"invoiceid"`
	Source    string `json:"source"`
	Status    string `json:"status"`
}
