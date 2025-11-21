package locales

import (
	"fmt"
	"strings"

	"github.com/pariz/gountries"
	"golang.org/x/text/language"
)

func TranslateCountry(countryOrCode, languageCode string) (string, error) {
	countryOrCode = strings.TrimSpace(countryOrCode)
	languageCode = strings.TrimSpace(languageCode)

	if countryOrCode == "" {
		return "", fmt.Errorf("empty country")
	}
	if languageCode == "" {
		return "", fmt.Errorf("empty language code")
	}
	query := gountries.New()
	var (
		c   gountries.Country
		err error
	)

	upper := strings.ToUpper(countryOrCode)

	if len(upper) == 2 || len(upper) == 3 {
		c, err = query.FindCountryByAlpha(upper)
		if err != nil {
			c, err = query.FindCountryByName(countryOrCode)
		}
	} else {
		c, err = query.FindCountryByName(countryOrCode)
	}

	if err != nil {
		return "", fmt.Errorf("country %q not found: %w", countryOrCode, err)
	}

	base, err := language.ParseBase(languageCode)
	if err != nil {
		return "", fmt.Errorf("invalid language code %q: %w", languageCode, err)
	}
	iso3 := strings.ToUpper(base.ISO3())

	if tr, ok := c.Translations[iso3]; ok && tr.Common != "" {
		return tr.Common, nil
	}

	if strings.ToLower(languageCode) == strings.ToLower(c.Alpha2) {
		for _, tr := range c.Name.Native {
			if tr.Common != "" {
				return tr.Common, nil
			}
		}
	}
	return c.Name.Common, nil
}

func TranslateCountryMust(countryOrCode, languageCode string) string {
	name, err := TranslateCountry(countryOrCode, languageCode)
	if err != nil {
		return countryOrCode
	}
	return name
}
