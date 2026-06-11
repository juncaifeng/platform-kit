package registry

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/your-org/platform-kit/internal/config"
)

type ConsulRegistry struct {
	client *api.Client
	config *config.Config
}

func NewConsulRegistry(cfg *config.Config) (*ConsulRegistry, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.RegistryAddr

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulRegistry{
		client: client,
		config: cfg,
	}, nil
}

func (r *ConsulRegistry) Register() error {
	serviceID := fmt.Sprintf("%s-%s", r.config.ServiceName, r.config.ServiceAddr)

	meta := map[string]string{
		"type":            r.config.ServiceType,
		"events_produced": r.config.EventsProduced,
		"events_consumed": r.config.EventsConsumed,
		"gateway_exposed": r.config.GatewayExposed,
		"gateway_prefix":  r.config.GatewayPrefix,
	}

	check := &api.AgentServiceCheck{
		HTTP:     fmt.Sprintf("http://%s:%d%s", r.config.ServiceAddr, r.config.ServicePort, r.config.HealthCheckPath),
		Interval: r.config.HealthCheckInterval,
		Timeout:  "5s",
	}

	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    r.config.ServiceName,
		Address: r.config.ServiceAddr,
		Port:    r.config.ServicePort,
		Meta:    meta,
		Check:   check,
	}

	if err := r.client.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	slog.Info("service registered",
		"service", r.config.ServiceName,
		"id", serviceID,
		"address", r.config.ServiceAddr,
		"port", r.config.ServicePort,
	)

	return nil
}

func (r *ConsulRegistry) Deregister() error {
	serviceID := fmt.Sprintf("%s-%s", r.config.ServiceName, r.config.ServiceAddr)

	if err := r.client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	slog.Info("service deregistered",
		"service", r.config.ServiceName,
		"id", serviceID,
	)

	return nil
}

func (r *ConsulRegistry) Heartbeat() error {
	serviceID := fmt.Sprintf("%s-%s", r.config.ServiceName, r.config.ServiceAddr)
	checkID := "service:" + serviceID

	if err := r.client.Agent().UpdateTTL(checkID, "service is healthy", api.HealthPassing); err != nil {
		return fmt.Errorf("failed to update TTL: %w", err)
	}

	return nil
}

func (r *ConsulRegistry) StartHeartbeat(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.Heartbeat(); err != nil {
				slog.Error("heartbeat failed", "error", err)
			}
		case <-stopCh:
			return
		}
	}
}
