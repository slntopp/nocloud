package graph

import (
	"context"
	"testing"

	"github.com/slntopp/nocloud-proto/ansible"
	"github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud-proto/instances"
	"go.uber.org/zap"
)

func TestValidateBillingPlan(t *testing.T) {
	test_cases := []struct {
		name     string
		instance *instances.Instance

		result bool
		err    string
	}{
		{
			name: "Check software not matching single",
			instance: &instances.Instance{
				BillingPlan: &billing.Plan{
					Software: []*ansible.Software{
						{
							Playbook: "Y",
						},
					},
				},
				Software: []*ansible.Software{
					{
						Playbook: "X",
					},
				},
			},
			result: false,
			err:    "software Y is not defined in Instance",
		},
		{
			name: "Check software not matching multiple",
			instance: &instances.Instance{
				BillingPlan: &billing.Plan{
					Software: []*ansible.Software{
						{
							Playbook: "X",
						},
						{
							Playbook: "Y",
						},
					},
				},
				Software: []*ansible.Software{
					{
						Playbook: "X",
					},
					{
						Playbook: "Z",
					},
				},
			},
			result: false,
			err:    "software Y is not defined in Instance",
		},
		{
			name: "Check software matching single",
			instance: &instances.Instance{
				BillingPlan: &billing.Plan{
					Software: []*ansible.Software{
						{
							Playbook: "X",
						},
					},
				},
				Software: []*ansible.Software{
					{
						Playbook: "X",
					},
				},
			},
			result: true,
			err:    "",
		},
		{
			name: "Check software matching multiple",
			instance: &instances.Instance{
				BillingPlan: &billing.Plan{
					Software: []*ansible.Software{
						{
							Playbook: "X",
						},
						{
							Playbook: "Y",
						},
					},
				},
				Software: []*ansible.Software{
					{
						Playbook: "X",
					},
					{
						Playbook: "Y",
					},
				},
			},
			result: true,
			err:    "",
		},
	}
	ctrl := &instancesController{log: zap.NewExample()}

	for _, test := range test_cases {
		t.Logf("Running test %s", test.name)

		err := ctrl.ValidateBillingPlan(context.Background(), "", test.instance)
		if test.result && err != nil {
			t.Errorf("Test %s expected to PASS but failed with: %v", test.name, err)
			continue
		} else if test.result && err == nil {
			t.Logf("Test %s passed", test.name)
			continue
		}

		if err == nil {
			t.Errorf("Test %s expected to FAIL but passed", test.name)
			continue
		}

		if err.Error() != test.err {
			t.Errorf("Test %s expected to FAIL with:\n\t%v\nbut failed with:\n\t%v", test.name, test.err, err)
		}
	}
}
