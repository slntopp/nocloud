package types

import pb "github.com/slntopp/nocloud-proto/billing"

type CreateUserParams struct {
	Firstname       string       `json:"firstname" validate:"required"`
	Lastname        string       `json:"lastname" validate:"required"`
	Email           string       `json:"email" validate:"required,email"`
	Language        string       `json:"language"`
	Country         string       `json:"country" validate:"omitempty,iso3166_1_alpha2"`
	PhoneNumber     string       `json:"phonenumber" validate:"required"`
	Password        string       `json:"password2" validate:"required"`
	Currency        *pb.Currency `json:"currency"`
	CompanyName     string       `json:"companyname"`
	City            string       `json:"city" validate:"required"`
	Postcode        string       `json:"postcode" validate:"required"`
	Address         string       `json:"address1" validate:"required"`
	NoEmail         bool         `json:"noemail"`
	AccountNumber   string       `json:"account_number"`
	BankName        string       `json:"bankname"`
	BIC             string       `json:"bic"`
	CheckingAccount string       `json:"checking_account"`
	TaxID           string       `json:"tax_id"`
}
