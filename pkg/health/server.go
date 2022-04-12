/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
package health

import (
	"context"
	"strings"
	"sync"

	pb "github.com/slntopp/nocloud/pkg/health/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	registryHost, servicesHost, servicesProvidersHost,
	settingsHost, dnsHost, statesHost, edgeHost string
)

var grpc_services []string

func init() {
	viper.AutomaticEnv()

	viper.SetDefault("PROBABLES", "registry:8080,services-registry:8080,sp-registry:8080,settings:8080,dns-mgmt:8080,states:8080,edge:8080")
	grpc_services = strings.Split(viper.GetString("PROBABLES"), ",")
}

type HealthServiceServer struct {
	pb.UnimplementedHealthServiceServer
	ctx context.Context
	log *zap.Logger
}

func NewServer(log *zap.Logger, ctx context.Context) *HealthServiceServer {
	return &HealthServiceServer{log: log, ctx: ctx}
}

func (s *HealthServiceServer) Probe(ctx context.Context, request *pb.ProbeRequest) (*pb.ProbeResponse, error) {
	log := s.log.Named("Health Probe")
	log.Info("Probe received", zap.String("Type", request.ProbeType))

	switch(request.ProbeType) {
	case "PING":
		return &pb.ProbeResponse{
			Response: "PONG",
			Status: pb.Status_SERVING,
			Serving: []*pb.ServingStatus{{
				Service: "health",
				Status: pb.Status_SERVING,
			}},
		}, nil
	case "services":
		return s.CheckServices(ctx, request)
	case "routines":
		return s.CheckRoutines(ctx, request)
	}

	return &pb.ProbeResponse{
		Response: "ok",
		Status: pb.Status_SERVING,
		Serving: []*pb.ServingStatus{{
			Service: "health",
			Status: pb.Status_SERVING,
		}},
	}, nil
}

func (s *HealthServiceServer) CheckServices(ctx context.Context, request *pb.ProbeRequest) (*pb.ProbeResponse, error) {
	check_services_ch := make(chan *pb.ServingStatus, len(grpc_services))
	var wg sync.WaitGroup
	
	for _, service := range grpc_services {
		wg.Add(1)
		go func(service string) {
			defer wg.Done()
			s.log.Debug("Dialing Service", zap.String("service", service))
			conn, err := grpc.Dial(service, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				s.log.Error("Dial returned Error", zap.Error(err))
				err_string := err.Error()
				check_services_ch <- &pb.ServingStatus{
					Service: service,
					Status: pb.Status_OFFLINE,
					Error: &err_string,
				}
				s.log.Debug("Sent to channel", zap.String("service", service))
				return
			}
			client := pb.NewInternalProbeServiceClient(conn)

			s.log.Debug("Testing Service", zap.String("service", service))
			r, err := client.Service(s.ctx, request)
			if err != nil {
				s.log.Error("Testing returned Error", zap.Error(err))
				err_string := err.Error()
				check_services_ch <- &pb.ServingStatus{
					Service: service,
					Status: pb.Status_INTENAL,
					Error: &err_string,
				}
				s.log.Debug("Sent to channel", zap.String("service", service))
				return
			}
			s.log.Debug("Service tested", zap.String("service", service))
			check_services_ch <- r
			s.log.Debug("Sent to channel", zap.String("service", service))
		}(service)
	}

	s.log.Debug("Waiting for tests")
	wg.Wait()
	s.log.Debug("Tests completed, processing")

	res := &pb.ProbeResponse{Status: pb.Status_SERVING}
	for i := 0; i < len(grpc_services); i++ {
		r := <- check_services_ch
		s.log.Debug("Received response", zap.String("service", r.GetService()))
		res.Serving = append(res.Serving, r)
		if r.Status != pb.Status_SERVING {
			res.Status = pb.Status_HASERRS
		}
	}

	return res, nil
}

func (s *HealthServiceServer) CheckRoutines(ctx context.Context, request *pb.ProbeRequest) (*pb.ProbeResponse, error) {
	log := s.log.Named("CheckRoutines")
	log.Debug("Checking Services Routines", zap.Strings("services", grpc_services))
	check_routines_ch := make(chan *pb.RoutineStatus)
	var wg sync.WaitGroup

	for _, service := range grpc_services {
		wg.Add(1)
		go func(service string) {
			defer wg.Done()
			log.Debug("Dialing Service", zap.String("service", service))
			conn, err := grpc.Dial(service, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Error("Dial returned Error", zap.Error(err))
				err_string := err.Error()
				check_routines_ch <- &pb.RoutineStatus{
					Status: &pb.ServingStatus{
						Service: service,
						Status: pb.Status_OFFLINE,
						Error: &err_string,
					},
				}
				log.Debug("Sent to channel", zap.String("service", service))
				return
			}
			client := pb.NewInternalProbeServiceClient(conn)

			log.Debug("Testing Service", zap.String("service", service))
			r, err := client.Routine(s.ctx, request)
			if err != nil {
				log.Error("Testing returned Error", zap.Error(err))
				err_string := err.Error()
				check_routines_ch <- &pb.RoutineStatus{
					Status: &pb.ServingStatus{
						Service: service,
						Status: pb.Status_INTENAL,
						Error: &err_string,
					},
				}
				log.Debug("Sent to channel", zap.String("service", service))
				return
			}
			log.Debug("Service tested", zap.String("service", service))
			for _, rt := range r.Routines {
				if rt.Status.Service == "" {
					rt.Status.Service = service
				}
				check_routines_ch <- rt
			}
			log.Debug("Sent to channel", zap.String("service", service))
		}(service)
	}

	go func() {
		log.Debug("Waiting for tests")
		wg.Wait()
		log.Debug("Tests completed, closing channel")
		close(check_routines_ch)
	}()

	res := &pb.ProbeResponse{Status: pb.Status_RUNNING}
	for r := range check_routines_ch {
		log.Debug("Received response", zap.String("service", r.GetStatus().GetService()))
		res.Routines = append(res.Routines, r)
		if r.GetStatus().GetStatus() != pb.Status_RUNNING && r.GetStatus().GetStatus() != pb.Status_NOEXIST {
			res.Status = pb.Status_HASERRS
		}
	}

	return res, nil
}