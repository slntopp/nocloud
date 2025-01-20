package billing

import (
	"bytes"
	"connectrpc.com/connect"
	"context"
	"encoding/csv"
	"github.com/arangodb/go-driver"
	instancespb "github.com/slntopp/nocloud-proto/instances"
	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"github.com/tiendc/go-csvlib"
	"go.uber.org/zap"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

const clientsFilePrefix = "users"
const servicesFilePrefix = "services"
const billsFilePrefix = "bills"

var (
	reportsLocation    string
	forbiddenGateways  []string
	forbiddenUserGroup int
	allowedGateways    []string // If empty, then ALL allowed
)

func init() {
	viper.SetDefault("REPORTS_LOCATION", "/reports")
	viper.SetDefault("FORBIDDEN_REPORTS_PAYMENT_GATEWAYS", "")
	viper.SetDefault("FORBIDDEN_REPORTS_USER_GROUP", "")
	viper.SetDefault("ALLOWED_REPORTS_PAYMENT_GATEWAYS", "")

	reportsLocation = viper.GetString("REPORTS_LOCATION")
	forbiddenGateways = viper.GetStringSlice("FORBIDDEN_REPORTS_PAYMENT_GATEWAYS")
	forbiddenUserGroup = viper.GetInt("FORBIDDEN_REPORTS_USER_GROUP")
	allowedGateways = viper.GetStringSlice("ALLOWED_REPORTS_PAYMENT_GATEWAYS")
}

type ClientsReport struct {
	WhmcsID        int    `csv:"WHMCS ID"`
	FirstName      string `csv:"Имя"`
	LastName       string `csv:"Фамилия"`
	CompanyName    string `csv:"Компания"`
	Email          string `csv:"Электронная почта"`
	Status         string `csv:"Статус"`
	Address1       string `csv:"Адрес1"`
	Address2       string `csv:"Адрес2"`
	City           string `csv:"Город"`
	State          string `csv:"Регион"`
	PostCode       string `csv:"Почтовый индекс"`
	Country        string `csv:"Страна"`
	Phone          string `csv:"Телефон"`
	PaymentMethod  string `csv:"Способ оплаты"`
	LastLogin      string `csv:"Последняя активность"`
	PAN            string `csv:"УНП"`
	ClientType     string `csv:"Тип клиента"`
	CurrentAccount string `csv:"Расчетный счет"`
	CompanyPAN     string `csv:"УНП организации"`
	GovRegDate     string `csv:"Дата государственной регистрации"`
	GovRegCert     string `csv:"Свидетельство о регистрации"`
}

type ServiceReport struct {
	WhmcsID     int    `csv:"WHMCS ID"`
	OrderID     int    `csv:"WHMCS ORDER ID"`
	ClientID    int    `csv:"WHMCS CLIENT ID"`
	ClientName  string `csv:"Имя клиента"`
	ProductName string `csv:"Название продукта"`
	IP          string `csv:"IP адрес"`
	DateCreate  string `csv:"Дата создания"`
	Status      string `csv:"Статус"`
	Price       string `csv:"Цена"`
}

type PaymentReport struct {
	WhmcsID       int    `csv:"WHMCS ID"`
	ClientID      int    `csv:"WHMCS CLIENT ID"`
	ClientName    string `csv:"Имя клиента"`
	Number        string `csv:"Номер счёта"`
	DatePaid      string `csv:"Дата платежа"`
	Amount        string `csv:"Сумма платежа"`
	Currency      string `csv:"Валюта"`
	PaymentMethod string `csv:"Способ платежа"`
	Status        string `csv:"Статус платежа"`
}

func (s *BillingServiceServer) CollectSystemReport(ctx context.Context, log *zap.Logger) {
	log = log.Named("CollectSystemReport")
	log.Info("Starting CollectSystemReport")

	accounts, err := s.accounts.List(ctx, graph.Account{
		Account: &accountspb.Account{Uuid: schema.ROOT_ACCOUNT_KEY},
		DocumentMeta: driver.DocumentMeta{
			ID:  driver.NewDocumentID(schema.ACCOUNTS_COL, schema.ROOT_ACCOUNT_KEY),
			Key: schema.ROOT_ACCOUNT_KEY,
		},
	}, 10000)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return
	}
	log.Debug("Got accounts", zap.Int("count", len(accounts)))

	page, limit := uint64(1), uint64(0)
	req := connect.NewRequest(&instancespb.ListInstancesRequest{Page: &page, Limit: &limit})
	req.Header().Set("Authorization", "Bearer "+ctx.Value(nocloud.NoCloudToken).(string))
	resp, err := s.instancesClient.List(ctx, req)
	if err != nil {
		log.Error("Failed to list instances", zap.Error(err))
		return
	}
	log.Debug("Got instances", zap.Int64("count", resp.Msg.Count))

	accToWhmcsId := make(map[string]int)
	for _, acc := range accounts {
		data := acc.GetData().AsMap()
		idVal, _ := data["whmcs_id"].(float64)
		id := int(idVal)
		if id > 0 {
			accToWhmcsId[acc.Key] = id
		}
	}

	_plans, err := s.plans.List(ctx, "")
	if err != nil {
		log.Error("Failed to list plans", zap.Error(err))
		return
	}
	plans := make(map[string]*graph.BillingPlan)
	for _, p := range _plans {
		plans[p.GetUuid()] = p
	}

	instances := make(map[int][]*instancespb.Instance)
	for _, obj := range resp.Msg.Pool {
		whmcsId, ok := accToWhmcsId[obj.GetAccount()]
		if ok {
			if val, ok := instances[whmcsId]; !ok || val == nil {
				instances[whmcsId] = make([]*instancespb.Instance, 0)
			}
			instances[whmcsId] = append(instances[whmcsId], obj.Instance)
		}
	}

	whmcsClients, err := s.whmcsGateway.GetClients(ctx)
	if err != nil {
		log.Error("Failed to get whmcs clients", zap.Error(err))
		return
	}
	whmcsInvoices, err := s.whmcsGateway.GetInvoices(ctx)
	if err != nil {
		log.Error("Failed to get whmcs invoices", zap.Error(err))
		return
	}
	methods, err := s.whmcsGateway.GetPaymentMethods(ctx)
	if err != nil {
		log.Error("Failed to get payment methods", zap.Error(err))
		return
	}
	methodsNames := make(map[string]string)
	for _, method := range methods {
		methodsNames[method.Module] = method.Name
	}

	clients := make([]whmcs_gateway.Client, 0)
	products := make(map[int][]whmcs_gateway.ListProduct)
	for _, c := range whmcsClients {
		if c.GroupID == forbiddenUserGroup {
			continue
		}
		client := c

		details, err := s.whmcsGateway.GetClientsDetails(ctx, client.ID)
		if err != nil {
			log.Error("Failed to get client details", zap.Error(err))
			continue
		}
		if int(details.GroupID) == forbiddenUserGroup {
			continue
		}
		if details.PaymentMethod != "" {
			if slices.Contains(forbiddenGateways, details.PaymentMethod) {
				continue
			}
			if len(allowedGateways) > 0 && !slices.Contains(allowedGateways, details.PaymentMethod) {
				continue
			}
		}
		clients = append(clients, details)

		prods, err := s.whmcsGateway.GetClientsProducts(ctx, client.ID)
		if err != nil {
			if !strings.Contains(err.Error(), "json: cannot unmarshal string into Go struct") {
				log.Error("Failed to get client products", zap.Error(err))
			}
			continue
		}
		if len(prods) == 0 {
			continue
		}
		products[client.ID] = prods
	}
	clientsMap := make(map[int]whmcs_gateway.Client)
	for _, c := range clients {
		clientsMap[c.ID] = c
	}

	reportsBills := make([]PaymentReport, 0)
	reportsServices := make([]ServiceReport, 0)
	reportsClients := make([]ClientsReport, 0)

	// Create clients reports
	for _, c := range clients {
		reportsClients = append(reportsClients, ClientsReport{
			WhmcsID:        c.ID,
			FirstName:      c.FirstName,
			LastName:       c.LastName,
			CompanyName:    c.CompanyName,
			Email:          c.Email,
			Status:         c.Status,
			Address1:       c.Address1,
			Address2:       c.Address2,
			City:           c.City,
			State:          c.State,
			PostCode:       c.PostCode,
			Country:        c.Country,
			Phone:          c.Phone,
			PaymentMethod:  methodsNames[c.PaymentMethod],
			LastLogin:      c.LastLogin,
			PAN:            c.GetPAN(),
			ClientType:     c.GetClientType(),
			CurrentAccount: c.GetCurrentAccount(),
			CompanyPAN:     c.GetCompanyPAN(),
			GovRegDate:     c.GetDateOfGovRegistration(),
			GovRegCert:     c.GetCertOfRegistration(),
		})
	}
	// Create bills reports
	const dateBorder = 1546300800 // 1 Jan 2019
	for _, i := range whmcsInvoices {
		if i.Status != "Paid" {
			continue
		}
		date, err := time.Parse(time.DateOnly, i.Date)
		if err != nil || date.Unix() < dateBorder {
			continue
		}
		if slices.Contains(forbiddenGateways, i.PaymentMethod) {
			continue
		}
		if len(allowedGateways) > 0 && !slices.Contains(allowedGateways, i.PaymentMethod) {
			continue
		}
		var number = strconv.Itoa(int(i.Id))
		if i.Number != "" {
			number = i.Number
		}
		reportsBills = append(reportsBills, PaymentReport{
			WhmcsID:       int(i.Id),
			ClientID:      int(i.UserID),
			ClientName:    clientsMap[int(i.UserID)].LastName + " " + clientsMap[int(i.UserID)].FirstName,
			Number:        number,
			DatePaid:      i.DatePaid,
			Amount:        strconv.FormatFloat(float64(i.Total), 'f', 2, 64),
			Currency:      i.Currency,
			PaymentMethod: methodsNames[i.PaymentMethod],
			Status:        i.Status,
		})
	}
	// Create services report
	for clID, services := range products {
		for _, srv := range services {
			rp := strconv.FormatFloat(float64(srv.RecurringAmount), 'f', 2, 64)
			fp := strconv.FormatFloat(float64(srv.FirstPaymentAmount), 'f', 2, 64)
			reportsServices = append(reportsServices, ServiceReport{
				WhmcsID:     srv.ID,
				ClientID:    clID,
				ClientName:  clientsMap[clID].LastName + " " + clientsMap[clID].FirstName,
				OrderID:     srv.OrderID,
				ProductName: srv.Name,
				IP:          strings.Trim(strings.Join(removeDuplicates([]string{srv.DedicatedIP, srv.Domain, srv.ServerIP}), " "), " "),
				DateCreate:  srv.RegDate,
				Status:      srv.Status,
				Price:       strings.Trim(strings.Join([]string{fp, rp}, " "), " "),
			})
		}
	}
	for clID, services := range instances {
		for _, srv := range services {
			var product = "no_product"
			var price = "-1"
			if srv.Product != nil && *srv.Product != "" {
				product = *srv.Product
				if srv.BillingPlan != nil {
					plan, ok := plans[srv.BillingPlan.GetUuid()]
					if ok {
						prod, ok := plan.GetProducts()[product]
						if ok {
							price = strconv.FormatFloat(prod.Price, 'f', 2, 64)
						}
					}
				}
			}
			ips := make([]string, 0)
			if srv.GetState() != nil && srv.GetState().GetInterfaces() != nil {
				for _, inter := range srv.GetState().GetInterfaces() {
					ips = append(ips, inter.GetData()["host"])
				}
			}
			reportsServices = append(reportsServices, ServiceReport{
				WhmcsID:     -1,
				ClientID:    clID,
				ClientName:  clientsMap[clID].LastName + " " + clientsMap[clID].FirstName,
				OrderID:     -1,
				ProductName: product,
				DateCreate:  time.Unix(srv.Created, 0).Format(time.DateOnly),
				Status:      srv.Status.String(),
				Price:       price,
				IP:          strings.Trim(strings.Join(removeDuplicates(ips), " "), " "),
			})
		}
	}

	var (
		buf    bytes.Buffer
		output string
		writer *csv.Writer
	)
	// Write clients report
	buf = bytes.Buffer{}
	writer = csv.NewWriter(&buf)
	writer.Comma = '|'
	err = csvlib.NewEncoder(writer).Encode(reportsClients)
	if err != nil {
		log.Error("Failed to write csv file", zap.Error(err))
		return
	}
	writer.Flush()
	output = buf.String()
	output = strings.Replace(output, "|", "%|%", -1)
	_ = writeToFile(log, clientsFilePrefix+"_example", strings.Replace(output, "%|%", ",", -1))
	if err = writeToFile(log, clientsFilePrefix, output); err != nil {
		log.Error("Failed to write to file", zap.Error(err))
		return
	}
	// Write services report
	buf = bytes.Buffer{}
	writer = csv.NewWriter(&buf)
	writer.Comma = '|'
	err = csvlib.NewEncoder(writer).Encode(reportsServices)
	if err != nil {
		log.Error("Failed to write csv file", zap.Error(err))
		return
	}
	writer.Flush()
	output = buf.String()
	output = strings.Replace(output, "|", "%|%", -1)
	_ = writeToFile(log, servicesFilePrefix+"_example", strings.Replace(output, "%|%", ",", -1))
	if err = writeToFile(log, servicesFilePrefix, output); err != nil {
		log.Error("Failed to write to file", zap.Error(err))
		return
	}
	// Write bills report
	buf = bytes.Buffer{}
	writer = csv.NewWriter(&buf)
	writer.Comma = '|'
	err = csvlib.NewEncoder(writer).Encode(reportsBills)
	if err != nil {
		log.Error("Failed to write csv file", zap.Error(err))
		return
	}
	writer.Flush()
	output = buf.String()
	output = strings.Replace(output, "|", "%|%", -1)
	_ = writeToFile(log, billsFilePrefix+"_example", strings.Replace(output, "%|%", ",", -1))
	if err = writeToFile(log, billsFilePrefix, output); err != nil {
		log.Error("Failed to write to file", zap.Error(err))
		return
	}
}

func writeToFile(log *zap.Logger, prefix string, content string) error {
	log.Debug("File output with prefix "+prefix, zap.String("output", content))
	now := reverseDate(strings.Replace(time.Now().Format(time.DateOnly), "-", "", -1))
	filename := prefix + "_" + now + ".csv"
	if err := os.MkdirAll(reportsLocation, 0777); err != nil {
		return err
	}
	filepath := path.Join(reportsLocation, filename)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(content))
	if err != nil {
		return err
	}
	return nil
}

func removeDuplicates(strings []string) []string {
	unique := make(map[string]bool)
	result := make([]string, 0)
	for _, str := range strings {
		if _, exists := unique[str]; !exists {
			unique[str] = true
			result = append(result, str)
		}
	}
	return result
}

func reverseDate(date string) string {
	if len(date) != 8 {
		return date
	}
	return string(date[6]) + string(date[7]) + string(date[4]) + string(date[5]) + date[:4]
}
