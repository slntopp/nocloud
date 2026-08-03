/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package registry

import (
	"strings"

	"github.com/pariz/gountries"
)

const vatRate = 0.23

// VAT rules for a Polish IT company (see repo doc "НДС для польских IT-услуг в ЕС.pdf"):
//   - Poland: always 23%
//   - EU + VAT ID (B2B): NP — reverse charge, no Polish VAT
//   - EU + no VAT ID (B2C / individual): 23%
//   - Outside EU: always NP — no Polish VAT
//
// NP (Nie podlega) is stored as tax_rate = 0; invoice templates render it as "NP", not "0%".
var euCountries = map[string]struct{}{
	"AT": {}, "BE": {}, "BG": {}, "HR": {}, "CY": {}, "CZ": {}, "DK": {},
	"EE": {}, "FI": {}, "FR": {}, "DE": {}, "GR": {}, "HU": {}, "IE": {},
	"IT": {}, "LV": {}, "LT": {}, "LU": {}, "MT": {}, "NL": {}, "PL": {},
	"PT": {}, "RO": {}, "SK": {}, "SI": {}, "ES": {}, "SE": {},
}

func normalizeCountryCode(country string) string {
	country = strings.TrimSpace(country)
	if country == "" {
		return ""
	}

	query := gountries.New()
	upper := strings.ToUpper(country)

	if len(upper) == 2 || len(upper) == 3 {
		c, err := query.FindCountryByAlpha(upper)
		if err == nil {
			return strings.ToUpper(c.Alpha2)
		}
	}

	c, err := query.FindCountryByName(country)
	if err == nil {
		return strings.ToUpper(c.Alpha2)
	}

	return upper
}

func isEUCountry(code string) bool {
	_, ok := euCountries[code]
	return ok
}

// CalculateTaxRate returns the account VAT rate (0.23 or 0).
// A return value of 0 means NP on invoices (reverse charge or export outside PL VAT scope).
func CalculateTaxRate(country, taxID string) float64 {
	code := normalizeCountryCode(country)
	taxID = strings.TrimSpace(taxID)

	if code == "PL" {
		return vatRate
	}
	if !isEUCountry(code) {
		return 0 // NP: outside EU
	}
	if taxID != "" {
		return 0 // NP: EU B2B with VAT ID (reverse charge)
	}
	return vatRate // EU B2C / individual without VAT ID
}

func applyTaxRate(data map[string]interface{}) {
	if data == nil {
		return
	}
	country, _ := data["country"].(string)
	taxID, _ := data["tax_id"].(string)
	data["tax_rate"] = CalculateTaxRate(country, taxID)
}
