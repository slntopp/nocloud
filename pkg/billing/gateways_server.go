package billing

import (
	"bytes"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
	accesspb "github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/rest_auth"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"html"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PaymentGatewayType string

var (
	PaymentGatewayFacture PaymentGatewayType = "facture"
)

const gatewaysBase = "/billing/payments"

type PaymentGatewayServer struct {
	log          *zap.Logger
	db           driver.Database
	ctrl         graph.PaymentGatewaysController
	invoicesCtrl graph.InvoicesController
	caCtrl       graph.CommonActionsController
	accountsCtrl graph.AccountsController

	rdb redisdb.Client
	sk  []byte

	settingsClient settingspb.SettingsServiceClient
}

func NewPaymentGatewayServer(_log *zap.Logger, db driver.Database, rdb redisdb.Client, sighingKey []byte, sc settingspb.SettingsServiceClient) *PaymentGatewayServer {
	log := _log.Named("PaymentGatewayServer")
	ctrl := graph.NewPaymentGatewaysController(log, db)
	invoicesCtrl := graph.NewInvoicesController(log, db)
	caCtrl := graph.NewCommonActionsController(log, db)
	accountsCtrl := graph.NewAccountsController(log, db)
	return &PaymentGatewayServer{
		log:            log,
		db:             db,
		ctrl:           ctrl,
		invoicesCtrl:   invoicesCtrl,
		caCtrl:         caCtrl,
		rdb:            rdb,
		sk:             sighingKey,
		accountsCtrl:   accountsCtrl,
		settingsClient: sc,
	}
}

func (s *PaymentGatewayServer) RegisterRoutes(router *mux.Router) {
	interceptor := rest_auth.NewInterceptor(s.log, s.rdb, s.sk)
	subRouter := router.PathPrefix(gatewaysBase).Subrouter()
	subRouter.Handle("/{key}/{invoice_uuid}/action", interceptor.JwtMiddleWare(http.HandlerFunc(s.HandlePaymentAction))).Methods("POST")
	subRouter.Handle("/{invoice_uuid}/view", interceptor.JwtMiddleWare(http.HandlerFunc(s.HandleViewInvoice))).Methods("GET")
}

func (s *PaymentGatewayServer) HandleViewInvoice(writer http.ResponseWriter, request *http.Request) {
	invoiceUuid := mux.Vars(request)["invoice_uuid"]
	if invoiceUuid == "" {
		http.Error(writer, "Invoice UUID is required", http.StatusBadRequest)
		return
	}
	invoice, err := s.invoicesCtrl.Get(request.Context(), invoiceUuid)
	if err != nil {
		http.Error(writer, "Failed to get invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}
	requester, _ := request.Context().Value(nocloud.NoCloudAccount).(string)
	if !s.caCtrl.HasAccess(request.Context(), requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ADMIN) &&
		requester != invoice.Account {
		http.Error(writer, "Access denied to view this invoice", http.StatusForbidden)
		return
	}
	account, err := s.accountsCtrl.Get(request.Context(), invoice.Account)
	if err != nil {
		http.Error(writer, "Failed to get account: "+err.Error(), http.StatusInternalServerError)
		return
	}
	group, err := s.accountsCtrl.GetAccountClientGroupAlwaysFound(request.Context(), invoice.Account)
	if err != nil {
		http.Error(writer, "Failed to get account group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	invConf := MakeInvoicesConf(s.log, &s.settingsClient)

	gateways, err := s.ctrl.List(request.Context(), true)
	if err != nil {
		http.Error(writer, "Failed to list payment gateways: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Prepare Gateways HTMLs
	for _, gw := range gateways {
		rawHtml := gw.CheckoutPanelHtml
		rawHtml = strings.ReplaceAll(rawHtml, "$ACTION_URL", os.Getenv("BASE_HOST")+"/billing/payments/"+gw.Key+"/"+invoice.Uuid+"/action")
		token, _ := auth.MakeToken(invoice.Account)
		rawHtml = strings.ReplaceAll(rawHtml, "$ACCESS_TOKEN", token)
		gw.CheckoutPanelHtml = rawHtml
	}

	var (
		Supplier    string
		Buyer       = "Email: "
		LogoURL     string
		InvoiceBody = invoice.Invoice
		Gateways    = gateways
	)

	if group.HasOwnInvoiceBase && group.InvoiceParametersCustom != nil {
		Supplier = group.InvoiceParametersCustom.InvoiceFrom
		LogoURL = group.InvoiceParametersCustom.LogoUrl
	} else {
		Supplier = invConf.InvoiceFrom
		LogoURL = invConf.LogoURL
	}

	if account.Data != nil {
		dataMap := account.Data.AsMap()
		if email, ok := dataMap["email"].(string); ok {
			Buyer += email
		} else {
			Buyer += "N/A"
		}
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(generateViewInvoiceHTML(InvoiceBody, Gateways, Supplier, Buyer, LogoURL)))
}

func (s *PaymentGatewayServer) HandlePaymentAction(writer http.ResponseWriter, request *http.Request) {
	pgKey := mux.Vars(request)["key"]
	if pgKey == "" {
		http.Error(writer, "Payment Gateway Key is required", http.StatusBadRequest)
		return
	}
	invoiceUuid := mux.Vars(request)["invoice_uuid"]
	if invoiceUuid == "" {
		http.Error(writer, "Invoice UUID is required", http.StatusBadRequest)
		return
	}
	invoice, err := s.invoicesCtrl.Get(request.Context(), invoiceUuid)
	if err != nil {
		http.Error(writer, "Failed to get invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}
	requester, _ := request.Context().Value(nocloud.NoCloudAccount).(string)
	if !s.caCtrl.HasAccess(request.Context(), requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ADMIN) &&
		requester != invoice.Account {
		http.Error(writer, "Access denied to perform action on this invoice", http.StatusForbidden)
		return
	}
	account, err := s.accountsCtrl.Get(request.Context(), invoice.Account)
	if err != nil {
		http.Error(writer, "Failed to get account: "+err.Error(), http.StatusInternalServerError)
		return
	}
	group, err := s.accountsCtrl.GetAccountClientGroupAlwaysFound(request.Context(), invoice.Account)
	if err != nil {
		http.Error(writer, "Failed to get account group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	invConf := MakeInvoicesConf(s.log, &s.settingsClient)

	var (
		Supplier    string
		Buyer       = "Email: "
		LogoURL     string
		InvoiceBody = invoice.Invoice
	)

	if group.HasOwnInvoiceBase && group.InvoiceParametersCustom != nil {
		Supplier = group.InvoiceParametersCustom.InvoiceFrom
		LogoURL = group.InvoiceParametersCustom.LogoUrl
	} else {
		Supplier = invConf.InvoiceFrom
		LogoURL = invConf.LogoURL
	}

	if account.Data != nil {
		dataMap := account.Data.AsMap()
		if email, ok := dataMap["email"].(string); ok {
			Buyer += email
		} else {
			Buyer += "N/A"
		}
	}

	switch PaymentGatewayType(pgKey) {
	case PaymentGatewayFacture:
		s.log.Debug("Handling Facture Payment Action", zap.String("invoice_uuid", invoiceUuid))
		gw, err := s.ctrl.Get(request.Context(), string(PaymentGatewayFacture))
		if err != nil {
			http.Error(writer, "Failed to get payment gateway: "+err.Error(), http.StatusInternalServerError)
			return
		}
		pdfBytes, err := generateInvoicePDF(InvoiceBody, gw.CheckoutExtraText, Supplier, Buyer, LogoURL)
		if err != nil {
			http.Error(writer, "Failed to generate invoice PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/pdf")
		writer.Header().Set("Content-Disposition", `attachment; filename="invoice-`+InvoiceBody.GetNumber()+`.pdf"`)
		_, _ = writer.Write(pdfBytes)
		return

	default:
		http.Error(writer, "Unsupported Payment Gateway Key", http.StatusBadRequest)
		return
	}
}

func generateInvoicePDF(invoiceBody *pb.Invoice, gwExtraText string, supplier string, buyer string, logoURL string) ([]byte, error) {
	if invoiceBody == nil {
		return nil, fmt.Errorf("invoiceBody is nil")
	}

	tsToTime := func(ts int64) time.Time {
		if ts <= 0 {
			return time.Time{}
		}
		if ts > 1_000_000_000_000 {
			return time.UnixMilli(ts)
		}
		return time.Unix(ts, 0)
	}
	formatDate := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("02/01/2006")
	}
	intOrDefault := func(v, d int32) int32 {
		if v <= 0 {
			return d
		}
		return v
	}
	formatMoney := func(cur *pb.Currency, amount float64) string {
		precision := int(intOrDefault(cur.GetPrecision(), 2))
		pow := math.Pow(10, float64(precision))
		amount = math.Round(amount*pow) / pow

		num := fmt.Sprintf("%.*f", precision, amount)

		format := "{amount}"
		if cur != nil && strings.TrimSpace(cur.GetFormat()) != "" {
			format = cur.GetFormat()
		} else if cur != nil && strings.TrimSpace(cur.GetCode()) != "" {
			format = "{amount} " + cur.GetCode()
		}
		return strings.ReplaceAll(format, "{amount}", num)
	}
	statusText := func(st pb.BillingStatus) string {
		switch st {
		case pb.BillingStatus_PAID:
			return "PAID"
		case pb.BillingStatus_UNPAID:
			return "UNPAID"
		case pb.BillingStatus_CANCELED:
			return "CANCELED"
		case pb.BillingStatus_TERMINATED:
			return "TERMINATED"
		case pb.BillingStatus_DRAFT:
			return "DRAFT"
		case pb.BillingStatus_RETURNED:
			return "RETURNED"
		default:
			return "UNKNOWN"
		}
	}
	loadLogo := func(pdf *gofpdf.Fpdf, loc string) (name string, w float64, err error) {
		if strings.TrimSpace(loc) == "" {
			return "", 0, nil
		}
		u, _ := url.Parse(loc)
		if u != nil && (u.Scheme == "http" || u.Scheme == "https") {
			resp, err := http.Get(loc)
			if err != nil {
				return "", 0, nil
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 300 {
				return "", 0, nil
			}
			b, _ := io.ReadAll(resp.Body)
			if len(b) == 0 {
				return "", 0, nil
			}
			name = "logo" + filepath.Ext(u.Path)
			opt := gofpdf.ImageOptions{ImageType: strings.TrimPrefix(strings.ToUpper(filepath.Ext(u.Path)), "."), ReadDpi: true}
			pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(b))
			return name, 40, nil
		}
		return loc, 40, nil
	}

	var (
		subtotal    float64
		totalTax    float64
		grandTotal  float64
		taxRate     = 0.0
		taxIncluded = false
	)
	if to := invoiceBody.GetTaxOptions(); to != nil {
		taxRate = to.GetTaxRate()
		taxIncluded = to.GetTaxIncluded()
	}

	type row struct {
		idx, item, qty, unit, unitPrice, vat, amount, total string
	}
	var rows []row

	for i, it := range invoiceBody.GetItems() {
		qty := float64(it.GetAmount())
		unitPrice := it.GetPrice()
		line := unitPrice * qty

		applyTax := it.GetApplyTax() && taxRate > 0
		var taxAmt, lineSubtotal, lineTotal float64
		var vatLabel string

		if applyTax {
			if taxIncluded {
				base := line / (1.0 + taxRate)
				taxAmt = line - base
				lineSubtotal = base
				lineTotal = line
			} else {
				taxAmt = line * taxRate
				lineSubtotal = line
				lineTotal = line + taxAmt
			}
			vatLabel = fmt.Sprintf("%.0f%%", taxRate*100)
		} else {
			taxAmt = 0
			lineSubtotal = line
			lineTotal = line
			vatLabel = "NP"
		}

		subtotal += lineSubtotal
		totalTax += taxAmt
		grandTotal += lineTotal

		rows = append(rows, row{
			idx:       fmt.Sprintf("%d", i+1),
			item:      it.GetDescription(),
			qty:       fmt.Sprintf("%d", it.GetAmount()),
			unit:      it.GetUnit(),
			unitPrice: formatMoney(invoiceBody.GetCurrency(), unitPrice),
			vat:       vatLabel,
			amount:    formatMoney(invoiceBody.GetCurrency(), lineSubtotal),
			total:     formatMoney(invoiceBody.GetCurrency(), lineTotal),
		})
	}

	if invoiceBody.GetSubtotal() > 0 {
		subtotal = invoiceBody.GetSubtotal()
	}
	if invoiceBody.GetTotal() > 0 {
		grandTotal = invoiceBody.GetTotal()
		if subtotal > 0 && grandTotal >= subtotal {
			totalTax = grandTotal - subtotal
		}
	}
	amountDue := grandTotal
	if invoiceBody.GetStatus() == pb.BillingStatus_PAID {
		amountDue = 0
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	if name, w, _ := loadLogo(pdf, logoURL); name != "" {
		pdf.ImageOptions(name, 15, 15, w, 0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	}

	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetXY(15, 18)
	pdf.CellFormat(0, 10, fmt.Sprintf("Invoice # %s", nonEmpty(invoiceBody.GetNumber(), invoiceBody.GetUuid())), "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	// Status + dates
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(35, 5, "Status:", "", 0, "", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(40, 5, statusText(invoiceBody.GetStatus()), "", 0, "", false, 0, "")
	pdf.Ln(6)

	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(35, 5, "Issue date:", "", 0, "", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(40, 5, formatDate(tsToTime(invoiceBody.GetCreated())), "", 0, "", false, 0, "")
	if t := tsToTime(invoiceBody.GetPayment()); !t.IsZero() {
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(35, 5, "Payment date:", "", 0, "", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(40, 5, formatDate(t), "", 0, "", false, 0, "")
	}
	pdf.Ln(10)

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(229, 231, 235)
	pdf.SetLineWidth(0.2)

	startY := pdf.GetY()
	wFull := 180.0
	wCol := (wFull - 5) / 2

	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(wCol, 6, "Supplier:", "1", 0, "", true, 0, "")
	pdf.CellFormat(5, 6, "", "", 0, "", false, 0, "")
	pdf.CellFormat(wCol, 6, "Buyer:", "1", 1, "", true, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.MultiCell(wCol, 5, unescapeBR(supplier), "1", "L", false)
	x := 15 + wCol + 5
	pdf.SetXY(x, startY+6)
	pdf.MultiCell(wCol, 5, unescapeBR(buyer), "1", "L", false)
	pdf.Ln(2)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(wCol, 6, "Due date", "1", 0, "", true, 0, "")
	pdf.CellFormat(5, 6, "", "", 0, "", false, 0, "")
	pdf.CellFormat(wCol, 6, "Account:", "1", 1, "", true, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(wCol, 8, formatDate(tsToTime(invoiceBody.GetDeadline())), "1", 0, "", false, 0, "")
	pdf.CellFormat(5, 8, "", "", 0, "", false, 0, "")
	pdf.MultiCell(wCol, 8, strings.TrimSpace(gwExtraText), "1", "L", false)
	pdf.Ln(2)

	colW := []float64{8, 76, 12, 14, 22, 12, 18, 18}
	align := []string{"C", "L", "C", "C", "R", "C", "R", "R"}

	pdf.SetFont("Helvetica", "B", 10)
	headers := []string{"#", "Item", "Qty", "Unit", "Unit price", "VAT", "Amount", "Total"}
	for i, h := range headers {
		pdf.CellFormat(colW[i], 7, h, "1", 0, align[i], true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 10)
	for _, r := range rows {
		data := []string{r.idx, r.item, r.qty, r.unit, r.unitPrice, r.vat, r.amount, r.total}
		lineCounts := make([]int, len(data))
		hMax := 0.0
		for i, txt := range data {
			lines := pdf.SplitLines([]byte(txt), colW[i]-1)
			lineCounts[i] = len(lines)
			if float64(lineCounts[i])*6 > hMax {
				hMax = float64(lineCounts[i]) * 6
			}
		}
		y := pdf.GetY()
		x := 15.0
		for i, txt := range data {
			pdf.SetXY(x, y)
			pdf.MultiCell(colW[i], 6, txt, "1", align[i], false)
			x += colW[i]
		}
		if pdf.GetY() < y+hMax {
			pdf.SetY(y + hMax)
		}
	}

	pdf.Ln(2)
	rightBoxW := 90.0
	pdf.SetX(15 + 180 - rightBoxW)
	pdf.SetFont("Helvetica", "", 10)
	type totalRow struct{ label, val string }
	tRows := []totalRow{
		{"Subtotal", formatMoney(invoiceBody.GetCurrency(), subtotal)},
		{"Tax", formatMoney(invoiceBody.GetCurrency(), totalTax)},
		{"Grand Total", formatMoney(invoiceBody.GetCurrency(), grandTotal)},
	}
	for i, tr := range tRows {
		style := ""
		if i == len(tRows)-1 {
			style = "B"
		}
		pdf.SetFont("Helvetica", style, 10)
		pdf.CellFormat(rightBoxW*0.55, 7, tr.label, "1", 0, "L", false, 0, "")
		pdf.CellFormat(rightBoxW*0.45, 7, tr.val, "1", 1, "R", false, 0, "")
	}

	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 7, "Amount Due: "+formatMoney(invoiceBody.GetCurrency(), amountDue), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 6, "To pay: "+formatMoney(invoiceBody.GetCurrency(), grandTotal), "", 1, "L", false, 0, "")

	pdf.Ln(4)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 5, "Invoice ID: "+nonEmpty(invoiceBody.GetUuid(), invoiceBody.GetNumber()), "", 1, "L", false, 0, "")

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func nonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func unescapeBR(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	return html.UnescapeString(s)
}

type pg struct {
	Key   string
	Name  string
	Extra string
	HTML  string
}

func generateViewInvoiceHTML(invoiceBody *pb.Invoice, paymentGateways []*pb.PaymentGateway, supplier string, buyer string, logoURL string) string {
	if invoiceBody == nil {
		return "<!doctype html><html><head><meta charset='utf-8'><title>Invoice</title></head><body><p>No invoice data</p></body></html>"
	}

	tsToTime := func(ts int64) time.Time {
		if ts <= 0 {
			return time.Time{}
		}
		if ts > 1_000_000_000_000 {
			return time.UnixMilli(ts)
		}
		return time.Unix(ts, 0)
	}
	formatDate := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("02/01/2006")
	}
	intOrDefault := func(v, d int32) int32 {
		if v <= 0 {
			return d
		}
		return v
	}
	formatMoney := func(cur *pb.Currency, amount float64) string {
		precision := int(intOrDefault(cur.GetPrecision(), 2))
		pow := math.Pow(10, float64(precision))
		amount = math.Round(amount*pow) / pow

		num := fmt.Sprintf("%.*f", precision, amount)

		format := "{amount}"
		if cur != nil && strings.TrimSpace(cur.GetFormat()) != "" {
			format = cur.GetFormat()
		} else if cur != nil && strings.TrimSpace(cur.GetCode()) != "" {
			format = "{amount} " + cur.GetCode()
		}
		return strings.ReplaceAll(format, "{amount}", num)
	}
	escapeWithBR := func(s string) string {
		if s == "" {
			return ""
		}
		return strings.ReplaceAll(html.EscapeString(s), "\n", "<br>")
	}
	statusText := func(st pb.BillingStatus) string {
		switch st {
		case pb.BillingStatus_PAID:
			return "PAID"
		case pb.BillingStatus_UNPAID:
			return "UNPAID"
		case pb.BillingStatus_CANCELED:
			return "CANCELED"
		case pb.BillingStatus_TERMINATED:
			return "TERMINATED"
		case pb.BillingStatus_DRAFT:
			return "DRAFT"
		case pb.BillingStatus_RETURNED:
			return "RETURNED"
		default:
			return "UNKNOWN"
		}
	}

	var (
		rowsBuf     bytes.Buffer
		subtotal    float64
		totalTax    float64
		grandTotal  float64
		taxRate     = 0.0
		taxIncluded = false
	)
	if to := invoiceBody.GetTaxOptions(); to != nil {
		taxRate = to.GetTaxRate()
		taxIncluded = to.GetTaxIncluded()
	}

	for i, it := range invoiceBody.GetItems() {
		qty := float64(it.GetAmount())
		unitPrice := it.GetPrice()
		line := unitPrice * qty

		applyTax := it.GetApplyTax() && taxRate > 0
		var taxAmt, lineSubtotal, lineTotal float64
		var vatLabel string

		if applyTax {
			if taxIncluded {
				base := line / (1.0 + taxRate)
				taxAmt = line - base
				lineSubtotal = base
				lineTotal = line
			} else {
				taxAmt = line * taxRate
				lineSubtotal = line
				lineTotal = line + taxAmt
			}
			vatLabel = fmt.Sprintf("%.0f%%", taxRate*100)
		} else {
			taxAmt = 0
			lineSubtotal = line
			lineTotal = line
			vatLabel = "NP"
		}

		subtotal += lineSubtotal
		totalTax += taxAmt
		grandTotal += lineTotal

		rowsBuf.WriteString(fmt.Sprintf(
			`<tr>
				<td class="c">%d</td>
				<td class="item"><div class="descr">%s</div></td>
				<td class="c">%d</td>
				<td class="c">%s</td>
				<td class="r">%s</td>
				<td class="r">%s</td>
				<td class="c">%s</td>
				<td class="r">%s</td>
				<td class="r">%s</td>
			</tr>`,
			i+1,
			html.EscapeString(it.GetDescription()),
			it.GetAmount(),
			html.EscapeString(it.GetUnit()),
			formatMoney(invoiceBody.GetCurrency(), unitPrice),
			formatMoney(invoiceBody.GetCurrency(), unitPrice),
			vatLabel,
			formatMoney(invoiceBody.GetCurrency(), lineSubtotal),
			formatMoney(invoiceBody.GetCurrency(), lineTotal),
		))
	}

	if invoiceBody.GetSubtotal() > 0 {
		subtotal = invoiceBody.GetSubtotal()
	}
	if invoiceBody.GetTotal() > 0 {
		grandTotal = invoiceBody.GetTotal()
		if subtotal > 0 && grandTotal >= subtotal {
			totalTax = grandTotal - subtotal
		}
	}

	amountDue := grandTotal
	if invoiceBody.GetStatus() == pb.BillingStatus_PAID {
		amountDue = 0
	}

	var enabled []pg
	for _, g := range paymentGateways {
		if g.GetEnabled() {
			enabled = append(enabled, pg{
				Key:   g.GetKey(),
				Name:  g.GetDisplayName(),
				Extra: g.GetCheckoutExtraText(),
				HTML:  g.GetCheckoutPanelHtml(),
			})
		}
	}
	if len(enabled) == 0 {
		for _, g := range paymentGateways {
			enabled = append(enabled, pg{
				Key:   g.GetKey(),
				Name:  g.GetDisplayName(),
				Extra: g.GetCheckoutExtraText(),
				HTML:  g.GetCheckoutPanelHtml(),
			})
		}
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Invoice %s</title>
<style>
:root{
	--fg:#111827;--muted:#6b7280;--line:#e5e7eb;--bg:#ffffff;--accent:#2563eb;--soft:#f9fafb;
	--chip:#f3f4f6;--ok:#16a34a;--warn:#ef4444;
}
*{box-sizing:border-box}
body{margin:0;background:var(--soft);color:var(--fg);font:14px/1.45 -apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Inter,Helvetica,Arial,sans-serif}
.wrapper{max-width:960px;margin:24px auto;padding:16px}
.card{background:var(--bg);box-shadow:0 1px 2px rgba(0,0,0,.04);border:1px solid var(--line);border-radius:10px;overflow:hidden}
.header{display:flex;gap:16px;align-items:center;justify-content:space-between;padding:18px 20px;border-bottom:1px solid var(--line);flex-wrap:wrap}
.brand{display:flex;align-items:center;gap:12px;min-width:220px}
.brand img{height:40px;object-fit:contain}
.h-meta{display:grid;gap:4px 16px;grid-template-columns:auto auto;align-items:center}
.h-meta .k{color:var(--muted)}
.h-actions{display:flex;gap:8px;align-items:center}
select,button{border:1px solid var(--line);background:#fff;border-radius:8px;padding:8px 10px;font:inherit}
button.primary{background:var(--accent);border-color:var(--accent);color:#fff}
.badge{padding:2px 8px;border-radius:999px;background:var(--chip);font-size:12px}
.badge.paid{background:#dcfce7;color:#065f46}
.badge.unpaid{background:#fee2e2;color:#991b1b}
.grid-2{display:grid;grid-template-columns:1fr 1fr;gap:16px;padding:16px 20px;border-bottom:1px solid var(--line)}
@media (max-width:700px){.grid-2{grid-template-columns:1fr}}
.box h4{margin:0 0 6px 0}
.box .muted{color:var(--muted);font-size:12px}
.info{display:grid;grid-template-columns:1fr 1fr;gap:16px;padding:0 20px 12px;border-bottom:1px solid var(--line)}
@media (max-width:700px){.info{grid-template-columns:1fr}}
.info .block{padding:12px;background:#fff}
.info .block h4{margin:0 0 8px 0}
.table{width:100%%;border-collapse:collapse;margin:0;padding:0}
.table-wrap{padding:12px 20px}
th,td{border:1px solid var(--line);padding:8px 10px;vertical-align:top}
th{background:#f8fafc;text-align:left;font-weight:600}
td.c{text-align:center}
td.r{text-align:right}
td.item .descr{white-space:pre-wrap}
tfoot td{font-weight:600}
.totals{display:grid;grid-template-columns:2fr 1fr;gap:16px;padding:8px 20px;border-top:1px solid var(--line);align-items:end}
@media (max-width:700px){.totals{grid-template-columns:1fr}}
.summary{display:grid;gap:8px}
.summary .row{display:flex;justify-content:space-between;border-bottom:1px dashed var(--line);padding:8px 0}
.pay{display:flex;justify-content:space-between;align-items:center;padding:12px 20px;border-top:1px solid var(--line);flex-wrap:wrap;gap:8px}
.pay .due{font-weight:700}
.small{font-size:12px;color:var(--muted)}
footer{padding:10px 20px}
hr.sep{border:0;border-top:1px solid var(--line);margin:0}
.note{padding:12px 20px}
</style>
</head>
<body>
<div class="wrapper">
<div class="card">

	<div class="header">
		<div class="brand">
			<img src="%s" alt="Logo" />
			<div>
				<div style="font-weight:700;font-size:18px;">Invoice # %s</div>
				<div class="small">Status: <span class="badge %s">%s</span></div>
			</div>
		</div>

		<div class="h-meta">
			<div class="k">Issue date:</div><div>%s</div>
			%s
			<div class="k">Payment method:</div>
			<div>
				<select id="paymentMethod"></select>
			</div>
            <div id="gatewayPanel"></div>
		</div>
	</div>

	<div class="grid-2">
		<div class="box">
			<h4>Supplier:</h4>
			<div>%s</div>
		</div>
		<div class="box">
			<h4>Buyer:</h4>
			<div>%s</div>
		</div>
	</div>

	<div class="info">
		<div class="block">
			<h4>Due date</h4>
			<div>%s</div>
		</div>
		<div class="block">
			<div id="gatewayExtra" class="small">—</div>
		</div>
	</div>

	<div class="table-wrap">
		<table class="table">
			<thead>
				<tr>
					<th class="c" style="width:36px">#</th>
					<th>Item</th>
					<th class="c" style="width:64px">Qty</th>
					<th class="c" style="width:64px">Unit</th>
					<th class="r" style="width:120px">Price</th>
					<th class="r" style="width:120px">Unit Price</th>
					<th class="c" style="width:80px">VAT</th>
					<th class="r" style="width:120px">Amount</th>
					<th class="r" style="width:120px">Total</th>
				</tr>
			</thead>
			<tbody>
				%s
			</tbody>
		</table>
	</div>

	<div class="totals">
		<div class="summary">
			<div class="row"><span>Subtotal</span><span>%s</span></div>
			<div class="row"><span>Tax</span><span>%s</span></div>
			<div class="row"><span>Grand Total</span><span>%s</span></div>
		</div>
	</div>

	<div class="pay">
		<div class="due">Amount Due: %s</div>
		<div>To pay: <strong>%s</strong></div>
	</div>

	<footer class="small">
		Invoice ID: %s
	</footer>
</div>
</div>

<script>
(function(){
	const currency = %q;
	const gateways = %s;

	function byId(id){return document.getElementById(id)}
	const sel = byId('paymentMethod');
	gateways.forEach(function(g,i){
		const opt = document.createElement('option');
		opt.value = g.key; opt.textContent = g.name || g.key;
		sel.appendChild(opt);
	});
	function renderGateway(key){
		const g = gateways.find(x=>x.key===key) || gateways[0];
		if(!g) return;
		byId('gatewayExtra').innerHTML = g.extra ? g.extra : '—';
		byId('gatewayPanel').innerHTML = g.html || '';
	}
	sel.addEventListener('change', function(){ renderGateway(this.value); });
	// init
	if(gateways.length>0){ sel.value = gateways[0].key; renderGateway(sel.value); }
})();
</script>

</body>
</html>`,
		"Title TODO",
		html.EscapeString(coalesce(logoURL, "")),
		html.EscapeString(coalesce(invoiceBody.GetNumber(), invoiceBody.GetUuid())),
		statusClass(invoiceBody.GetStatus()), statusText(invoiceBody.GetStatus()),
		formatDate(tsToTime(invoiceBody.GetCreated())),
		paymentDateHTML(invoiceBody.GetPayment(), tsToTime, formatDate),
		escapeWithBR(supplier),
		escapeWithBR(buyer),
		formatDate(tsToTime(invoiceBody.GetDeadline())),
		rowsBuf.String(),
		formatMoney(invoiceBody.GetCurrency(), subtotal),
		formatMoney(invoiceBody.GetCurrency(), totalTax),
		formatMoney(invoiceBody.GetCurrency(), grandTotal),
		formatMoney(invoiceBody.GetCurrency(), amountDue),
		formatMoney(invoiceBody.GetCurrency(), grandTotal),
		html.EscapeString(coalesce(invoiceBody.GetUuid(), "")),
		// JS data
		invoiceBody.GetCurrency().GetCode(),
		jsGateways(enabled),
	)

	return b.String()
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func statusClass(st pb.BillingStatus) string {
	switch st {
	case pb.BillingStatus_PAID:
		return "badge paid"
	case pb.BillingStatus_UNPAID, pb.BillingStatus_DRAFT:
		return "badge unpaid"
	default:
		return "badge"
	}
}

func paymentDateHTML(ts int64, tsToTime func(int64) time.Time, formatDate func(time.Time) string) string {
	t := tsToTime(ts)
	if t.IsZero() {
		return ""
	}
	return `<div class="k">Payment date:</div><div>` + formatDate(t) + `</div>`
}

func jsGateways(gws []pg) string {
	var parts []string
	for _, g := range gws {
		parts = append(parts, fmt.Sprintf(`{"key":%q,"name":%q,"extra":%q,"html":%q}`,
			g.Key, g.Name, g.Extra, g.HTML))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
