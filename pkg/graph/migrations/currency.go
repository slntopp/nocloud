package migrations

import (
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

var LEGACY_CURRENCIES = []*pb.Currency{
	{Id: schema.DEFAULT_CURRENCY_ID, Name: schema.DEFAULT_CURRENCY_NAME},
	{Id: 1, Name: "USD"},
	{Id: 2, Name: "EUR"},
	{Id: 3, Name: "BYN"},
	{Id: 4, Name: "PLN"},
}

func UpdateNumericCurrencyToDynamic(col driver.Collection) {

}
