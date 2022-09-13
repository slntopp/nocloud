package services

import (
	"context"

	"github.com/slntopp/nocloud/pkg/instances/proto"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
)

const MIN_DRIVE_SIZE = "min_drive_size"
const MAX_DRIVE_SIZE = "max_drive_size"

// Check whether instance satisfies SP vars limitations
// e.g. disk size
func (s *ServicesServer) ensureSPBounds(ctx context.Context, instance *proto.Instance, ig *proto.InstancesGroup) *pb.TestConfigError {
	log := s.log.Named("Ensure Services Provider Bounds")
	log.Info("Running bounds check")

	sp, err := s.sp_ctrl.Get(ctx, ig.GetSp())
	if err != nil {
		log.Warn("Cannot gather sp by provided id from instance group", zap.Error(err))
		return &pb.TestConfigError{
			InstanceGroup: ig.GetTitle(),
			Instance:      instance.GetTitle(),
			Error:         "Cannot gather sp by provided id from instance group",
		}
	}
	resources := instance.GetResources()
	vars := sp.GetVars()

	size, ok := resources["drive_size"]
	if !ok {
		log.Warn("No drive_size resource field provided")
		return &pb.TestConfigError{
			InstanceGroup: ig.GetTitle(),
			Instance:      instance.GetTitle(),
			Error:         "No drive_size resource field provided",
		}
	}
	driveType, ok := resources["drive_type"]
	if !ok {
		log.Warn("No drive_type resource field provided")
		return &pb.TestConfigError{
			InstanceGroup: ig.GetTitle(),
			Instance:      instance.GetTitle(),
			Error:         "No drive_type resource field provided",
		}
	}
	drive := driveType.String()

	min, ok := vars[MIN_DRIVE_SIZE]
	if !ok {
		return nil
	}
	if minSize, ok := min.GetValue()[drive]; ok && minSize.GetNumberValue() > size.GetNumberValue() {
		log.Warn("Out of bounds", zap.Float64("min_size", minSize.GetNumberValue()))
		return &pb.TestConfigError{
			InstanceGroup: ig.GetTitle(),
			Instance:      instance.GetTitle(),
			Error:         "Provided drive_size is less than sp minimum",
		}
	} else {
		if minSize, ok := min.GetValue()["default"]; ok && minSize.GetNumberValue() > size.GetNumberValue() {
			log.Warn("Out of bounds", zap.Float64("min_size", minSize.GetNumberValue()))
			return &pb.TestConfigError{
				InstanceGroup: ig.GetTitle(),
				Instance:      instance.GetTitle(),
				Error:         "Provided drive_size is less than sp minimum",
			}
		}
	}

	max, ok := vars[MAX_DRIVE_SIZE]
	if !ok {
		return nil
	}
	if maxSize, ok := max.GetValue()[drive]; ok && maxSize.GetNumberValue() < size.GetNumberValue() {
		log.Warn("Out of bounds", zap.Float64("max_size", maxSize.GetNumberValue()))
		return &pb.TestConfigError{
			InstanceGroup: ig.GetTitle(),
			Instance:      instance.GetTitle(),
			Error:         "Provided drive_size is bigger than sp maximum",
		}
	} else {
		if maxSize, ok := max.GetValue()["default"]; ok && maxSize.GetNumberValue() < size.GetNumberValue() {
			log.Warn("Out of bounds", zap.Float64("max_size", maxSize.GetNumberValue()))
			return &pb.TestConfigError{
				InstanceGroup: ig.GetTitle(),
				Instance:      instance.GetTitle(),
				Error:         "Provided drive_size is bigger than sp maximum",
			}
		}
	}

	return nil
}
