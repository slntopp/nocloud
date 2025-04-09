/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	redis "github.com/go-redis/redis/v8"
	"github.com/heltonmarx/goami/ami"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"google.golang.org/protobuf/types/known/structpb"
	"math/rand"
	"slices"
	"time"

	elpb "github.com/slntopp/nocloud-proto/events_logging"
	"github.com/slntopp/nocloud-proto/notes"
	"github.com/slntopp/nocloud/pkg/nocloud/sessions"

	"github.com/arangodb/go-driver"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"

	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/credentials"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"

	pb "github.com/slntopp/nocloud-proto/registry"
	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	servicespb "github.com/slntopp/nocloud-proto/services"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	sc "github.com/slntopp/nocloud/pkg/settings/client"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	servicesClient servicespb.ServicesServiceClient
	amiService     string
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("SERVICES_HOST", "services-registry:8000")
	viper.SetDefault("AMI_SERVICE", "SomeService")
	servicesHost := viper.GetString("SERVICES_HOST")
	amiService = viper.GetString("AMI_SERVICE")

	servicesConn, err := grpc.Dial(servicesHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	servicesClient = servicespb.NewServicesServiceClient(servicesConn)
}

type AccountsServiceServer struct {
	pb.UnimplementedAccountsServiceServer
	db      driver.Database
	ctrl    graph.AccountsController
	ns_ctrl graph.NamespacesController
	ca      graph.CommonActionsController

	log         *zap.Logger
	SIGNING_KEY []byte

	rdb       redisdb.Client
	amiSocket *ami.Socket
}

func NewAccountsServer(log *zap.Logger, db driver.Database, rdb redisdb.Client, amiSocket *ami.Socket) *AccountsServiceServer {
	return &AccountsServiceServer{
		log: log, db: db,
		ctrl: graph.NewAccountsController(
			log.Named("AccountsController"), db,
		),
		ns_ctrl: graph.NewNamespacesController(
			log.Named("NamespacesController"), db,
		),
		ca: graph.NewCommonActionsController(
			log.Named("CommonActionsController"), db,
		),
		rdb:       rdb,
		amiSocket: amiSocket,
	}
}

func (s *AccountsServiceServer) SetupSettingsClient(settingsClient settingspb.SettingsServiceClient, internal_token string) {
	sc.Setup(
		s.log, metadata.AppendToOutgoingContext(
			context.Background(), "authorization", "bearer "+internal_token,
		), &settingsClient,
	)
	var settings AccountPostCreateSettings
	if scErr := sc.Fetch(accountPostCreateSettingsKey, &settings, defaultSettings); scErr != nil {
		s.log.Warn("Cannot fetch settings", zap.Error(scErr))
	}

	var stdSettings SignUpSettings
	if scErr := sc.Fetch(signupKey, &stdSettings, standartSettings); scErr != nil {
		s.log.Warn("Cannot fetch standart settings", zap.Error(scErr))
	}
}

const getOwnServices = `
FOR node, edge, path IN 2 OUTBOUND @account
	GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@services)
    RETURN node._key
`

const phoneVerificationDataKeyTemplate = "registry-phone-verification-%s"
const emailVerificationDataKeyTemplate = "registry-email-verification-%s"

type VerificationData struct {
	Code    string `json:"code"`
	Sent    int64  `json:"sent"`
	Expires int64  `json:"expires"`
}

func generateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *AccountsServiceServer) Verify(ctx context.Context, req *pb.VerificationRequest) (*pb.VerificationResponse, error) {
	log := s.log.Named("Verify")

	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	log = log.With(zap.String("requester", requester))
	log.Debug("Verify request received", zap.Any("request", req))

	rootNs := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	rootAccess := s.ca.HasAccess(ctx, requester, rootNs, access.Level_ROOT)
	if !rootAccess && (req.GetAccount() != requester) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	acc, err := s.ctrl.Get(ctx, req.GetAccount())
	if err != nil {
		log.Error("error obtaining account", zap.Error(err))
		return nil, err
	}
	if req.Type == pb.VerificationType_PHONE && acc.GetIsPhoneVerified() ||
		req.Type == pb.VerificationType_EMAIL && acc.GetIsEmailVerified() {
		return nil, fmt.Errorf("already verified")
	}
	dataVal := acc.GetData()
	if dataVal == nil {
		dataVal, _ = structpb.NewStruct(map[string]any{})
	}
	data := dataVal.AsMap()
	accountEmail, _ := data["email"].(string)
	accountPhone, _ := data["phone"].(string)
	log.Debug("account's phone and email", zap.String("phone", accountPhone), zap.String("email", accountEmail))

	var vData VerificationData
	var _key string
	switch req.Type {
	case pb.VerificationType_PHONE:
		_key = phoneVerificationDataKeyTemplate
	case pb.VerificationType_EMAIL:
		_key = emailVerificationDataKeyTemplate
	default:
		return nil, fmt.Errorf("unsupported verification type")
	}
	res, err := s.rdb.Get(ctx, fmt.Sprintf(_key, acc.GetUuid())).Result()
	if err != nil {
		if err.Error() == redis.Nil.Error() {
			res = ""
		} else {
			log.Error("failed to obtain verification data from redis", zap.Error(err))
			return nil, fmt.Errorf("internal error")
		}
	}
	if res == "" {
		res = "{}"
	}
	if err = json.Unmarshal([]byte(res), &vData); err != nil {
		log.Error("failed to unmarshal verification data", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	now := time.Now().Unix()
	if req.Type == pb.VerificationType_PHONE {

		if req.Action == pb.VerificationAction_BEGIN {

			if now-vData.Sent < 150 {
				return nil, fmt.Errorf("too many requests. Try again later")
			}

			code := generateCode()
			vData.Code = code
			vData.Expires = now + 600
			vData.Sent = now
			encoded, err := json.Marshal(vData)
			if err != nil {
				log.Error("failed to marshal verification data", zap.Error(err))
				return nil, fmt.Errorf("internal error")
			}
			if err = s.rdb.Set(ctx, fmt.Sprintf(phoneVerificationDataKeyTemplate, acc.GetUuid()), string(encoded), 0).Err(); err != nil {
				log.Error("failed to save verification data", zap.Error(err))
				return nil, fmt.Errorf("internal error")
			}

			if s.amiSocket != nil {
				uuid, _ := ami.GetUUID()
				to := accountPhone
				smsBody := fmt.Sprintf("Ваш проверочный код: %s", code)
				command := fmt.Sprintf("dongle sms %s %s %s", amiService, to, smsBody)
				log.Debug("trying to send message via AMI")
				if resp, err := ami.Command(s.amiSocket, uuid, command); err != nil {
					log.Error("failed to send ami message", zap.Error(err), zap.Any("response", resp))
					return nil, fmt.Errorf("internal error")
				} else {
					log.Debug("AMI response", zap.Any("response", resp))
				}
			} else {
				log.Error("Cannot send SMS. AMI is not initialized")
			}

		}

		if req.Action == pb.VerificationAction_APPROVE {
			if vData.Code == "" {
				log.Error("no saved code")
				return nil, fmt.Errorf("can't approve. You must request code first")
			}
			if vData.Code != req.GetSecureCode() || now > vData.Expires {
				return nil, fmt.Errorf("invalid code")
			}
			if err = s.ctrl.Update(ctx, acc, map[string]interface{}{"is_phone_verified": true}); err != nil {
				log.Error("failed to update account", zap.Error(err))
				return nil, fmt.Errorf("internal error")
			}
		}
	} else if req.Type == pb.VerificationType_EMAIL {
		return nil, fmt.Errorf("email verification currently disabled")
	} else {
		return nil, fmt.Errorf("unsupported verification type")
	}

	return &pb.VerificationResponse{Result: true}, nil
}

func (s *AccountsServiceServer) Suspend(ctx context.Context, req *accountspb.SuspendRequest) (*accountspb.SuspendResponse, error) {
	log := s.log.Named("SuspendAccount")
	log.Debug("Suspend request received", zap.Any("request", req), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	accId := driver.NewDocumentID(schema.ACCOUNTS_COL, req.Uuid)

	cursor, err := s.db.Query(ctx, getOwnServices, map[string]interface{}{
		"account":     accId,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@services":   schema.SERVICES_COL,
	})
	if err != nil {
		log.Error("Error Quering Services to Suspend", zap.Error(err))
		return &accountspb.SuspendResponse{Result: false}, err
	}
	defer cursor.Close()

	for {
		var serviceUuid string
		meta, err := cursor.ReadDocument(ctx, &serviceUuid)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Error("Error Reading Service UUID", zap.Error(err), zap.Any("meta", meta))
			continue
		}
		if _, err := servicesClient.Suspend(ctx, &servicespb.SuspendRequest{Uuid: serviceUuid}); err != nil {
			log.Error("Error Suspending Service", zap.Error(err))
		}
	}

	acc, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		log.Debug("Error getting account", zap.String("requested_id", req.Uuid), zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if err := s.ctrl.Update(ctx, acc, map[string]interface{}{"suspended": true}); err != nil {
		log.Debug("Error updating account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while updating account")
	}

	nocloud.Log(log, &elpb.Event{
		Entity:    schema.ACCOUNTS_COL,
		Uuid:      acc.Key,
		Scope:     "database",
		Action:    "suspended",
		Rc:        0,
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	})

	log.Debug("Suspend request end")
	return &accountspb.SuspendResponse{Result: true}, nil
}

func (s *AccountsServiceServer) Unsuspend(ctx context.Context, req *accountspb.UnsuspendRequest) (*accountspb.UnsuspendResponse, error) {
	log := s.log.Named("UnsuspendAccount")
	log.Debug("Unsuspend request received", zap.Any("request", req), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	accId := driver.NewDocumentID(schema.ACCOUNTS_COL, req.Uuid)

	cursor, err := s.db.Query(ctx, getOwnServices, map[string]interface{}{
		"account":     accId,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@services":   schema.SERVICES_COL,
	})
	if err != nil {
		log.Error("Error Quering Services to Unsuspend", zap.Error(err))
		return &accountspb.UnsuspendResponse{Result: false}, err
	}
	defer cursor.Close()

	for {
		var serviceUuid string
		meta, err := cursor.ReadDocument(ctx, &serviceUuid)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Error("Error Reading Service UUID", zap.Error(err), zap.Any("meta", meta))
			continue
		}
		if _, err := servicesClient.Unsuspend(ctx, &servicespb.UnsuspendRequest{Uuid: serviceUuid}); err != nil {
			log.Error("Error Unsuspending Service", zap.Error(err))
		}
	}

	acc, err := s.ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		log.Debug("Error getting account", zap.String("requested_id", req.GetUuid()), zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if err := s.ctrl.Update(ctx, acc, map[string]interface{}{"suspended": false}); err != nil {
		log.Debug("Error updating account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while updating account")
	}

	nocloud.Log(log, &elpb.Event{
		Entity:    schema.ACCOUNTS_COL,
		Uuid:      acc.Key,
		Scope:     "database",
		Action:    "unsuspended",
		Rc:        0,
		Requestor: requestor,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	})

	log.Debug("Unsuspend request end")
	return &accountspb.UnsuspendResponse{Result: true}, nil
}

func (s *AccountsServiceServer) Get(ctx context.Context, request *accountspb.GetRequest) (*accountspb.Account, error) {
	log := s.log.Named("GetAccount")
	log.Debug("Get request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	requested := request.GetUuid()
	if requested == "me" {
		requested = requestor
	}

	log.Debug("Retrieving account", zap.String("uuid", requested))
	acc, err := s.ctrl.GetWithAccess(ctx, requestorId, requested)
	if err != nil || acc.Access == nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if acc.Account == nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	log.Debug("Retrieved account", zap.Any("account", acc))

	// Provide public information without access check
	if request.GetPublic() {
		return &accountspb.Account{Title: acc.Account.GetTitle()}, nil
	}

	if acc.Key == requestor {
		return acc.Account, nil
	}

	if acc.Access.Level < access.Level_READ {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	if acc.Access.Level < access.Level_ROOT {
		acc.SuspendConf = nil
	}

	return acc.Account, nil
}

func (s *AccountsServiceServer) List(ctx context.Context, request *accountspb.ListRequest) (*accountspb.ListResponse, error) {
	log := s.log.Named("ListAccounts")
	log.Debug("List request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.ctrl.Get(ctx, requestor)
	if err != nil {
		log.Debug("Error getting requestor account", zap.Any("error", err))
		return nil, status.Error(codes.PermissionDenied, "Requestor Account not found")
	}

	depth := request.GetDepth()
	if request.GetDepth() == 0 {
		depth = 4
	}

	page := request.GetPage()
	limit := request.GetLimit()

	offset := (page - 1) * limit

	pool, count, active, err := s.ctrl.ListImproved(ctx, acc.GetUuid(), depth, offset, limit, request.GetField(), request.GetSort(), request.GetFilters())
	if err != nil {
		log.Debug("Error listing accounts", zap.Any("error", err))
		return nil, status.Error(codes.Internal, "Error listing accounts")
	}
	log.Debug("List result", zap.Any("pool", pool))

	result := make([]*accountspb.Account, len(pool))
	for i, acc := range pool {
		if acc.Access.Level < access.Level_ROOT {
			acc.Account.SuspendConf = nil
		}
		result[i] = acc.Account
	}
	log.Debug("Convert result", zap.Any("pool", result))

	return &accountspb.ListResponse{
		Pool:   result,
		Count:  count,
		Active: active,
	}, nil
}

func (s *AccountsServiceServer) Token(ctx context.Context, request *accountspb.TokenRequest) (*accountspb.TokenResponse, error) {
	log := s.log.Named("Token")
	log.Debug("Token request received", zap.Any("request", request))

	var acc graph.Account
	var ok bool

	requestor := ctx.Value(nocloud.NoCloudAccount)

	if requestor != nil && request.Uuid != nil {
		var err error
		acc, err = s.ctrl.Get(ctx, *request.Uuid)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		reqAcc, err := s.ctrl.Get(ctx, requestor.(string))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if acc.ID == reqAcc.ID {
			acc.Access.Level = reqAcc.Access.Level
		}
		if acc.Access.Level < access.Level(access.Level_ROOT) {
			log.Warn("Need SUDO Access token", zap.String("requestor", requestor.(string)))
			return nil, status.Error(codes.Unauthenticated, "Wrong credentials")
		}
		request.Exp = int32(time.Now().Unix() + int64(time.Minute.Seconds())*10)
	} else {
		if request.GetAuth() == nil {
			return nil, status.Error(codes.InvalidArgument, "Auth data was not presented")
		}
		acc, ok = s.ctrl.Authorize(ctx, request.Auth.Type, request.Auth.Data...)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "Wrong credentials given")
		}
	}

	log.Debug("Authorized user", zap.String("ID", acc.ID.String()))

	session := sessions.New(int64(request.GetExp()), acc.Key)
	if err := sessions.Store(s.rdb, acc.Key, session); err != nil {
		log.Error("Failed to store session", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to issue token: session")
	}

	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = acc.Key
	claims[nocloud.NOCLOUD_SESSION_CLAIM] = session.GetId()
	claims["expires"] = request.GetExp()

	if request.GetRootClaim() {
		ns := driver.NewDocumentID(schema.NAMESPACES_COL, "0")
		ok, lvl := s.ca.AccessLevel(ctx, acc.Key, ns)
		if !ok {
			lvl = access.Level_NONE
		}
		log.Debug("Adding Root claim to the token", zap.Int32("access_lvl", int32(lvl)))
		claims[nocloud.NOCLOUD_ROOT_CLAIM] = lvl
	}

	if sp := request.GetSpClaim(); sp != "" {
		ok, lvl := s.ca.AccessLevel(ctx, acc.Key, driver.NewDocumentID(schema.SERVICES_PROVIDERS_COL, sp))
		if !ok {
			lvl = access.Level_NONE
		}
		log.Debug("Adding ServicesProvider claim to the token", zap.Int32("access_lvl", int32(lvl)))
		claims[nocloud.NOCLOUD_SP_CLAIM] = lvl
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token_string, err := token.SignedString(s.SIGNING_KEY)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to issue token")
	}

	var event = &elpb.Event{
		Entity:    schema.ACCOUNTS_COL,
		Uuid:      acc.Key,
		Scope:     "database",
		Action:    "login",
		Rc:        0,
		Requestor: acc.Key,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: "",
		},
	}

	nocloud.Log(log, event)

	return &accountspb.TokenResponse{Token: token_string}, nil
}

func (s *AccountsServiceServer) Create(ctx context.Context, request *accountspb.CreateRequest) (*accountspb.CreateResponse, error) {
	log := s.log.Named("CreateAccount")
	log.Debug("Create request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	var nsToConnect = request.GetNamespace()
	isSubaccount := request.GetAccountOwner() != ""
	var motherAcc graph.Account
	if isSubaccount {
		var err error
		motherAcc, err = s.ctrl.Get(ctx, request.GetAccountOwner())
		if err != nil {
			log.Error("Error getting account owner", zap.Error(err), zap.String("account", request.GetAccountOwner()))
			return nil, err
		}
		ok, access_lvl := s.ca.AccessLevel(ctx, requestor, motherAcc.ID)
		if !ok {
			log.Warn("No Access", zap.String("requestor", requestor), zap.String("mother", motherAcc.ID.String()))
			return nil, status.Error(codes.PermissionDenied, "No Access")
		} else if access_lvl < access.Level_ADMIN {
			log.Warn("No Enough Rights", zap.String("requestor", requestor), zap.String("mother", motherAcc.ID.String()))
			return nil, status.Error(codes.PermissionDenied, "No Enough Rights")
		}
		if motherAcc.GetAccountOwner() != "" {
			log.Warn("Subaccounts cannot create subaccounts", zap.String("account", motherAcc.GetUuid()))
			return nil, status.Error(codes.InvalidArgument, "Subaccounts cannot create subaccounts")
		}
		existNs, err := s.ctrl.GetNamespace(ctx, motherAcc)
		if err != nil {
			if err.Error() == "no more documents" {
				personal_ctx := context.WithValue(ctx, nocloud.NoCloudAccount, request.GetAccountOwner())
				createdNs, err := s.ns_ctrl.Create(personal_ctx, motherAcc.GetTitle())
				if err != nil {
					log.Error("Error creating namespace for account owner", zap.Error(err), zap.String("namespace", request.Namespace))
					return nil, err
				}
				nsToConnect = createdNs.GetUuid()
			} else {
				log.Error("Error getting namespace", zap.Error(err), zap.String("namespace", request.Namespace))
				return nil, err
			}
		} else {
			nsToConnect = existNs.GetUuid()
		}
	}

	ns, err := s.ns_ctrl.Get(ctx, nsToConnect)
	if err != nil {
		log.Error("Error getting namespace", zap.Error(err), zap.String("namespace", request.Namespace))
		return nil, err
	}

	ok, access_lvl := s.ca.AccessLevel(ctx, requestor, ns.ID)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "No Access")
	} else if access_lvl < access.Level_MGMT {
		return nil, status.Error(codes.PermissionDenied, "No Enough Rights")
	}

	cred, err := credentials.MakeCredentials(request.Auth, log)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if cred.Find(ctx, s.db) {
		return nil, status.Error(codes.AlreadyExists, "Such username also exists")
	}

	creationAccount := accountspb.Account{
		Title:        request.Title,
		Currency:     request.Currency,
		Data:         request.GetData(),
		AccountOwner: request.GetAccountOwner(),
	}

	acc, err := s.ctrl.Create(ctx, creationAccount)
	if err != nil {
		log.Error("Error creating account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while creating account")
	}
	res := &accountspb.CreateResponse{Uuid: acc.Key}

	if request.Access != nil && access.Level(*request.Access) < access_lvl {
		access_lvl = access.Level(*request.Access)
	}

	var settings AccountPostCreateSettings
	if scErr := sc.Fetch(accountPostCreateSettingsKey, &settings, defaultSettings); scErr != nil {
		log.Warn("Cannot fetch settings", zap.Error(scErr))
	}
	settings.CreateNamespace = true // Always create namespace

	if settings.CreateNamespace {
		personal_ctx := context.WithValue(ctx, nocloud.NoCloudAccount, acc.GetUuid())

		createdNs, err := s.ns_ctrl.Create(personal_ctx, acc.GetTitle())
		if err != nil {
			log.Error("Cannot create a namespace for new Account", zap.String("account", acc.GetUuid()), zap.Error(err))
		} else {
			res.Namespace = createdNs.ID.Key()
		}
	}

	col, _ := s.db.Collection(ctx, schema.NS2ACC)
	err = acc.JoinNamespace(ctx, col, ns, access_lvl, roles.OWNER)
	if err != nil {
		log.Error("Error linking to namespace")
		return res, err
	}

	col, _ = s.db.Collection(ctx, schema.CREDENTIALS_EDGE_COL)
	err = s.ctrl.SetCredentials(ctx, acc, col, cred, roles.OWNER)
	if err != nil {
		return res, err
	}

	// Patch mother account with new subaccount and link namespace
	if isSubaccount {
		subaccounts := motherAcc.GetSubaccounts()
		subaccounts = append(subaccounts, acc.GetUuid())
		if err = s.ctrl.Update(ctx, motherAcc, map[string]interface{}{
			"subaccounts": subaccounts,
		}); err != nil {
			log.Error("Error updating mother account with new subaccount", zap.Error(err))
			return res, err
		}
		col, _ := s.db.Collection(ctx, schema.NS2ACC)
		accNs, err := s.ctrl.GetNamespace(ctx, acc)
		if err != nil {
			log.Error("Error getting personal namespace", zap.Error(err))
			return res, err
		}
		if err := motherAcc.JoinNamespace(ctx, col, accNs, access.Level_MGMT, roles.DEFAULT); err != nil {
			log.Error("Error joining child namespace to mother account", zap.Error(err))
			return res, err
		}
		log.Debug("Subaccount created and linked", zap.String("subaccount", acc.GetUuid()))
	}

	return res, nil
}

func (s *AccountsServiceServer) SignUp(ctx context.Context, request *accountspb.CreateRequest) (*accountspb.CreateResponse, error) {
	log := s.log.Named("SignUp")
	log.Debug("Create request received", zap.Any("request", request), zap.Any("context", ctx))

	if request.GetAccountOwner() != "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot create subaccount during SignUp")
	}

	var stdSettings SignUpSettings
	if scErr := sc.Fetch(signupKey, &stdSettings, standartSettings); scErr != nil {
		log.Warn("Cannot fetch settings", zap.Error(scErr))
	}

	if !stdSettings.Enabled {
		return nil, status.Error(codes.Unavailable, "SignUp is disabled")
	}

	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY)

	ns, err := s.ns_ctrl.Get(ctx, stdSettings.Namespace)
	if err != nil {
		log.Debug("Error getting namespace", zap.Error(err), zap.String("namespace", stdSettings.Namespace))
		return nil, err
	}

	ok, access_lvl := s.ca.AccessLevel(ctx, schema.ROOT_ACCOUNT_KEY, ns.ID)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "No Access")
	} else if access_lvl < access.Level_MGMT {
		return nil, status.Error(codes.PermissionDenied, "No Enough Rights")
	}

	cred, err := credentials.MakeCredentials(request.Auth, log)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !slices.Contains(stdSettings.AllowedTypes, cred.Type()) {
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("Such auth type not allowed. Type: %s", cred.Type()))
	}

	if cred.Find(ctx, s.db) {
		return nil, status.Error(codes.AlreadyExists, "Such username also exists")
	}

	var accStatus accountspb.AccountStatus

	if stdSettings.EnabledAccount {
		accStatus = accountspb.AccountStatus_ACTIVE
	} else {
		accStatus = accountspb.AccountStatus_LOCK
	}

	creationAccount := accountspb.Account{
		Title:    request.Title,
		Currency: request.Currency,
		Data:     request.GetData(),
		Status:   accStatus,
	}

	acc, err := s.ctrl.Create(ctx, creationAccount)
	if err != nil {
		log.Debug("Error creating account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while creating account")
	}
	res := &accountspb.CreateResponse{Uuid: acc.Key}

	if request.Access != nil && access.Level(*request.Access) < access_lvl {
		access_lvl = access.Level(*request.Access)
	}

	var settings AccountPostCreateSettings
	if scErr := sc.Fetch(accountPostCreateSettingsKey, &settings, defaultSettings); scErr != nil {
		log.Warn("Cannot fetch settings", zap.Error(scErr))
	}

	if settings.CreateNamespace {
		personal_ctx := context.WithValue(ctx, nocloud.NoCloudAccount, acc.GetUuid())

		createdNs, err := s.ns_ctrl.Create(personal_ctx, acc.GetTitle())
		if err != nil {
			log.Warn("Cannot create a namespace for new Account", zap.String("account", acc.GetUuid()), zap.Error(err))
		} else {
			res.Namespace = createdNs.ID.Key()
		}
	}

	col, _ := s.db.Collection(ctx, schema.NS2ACC)
	err = acc.JoinNamespace(ctx, col, ns, access_lvl, roles.OWNER)
	if err != nil {
		log.Debug("Error linking to namespace")
		return res, err
	}

	col, _ = s.db.Collection(ctx, schema.CREDENTIALS_EDGE_COL)

	err = s.ctrl.SetCredentials(ctx, acc, col, cred, roles.OWNER)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Update Account
// Supports updating Title and Data
//
//	Updating Data rules:
//		1. If Data is nil - it'll be skipped
//		2. If Data is not nil but has no keys - it'll be wiped
//		3. If value of one of the Data keys is nil - it'll be deleted from Data
func (s *AccountsServiceServer) Update(ctx context.Context, request *accountspb.Account) (*accountspb.UpdateResponse, error) {
	log := s.log.Named("UpdateAccount")
	log.Debug("Update request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.ctrl.GetWithAccess(ctx, requestorId, request.GetUuid())
	if err != nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if acc.Access == nil {
		log.Warn("Error Access is nil")
		return nil, status.Error(codes.PermissionDenied, "Error Access is nil")
	}

	if requestor == request.GetUuid() {
		acc.Access.Level = access.Level_ROOT
	}

	if acc.Access.Level < access.Level_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	if acc.Access.Level < access.Level_ROOT {
		request.SuspendConf = nil
		request.Suspended = nil
	}

	patch := make(map[string]interface{})

	if acc.Title != request.Title && request.Title != "" {
		log.Debug("Title patch detected")
		patch["title"] = request.Title
	}

	if request.Currency != nil {
		if acc.GetCurrency() != nil && (acc.GetCurrency().GetId() != request.GetCurrency().GetId()) {
			log.Debug("Currency patch detected")
			patch["currency"] = request.GetCurrency()
		}
	}

	if acc.GetStatus() != request.GetStatus() {
		log.Debug("Status patch detected")
		patch["status"] = request.GetStatus()
	}

	if request.Data == nil {
		log.Debug("Data patch is not present, skipping")
		goto patch
	}

	if len(request.Data.AsMap()) == 0 {
		log.Debug("Data patch is empty, wiping data")
		patch["data"] = nil
		goto patch
	}

	log.Debug("Merging data")
	patch["data"] = MergeMaps(acc.Data.AsMap(), request.Data.AsMap())

patch:
	if len(patch) == 0 {
		log.Debug("Resulting patch is empty")
		return nil, status.Error(codes.InvalidArgument, "Nothing changed")
	}

	log.Debug("Updating Account", zap.String("account", request.Uuid), zap.Any("patch", patch))
	err = s.ctrl.Update(ctx, acc, patch)
	if err != nil {
		log.Debug("Error updating account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while updating account")
	}

	return &accountspb.UpdateResponse{Result: true}, nil
}

func (s *AccountsServiceServer) EnsureRootExists(passwd string) error {
	return s.ctrl.EnsureRootExists(passwd)
}

func (s *AccountsServiceServer) SetCredentials(ctx context.Context, request *accountspb.SetCredentialsRequest) (*accountspb.SetCredentialsResponse, error) {
	log := s.log.Named("SetCredentials")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.ctrl.Get(ctx, request.Account)
	if err != nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if !s.ca.HasAccess(ctx, requestor, acc.ID, access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "NoAccess")
	}

	auth := request.GetAuth()

	edge, _ := s.db.Collection(ctx, schema.ACC2CRED)
	old_cred_key, has_credentials := s.ctrl.GetCredentials(ctx, edge, acc, auth.Type)
	log.Debug("Checking if has credentials", zap.Bool("has_credentials", has_credentials), zap.Any("old_credentials", old_cred_key))

	cred, err := credentials.MakeCredentials(auth, log)
	if err != nil {
		log.Debug("Error creating new credentials", zap.String("type", auth.Type), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error creading new credentials")
	}

	if has_credentials {
		err = s.ctrl.UpdateCredentials(ctx, old_cred_key, cred)
	} else {
		err = s.ctrl.SetCredentials(ctx, acc, edge, cred, roles.OWNER)
	}

	if err != nil {
		log.Debug("Error updating/setting credentials", zap.String("type", auth.Type), zap.Error(err))
		return nil, status.Error(codes.Internal, "Error updating/setting credentials")
	}

	return &accountspb.SetCredentialsResponse{Result: true}, nil
}

func (s *AccountsServiceServer) Delete(ctx context.Context, request *accountspb.DeleteRequest) (*accountspb.DeleteResponse, error) {
	log := s.log.Named("Delete")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.ctrl.Get(ctx, request.Uuid)
	if err != nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if !s.ca.HasAccess(ctx, requestor, acc.ID, access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "NoAccess")
	}

	err = acc.Delete(ctx, s.db)
	if err != nil {
		log.Debug("Error deleting account and it's children", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting account")
	}

	return &accountspb.DeleteResponse{Result: true}, nil
}

func (s *AccountsServiceServer) AddNote(ctx context.Context, request *notes.AddNoteRequest) (*notes.NoteResponse, error) {
	log := s.log.Named("AddNote")
	log.Debug("AddNote request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.ctrl.GetWithAccess(ctx, requestorId, request.GetUuid())
	if err != nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if acc.Access == nil {
		log.Warn("Error Access is nil")
		return nil, status.Error(codes.PermissionDenied, "Error Access is nil")
	}

	if requestor == request.GetUuid() {
		acc.Access.Level = access.Level_ROOT
	}

	if acc.Access.Level < access.Level_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	acc.AdminNotes = append(acc.GetAdminNotes(), &notes.AdminNote{
		Admin:   requestor,
		Msg:     request.GetMsg(),
		Created: time.Now().Unix(),
	})

	patch := map[string]any{
		"admin_notes": acc.GetAdminNotes(),
	}

	err = s.ctrl.Update(ctx, acc, patch)
	if err != nil {
		log.Error("Failed to add note", zap.Error(err))
		return nil, err
	}

	return &notes.NoteResponse{Result: true, AdminNotes: acc.GetAdminNotes()}, nil
}

func (s *AccountsServiceServer) PatchNote(ctx context.Context, request *notes.PatchNoteRequest) (*notes.NoteResponse, error) {
	log := s.log.Named("PatchNote")
	log.Debug("PatchNote request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.ctrl.GetWithAccess(ctx, requestorId, request.GetUuid())
	if err != nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if acc.Access == nil {
		log.Warn("Error Access is nil")
		return nil, status.Error(codes.PermissionDenied, "Error Access is nil")
	}

	if requestor == request.GetUuid() {
		acc.Access.Level = access.Level_ROOT
	}

	if acc.Access.Level < access.Level_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	note := acc.GetAdminNotes()[request.GetIndex()]
	if (note.GetAdmin() == requestor) || (note.GetAdmin() != requestor && acc.Access.GetLevel() == access.Level_ROOT) {
		note.Admin = requestor
		note.Msg = request.GetMsg()
		note.Updated = time.Now().Unix()

		patch := map[string]any{
			"admin_notes": acc.GetAdminNotes(),
		}

		err = s.ctrl.Update(ctx, acc, patch)
		if err != nil {
			log.Error("Failed to patch note", zap.Error(err))
			return nil, err
		}
		return &notes.NoteResponse{Result: true, AdminNotes: acc.GetAdminNotes()}, nil
	}

	return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
}

func (s *AccountsServiceServer) RemoveNote(ctx context.Context, request *notes.RemoveNoteRequest) (*notes.NoteResponse, error) {
	log := s.log.Named("RemoveNote")
	log.Debug("RemoveNote request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	requestorId := driver.NewDocumentID(schema.ACCOUNTS_COL, requestor)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.ctrl.GetWithAccess(ctx, requestorId, request.GetUuid())
	if err != nil {
		log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if acc.Access == nil {
		log.Warn("Error Access is nil")
		return nil, status.Error(codes.PermissionDenied, "Error Access is nil")
	}

	if requestor == request.GetUuid() {
		acc.Access.Level = access.Level_ROOT
	}

	if acc.Access.Level < access.Level_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	note := acc.GetAdminNotes()[request.GetIndex()]
	if (note.GetAdmin() == requestor) || (note.GetAdmin() != requestor && acc.Access.GetLevel() == access.Level_ROOT) {
		acc.AdminNotes = slices.Delete(acc.AdminNotes, int(request.GetIndex()), int(request.GetIndex()+1))

		patch := map[string]any{
			"admin_notes": acc.GetAdminNotes(),
		}

		err = s.ctrl.Update(ctx, acc, patch)
		if err != nil {
			log.Error("Failed to remove note", zap.Error(err))
			return nil, err
		}
		return &notes.NoteResponse{Result: true, AdminNotes: acc.GetAdminNotes()}, nil
	}

	return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account notes")
}
