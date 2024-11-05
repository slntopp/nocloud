package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/billing/promocodes"
	ipb "github.com/slntopp/nocloud-proto/instances"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"math"
	"slices"
	"time"
)

type PromocodesController interface {
	Create(ctx context.Context, promo *pb.Promocode) (*pb.Promocode, error)
	Update(ctx context.Context, promo *pb.Promocode) (*pb.Promocode, error)
	Delete(ctx context.Context, uuid string) error
	Get(ctx context.Context, uuid string) (*pb.Promocode, error)
	GetByCode(ctx context.Context, code string) (*pb.Promocode, error)
	List(ctx context.Context, req *pb.ListPromocodesRequest) ([]*pb.Promocode, error)
	Count(ctx context.Context, req *pb.CountPromocodesRequest) (int64, error)
	AddEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error
	RemoveEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error
}

type promocodesController struct {
	log        *zap.Logger
	col        driver.Collection
	instances  InstancesController
	invoices   InvoicesController
	currencies CurrencyController
	plans      BillingPlansController
}

func NewPromocodesController(logger *zap.Logger, db driver.Database, conn rabbitmq.Connection) PromocodesController {
	ctx := context.TODO()
	log := logger.Named("PromocodesController")
	promos := GetEnsureCollection(log, ctx, db, schema.PROMOCODES_COL)

	if _, _, err := promos.EnsureHashIndex(ctx, []string{"code"}, &driver.EnsureHashIndexOptions{
		Unique: true,
	}); err != nil {
		log.Error("Failed to create index on 'code'", zap.Error(err))
		panic(err)
	}

	instances := NewInstancesController(log, db, conn)
	invoices := NewInvoicesController(log, db)
	currencies := NewCurrencyController(log, db)
	plans := NewBillingPlansController(log, db)

	return &promocodesController{
		log: log, col: promos,
		instances: instances, invoices: invoices,
		currencies: currencies,
		plans:      plans,
	}
}

func buildFiltersQuery(filters map[string]*structpb.Value, vars map[string]any) string {
	query := ""
	for key, value := range filters {
		if key == "resources" {
			values := value.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` LET result = (
			           FOR use in p.uses
                       LET bools = (
				           FOR r in @resources
				           LET parts = SPLIT(r, "/")
                           FILTER LENGTH(parts) == 2
                           FILTER parts[1] == use.invoice or parts[1] == use.instance
                           RETURN true
                       )
                       FILTER LENGTH(bools) > 0
                       RETURN true
                    )
                    FILTER LENGTH(result) > 0`)
			vars[key] = values
		}
	}
	return query
}

func (c *promocodesController) Create(ctx context.Context, promo *pb.Promocode) (*pb.Promocode, error) {
	log := c.log.Named("Create")

	document, err := c.col.CreateDocument(ctx, promo)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	promo.Uuid = document.Key
	return applyCurrentState(promo), nil
}

func (c *promocodesController) Update(ctx context.Context, promo *pb.Promocode) (*pb.Promocode, error) {
	log := c.log.Named("Update")

	_, err := c.col.UpdateDocument(ctx, promo.GetUuid(), promo)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	return applyCurrentState(promo), nil
}

func (c *promocodesController) Delete(ctx context.Context, uuid string) error {
	log := c.log.Named("Delete")

	var promo pb.Promocode
	meta, err := c.col.ReadDocument(ctx, uuid, &promo)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return err
	}
	promo.Uuid = meta.Key

	promo.Status = pb.PromocodeStatus_DELETED
	_, err = c.col.UpdateDocument(ctx, promo.GetUuid(), &promo)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return err
	}

	return nil
}

func (c *promocodesController) Get(ctx context.Context, uuid string) (*pb.Promocode, error) {
	log := c.log.Named("Get")

	var promo pb.Promocode

	meta, err := c.col.ReadDocument(ctx, uuid, &promo)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	promo.Uuid = meta.Key
	return applyCurrentState(&promo), nil
}

const getByCodeQuery = `FOR p IN @@promocodes FILTER p.code == @code RETURN p`

func (c *promocodesController) GetByCode(ctx context.Context, code string) (*pb.Promocode, error) {
	log := c.log.Named("GetByCode")

	cur, err := c.col.Database().Query(ctx, getByCodeQuery, map[string]interface{}{
		"@promocodes": schema.PROMOCODES_COL,
		"code":        code,
	})
	if err != nil {
		log.Error("Failed to execute query", zap.Error(err))
		return nil, err
	}
	if !cur.HasMore() {
		return nil, fmt.Errorf("promocode with code %s not found", code)
	}
	var promo pb.Promocode
	meta, err := cur.ReadDocument(ctx, &promo)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	promo.Uuid = meta.Key
	return applyCurrentState(&promo), nil
}

func (c *promocodesController) List(ctx context.Context, req *pb.ListPromocodesRequest) ([]*pb.Promocode, error) {
	log := c.log.Named("List")

	query := "LET promo = (FOR p in @@promocodes "
	vars := map[string]any{
		"@promocodes": schema.PROMOCODES_COL,
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT a.%s %s`
		field, sort := req.GetField(), req.GetSort()

		query += fmt.Sprintf(subQuery, field, sort)
	}

	query += buildFiltersQuery(req.GetFilters(), vars)

	if req.Page != nil && req.Limit != nil {
		if req.GetLimit() != 0 {
			limit, page := req.GetLimit(), req.GetPage()
			offset := (page - 1) * limit

			query += ` LIMIT @offset, @count`
			vars["offset"] = offset
			vars["count"] = limit
		}
	}

	query += " RETURN merge(p, {uuid: p._key})) RETURN promo"
	log.Debug("Query", zap.String("q", query))
	cur, err := c.col.Database().Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to get documents", zap.Error(err))
		return nil, err
	}

	var promo []*pb.Promocode
	_, err = cur.ReadDocument(ctx, &promo)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return nil, err
	}

	for _, p := range promo {
		p = applyCurrentState(p)
	}

	return promo, nil
}

func (c *promocodesController) Count(ctx context.Context, req *pb.CountPromocodesRequest) (int64, error) {
	log := c.log.Named("Count")

	query := "LET promo = (FOR p in @@promocodes "
	vars := map[string]any{
		"@promocodes": schema.PROMOCODES_COL,
	}

	query += buildFiltersQuery(req.GetFilters(), vars)

	query += " RETURN merge(p, {uuid: p._key})) RETURN promo"
	log.Debug("Query", zap.String("q", query))
	cur, err := c.col.Database().Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to get documents", zap.Error(err))
		return 0, err
	}

	var promo []*pb.Promocode
	_, err = cur.ReadDocument(ctx, &promo)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return 0, err
	}

	return int64(len(promo)), nil
}

func (c *promocodesController) AddEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error {
	log := c.log.Named("AddEntry")

	if err := validateEntry(entry); err != nil {
		log.Error("Failed to validate entry", zap.Error(err))
		return err
	}

	if entry.Account == "" {
		accId, err := findEntryAccount(entry, c.instances, c.invoices)
		if err != nil {
			log.Error("Couldn't find resource owner account for entry", zap.Error(err))
			return fmt.Errorf("invalid entry: resource owner not found")
		}
		entry.Account = accId
	}

	db := c.col.Database()
	trID, err := db.BeginTransaction(ctx, driver.TransactionCollections{
		Write:     []string{schema.PROMOCODES_COL},
		Exclusive: []string{schema.PROMOCODES_COL},
	}, &driver.BeginTransactionOptions{})
	if err != nil {
		log.Error("Failed to start transaction", zap.Error(err))
		return fmt.Errorf("internal error: start transaction. Try again later")
	}
	ctx = driver.WithTransactionID(ctx, trID)

	var promo pb.Promocode
	meta, err := c.col.ReadDocument(ctx, uuid, &promo)
	if err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		log.Error("Failed to get document", zap.Error(err))
		return fmt.Errorf("internal error: reading promocode. Try again later")
	}
	promo.Uuid = meta.Key

	if err = invalidStatusError(promo.Status); err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return err
	}
	promo = *applyCurrentStateWithUsed(&promo, entry)
	if err = invalidStateError(promo.State); err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return err
	}

	_, ok := findEntry(promo.Uses, entry)
	if ok {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return fmt.Errorf("promocode already used on given resource")
	}

	entry.Exec = time.Now().Unix()
	promo.Uses = append(promo.Uses, entry)

	_, err = c.col.ReplaceDocument(ctx, promo.GetUuid(), &promo)
	if err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		log.Error("Failed to update document", zap.Error(err))
		return fmt.Errorf("internal error: applying promocode. Try again later")
	}

	if err = db.CommitTransaction(ctx, trID, &driver.CommitTransactionOptions{}); err != nil {
		return fmt.Errorf("internal error: commiting changes. Try again later")
	}
	return nil
}

func (c *promocodesController) RemoveEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error {
	log := c.log.Named("RemoveEntry")

	if err := validateEntry(entry); err != nil {
		log.Error("Failed to validate entry", zap.Error(err))
		return err
	}

	db := c.col.Database()
	trID, err := db.BeginTransaction(ctx, driver.TransactionCollections{
		Write:     []string{schema.PROMOCODES_COL},
		Exclusive: []string{schema.PROMOCODES_COL},
	}, &driver.BeginTransactionOptions{})
	if err != nil {
		log.Error("Failed to start transaction", zap.Error(err))
		return err
	}
	ctx = driver.WithTransactionID(ctx, trID)

	var promo pb.Promocode
	meta, err := c.col.ReadDocument(ctx, uuid, &promo)
	if err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		log.Error("Failed to get document", zap.Error(err))
		return err
	}
	promo.Uuid = meta.Key

	if err = invalidStatusError(promo.Status); err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return err
	}

	ind, ok := findEntry(promo.Uses, entry)
	if !ok {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return fmt.Errorf("promocode is not applied on given resource")
	}

	entry.Exec = time.Now().Unix()
	promo.Uses = slices.Delete(promo.Uses, ind, ind+1)

	_, err = c.col.ReplaceDocument(ctx, promo.GetUuid(), &promo)
	if err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		log.Error("Failed to update document", zap.Error(err))
		return err
	}

	return db.CommitTransaction(ctx, trID, &driver.CommitTransactionOptions{})
}

func (c *promocodesController) GetDiscountPriceByResource(i *ipb.Instance, defCurrency *bpb.Currency, initCost float64, initCurrency *bpb.Currency, resType string, resource string) (float64, error) {
	if resType != "addon" && resType != "product" && resType != "resource" {
		return initCost, fmt.Errorf("invalid resource type")
	}

	ctx := context.Background()
	// Get all associated with instance promocodes
	promos, err := c.List(ctx, &pb.ListPromocodesRequest{
		Filters: map[string]*structpb.Value{
			"resources": structpb.NewListValue(&structpb.ListValue{
				Values: []*structpb.Value{
					structpb.NewStringValue("instances/" + i.GetUuid()),
				},
			}),
		},
	})
	if err != nil {
		return initCost, fmt.Errorf("failed to get promocodes: %w", err)
	}
	if len(promos) == 0 {
		return initCost, nil
	}

	cost := initCost
	if defCurrency.GetId() != initCurrency.GetId() {
		if cost, err = c.currencies.Convert(ctx, initCurrency, defCurrency, initCost); err != nil {
			return initCost, fmt.Errorf("failed to convert initial cost: %w", err)
		}
	}

	maxDiscount := float64(0)
	for _, promo := range promos {
		for _, item := range promo.GetPromoItems() {
			if item.GetPlanPromo().GetBillingPlan() != i.GetBillingPlan().GetUuid() &&
				(item.AddonPromo != nil || item.PlanPromo != nil) {
				continue
			}
			var applyToAll = item.GetPlanPromo().Addon == nil &&
				item.GetPlanPromo().Product == nil &&
				item.GetPlanPromo().Resource == nil
			if !applyToAll && resType == "addon" && item.GetPlanPromo().GetAddon() != resource {
				continue
			}
			if !applyToAll && resType == "product" && item.GetPlanPromo().GetProduct() != resource {
				continue
			}
			if !applyToAll && resType == "resource" && item.GetPlanPromo().GetResource() != resource {
				continue
			}

			discount := float64(0)
			sch := item.GetSchema()
			if sch.DiscountPercent != nil {
				discount = cost - cost*sch.GetDiscountPercent()
			} else if sch.DiscountAmount != nil {
				discount = math.Min(cost, float64(sch.GetDiscountAmount()))
			} else if sch.FixedPrice != nil {
				discount = cost - float64(sch.GetFixedPrice())
			}

			if discount > maxDiscount {
				maxDiscount = discount
			}
		}
	}

	if maxDiscount == 0 {
		return initCost, nil
	}

	cost = cost - maxDiscount
	if defCurrency.GetId() != initCurrency.GetId() {
		if cost, err = c.currencies.Convert(ctx, defCurrency, initCurrency, cost); err != nil {
			return initCost, fmt.Errorf("failed to convert final cost: %w", err)
		}
	}

	return cost, nil
}

func validateEntry(entry *pb.EntryResource) error {
	if entry == nil {
		return fmt.Errorf("invalid entry: entry cannot be nil")
	}
	if entry.Instance == nil && entry.Invoice == nil {
		return fmt.Errorf("invalid entry: one of the entry fields must be set")
	}
	if entry.Instance != nil && entry.Invoice != nil {
		return fmt.Errorf("invalid entry: only one of the entry fields must be set")
	}
	return nil
}

func findEntry(uses []*pb.EntryResource, entry *pb.EntryResource) (int, bool) {
	for i, u := range uses {
		if u.Instance != nil && entry.Instance != nil && u.Instance == entry.Instance {
			return i, true
		}
		if u.Invoice != nil && entry.Invoice != nil && u.Invoice == entry.Invoice {
			return i, true
		}
	}
	return -1, false
}

func applyCurrentState(promo *pb.Promocode) *pb.Promocode {
	expired := time.Now().Unix() >= promo.GetDueDate()
	taken := int64(len(promo.GetUses())) >= promo.GetLimit()

	if taken {
		promo.State = pb.PromocodeState_ALL_TAKEN
	}
	if expired {
		promo.State = pb.PromocodeState_EXPIRED
	}

	if !taken && !expired {
		promo.State = pb.PromocodeState_USABLE
	}

	return promo
}

func applyCurrentStateWithUsed(promo *pb.Promocode, newEntry *pb.EntryResource) *pb.Promocode {
	promo = applyCurrentState(promo)
	newEntryAccount := newEntry.Account
	maxUsesPerUser := promo.GetUsesPerUser()

	current := int64(0)
	for _, use := range promo.GetUses() {
		if use.GetAccount() == newEntryAccount {
			current++
			if current == maxUsesPerUser {
				promo.State = pb.PromocodeState_USED
				break
			}
		}
	}

	return promo
}

func invalidStateError(state pb.PromocodeState) error {
	switch state {
	case pb.PromocodeState_EXPIRED:
		return fmt.Errorf("promocode is expired")
	case pb.PromocodeState_ALL_TAKEN:
		return fmt.Errorf("no promocodes left")
	case pb.PromocodeState_USED:
		return fmt.Errorf("already got maximum uses of this promocode")
	}
	return nil
}

func invalidStatusError(status pb.PromocodeStatus) error {
	switch status {
	case pb.PromocodeStatus_DELETED:
		return fmt.Errorf("promocode is deleted")
	case pb.PromocodeStatus_SUSPENDED:
		return fmt.Errorf("promocode is inactive")
	}
	return nil
}

func findEntryAccount(entry *pb.EntryResource, instances InstancesController, invoices InvoicesController) (string, error) {
	if entry.Instance != nil {
		acc, err := instances.GetInstanceOwner(context.Background(), entry.GetInstance())
		if err != nil {
			return "", err
		}
		return acc.GetUuid(), nil
	}
	if entry.Invoice != nil {
		inv, err := invoices.Get(context.Background(), entry.GetInvoice())
		if err != nil {
			return "", err
		}
		return inv.GetAccount(), nil
	}
	return "", fmt.Errorf("fatal: entry resource is not set")
}
