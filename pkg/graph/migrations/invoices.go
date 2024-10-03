package migrations

import (
	"context"
	"encoding/csv"
	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
	"os"
)

const createInvoicesBasedOnOldTransactions = `
FOR t IN @@transactions
    FILTER CONTAINS(t.meta.transactionType, "invoice")
    LET newKey = t._key
    LET existingInvoice = DOCUMENT(@@invoices, newKey)
    FILTER existingInvoice == null

    LET deadline = TO_NUMBER(t.meta.duedate)
    LET account = t.account
    LET created = TO_NUMBER(t.created)
    LET currency = t.currency
    LET total = t.total < 0 ? -t.total : t.total
    LET processed = t.proc > 0 ? TO_NUMBER(t.proc) : null
    LET payment = t.exec > 0 ? TO_NUMBER(t.exec) : null
    LET note = t.meta.note
    LET status = t.meta.status == "Paid" ? 1 : 2

    LET typesObject = { "invoice top-up": 4, "invoice payment": 1, "invoice for service": 1 }
    LET type = typesObject[t.meta.transactionType]

    LET transactions = TO_ARRAY(t._key)

    LET instance = LENGTH(t.meta.instances) > 0 ? TO_STRING(t.meta.instances[0]) : null
    LET items = (
        FILTER IS_ARRAY(t.items)
        FOR item IN t.items
            RETURN { amount: 1, price: TO_NUMBER(item.amount), description: item.description, unit: "Pcs", instance: instance }
    )

	LET invoice = {
		_key: newKey,
        deadline: deadline,
		created: created,
		currency: currency,
		total: total,
		processed: processed,
		payment: payment,
		meta: { note: note, whmcs_invoice_id: @whmcsIds[newKey] },
		status: status,
		type: type,
		account: account,
		transactions: transactions,
		items: items,
        number: "Legacy invoice",
        numeric_number: 0,
        number_template: "Legacy invoice",
	}

	INSERT invoice INTO @@invoices
`

const clearNulls = `
FOR invoice IN @@invoices
UPDATE invoice._key WITH invoice IN @@invoices
OPTIONS { keepNull: false }
`

func MigrateOldInvoicesToNew(log *zap.Logger, invoices driver.Collection, transactions driver.Collection, whmcsInvoicesFile string) {
	log.Info("Migrating old transaction invoices to new")
	idToWhmcsId := make(map[string]string)
	file, err := os.Open(whmcsInvoicesFile)
	if err != nil {
		log.Fatal("Error migrating old invoices to new", zap.Error(err))
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+whmcsInvoicesFile, zap.Error(err))
	}
	const id, whmcsId = 1, 2
	log.Debug("First record", zap.Any("record", records[0]))
	for i := 1; i < len(records); i++ {
		log.Debug("Record", zap.Any("record", records[i]))
		idToWhmcsId[records[i][id]] = records[i][whmcsId]
	}

	db := invoices.Database()
	_, err = db.Query(context.TODO(), createInvoicesBasedOnOldTransactions, map[string]interface{}{
		"@invoices":     invoices.Name(),
		"@transactions": transactions.Name(),
		"@whmcsIds":     idToWhmcsId,
	})
	if err != nil {
		log.Fatal("Error migrating old invoices to new", zap.Error(err))
	}
	_, err = db.Query(context.TODO(), clearNulls, map[string]interface{}{
		"@invoices": invoices.Name(),
	})
	if err != nil {
		log.Fatal("Error migrating old invoices to new", zap.Error(err))
	}
	log.Info("Migrated old invoices to new")
}
