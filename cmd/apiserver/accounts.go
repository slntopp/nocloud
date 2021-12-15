/*
Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz

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
package main

import (
	"context"

	"github.com/slntopp/nocloud/pkg/accounting/accountspb"
	"github.com/slntopp/nocloud/pkg/api/apipb"
)

type accountsAPI struct {
	client accountspb.AccountsServiceClient
	apipb.UnimplementedAccountsServiceServer
}

func (acc *accountsAPI) mustEmbedUnimplementedAccountsServiceServer() {}

func (acc *accountsAPI) Token(ctx context.Context, request *accountspb.TokenRequest) (*accountspb.TokenResponse, error) {
	return acc.client.Token(ctx, request)
}

func (acc *accountsAPI) Get(ctx context.Context, request *accountspb.GetRequest) (*accountspb.Account, error) {
	return acc.client.Get(ctx, request)
}

func (acc *accountsAPI) List(ctx context.Context, request *accountspb.ListRequest) (*accountspb.ListResponse, error) {
	return acc.client.List(ctx, request)
}

func (acc *accountsAPI) Create(ctx context.Context, request *accountspb.CreateRequest) (*accountspb.CreateResponse, error) {
	return acc.client.Create(ctx, request)
}

func (acc *accountsAPI) Update(ctx context.Context, request *accountspb.Account) (*accountspb.UpdateResponse, error) {
	return acc.client.Update(ctx, request)
}

func (acc *accountsAPI) SetCredentials(ctx context.Context, request *accountspb.SetCredentialsRequest) (*accountspb.SetCredentialsResponse, error) {
	return acc.client.SetCredentials(ctx, request)
}

func (acc *accountsAPI) Delete(ctx context.Context, request *accountspb.DeleteRequest) (*accountspb.DeleteResponse, error) {
	return acc.client.Delete(ctx, request)
}