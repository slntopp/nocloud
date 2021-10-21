package graph

import (
	"context"
	"errors"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/accounting/accountspb"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ACCOUNTS_COL = "Accounts"
	ACC2NS = ACCOUNTS_COL + "2" + NAMESPACES_COL
	ACC2CRED = ACCOUNTS_COL + "2" + CREDENTIALS_COL
)

type Account struct {
	Title string `json:"title"`
	driver.DocumentMeta
}

type AccountsController struct {
	col driver.Collection // Accounts Collection
	cred driver.Collection // Credentials Collection

	log *zap.Logger
}

func NewAccountsController(log *zap.Logger, col, cred driver.Collection) AccountsController {
	return AccountsController{log: log, col: col, cred: cred}
}

func (ctrl *AccountsController) Get(ctx context.Context, id string) (Account, error) {
	var r Account
	_, err := ctrl.col.ReadDocument(nil, id, &r)
	return r, err
}

func (ctrl *AccountsController) Exists(ctx context.Context, id string) (bool, error) {
	return ctrl.col.DocumentExists(nil, id)
}

func (ctrl *AccountsController) Create(ctx context.Context, title string) (Account, error) {
	acc := Account{
		Title: title,
	}
	meta, err := ctrl.col.CreateDocument(ctx, acc)
	acc.DocumentMeta = meta
	return acc, err
}

// Grant account access to namespace
func (acc *Account) LinkNamespace(ctx context.Context, edge driver.Collection, ns Namespace, level int32) (error) {
	_, err := edge.CreateDocument(ctx, Access{
		From: acc.ID,
		To: ns.ID,
		Level: level,
		DocumentMeta: driver.DocumentMeta {
			Key: acc.Key + "-" + ns.Key,
		},
	})
	return err
}

// Grant namespace access to account
func (acc *Account) JoinNamespace(ctx context.Context, edge driver.Collection, ns Namespace, level int32) (error) {
	_, err := edge.CreateDocument(ctx, Access{
		From: ns.ID,
		To: acc.ID,
		Level: level,
		DocumentMeta: driver.DocumentMeta {
			Key: acc.Key + "-" + ns.Key,
		},
	})
	return err
}

// Set Account Credentials, ensure account has only one credentials document linked per credentials type
func (ctrl *AccountsController) SetCredentials(ctx context.Context, acc Account, edge driver.Collection, c Credentials) (error) {
	requestor, ok := ctx.Value(nocloud.NoCloudAccount).(string)
	if !ok {
		return status.Error(codes.Internal, "Account ID is not given")
	}
	if !HasAccess(ctx, ctrl.col.Database(), requestor, acc.ID.String(), access.ADMIN) {
		return status.Error(codes.PermissionDenied, "NoAccess")
	}

	cred, err := ctrl.cred.CreateDocument(ctx, c)	
	_, err = edge.CreateDocument(ctx, CredentialsLink{
		From: acc.ID,
		To: cred.ID,
		Type: c.Type(),
		DocumentMeta: driver.DocumentMeta {
			Key: c.Type() + "-" + acc.Key, // Ensure only one credentials vertex per type
		},
	})
	if err != nil {
		return status.Error(codes.Internal, "Couldn't create credentials")
	}
	return nil
}

func MakeCredentials(credentials accountspb.Credentials) (Credentials, error) {
	var cred Credentials;
	var err error;
	switch credentials.Type {
	case "standard":
		cred, err = NewStandardCredentials(credentials.Data[0], credentials.Data[1])
	default:
		return nil, errors.New("Auth type is wrong")
	}

	return cred, err
}

func (ctrl *AccountsController) Authorize(ctx context.Context, auth_type string, args ...string) (Account, bool) {
	var credentials Credentials;
	var ok bool;
	ctrl.log.Debug("Authorization request", zap.String("type", auth_type))
	switch auth_type {
	case "standard":
		credentials = &StandardCredentials{Username: args[0]}
		ok = credentials.Find(ctx, ctrl.col.Database())
	default:
		return Account{}, false
	}
	// Check if could find Credentials
	if !ok {
		ctrl.log.Info("Coudn't find credentials", zap.Skip())
		return Account{}, false
	}

	ok = credentials.Authorize(args...)
	// Check if could authorize
	if !ok {
		ctrl.log.Info("Coudn't authorize", zap.Skip())
		return Account{}, false
	}

	account, ok := credentials.Account(ctx, ctrl.col.Database())
	ctrl.log.Debug("Authorized account", zap.Bool("result", ok), zap.Any("account", account))
	return account, ok
}

func (ctrl *AccountsController) EnsureRootExists(passwd string) (err error) {
	exists, err := ctrl.col.DocumentExists(nil, "0")
	if err != nil {
		return err
	}

	var meta driver.DocumentMeta
	if !exists {
		meta, err = ctrl.col.CreateDocument(nil, Account{ 
			Title: "nocloud",
			DocumentMeta: driver.DocumentMeta { Key: "0" },
		})
		if err != nil {
			return err
		}
		ctrl.log.Debug("Created root Account", zap.Any("result", meta))
	}
	var root Account
	meta, err = ctrl.col.ReadDocument(nil, "0", &root)
	if err != nil {
		return err
	}
	root.DocumentMeta = meta

	ns_col, _ := ctrl.col.Database().Collection(nil, NAMESPACES_COL)
	exists, err = ns_col.DocumentExists(nil, "0")
	if !exists {
		meta, err := ns_col.CreateDocument(nil, Namespace{ 
			Title: "platform",
			DocumentMeta: driver.DocumentMeta { Key: "0" },
		})
		if err != nil {
			return err
		}
		ctrl.log.Debug("Created root Namespace", zap.Any("result", meta))
	}

	var rootNS Namespace
	meta, err = ns_col.ReadDocument(nil, "0", &rootNS)
	if err != nil {
		return err
	}
	rootNS.DocumentMeta = meta

	edge_col, _ := ctrl.col.Database().Collection(nil, ACC2NS)
	exists, err = edge_col.DocumentExists(nil, "0-0")
	if !exists {
		err = root.LinkNamespace(nil, edge_col, rootNS, 4)
		if err != nil {
			return err
		}
	}

	ctx := context.WithValue(context.Background(), nocloud.NoCloudAccount, "0")
	cred_edge_col, _ := ctrl.col.Database().Collection(nil, ACC2CRED)
	cred, err := NewStandardCredentials("nocloud", passwd)
	if err != nil {
		return err
	}

	exists, err = cred_edge_col.DocumentExists(nil, "standard-0")
	if !exists {
		err = ctrl.SetCredentials(ctx, root, cred_edge_col, cred)
		if err != nil {
			return err
		}
	}
	_, r := ctrl.Authorize(ctx, "standard", "nocloud", passwd)
	if !r {
		return errors.New("Cannot authorize nocloud")
	}
	return nil
}