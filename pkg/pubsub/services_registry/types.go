package services_registry

const SERVICES_EVENTS = "services_registry"

const (
	InstanceCreated       = "instance_created"
	CommandInstanceInvoke = "command_instance_invoke"
)

func Topic(key string) string {
	return SERVICES_EVENTS + "." + key
}
