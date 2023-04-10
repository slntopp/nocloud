package events_logging

import (
	"context"
	"database/sql"
	"fmt"
	epb "github.com/slntopp/nocloud-proto/events_logging"
	"go.uber.org/zap"
)

type SqliteRepository struct {
	*sql.DB
	log *zap.Logger
}

func NewSqliteRepository(_log *zap.Logger, datasource string) *SqliteRepository {
	log := _log.Named("SqliteRep")

	log.Info("Creating SqliteRep")

	db, err := sql.Open("sqlite", datasource)
	if err != nil {
		log.Fatal("Failed to open connection", zap.Error(err))
		return nil
	}

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS EVENTS (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    ENTITY TEXT,
    UUID TEXT,
    SCOPE TEXT,
    ACTION TEXT,
    RC INTEGER,
    REQUESTOR TEXT
);

CREATE TABLE IF NOT EXISTS SNAPSHOTS (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    DIFF TEXT,
    EVENT_ID INTEGER,

    FOREIGN KEY (EVENT_ID) REFERENCES EVENTS(ID) ON DELETE CASCADE
);
`)
	if err != nil {
		log.Fatal("Failed to exec query", zap.Error(err))
		return nil
	}
	return &SqliteRepository{DB: db, log: log}
}

func (r *SqliteRepository) CreateEvent(ctx context.Context, eventMessage *ShortLogMessage) error {
	log := r.log.Named("Create Event")

	insertEventQuery := `INSERT INTO EVENTS (ENTITY, UUID, SCOPE, ACTION, RC, REQUESTOR) VALUES ($1, $2, $3, $4, $5, $6) RETURNING ID`

	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		log.Error("Failed to start transaction", zap.Error(err))
		return err
	}

	row := tx.QueryRow(insertEventQuery, eventMessage.Entity, eventMessage.Uuid, eventMessage.Scope, eventMessage.Action, eventMessage.Rc, eventMessage.Requestor)

	var createdEventId int32
	err = row.Scan(&createdEventId)
	if err != nil {
		log.Error("Failed to create event", zap.Error(err))
		tx.Rollback()
		return err
	}

	if eventMessage.Diff != "" {
		insertSnapshotRow := `INSERT INTO SNAPSHOTS (DIFF, EVENT_ID) VALUES ($1, $2);`

		row = tx.QueryRow(insertSnapshotRow, eventMessage.Diff, createdEventId)
		err := row.Scan()
		if err != nil {
			log.Error("Failed to create event", zap.Error(err))
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *SqliteRepository) GetEvents(ctx context.Context, req *epb.GetEventsRequest) ([]*epb.Event, error) {
	log := r.log.Named("GetEvents")

	selectQuery := fmt.Sprintf(`SELECT E.ID, E.ENTITY, E.UUID, E.SCOPE, E.ACTION, E.RC, E.REQUESTOR, S.ID, S.DIFF FROM EVENTS E LEFT OUTER JOIN SNAPSHOTS S on E.ID = S.EVENT_ID WHERE E.ENTITY = %s AND E.UUID = %s`, req.GetEntity(), req.GetUuid())

	if req.Scope != nil {
		selectQuery += fmt.Sprintf(`AND E.SCOPE = %s`, req.GetScope())
	}

	if req.Page != nil && req.Limit != nil {
		limit, page := req.GetLimit(), req.GetPage()
		offset := (page - 1) * limit

		selectQuery += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)
	}

	log.Debug("Query", zap.String("q", selectQuery))

	var events []*epb.Event

	rows, err := r.Query(selectQuery)
	if err != nil {
		log.Error("Error query events", zap.Error(err))
		return nil, err
	}

	for rows.Next() {
		var event = epb.Event{Snapshot: &epb.Snapshot{}}
		rows.Scan(
			&event.Id,
			&event.Entity,
			&event.Uuid,
			&event.Scope,
			&event.Action,
			&event.Rc,
			&event.Requestor,
			&event.Snapshot.Id,
			&event.Snapshot.Diff,
		)
		events = append(events, &event)
	}

	return events, nil
}

func (r *SqliteRepository) GetTrace(ctx context.Context, req *epb.GetTraceRequest) ([]*epb.Event, error) {
	log := r.log.Named("GetTrace")

	selectQuery := `SELECT E.ID, E.ENTITY, E.UUID, E.SCOPE, E.ACTION, E.RC, E.REQUESTOR, S.ID, S.DIFF FROM EVENTS E LEFT OUTER JOIN SNAPSHOTS S on E.ID = S.EVENT_ID WHERE E.REQUESTOR=$1`

	if req.Page != nil && req.Limit != nil {
		limit, page := req.GetLimit(), req.GetPage()
		offset := (page - 1) * limit

		selectQuery += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)
	}

	log.Debug("Query", zap.String("q", selectQuery))

	var events []*epb.Event

	rows, err := r.Query(selectQuery, req.GetRequestor())
	if err != nil {
		log.Error("Error query events", zap.Error(err))
		return nil, err
	}

	for rows.Next() {
		var event = epb.Event{Snapshot: &epb.Snapshot{}}
		rows.Scan(
			&event.Id,
			&event.Entity,
			&event.Uuid,
			&event.Scope,
			&event.Action,
			&event.Rc,
			&event.Requestor,
			&event.Snapshot.Id,
			&event.Snapshot.Diff,
		)
		events = append(events, &event)
	}

	return events, nil
}
