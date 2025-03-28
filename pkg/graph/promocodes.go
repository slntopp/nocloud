package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/billing/promocodes"
	ipb "github.com/slntopp/nocloud-proto/instances"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"math"
	"slices"
	"strings"
	"sync"
	"time"
)

type PromocodesController interface {
	Create(ctx context.Context, promo *pb.Promocode) (*pb.Promocode, error)
	Update(ctx context.Context, promo *pb.Promocode) (*pb.Promocode, error)
	Delete(ctx context.Context, uuid string) error
	Get(ctx context.Context, uuid string) (*pb.Promocode, error)
	GetByCode(ctx context.Context, code string, requester ...string) (*pb.Promocode, error)
	List(ctx context.Context, req *pb.ListPromocodesRequest) ([]*pb.Promocode, error)
	Count(ctx context.Context, req *pb.CountPromocodesRequest) (int64, error)
	AddEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error
	RemoveEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error
	IsPlanAffected(ctx context.Context, promo *pb.Promocode, _plan string) (bool, error)
	GetDiscountPriceByInstance(i *ipb.Instance, includeOneTimePayments bool, filterOneTime ...bool) (float64, Summary, error)
	GetDiscountPriceByResource(i *ipb.Instance, defCurrency *bpb.Currency, initCost float64, initCurrency *bpb.Currency, resType string, resource string) (float64, PromoSummary, error)
	CalculateResourceDiscount(promos []*pb.Promocode, planId string, resType string, resource string, cost float64) (float64, PromoSummary)
}

type promocodesController struct {
	log        *zap.Logger
	col        driver.Collection
	instances  InstancesController
	invoices   InvoicesController
	currencies CurrencyController
	plans      BillingPlansController
	addons     AddonsController
	showcases  ShowcasesController
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
	showcases := NewShowcasesController(log, db)

	return &promocodesController{
		log: log, col: promos,
		instances: instances, invoices: invoices,
		currencies: currencies,
		plans:      plans,
		addons:     addons,
		showcases:  showcases,
	}
}

type Summary map[string]PromoSummary

type PromoSummary struct {
	Promocode      string
	Code           string
	DiscountAmount float64
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
			           FOR use in uses
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
		} else if key == "plan" {
			values := value.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` LET result = (
                       FILTER IS_ARRAY(p.promo_items)
                       LET has = (
                          FOR item IN p.promo_items
                          FILTER item.plan_promo
                          FILTER item.plan_promo.billing_plan IN @billingPlans
                          RETURN true
                       )
                       FILTER LENGTH(has) > 0
                       RETURN true
                    )
                    FILTER LENGTH(result) > 0`)
			vars["billingPlans"] = values
		} else if key == "showcase" {
			values := value.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` LET result = (
                       FILTER IS_ARRAY(p.promo_items)
                       LET has = (
                          FOR item IN p.promo_items
                          FILTER item.showcase_promo
                          FILTER item.showcase_promo.showcase IN @showcasesList
                          RETURN true
                       )
                       FILTER LENGTH(has) > 0
                       RETURN true
                    )
                    FILTER LENGTH(result) > 0`)
			vars["showcasesList"] = values
		} else if key == "uuids" {
			values := value.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER p._key IN @%s`, key)
			vars[key] = values
		} else {
			values := value.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER p["%s"] IN @%s`, key, key)
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

	query := `
UPDATE @promocode WITH {
title: @title,
description: @description,
code: @code,
status: @status,
condition: @condition,
due_date: @due_date,
uses_per_user: @uses_per_user,
limit: @limit,
one_time: @one_time,
active_time: @active_time,
meta: @meta,
promo_items: @promo_items,
} IN @@promocodes RETURN NEW
`
	cur, err := c.col.Database().Query(ctx, query, map[string]interface{}{
		"promocode":     promo.GetUuid(),
		"@promocodes":   schema.PROMOCODES_COL,
		"title":         promo.GetTitle(),
		"description":   promo.GetDescription(),
		"code":          promo.GetCode(),
		"status":        promo.GetStatus(),
		"condition":     promo.GetCondition(),
		"due_date":      promo.GetDueDate(),
		"uses_per_user": promo.GetUsesPerUser(),
		"limit":         promo.GetLimit(),
		"one_time":      promo.GetOneTime(),
		"active_time":   promo.GetActiveTime(),
		"meta":          promo.GetMeta(),
		"promo_items":   promo.GetPromoItems(),
	})
	if err != nil {
		log.Error("Failed to update promocode", zap.Error(err))
		return nil, err
	}

	var _promo *pb.Promocode
	if _, err = cur.ReadDocument(ctx, _promo); err != nil {
		log.Error("Failed to read result promocode")
		return applyCurrentState(promo), nil
	}

	return applyCurrentState(_promo), nil
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

	var promo = &pb.Promocode{}

	meta, err := c.col.ReadDocument(ctx, uuid, promo)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	promo.Uuid = meta.Key
	return applyCurrentState(promo), nil
}

const getByCodeQuery = `FOR p IN @@promocodes FILTER p.code == @code RETURN p`

func (c *promocodesController) GetByCode(ctx context.Context, code string, requester ...string) (*pb.Promocode, error) {
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
	var promo = &pb.Promocode{}
	meta, err := cur.ReadDocument(ctx, promo)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	var acc *pb.EntryResource
	if len(requester) > 0 {
		acc = &pb.EntryResource{Account: requester[0]}
	}

	promo.Uuid = meta.Key
	return applyCurrentStateWithUsed(promo, acc), nil
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

func (c *promocodesController) IsPlanAffected(ctx context.Context, promo *pb.Promocode, _plan string) (bool, error) {
	if promo == nil || _plan == "" {
		return false, nil
	}
	_, err := c.plans.Get(ctx, &bpb.Plan{Uuid: _plan})
	if err != nil {
		return false, err
	}
	for _, item := range promo.GetPromoItems() {
		if item.ShowcasePromo != nil {
			showcasesPlans := c.getShowcasesPlansCached()
			plansMap, scOk := showcasesPlans[item.GetShowcasePromo().GetShowcase()]
			if scOk {
				if _, ok := plansMap[_plan]; ok {
					return true, nil
				}
			}
		}
		if item.PlanPromo != nil {
			if item.GetPlanPromo().GetBillingPlan() == _plan {
				return true, nil
			}
		}
		if item.AddonPromo != nil {
			return true, nil
		}
	}
	return false, nil
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
	if err = invalidStateError(promo.Condition); err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return err
	}

	_, ok := findEntry(promo.Uses, entry)
	if ok {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return fmt.Errorf("%w: promocode already used on given resource", ErrAlreadyExists)
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
func (c *promocodesController) GetDiscountPriceByInstance(i *ipb.Instance, includeOneTimePayments bool, filterOneTime ...bool) (float64, Summary, error) {
	ctx := context.Background()
	cost, err := c.instances.CalculateInstanceEstimatePrice(i, includeOneTimePayments)
	if err != nil {
		return 0, Summary{}, fmt.Errorf("failed to calculate instance estimate price: %w", err)
	}

	promos, err := c.listAssociated("instances/" + i.GetUuid())
	if err != nil {
		return cost, Summary{}, fmt.Errorf("failed to get promocodes: %w", err)
	}
	if len(promos) == 0 {
		return cost, Summary{}, nil
	}
	promos = filterInactivePromos(promos, i.GetUuid(), time.Now().Unix())
	if len(filterOneTime) > 0 && filterOneTime[0] {
		promos = filterOneTimePromos(promos)
	}

	bp, err := c.plans.Get(ctx, i.GetBillingPlan())
	if err != nil {
		return cost, Summary{}, fmt.Errorf("failed to get billing plan: %w", err)
	}

	handleSummary := func(sum Summary, ps PromoSummary) {
		if ps.Promocode == "" {
			return
		}
		old, ok := sum[ps.Promocode]
		if !ok {
			sum[ps.Promocode] = ps
			return
		}
		old.DiscountAmount += ps.DiscountAmount
		sum[ps.Promocode] = old
	}

	sum := Summary{}
	_checkPrice := float64(0)
	discountPrice := float64(0)
	// Calculate product
	product, hasProduct := bp.GetProducts()[i.GetProduct()]
	oneTime := product.GetPeriod() == 0
	if hasProduct && (!oneTime || includeOneTimePayments) {
		_checkPrice += product.GetPrice()
		discount, ps := c.calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), "product", i.GetProduct(), product.GetPrice())
		discountPrice += product.GetPrice() - discount
		handleSummary(sum, ps)
	}
	// Calculate addons
	if hasProduct && (!oneTime || includeOneTimePayments) {
		for _, a := range i.GetAddons() {
			addon, err := c.addons.Get(ctx, a)
			if err != nil {
				return cost, Summary{}, fmt.Errorf("failed to get addon: %w", err)
			}
			price := addon.GetPeriods()[product.GetPeriod()] // Assume this was checked by CalculateInstanceEstimatePrice
			_checkPrice += price
			discount, ps := c.calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), "addon", addon.GetUuid(), price)
			discountPrice += price - discount
			handleSummary(sum, ps)
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
		discount, ps := c.calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), "resource", res.GetKey(), price)
		discountPrice += price - discount
		handleSummary(sum, ps)
	}

	const EqualityThreshold = 0.001
	if math.Abs(_checkPrice-cost) > EqualityThreshold {
		return cost, Summary{}, fmt.Errorf("unexpected price: %f != %f", _checkPrice, cost)
	}

	return discountPrice, sum, nil
}

// GetDiscountPriceByResource returns resource discounted price. It returns cost with initCurrency which you pass as parameter
// Returning initial price if promocode is one time
func (c *promocodesController) GetDiscountPriceByResource(i *ipb.Instance, defCurrency *bpb.Currency, initCost float64, initCurrency *bpb.Currency, resType string, resource string) (float64, PromoSummary, error) {
	ctx := context.Background()
	if resType != "addon" && resType != "product" && resType != "resource" {
		return initCost, PromoSummary{}, fmt.Errorf("invalid resource type")
	}

	promos, err := c.listAssociated("instances/" + i.GetUuid())
	if err != nil {
		return initCost, PromoSummary{}, fmt.Errorf("failed to get promocodes: %w", err)
	}
	if len(promos) == 0 {
		return initCost, PromoSummary{}, nil
	}
	promos = filterInactivePromos(promos, i.GetUuid(), time.Now().Unix())
	promos = filterOneTimePromos(promos)

	rate, _, err := c.currencies.GetExchangeRate(ctx, initCurrency, defCurrency)
	if err != nil {
		return initCost, PromoSummary{}, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	cost := initCost
	if defCurrency.GetId() != initCurrency.GetId() {
		cost *= rate
	}

	discount, ps := c.calculateResourceDiscount(promos, i.GetBillingPlan().GetUuid(), resType, resource, cost)
	if discount == 0 {
		return initCost, PromoSummary{}, nil
	}

	ps.DiscountAmount *= 1 / rate
	cost = cost - discount
	if defCurrency.GetId() != initCurrency.GetId() {
		cost *= 1 / rate
	}

	return cost, ps, nil
}

func (c *promocodesController) CalculateResourceDiscount(promos []*pb.Promocode, planId string, resType string, resource string, cost float64) (float64, PromoSummary) {
	return c.calculateResourceDiscount(promos, planId, resType, resource, cost)
}

func (c *promocodesController) calculateResourceDiscount(promos []*pb.Promocode, planId string, resType string, resource string, cost float64) (float64, PromoSummary) {
	showcasesPlans := c.getShowcasesPlansCached()

	maxDiscount := float64(0)
	ps := PromoSummary{}

	for _, promo := range promos {
		for _, item := range promo.GetPromoItems() {
			// If it has plan promo and no items specified, then apply to plan
			var applyToAll = item.PlanPromo != nil &&
				item.GetPlanPromo().Addon == nil &&
				item.GetPlanPromo().Product == nil &&
				item.GetPlanPromo().Resource == nil

			// Apply to addon if specified
			if resType == "addon" && item.AddonPromo != nil && item.GetAddonPromo().GetAddon() == resource {
				goto proceed
			}

			// If showcase promo specified and current plan is in it then proceed
			if item.ShowcasePromo != nil {
				sc := item.GetShowcasePromo().GetShowcase()
				m.Lock()
				if _, ok := showcasesPlans[sc]; !ok {
					m.Unlock()
					continue
				}
				if _, ok := showcasesPlans[sc][planId]; !ok {
					m.Unlock()
					continue
				}
				m.Unlock()
				goto proceed
			}

			if item.GetPlanPromo().GetBillingPlan() != planId &&
				(item.AddonPromo != nil || item.PlanPromo != nil || item.ShowcasePromo != nil) {
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
				discount = cost * sch.GetDiscountPercent()
			} else if sch.DiscountAmount != nil {
				discount = math.Min(cost, float64(sch.GetDiscountAmount()))
			} else if sch.FixedPrice != nil {
				discount = cost - float64(sch.GetFixedPrice())
			}

			if discount > maxDiscount {
				maxDiscount = discount
				ps.DiscountAmount = maxDiscount
				ps.Promocode = promo.GetUuid()
				ps.Code = promo.GetCode()
			}
		}
	}
	return maxDiscount, ps
}

var _showcasesPlans = map[string]map[string]struct{}{}

const showcasesPlansCacheTTL = 60 * time.Minute

var showcasesPlansLastUpdate = int64(0)

var m = &sync.Mutex{}

func (c *promocodesController) getShowcasesPlansCached() map[string]map[string]struct{} {
	m.Lock()
	defer m.Unlock()

	now := time.Now().Unix()
	if showcasesPlansLastUpdate+int64(showcasesPlansCacheTTL.Seconds()) > now {
		return _showcasesPlans
	}

	scs, err := c.showcases.List(context.Background(), schema.ROOT_ACCOUNT_KEY, true, &sppb.ListRequest{})
	if err != nil {
		c.log.Named("getShowcasesPlansCached").Error("FATAL: failed to list showcases", zap.Error(err))
		return _showcasesPlans
	}

	_showcasesPlans = make(map[string]map[string]struct{})
	for _, s := range scs {
		_showcasesPlans[s.GetUuid()] = make(map[string]struct{})
		for _, p := range s.GetItems() {
			_showcasesPlans[s.GetUuid()][p.GetPlan()] = struct{}{}
		}
	}

	showcasesPlansLastUpdate = now
	return _showcasesPlans
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
	if (entry.Instance == nil || entry.GetInstance() == "") &&
		(entry.Invoice == nil || entry.GetInvoice() == "") {
		return fmt.Errorf("invalid entry: one of the entry fields must be set")
	}
	if entry.Instance != nil && entry.Invoice != nil {
		return fmt.Errorf("invalid entry: only one of the entry fields must be set")
	}
	return nil
}

func findEntry(uses []*pb.EntryResource, entry *pb.EntryResource) (int, bool) {
	for i, u := range uses {
		if u.Instance != nil && entry.Instance != nil && *u.Instance == *entry.Instance {
			return i, true
		}
		if u.Invoice != nil && entry.Invoice != nil && *u.Invoice == *entry.Invoice {
			return i, true
		}
	}
	return -1, false
}

func applyCurrentState(promo *pb.Promocode) *pb.Promocode {
	expired := promo.GetDueDate() > 0 && time.Now().Unix() >= promo.GetDueDate()
	taken := promo.GetLimit() > 0 && int64(len(promo.GetUses())) >= promo.GetLimit()

	if taken {
		promo.Condition = pb.PromocodeCondition_ALL_TAKEN
	}
	if expired {
		promo.Condition = pb.PromocodeCondition_EXPIRED
	}
	if !taken && !expired {
		promo.Condition = pb.PromocodeCondition_USABLE
	}

	return promo
}

func applyCurrentStateWithUsed(promo *pb.Promocode, newEntry *pb.EntryResource) *pb.Promocode {
	promo = applyCurrentState(promo)
	maxUsesPerUser := promo.GetUsesPerUser()
	if maxUsesPerUser <= 0 {
		return promo
	}
	if newEntry == nil || newEntry.Account == "" {
		return promo
	}
	newEntryAccount := newEntry.Account

	current := int64(0)
	for _, use := range promo.GetUses() {
		if use.GetAccount() == newEntryAccount {
			current++
			if current == maxUsesPerUser {
				promo.Condition = pb.PromocodeCondition_USED
				break
			}
		}
	}

	return promo
}

func invalidStateError(state pb.PromocodeCondition) error {
	switch state {
	case pb.PromocodeCondition_EXPIRED:
		return fmt.Errorf("promocode is expired")
	case pb.PromocodeCondition_ALL_TAKEN:
		return fmt.Errorf("no promocodes left")
	case pb.PromocodeCondition_USED:
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

func filterInactivePromos(promos []*pb.Promocode, instance string, now int64) []*pb.Promocode {
	var promosFiltered []*pb.Promocode
	for _, promo := range promos {
		if promo.GetActiveTime() == 0 {
			promosFiltered = append(promosFiltered, promo)
			continue
		}
		if promo.GetStatus() == pb.PromocodeStatus_DELETED || promo.GetStatus() == pb.PromocodeStatus_SUSPENDED {
			continue
		}
		var entry *pb.EntryResource
		for _, e := range promo.GetUses() {
			if e.Instance != nil && *e.Instance == instance {
				entry = e
				break
			}
		}
		if entry == nil {
			continue
		}
		if entry.Exec+promo.GetActiveTime() < now {
			continue
		}
		promosFiltered = append(promosFiltered, promo)
	}
	return promosFiltered
}
func filterOneTimePromos(promos []*pb.Promocode) []*pb.Promocode {
	var promosFiltered []*pb.Promocode
	for _, promo := range promos {
		if promo.GetOneTime() {
			continue
		}
		promosFiltered = append(promosFiltered, promo)
	}
	return promosFiltered
}
