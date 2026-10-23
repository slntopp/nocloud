package billing

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

type lineInst struct{ title string }

func (i lineInst) GetTitle() string                         { return i.title }
func (i lineInst) GetResources() map[string]*structpb.Value { return nil }
func (i lineInst) GetConfig() map[string]*structpb.Value    { return nil }

func TestInvoiceLineServiceName(t *testing.T) {
	cases := []struct{ prefix, product, inst, want string }{
		// Callers pass the prefix with a trailing space, and an empty one as " ".
		{"Услуга ИИ-коннект - ", "AI СТАРТ ", "AI СТАРТ", "Услуга ИИ-коннект - AI СТАРТ"},
		{"Услуга ИИ-коннект - ", "AI СТАРТ ", "my-bot", "Услуга ИИ-коннект - AI СТАРТ - my-bot"},
		{" ", "AI СТАРТ ", "AI СТАРТ", "AI СТАРТ - AI СТАРТ"},
		{" ", "VDS Start ", "my-server", "VDS Start - my-server"},
	}
	for _, c := range cases {
		if got := invoiceLineServiceName(c.prefix, c.product, lineInst{c.inst}); got != c.want {
			t.Errorf("(%q, %q, %q) = %q, want %q", c.prefix, c.product, c.inst, got, c.want)
		}
	}
}
