package services_registry

const SERVICES_EVENTS = "services_registry"

const (
	InstanceCreated = "instance_created"
)

func Topic(key string) string {
	return SERVICES_EVENTS + "." + key
}
