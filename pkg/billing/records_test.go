package billing

import (
	pb "github.com/slntopp/nocloud-proto/billing"
	"slices"
	"testing"
)

func TestConsume(t *testing.T) {
	testCases := []struct {
		Record   pb.Record
		Plan     pb.Plan
		Expected float64
	}{
		{
			Record: pb.Record{
				Resource: "ip",
				Total:    1,
			},
			Plan: pb.Plan{
				Resources: []*pb.ResourceConf{
					{
						Key:    "ram",
						Price:  50,
						Period: 3600,
					},
					{
						Key:    "cpu",
						Price:  50,
						Period: 3600,
					},
					{
						Key:    "ip",
						Price:  100,
						Period: 7200,
					},
				},
			},
			Expected: 100,
		},
		{
			Record: pb.Record{
				Resource: "cpu",
				Total:    2,
			},
			Plan: pb.Plan{
				Resources: []*pb.ResourceConf{
					{
						Key:    "ram",
						Price:  36,
						Period: 3600,
					},
					{
						Key:    "cpu",
						Price:  72,
						Period: 3600,
					},
					{
						Key:    "ip",
						Price:  100,
						Period: 7200,
					},
				},
			},
			Expected: 144,
		},
		{
			Record: pb.Record{
				Product: "S",
				Total:   1,
			},
			Plan: pb.Plan{
				Resources: []*pb.ResourceConf{
					{
						Key:    "ram",
						Price:  36,
						Period: 3600,
					},
					{
						Key:    "cpu",
						Price:  72,
						Period: 3600,
					},
					{
						Key:    "ip",
						Price:  100,
						Period: 7200,
					},
				},
				Products: map[string]*pb.Product{
					"M": {
						Price: 260,
					},
					"S": {
						Price: 130,
					},
					"L": {
						Price: 360,
					},
				},
			},
			Expected: 130,
		},
	}

	for i, tc := range testCases {
		rec := tc.Record
		plan := tc.Plan

		if rec.GetResource() == "" {
			product, ok := plan.GetProducts()[rec.GetProduct()]
			if !ok {
				t.Errorf("Billing plan has no product from record. Index %d", i)
				continue
			}
			rec.Total = rec.GetTotal() * product.GetPrice()
			if rec.GetTotal() != tc.Expected {
				t.Errorf("Wrong total. Index %d. Got %f, expected %f", i, rec.GetTotal(), tc.Expected)
			}
		} else {
			resources := plan.GetResources()

			if !slices.ContainsFunc(resources, func(conf *pb.ResourceConf) bool {
				return conf.GetKey() == rec.GetResource()
			}) {
				t.Errorf("Billling plan has no resource. Index %d", i)
				continue
			}

			for _, res := range resources {
				if res.GetKey() == rec.GetResource() {
					rec.Total = rec.GetTotal() * res.GetPrice()
					if rec.GetTotal() != tc.Expected {
						t.Errorf("Wrong total. Index %d. Got %f, expected %f", i, rec.GetTotal(), tc.Expected)
					}
					break
				}
			}
		}
	}
}
