package aswords

import (
	"fmt"
	"strconv"
	"strings"

	ntw "github.com/maitres/number-to-words"
)

type Language string

const (
	LangEn Language = "en"
	LangRu Language = "ru"
	LangPl Language = "pl"
)

type FractionStyle int

const (
	FractionStylePoint FractionStyle = iota
	FractionStyleFractional
)

type Currency struct {
	Code          string
	MajorSingular string
	MajorPaucal   string
	MajorPlural   string

	MinorSingular string
	MinorPaucal   string
	MinorPlural   string

	MinorUnit int
}

var (
	CurrencyRUB = Currency{
		Code:          "RUB",
		MajorSingular: "рубль",
		MajorPaucal:   "рубля",
		MajorPlural:   "рублей",

		MinorSingular: "копейка",
		MinorPaucal:   "копейки",
		MinorPlural:   "копеек",

		MinorUnit: 2,
	}

	CurrencyPLN = Currency{
		Code:          "PLN",
		MajorSingular: "złoty",
		MajorPaucal:   "złote",
		MajorPlural:   "złotych",

		MinorSingular: "grosz",
		MinorPaucal:   "grosze",
		MinorPlural:   "groszy",

		MinorUnit: 2,
	}

	CurrencyUSD = Currency{
		Code:          "USD",
		MajorSingular: "dollar",
		MajorPaucal:   "dollars",
		MajorPlural:   "dollars",

		MinorSingular: "cent",
		MinorPaucal:   "cents",
		MinorPlural:   "cents",

		MinorUnit: 2,
	}
)

func AmountToWords(amount string, lang Language, style FractionStyle) (string, error) {
	negative, intStr, fracStr, err := normalizeAmountString(amount)
	if err != nil {
		return "", err
	}

	intVal, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse integer part: %w", err)
	}

	intWords := intToWords(intVal, lang)
	prefix := negativePrefix(lang, negative)

	if fracStr == "" || isAllZeros(fracStr) {
		return prefix + intWords, nil
	}

	fracTrimmed := strings.TrimRight(fracStr, "0")
	if fracTrimmed == "" {
		return prefix + intWords, nil
	}

	fracVal, err := strconv.ParseInt(fracTrimmed, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse fractional part: %w", err)
	}
	fracWords := intToWords(fracVal, lang)

	switch style {
	case FractionStyleFractional:
		scale := len(fracTrimmed)
		denom := denomWord(lang, scale)
		if denom == "" {
			connector := pointWord(lang)
			return fmt.Sprintf("%s%s %s %s", prefix, intWords, connector, fracWords), nil
		}
		switch lang {
		case LangRu:
			return fmt.Sprintf("%s%s целых %s %s", prefix, intWords, fracWords, denom), nil
		case LangPl:
			return fmt.Sprintf("%s%s całych %s %s", prefix, intWords, fracWords, denom), nil
		default: // English
			// "five and thirty-four hundredths"
			return fmt.Sprintf("%s%s and %s %s", prefix, intWords, fracWords, denom), nil
		}

	default:
		connector := pointWord(lang)
		return fmt.Sprintf("%s%s %s %s", prefix, intWords, connector, fracWords), nil
	}
}

func CurrencyToWords(amount string, lang Language, cur Currency) (string, error) {
	if cur.MinorUnit < 0 || cur.MinorUnit > 9 {
		return "", fmt.Errorf("invalid currency MinorUnit: %d", cur.MinorUnit)
	}

	negative, intStr, fracStr, err := normalizeAmountString(amount)
	if err != nil {
		return "", err
	}

	majorVal, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse integer part: %w", err)
	}

	majorWords := intToWords(majorVal, lang)
	majorNoun := pluralForm(lang, majorVal,
		cur.MajorSingular, cur.MajorPaucal, cur.MajorPlural)

	minorStr := normalizeMinor(fracStr, cur.MinorUnit)

	var minorVal int64
	var minorWords, minorNoun string

	if minorStr != "" && !isAllZeros(minorStr) {
		minorVal, err = strconv.ParseInt(minorStr, 10, 64)
		if err != nil {
			return "", fmt.Errorf("parse minor part: %w", err)
		}
		minorWords = intToWords(minorVal, lang)
		minorNoun = pluralForm(lang, minorVal,
			cur.MinorSingular, cur.MinorPaucal, cur.MinorPlural)
	}

	prefix := negativePrefix(lang, negative)

	if minorNoun == "" || minorVal == 0 {
		return fmt.Sprintf("%s%s %s", prefix, majorWords, majorNoun), nil
	}

	switch lang {
	case LangEn:
		return fmt.Sprintf("%s%s %s and %s %s",
			prefix, majorWords, majorNoun, minorWords, minorNoun), nil
	default:
		return fmt.Sprintf("%s%s %s %s %s",
			prefix, majorWords, majorNoun, minorWords, minorNoun), nil
	}
}

func normalizeAmountString(amount string) (negative bool, intPart, fracPart string, err error) {
	s := strings.TrimSpace(amount)
	if s == "" {
		err = fmt.Errorf("empty amount")
		return
	}

	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	if s == "" {
		err = fmt.Errorf("invalid amount %q", amount)
		return
	}

	if s[0] == '-' {
		negative = true
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}
	if s == "" {
		err = fmt.Errorf("invalid amount %q", amount)
		return
	}

	parts := strings.SplitN(s, ".", 2)
	intPart = parts[0]
	if intPart == "" {
		intPart = "0"
	}
	if !allDigits(intPart) {
		err = fmt.Errorf("invalid integer part %q", intPart)
		return
	}

	if len(parts) == 2 {
		fracPart = parts[1]
		for i, r := range fracPart {
			if r < '0' || r > '9' {
				fracPart = fracPart[:i]
				break
			}
		}
	}

	return
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isAllZeros(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if r != '0' {
			return false
		}
	}
	return true
}

func intToWords(n int64, lang Language) string {
	v := int(n)

	switch lang {
	case LangRu:
		return ntw.IntegerToRuRu(v)
	case LangPl:
		return ntw.IntegerToPlPl(v)
	case LangEn:
		fallthrough
	default:
		return ntw.IntegerToEnUs(v)
	}
}

func negativePrefix(lang Language, negative bool) string {
	if !negative {
		return ""
	}
	switch lang {
	case LangRu:
		return "минус "
	case LangPl, LangEn:
		return "minus "
	default:
		return "minus "
	}
}

func pointWord(lang Language) string {
	switch lang {
	case LangRu:
		return "точка"
	case LangPl:
		return "przecinek"
	default:
		return "point"
	}
}

func denomWord(lang Language, scale int) string {
	if scale <= 0 {
		return ""
	}

	switch lang {
	case LangRu:
		switch scale {
		case 1:
			return "десятых"
		case 2:
			return "сотых"
		case 3:
			return "тысячных"
		default:
			return ""
		}
	case LangPl:
		switch scale {
		case 1:
			return "dziesiątych"
		case 2:
			return "setnych"
		case 3:
			return "tysięcznych"
		default:
			return ""
		}
	default:
		switch scale {
		case 1:
			return "tenths"
		case 2:
			return "hundredths"
		case 3:
			return "thousandths"
		default:
			return ""
		}
	}
}

func normalizeMinor(fracStr string, minorUnit int) string {
	if minorUnit <= 0 {
		return ""
	}
	for i, r := range fracStr {
		if r < '0' || r > '9' {
			fracStr = fracStr[:i]
			break
		}
	}

	if fracStr == "" {
		return strings.Repeat("0", minorUnit)
	}

	if len(fracStr) >= minorUnit {
		return fracStr[:minorUnit]
	}

	return fracStr + strings.Repeat("0", minorUnit-len(fracStr))
}

func pluralForm(lang Language, n int64, singular, paucal, plural string) string {
	if singular == "" && paucal == "" && plural == "" {
		return ""
	}

	switch lang {
	case LangRu, LangPl:
		lastTwo := n % 100
		last := n % 10

		if lastTwo >= 11 && lastTwo <= 14 {
			if plural != "" {
				return plural
			}
			if paucal != "" {
				return paucal
			}
			return singular
		}

		if last == 1 {
			if singular != "" {
				return singular
			}
			if paucal != "" {
				return paucal
			}
			return plural
		}

		if last >= 2 && last <= 4 {
			if paucal != "" {
				return paucal
			}
			if plural != "" {
				return plural
			}
			return singular
		}

		if plural != "" {
			return plural
		}
		if paucal != "" {
			return paucal
		}
		return singular

	default:
		if n == 1 {
			if singular != "" {
				return singular
			}
			if paucal != "" {
				return paucal
			}
			return plural
		}
		if plural != "" {
			return plural
		}
		if paucal != "" {
			return paucal
		}
		return singular
	}
}
