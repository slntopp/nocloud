package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing/promocodes"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
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
	Count(ctx context.Context, req *pb.CountPromocodesRequest) ([]*pb.Promocode, error)
	AddEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error
	RemoveEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error
}

type promocodesController struct {
	log *zap.Logger
	col driver.Collection
}

func NewPromocodesController(logger *zap.Logger, db driver.Database) PromocodesController {
	ctx := context.TODO()
	log := logger.Named("PromocodesController")
	promos := GetEnsureCollection(log, ctx, db, schema.PROMOCODES_COL)

	if _, _, err := promos.EnsureHashIndex(ctx, []string{"code"}, &driver.EnsureHashIndexOptions{
		Unique: true,
	}); err != nil {
		log.Error("Failed to create index on 'code'", zap.Error(err))
		panic(err)
	}

	return &promocodesController{
		log: log, col: promos,
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
	return promo, nil
}

func (c *promocodesController) Update(ctx context.Context, promo *pb.Promocode) (*pb.Promocode, error) {
	log := c.log.Named("Update")

	_, err := c.col.UpdateDocument(ctx, promo.GetUuid(), promo)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	return promo, nil
}

func (c *promocodesController) Delete(ctx context.Context, uuid string) error {
	log := c.log.Named("Delete")

	_, err := c.col.RemoveDocument(ctx, uuid)
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
	return &promo, nil
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
	return &promo, nil
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

	return promo, nil
}

func (c *promocodesController) Count(ctx context.Context, req *pb.CountPromocodesRequest) ([]*pb.Promocode, error) {
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
		return nil, err
	}

	var promo []*pb.Promocode
	_, err = cur.ReadDocument(ctx, &promo)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return nil, err
	}

	return promo, nil
}

func (c *promocodesController) AddEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error {
	log := c.log.Named("AddEntry")

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

	_, ok := findEntry(promo.Uses, entry)
	if ok {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		return fmt.Errorf("promocode already used on this given resource")
	}

	entry.Exec = time.Now().Unix()
	promo.Uses = append(promo.Uses, entry)

	_, err = c.col.ReplaceDocument(ctx, promo.GetUuid(), &promo)
	if err != nil {
		_ = db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{})
		log.Error("Failed to update document", zap.Error(err))
		return err
	}

	return db.CommitTransaction(ctx, trID, &driver.CommitTransactionOptions{})
}

func (c *promocodesController) RemoveEntry(ctx context.Context, uuid string, entry *pb.EntryResource) error {
	log := c.log.Named("AddEntry")

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

func validateEntry(entry *pb.EntryResource) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}
	if entry.Instance == nil && entry.Invoice == nil {
		return fmt.Errorf("one of the entry fields must be not nil")
	}
	if entry.Instance != nil && entry.Invoice != nil {
		return fmt.Errorf("only one of the entry fields must be not nil")
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
