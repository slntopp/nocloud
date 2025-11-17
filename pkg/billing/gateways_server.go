package billing

import (
	"bytes"
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/gorilla/mux"
	accesspb "github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/invoicei18n"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/aswords"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/rest_auth"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"html"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
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
		for k, v := range gw.GetLanguageDisplay() {
			rawHtml := v.CheckoutPanelHtml
			rawHtml = strings.ReplaceAll(rawHtml, "$ACTION_URL", os.Getenv("BASE_HOST")+"/billing/payments/"+gw.Key+"/"+invoice.Uuid+"/action")
			token, _ := auth.MakeToken(invoice.Account)
			rawHtml = strings.ReplaceAll(rawHtml, "$ACCESS_TOKEN", url.QueryEscape(token))
			v.CheckoutPanelHtml = rawHtml
			gw.GetLanguageDisplay()[k] = v
		}
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

	languageCode := viper.GetString("PRIMARY_LANGUAGE_CODE")
	if account.GetLanguageCode() != "" {
		languageCode = account.GetLanguageCode()
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(generateViewInvoiceHTML(InvoiceBody, Gateways, Supplier, Buyer, LogoURL, languageCode, invoice.GetStatus() != pb.BillingStatus_UNPAID, false)))
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

	languageCode := viper.GetString("PRIMARY_LANGUAGE_CODE")
	if account.GetLanguageCode() != "" {
		languageCode = account.GetLanguageCode()
	}

	switch PaymentGatewayType(pgKey) {
	case PaymentGatewayFacture:
		s.log.Debug("Handling Facture Payment Action", zap.String("invoice_uuid", invoiceUuid))
		gws, err := s.ctrl.List(context.Background(), true)
		if err != nil {
			http.Error(writer, "Failed to list gateways: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = s.ctrl.Get(request.Context(), string(PaymentGatewayFacture))
		if err != nil {
			http.Error(writer, "Failed to get payment gateway: "+err.Error(), http.StatusInternalServerError)
			return
		}
		pdfBytes, err := GenerateInvoicePDF(InvoiceBody, gws, Supplier, Buyer, LogoURL, languageCode)
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

func GenerateInvoicePDF(invoiceBody *pb.Invoice, paymentGateways []*pb.PaymentGateway, supplier, buyer, logoURL, lang string) ([]byte, error) {
	htmlRaw := generateViewInvoiceHTML(invoiceBody, paymentGateways, supplier, buyer, logoURL, lang, true, true)
	gotenbergHost := viper.GetString("GOTENBERG_HOST")
	if gotenbergHost == "" {
		return nil, fmt.Errorf("GOTENBERG_HOST is empty")
	}
	if !strings.HasPrefix(gotenbergHost, "http") {
		gotenbergHost = fmt.Sprintf("http://%s", gotenbergHost)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart file field: %w", err)
	}
	if _, err = io.Copy(part, strings.NewReader(htmlRaw)); err != nil {
		return nil, fmt.Errorf("failed to write HTML to multipart: %w", err)
	}

	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	urlPath := strings.TrimRight(gotenbergHost, "/") + "/forms/chromium/convert/html"
	req, err := http.NewRequest(http.MethodPost, urlPath, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to Gotenberg: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gotenberg: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gotenberg response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errText := strings.TrimSpace(string(body))
		if errText == "" {
			errText = fmt.Sprintf("gotenberg returned status %d with empty body", resp.StatusCode)
		}
		return nil, fmt.Errorf("gotenberg error: %s", errText)
	}

	return body, nil
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

func generateViewInvoiceHTML(invoiceBody *pb.Invoice, paymentGateways []*pb.PaymentGateway, supplier string, buyer string, logoURL string, lang string, omitPmPanel bool, omitGwPanel bool) string {
	l := invoicei18n.Lang(lang)

	statusKey := func(st pb.BillingStatus) string {
		switch st {
		case pb.BillingStatus_PAID:
			return "$invoice.status.paid"
		case pb.BillingStatus_UNPAID:
			return "$invoice.status.unpaid"
		case pb.BillingStatus_CANCELED:
			return "$invoice.status.canceled"
		case pb.BillingStatus_TERMINATED:
			return "$invoice.status.terminated"
		case pb.BillingStatus_DRAFT:
			return "$invoice.status.draft"
		case pb.BillingStatus_RETURNED:
			return "$invoice.status.returned"
		default:
			return "$invoice.status.unknown"
		}
	}

	if invoiceBody == nil {
		return "<!doctype html><html><head><meta charset='utf-8'><title>$invoice.title</title></head><body><p>$invoice.no_data</p></body></html>"
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
		return strings.ReplaceAll(html.EscapeString(s), "\n", "<br/>")
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
			var display = &pb.PaymentGatewayDisplay{}
			if languageDisplay, ok := g.GetLanguageDisplay()[lang]; ok {
				display = languageDisplay
			} else if languageDisplay, ok = g.GetLanguageDisplay()[string(invoicei18n.DefaultLang)]; ok {
				display = languageDisplay
			} else if len(g.GetLanguageDisplay()) > 0 {
				display = g.GetLanguageDisplay()[maps.Keys(g.GetLanguageDisplay())[0]]
			}
			enabled = append(enabled, pg{
				Key:   g.GetKey(),
				Name:  display.GetDisplayName(),
				Extra: display.GetCheckoutExtraText(),
				HTML:  display.GetCheckoutPanelHtml(),
			})
		}
	}

	gwPanelHtml := `<div class="h-actions" id="gatewayPanel"></div>`
	if omitGwPanel {
		gwPanelHtml = `<div style="display:none" class="h-actions" id="gatewayPanel"></div>`
	}
	pmHtml := `<div class="k">$invoice.payment_method</div>
<div>
<select id="paymentMethod"></select>
</div>`
	if omitPmPanel {
		pmHtml = `<div class="k">
<span>$invoice.payment_method: </span><span id="gatewayName"></span>
</div>`
	}

	var titleKey = "$invoice.title"
	if invoiceBody.GetStatus() == pb.BillingStatus_PAID || invoiceBody.GetStatus() == pb.BillingStatus_RETURNED {
		titleKey = "$invoice.title_paid"
	}

	format := func(x float64) string {
		floored := math.Floor(x*100) / 100
		return fmt.Sprintf("%.2f", floored)
	}
	var totalAsWords string
	totalAsWords, err := aswords.AmountToWords(format(invoiceBody.Total), aswords.Language(lang), aswords.FractionStylePoint)
	if err != nil {
		fmt.Printf("ERROR: Formatting total as words: %s\n", err.Error())
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, `<!doctype html>
<html lang="%s">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
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
				<div style="font-weight:700;font-size:18px;">$invoice.title # %s</div>
				<div class="small">$invoice.status_label <span class="badge %s">%s</span></div>
			</div>
		</div>

		<div class="h-meta">
			<div class="k">$invoice.issue_date</div><div>%s</div>
			%s
			%s
		</div>

        %s
	</div>

	<div class="grid-2">
		<div class="box">
			<h4>$invoice.supplier</h4>
			<div>%s</div>
		</div>
		<div class="box">
			<h4>$invoice.buyer</h4>
			<div>%s</div>
		</div>
	</div>

	<div class="info">
		<div class="block">
			<h4>$invoice.due_date</h4>
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
					<th>$table.item</th>
					<th class="c" style="width:64px">$table.qty</th>
					<th class="c" style="width:64px">$table.unit</th>
					<th class="r" style="width:120px">$table.price</th>
					<th class="r" style="width:120px">$table.unit_price</th>
					<th class="c" style="width:80px">$table.vat</th>
					<th class="r" style="width:120px">$table.amount</th>
					<th class="r" style="width:120px">$table.total</th>
				</tr>
			</thead>
			<tbody>
				%s
			</tbody>
		</table>
	</div>

	<div class="totals">
		<div class="summary">
			<div class="row"><span>$summary.subtotal</span><span>%s</span></div>
			<div class="row"><span>$summary.tax</span><span>%s</span></div>
			<div class="row"><span>$summary.grand_total</span><span>%s</span></div>
		</div>
	</div>

	<div class="pay">
		<div class="due">$summary.amount_due: %s</div>
		<div>$summary.to_pay: <strong>%s</strong></div>
	</div>

    <div class="pay">
		<div><strong>%s</strong></div>
	</div>

	<footer class="small">
		$footer.invoice_id: %s
	</footer>
</div>
</div>

<script>
(function(){
	const currency = %q;
	const gateways = %s;

    function byId(id){return document.getElementById(id)};

    const defaultGwKey = "%s";
    const defaultGw = gateways.find(x=>x.key===defaultGwKey) || gateways[0];
    let gwName = byId('gatewayName') ? byId('gatewayName') : {};
    gwName.value = defaultGw.name;

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
		l,
		titleKey,
		html.EscapeString(coalesce(logoURL, "")),
		html.EscapeString(coalesce(invoiceBody.GetNumber(), invoiceBody.GetUuid())),
		statusClass(invoiceBody.GetStatus()), statusKey(invoiceBody.GetStatus()),
		formatDate(tsToTime(invoiceBody.GetCreated())),
		paymentDateHTML(invoiceBody.GetPayment(), tsToTime, formatDate),
		pmHtml,
		gwPanelHtml,
		escapeWithBR(supplier),
		escapeWithBR(buyer),
		formatDate(tsToTime(invoiceBody.GetDeadline())),
		rowsBuf.String(),
		formatMoney(invoiceBody.GetCurrency(), subtotal),
		formatMoney(invoiceBody.GetCurrency(), totalTax),
		formatMoney(invoiceBody.GetCurrency(), grandTotal),
		formatMoney(invoiceBody.GetCurrency(), amountDue),
		formatMoney(invoiceBody.GetCurrency(), grandTotal),
		totalAsWords,
		html.EscapeString(coalesce(invoiceBody.GetUuid(), "")),
		// JS data
		invoiceBody.GetCurrency().GetCode(),
		jsGateways(enabled),
		invoiceBody.GetPaymentGateway(),
	)

	return invoicei18n.Replace(l, b.String())
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
	return `<div class="k">$invoice.payment_date:</div><div>` + formatDate(t) + `</div>`
}

func jsGateways(gws []pg) string {
	escapeWithBR := func(s string) string {
		if s == "" {
			return ""
		}
		return strings.ReplaceAll(html.EscapeString(s), "\n", "<br/>")
	}
	var parts []string
	for _, g := range gws {
		parts = append(parts, fmt.Sprintf(`{"key":%q,"name":%q,"extra":%q,"html":%q}`,
			g.Key, g.Name, escapeWithBR(g.Extra), g.HTML))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
