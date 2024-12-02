package billing

import pb "github.com/slntopp/nocloud-proto/billing"

const BILLING_EVENTS = "billing"

const (
	InvoiceCreated    = "invoice_created"
	InvoiceUpdated    = "invoice_updated"
	InvoicePaid       = "invoice_paid"
	InvoiceDraft      = "invoice_draft"
	InvoiceUnpaid     = "invoice_unpaid"
	InvoiceCancelled  = "invoice_cancelled"
	InvoiceTerminated = "invoice_terminated"
	InvoiceReturned   = "invoice_returned"
)

var stToKey = map[pb.BillingStatus]string{
	pb.BillingStatus_DRAFT:      InvoiceDraft,
	pb.BillingStatus_UNPAID:     InvoiceUnpaid,
	pb.BillingStatus_PAID:       InvoicePaid,
	pb.BillingStatus_CANCELED:   InvoiceCancelled,
	pb.BillingStatus_TERMINATED: InvoiceTerminated,
	pb.BillingStatus_RETURNED:   InvoiceReturned,
}

func InvoiceStatusToKey(status pb.BillingStatus) string {
	return stToKey[status]
}

func Topic(key string) string {
	return BILLING_EVENTS + "." + key
}
