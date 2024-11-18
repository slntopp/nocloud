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

    FILTER t.meta.transactionType != "transaction"

    LET deadline = TO_NUMBER(t.meta.duedate)
    LET account = t.account
    LET created = TO_NUMBER(t.created)
    LET currency = t.currency
    LET total = t.total < 0 ? -t.total : t.total
    LET processed = t.proc > 0 ? TO_NUMBER(t.proc) : null
    LET payment = t.exec > 0 ? TO_NUMBER(t.exec) : null
    LET note = t.meta.note
    LET status = t.meta.status == "Paid" || t.meta.status == "Paid balance" ? 1 : 2

    FILTER t.meta.transactionType == "invoice top-up" || t.meta.transactionType == "invoice payment" || t.meta.transactionType == "invoice for service"

    LET typesObject = { "invoice top-up": 4, "invoice payment": 1, "invoice for service": 3 }
    LET typeB = typesObject[t.meta.transactionType]
    LET type = typeB == 3 ? @invTypes[newKey] == "first" ? 2 : typeB : typeB

    LET transactions = TO_ARRAY(t._key)

    LET whmcs_invoice_id = TO_NUMBER(@whmcsIds[newKey])
    FILTER whmcs_invoice_id > 0

    LET instancesArr = TO_ARRAY(invInstances[TO_STRING(whmcs_invoice_id)])

    LET items0 = (
        FILTER IS_ARRAY(t.meta.items)
        FOR item IN t.meta.items
            RETURN { amount: 1, price: TO_NUMBER(item.amount), description: item.description, unit: "Pcs" }
    )

    LET items1 = LENGTH(instancesArr) != 1 || LENGTH(items0) != 1 ? items0 : (
		RETURN [ MERGE(items0[0], {instance: instancesArr[0]}) ]
    )
    LET items2 = LENGTH(instancesArr) != 2 || LENGTH(items1) != 2 ? items1 : (
		RETURN [ MERGE(items1[0], {instance: instancesArr[0]}), MERGE(items1[1], {instance: instancesArr[1]}) ]
    )
    LET items3 = LENGTH(instancesArr) != 3 || LENGTH(items2) != 3 ? items2 : (
		RETURN [ MERGE(items2[0], {instance: instancesArr[0]}), MERGE(items2[1], {instance: instancesArr[1]}), MERGE(items2[2], {instance: instancesArr[2]}) ]
    )
    LET items4 = LENGTH(instancesArr) != 1 || LENGTH(items3) != 2 ? items3 : (
		RETURN [ MERGE(items3[0], {instance: instancesArr[0]}), items3[1] ]
    )
    LET items5 = LENGTH(instancesArr) != 1 || LENGTH(items4) != 3 ? items4 : (
		RETURN [ MERGE(items4[0], {instance: instancesArr[0]}), items4[1], items4[2] ]
    )

	LET items = items5

	LET invoice = {
		_key: newKey,
        deadline: deadline,
		created: created,
		currency: currency,
		total: total,
		processed: processed,
		payment: payment,
		meta: { note: note, whmcs_invoice_id: whmcs_invoice_id },
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

func MigrateOldInvoicesToNew(log *zap.Logger, invoices driver.Collection, transactions driver.Collection, whmcsInvoicesFile string, whmcsInstancesFile string) {
	log.Info("Migrating old transaction invoices to new")
	idToWhmcsId := make(map[string]string)
	invTypes := make(map[string]string)
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
	if len(records) == 0 {
		log.Warn("No records found in " + whmcsInvoicesFile)
		return
	}
	const id, whmcsId, invType = 1, 2, 3
	log.Debug("First record", zap.Any("record", records[0]))
	for i := 1; i < len(records); i++ {
		log.Debug("Record", zap.Any("record", records[i]))
		idToWhmcsId[records[i][id]] = records[i][whmcsId]
		invTypes[records[i][id]] = records[i][invType]
	}

	invInstances := make(map[string][]string)
	file, err = os.Open(whmcsInstancesFile)
	if err != nil {
		log.Fatal("Error migrating old invoices to new", zap.Error(err))
	}
	defer file.Close()
	csvReader = csv.NewReader(file)
	records, err = csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+whmcsInstancesFile, zap.Error(err))
	}
	if len(records) == 0 {
		log.Warn("No records found in " + whmcsInstancesFile)
		return
	}
	const instance, invId = 2, 3
	log.Debug("First record", zap.Any("record", records[0]))
	for i := 1; i < len(records); i++ {
		log.Debug("Record", zap.Any("record", records[i]))
		arr := invInstances[records[i][invId]]
		if arr == nil {
			arr = make([]string, 0)
		}
		arr = append(arr, records[i][instance])
		invInstances[records[i][invId]] = arr
	}

	db := invoices.Database()
	_, err = db.Query(context.TODO(), createInvoicesBasedOnOldTransactions, map[string]interface{}{
		"@invoices":     invoices.Name(),
		"@transactions": transactions.Name(),
		"whmcsIds":      idToWhmcsId,
		"invTypes":      invTypes,
		"invInstances":  invInstances,
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
