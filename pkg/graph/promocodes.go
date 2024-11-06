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
	"strings"
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
	GetDiscountPriceByInstance(i *ipb.Instance, includeOneTimePayments bool) (float64, error)
	GetDiscountPriceByResource(i *ipb.Instance, defCurrency *bpb.Currency, initCost float64, initCurrency *bpb.Currency, resType string, resource string) (float64, error)
}

type promocodesController struct {
	log        *zap.Logger
	col        driver.Collection
	instances  InstancesController
	invoices   InvoicesController
	currencies CurrencyController
	plans      BillingPlansController
	addons     AddonsController
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
	addons := NewAddonsController(log, db)

	return &promocodesController{
		log: log, col: promos,
		instances: instances, invoices: invoices,
		currencies: currencies,
		plans:      plans,
		addons:     addons,
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
                       LET uses = p.uses == null ? [] : p.uses
			           FOR use in p.uses
                       LET bools = (
				           FOR r in @resources
				           LET parts = SPLIT(r, "/")
                           FILTER LENGTH(parts) == 2
                           FILTER parts[1] == TO_STRING(use.invoice) or parts[1] == TO_STRING(use.instance)
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
	log := c.log.Named("ListPromocodes")

	query := "LET promo = (FOR p in @@promocodes "
	vars := map[string]any{
		"@promocodes": schema.PROMOCODES_COL,
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT p.%s %s`
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

	log.Debug("Got promocodes", zap.Any("promocodes", promo))
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

// GetDiscountPriceByInstance returns instance estimate price (see graph.InstancesController) with applied discounts from linked promocodes
// Returning cost with default platform currency.
func (c *promocodesController) GetDiscountPriceByInstance(i *ipb.Instance, includeOneTimePayments bool) (float64, error) {
	ctx := context.Background()
	cost, err := c.instances.CalculateInstanceEstimatePrice(i, includeOneTimePayments)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate instance estimate price: %w", err)
	}

	promos, err := c.listAssociated("instances/" + i.GetUuid())
	if err != nil {
		return cost, fmt.Errorf("failed to get promocodes: %w", err)
	}
	if len(promos) == 0 {
		return cost, nil
	}

	// Filter inactive promocodes
	now := time.Now().Unix()
	var promosFiltered []*pb.Promocode
	for _, promo := range promos {
		var entry *pb.EntryResource
		for _, e := range promo.GetUses() {
			if e.Instance != nil && *e.Instance == i.GetUuid() {
				entry = e
				break
			}
		}
		if entry == nil {
			c.log.Error("Got no entry after listAssociated. Logic error")
			continue
		}
		if entry.Exec+promo.GetActiveTime() < now {
			continue
		}
		promosFiltered = append(promosFiltered, promo)
	}
	promos = promosFiltered

	bp, err := c.plans.Get(ctx, i.GetBillingPlan())
	if err != nil {
		return cost, fmt.Errorf("failed to get billing plan: %w", err)
	}

	_checkPrice := float64(0)
	discountPrice := float64(0)
	// Calculate product
	product, hasProduct := bp.GetProducts()[i.GetProduct()]
	oneTime := product.GetPeriod() == 0
	if hasProduct && (!oneTime || includeOneTimePayments) {
		_checkPrice += product.GetPrice()
		discountPrice += calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), "product", i.GetProduct(), product.GetPrice())
	}
	// Calculate addons
	if hasProduct && (!oneTime || includeOneTimePayments) {
		for _, a := range i.GetAddons() {
			addon, err := c.addons.Get(ctx, a)
			if err != nil {
				return cost, fmt.Errorf("failed to get addon: %w", err)
			}
			price := addon.GetPeriods()[product.GetPeriod()] // Assume this was checked by CalculateInstanceEstimatePrice
			_checkPrice += price
			discountPrice += calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), "addon", addon.GetUuid(), price)
		}
	}
	// Calculate resources
	for _, res := range bp.GetResources() {
		if res.GetPeriod() == 0 && !includeOneTimePayments {
			continue
		}
		var price float64
		// ram and drive_size calculates in GB for value
		if res.GetKey() == "ram" {
			value := i.GetResources()["ram"].GetNumberValue() / 1024
			price = value * res.GetPrice()
		} else if strings.Contains(res.GetKey(), "drive") {
			driveType := i.GetResources()["drive_type"].GetStringValue()
			if res.GetKey() != "drive_"+strings.ToLower(driveType) {
				continue
			}
			count := i.GetResources()["drive_size"].GetNumberValue() / 1024
			price = res.GetPrice() * count
		} else {
			count := i.GetResources()[res.GetKey()].GetNumberValue()
			price = res.GetPrice() * count
		}

		_checkPrice += price
		discountPrice += calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), "resource", res.GetKey(), price)
	}

	const EqualityThreshold = 0.001
	if math.Abs(_checkPrice-cost) > EqualityThreshold {
		return cost, fmt.Errorf("unexpected price: %f != %f", _checkPrice, cost)
	}

	return discountPrice, nil
}

// GetDiscountPriceByResource returns resource discounted price. It returns cost with initCurrency which you pass as parameter.
func (c *promocodesController) GetDiscountPriceByResource(i *ipb.Instance, defCurrency *bpb.Currency, initCost float64, initCurrency *bpb.Currency, resType string, resource string) (float64, error) {
	ctx := context.Background()
	if resType != "addon" && resType != "product" && resType != "resource" {
		return initCost, fmt.Errorf("invalid resource type")
	}

	promos, err := c.listAssociated("instances/" + i.GetUuid())
	if err != nil {
		return initCost, fmt.Errorf("failed to get promocodes: %w", err)
	}
	if len(promos) == 0 {
		return initCost, nil
	}

	// Filter inactive promocodes
	now := time.Now().Unix()
	var promosFiltered []*pb.Promocode
	for _, promo := range promos {
		var entry *pb.EntryResource
		for _, e := range promo.GetUses() {
			if e.Instance != nil && *e.Instance == i.GetUuid() {
				entry = e
				break
			}
		}
		if entry == nil {
			c.log.Error("Got no entry after listAssociated. Logic error")
			continue
		}
		if entry.Exec+promo.GetActiveTime() < now {
			continue
		}
		promosFiltered = append(promosFiltered, promo)
	}
	promos = promosFiltered

	cost := initCost
	if defCurrency.GetId() != initCurrency.GetId() {
		if cost, err = c.currencies.Convert(ctx, initCurrency, defCurrency, initCost); err != nil {
			return initCost, fmt.Errorf("failed to convert initial cost: %w", err)
		}
	}

	discount := calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), resType, resource, cost)
	if discount == 0 {
		return initCost, nil
	}

	cost = cost - discount
	if defCurrency.GetId() != initCurrency.GetId() {
		if cost, err = c.currencies.Convert(ctx, defCurrency, initCurrency, cost); err != nil {
			return initCost, fmt.Errorf("failed to convert final cost: %w", err)
		}
	}

	return cost, nil
}

func calculateResourceDiscount(promos []*pb.Promocode, planId string, resType string, resource string, cost float64) float64 {
	maxDiscount := float64(0)
	for _, promo := range promos {
		for _, item := range promo.GetPromoItems() {
			var applyToAll = item.GetPlanPromo().Addon == nil &&
				item.GetPlanPromo().Product == nil &&
				item.GetPlanPromo().Resource == nil

			if resType == "addon" && item.AddonPromo != nil && item.GetAddonPromo().GetAddon() == resource {
				goto proceed
			}

			if item.GetPlanPromo().GetBillingPlan() != planId &&
				(item.AddonPromo != nil || item.PlanPromo != nil) {
				continue
			}
			if !applyToAll && resType == "addon" && item.GetPlanPromo().GetAddon() != resource {
				continue
			}
			if !applyToAll && resType == "product" && item.GetPlanPromo().GetProduct() != resource {
				continue
			}
			if !applyToAll && resType == "resource" && item.GetPlanPromo().GetResource() != resource {
				continue
			}

		proceed:
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
	return maxDiscount
}

func (c *promocodesController) listAssociated(res string) ([]*pb.Promocode, error) {
	return c.List(context.Background(), &pb.ListPromocodesRequest{
		Filters: map[string]*structpb.Value{
			"resources": structpb.NewListValue(&structpb.ListValue{
				Values: []*structpb.Value{
					structpb.NewStringValue(res),
				},
			}),
		},
	})
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
