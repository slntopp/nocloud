package whmcs_gateway

import (
	"encoding/json"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"strconv"
	"strings"
)

type IntOrString int

func (i *IntOrString) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		if i != nil {
			*i = 0
		}
		return nil
	}

	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*i = IntOrString(num)
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		parsedNum, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("invalid string value for IntOrString: %s", str)
		}
		*i = IntOrString(parsedNum)
		return nil
	}

	return fmt.Errorf("IntOrString must be a string or number")
}

func (i *IntOrString) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(*i)))
}

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
	Action          string         `url:"action"`
	Username        string         `url:"username"`
	Password        string         `url:"password"` // md5 hash
	UserId          string         `url:"userid"`
	Status          string         `url:"status"`
	SendInvoice     string         `url:"sendinvoice"`
	PaymentMethod   *string        `url:"paymentmethod"`
	TaxRate         *floatAsString `url:"taxrate"`
	Date            string         `url:"date"`
	DueDate         string         `url:"duedate"`
	AutoApplyCredit string         `url:"autoapplycredit"`
	Notes           string         `url:"notes"`
	ResponseType    string         `url:"responsetype"`
}

type GetInvoiceQuery struct {
	Action       string `url:"action"`
	InvoiceId    int    `url:"invoiceid"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
}

type GetInvoicesQuery struct {
	Action       string  `url:"action"`
	ResponseType string  `url:"responsetype"`
	Username     string  `url:"username"`
	Password     string  `url:"password"` // md5 hash
	LimitStart   *int    `url:"limitstart"`
	LimitNum     *int    `url:"limitnum"`
	UserID       *int    `url:"userid"`
	Status       *string `url:"status"`
	OrderBy      *string `url:"orderby"`
	Order        *string `url:"order"`
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
	//Transactions  string        `json:"transactions"`
}

type GetInvoicesResponse struct {
	Result       string         `json:"result"`
	Message      string         `json:"message"`
	TotalResults int            `json:"totalresults"`
	StartNumber  int            `json:"startnumber"`
	NumReturned  int            `json:"numreturned"`
	Invoices     InvoicesHolder `json:"invoices"`
}

type InvoiceInList struct {
	Id            IntOrString   `json:"id"`
	CreatedAt     string        `json:"created_at"`
	Number        string        `json:"invoicenum"`
	UserID        IntOrString   `json:"userid"`
	DatePaid      string        `json:"datepaid"`
	Date          string        `json:"date"`
	Total         floatAsString `json:"total"`
	Currency      string        `json:"currencycode"`
	PaymentMethod string        `json:"paymentmethod"`
	Status        string        `json:"status"`
}

type InvoicesHolder struct {
	Invoice []InvoiceInList `json:"invoice"`
}

type InvoiceResponse struct {
	Status    string `json:"status"`
	InvoiceId int    `json:"invoiceid"`
	Result    string `json:"result"`
	Message   string `json:"message"`
}

type PaymentURIResponse struct {
	AccessToken string `json:"access_token"`
	RedirectUrl string `json:"redirect_url"`
	Result      string `json:"result"`
	Message     string `json:"message"`
}

type UpdateInvoiceQuery struct {
	Action       string `url:"action"`
	InvoiceId    int    `url:"invoiceid"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash

	Status              *string               `url:"status,omitempty"`
	PaymentMethod       *string               `url:"paymentmethod,omitempty"`
	TaxRate             floatAsString         `url:"taxrate"`
	TaxRate2            floatAsString         `url:"taxrate2"`
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
	Action          string `url:"action"`
	Username        string `url:"username"`
	Password        string `url:"password"`
	ClientID        int    `url:"client_id"`
	Destination     string `url:"destination"`
	SsoRedirectPath string `url:"sso_redirect_path"`
	ResponseType    string `url:"responsetype"`
}

type AddPaymentQuery struct {
	Action       string         `url:"action"`
	InvoiceId    int            `url:"invoiceid"`
	ResponseType string         `url:"responsetype"`
	Username     string         `url:"username"`
	Password     string         `url:"password"` // md5 hash
	TransId      string         `url:"transid"`
	Gateway      string         `url:"gateway"`
	Date         string         `url:"date"`
	Amount       *floatAsString `url:"amount"`
	Fees         *floatAsString `url:"fees"`
	NoEmail      *bool          `url:"noemail"`
}

type UpdateClientQuery struct {
	Action       string `url:"action"`
	ClientID     int    `url:"clientid"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
	Notes        string `url:"notes"`
}

type AddNoteQuery struct {
	Action       string `url:"action"`
	UserID       int    `url:"userid"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
	Notes        string `url:"notes"`
	Sticky       bool   `url:"sticky"`
}

type GetClientsQuery struct {
	Action       string `url:"action"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
	LimitNum     int    `url:"limitnum"`
}

type GetClientsDetailsQuery struct {
	Action       string `url:"action"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
	ClientID     int    `url:"clientid"`
}

type GetClientsProductsQuery struct {
	Action       string `url:"action"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
	ClientID     int    `url:"clientid"`
	LimitNum     int    `url:"limitnum"`
}

type GetPaymentMethodsQuery struct {
	Action       string `url:"action"`
	ResponseType string `url:"responsetype"`
	Username     string `url:"username"`
	Password     string `url:"password"` // md5 hash
}

type InvoicePaid struct {
	InvoiceId IntOrString `json:"invoiceid"`
}

type InvoiceCancelled struct {
	InvoiceId IntOrString `json:"invoiceid"`
}

type InvoiceRefunded struct {
	InvoiceId IntOrString `json:"invoiceid"`
}

type InvoiceModified struct {
	InvoiceId IntOrString `json:"invoiceid"`
}

type UpdateInvoiceTotal struct {
	InvoiceId IntOrString `json:"invoiceid"`
}

type InvoiceUnpaid struct {
	InvoiceId IntOrString `json:"invoiceid"`
}

type InvoiceDeleted struct {
	InvoiceId IntOrString `json:"invoiceid"`
}

type InvoiceCreated struct {
	InvoiceId IntOrString `json:"invoiceid"`
	Source    string      `json:"source"`
	Status    string      `json:"status"`
	User      interface{} `json:"user"`
}

type GetClientsResponse struct {
	Result       string        `json:"result"`
	Message      string        `json:"message"`
	TotalResults int           `json:"totalresults"`
	StartNumber  int           `json:"startnumber"`
	NumReturned  int           `json:"numreturned"`
	Clients      ClientsHolder `json:"clients"`
}

type ListClient struct {
	ID      int `json:"id"`
	GroupID int `json:"groupid"`
}

type ClientsHolder struct {
	Client []ListClient `json:"client"`
}

type GetClientsProductsResponse struct {
	Result       string         `json:"result"`
	Message      string         `json:"message"`
	TotalResults int            `json:"totalresults"`
	StartNumber  int            `json:"startnumber"`
	NumReturned  int            `json:"numreturned"`
	Products     ProductsHolder `json:"products"`
}

type ListProduct struct {
	ID                 int           `json:"id"`
	ClientID           int           `json:"clientid"`
	OrderID            int           `json:"orderid"`
	Name               string        `json:"name"`
	RegDate            string        `json:"regdate"`
	Domain             string        `json:"domain"`
	DedicatedIP        string        `json:"dedicatedip"`
	ServerIP           string        `json:"serverip"`
	FirstPaymentAmount floatAsString `json:"firstpaymentamount"`
	RecurringAmount    floatAsString `json:"recurringamount"`
	Status             string        `json:"status"`
}

type ProductsHolder struct {
	Product []ListProduct `json:"product"`
}

type GetClientsDetailsResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
	Client  Client `json:"client"`
}

type CustomField struct {
	ID    IntOrString `json:"id"`
	Value string      `json:"value"`
}

type Client struct {
	ID            int           `json:"id"`
	FirstName     string        `json:"firstname"`
	LastName      string        `json:"lastname"`
	CompanyName   string        `json:"companyname"`
	Email         string        `json:"email"`
	Status        string        `json:"status"`
	Address1      string        `json:"address1"`
	Address2      string        `json:"address2"`
	City          string        `json:"city"`
	State         string        `json:"state"`
	PostCode      string        `json:"postcode"`
	Country       string        `json:"country"`
	Phone         string        `json:"phonenumber"`
	PaymentMethod string        `json:"defaultgateway"`
	LastLogin     string        `json:"lastlogin"` // example: Date: 28/01/2022 11:13<br>IP Address: 82.209.222.112<br>Host: mm-112-222-209-82.static.mgts.by
	GroupID       IntOrString   `json:"groupid"`
	CustomFields  []CustomField `json:"customfields"`
}

func findCustomField(fields []CustomField, id int) string {
	if fields == nil {
		return ""
	}
	for _, field := range fields {
		if int(field.ID) == id {
			return field.Value
		}
	}
	return ""
}

func (c Client) GetClientType() string {
	return findCustomField(c.CustomFields, 3629) // Физ. лицо или предприниматель
}
func (c Client) GetPAN() string {
	return findCustomField(c.CustomFields, 3212) // УНП
}
func (c Client) GetCompanyPAN() string {
	return findCustomField(c.CustomFields, 3550) // УНП организации
}
func (c Client) GetCertOfRegistration() string {
	return findCustomField(c.CustomFields, 3628) // Домен ю/л - Свидетельство о регистрации №
}
func (c Client) GetDateOfGovRegistration() string {
	return findCustomField(c.CustomFields, 3552) // Домен ю/л - Дата государственной регистрации
}

// GetCurrentAccount - расчетный счет клиента
func (c Client) GetCurrentAccount() string {
	return findCustomField(c.CustomFields, 3409) // Расчетный счёт
}

type GetPaymentMethodsResponse struct {
	Result  string               `json:"result"`
	Message string               `json:"message"`
	Methods PaymentMethodsHolder `json:"paymentmethods"`
}

type PaymentMethodsHolder struct {
	Method []PaymentMethod `json:"paymentmethod"`
}

type PaymentMethod struct {
	Module string `json:"module"`
	Name   string `json:"displayname"`
}
