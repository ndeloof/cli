package opts

import (
	"encoding/csv"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/compose-go/types"
	"github.com/docker/docker/api/types/swarm"
)

const (
	portOptTargetPort    = "target"
	portOptPublishedPort = "published"
	portOptProtocol      = "protocol"
	portOptMode          = "mode"
)

// PortOpt represents a port config in swarm mode.
type PortOpt struct {
	ports []swarm.PortConfig
}

// Set a new port value
// nolint: gocyclo
func (p *PortOpt) Set(value string) error {
	longSyntax, err := regexp.MatchString(`\w+=\w+(,\w+=\w+)*`, value)
	if err != nil {
		return err
	}
	if longSyntax {
		csvReader := csv.NewReader(strings.NewReader(value))
		fields, err := csvReader.Read()
		if err != nil {
			return err
		}

		pConfig := swarm.PortConfig{}
		for _, field := range fields {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid field %s", field)
			}

			key := strings.ToLower(parts[0])
			value := strings.ToLower(parts[1])

			switch key {
			case portOptProtocol:
				if value != string(swarm.PortConfigProtocolTCP) && value != string(swarm.PortConfigProtocolUDP) && value != string(swarm.PortConfigProtocolSCTP) {
					return fmt.Errorf("invalid protocol value %s", value)
				}

				pConfig.Protocol = swarm.PortConfigProtocol(value)
			case portOptMode:
				if value != string(swarm.PortConfigPublishModeIngress) && value != string(swarm.PortConfigPublishModeHost) {
					return fmt.Errorf("invalid publish mode value %s", value)
				}

				pConfig.PublishMode = swarm.PortConfigPublishMode(value)
			case portOptTargetPort:
				tPort, err := strconv.ParseUint(value, 10, 16)
				if err != nil {
					return err
				}

				pConfig.TargetPort = uint32(tPort)
			case portOptPublishedPort:
				pPort, err := strconv.ParseUint(value, 10, 16)
				if err != nil {
					return err
				}

				pConfig.PublishedPort = uint32(pPort)
			default:
				return fmt.Errorf("invalid field key %s", key)
			}
		}

		if pConfig.TargetPort == 0 {
			return fmt.Errorf("missing mandatory field %q", portOptTargetPort)
		}

		if pConfig.PublishMode == "" {
			pConfig.PublishMode = swarm.PortConfigPublishModeIngress
		}

		if pConfig.Protocol == "" {
			pConfig.Protocol = swarm.PortConfigProtocolTCP
		}

		p.ports = append(p.ports, pConfig)
	} else {
		// short syntax
		config, err := types.ParsePortConfig(value)
		if err != nil {
			return err
		}
		converted, err := convertPort(config)
		if err != nil {
			return err
		}
		p.ports = append(p.ports, converted...)
	}
	return nil
}

func convertPort(configs []types.ServicePortConfig) ([]swarm.PortConfig, error) {
	portConfigs := []swarm.PortConfig{}
	for _, config := range configs {
		if config.HostIP != "" && config.HostIP != "0.0.0.0" {
			return nil, fmt.Errorf("hostip is not supported")
		}
		portConfigs = append(portConfigs, swarm.PortConfig{
			// TODO Name: ?
			Protocol:      swarm.PortConfigProtocol(strings.ToLower(config.Protocol)),
			TargetPort:    config.Target,
			PublishedPort: config.Published,
			PublishMode:   swarm.PortConfigPublishMode(config.Mode),
		})
	}
	return portConfigs, nil
}

// Type returns the type of this option
func (p *PortOpt) Type() string {
	return "port"
}

// String returns a string repr of this option
func (p *PortOpt) String() string {
	ports := []string{}
	for _, port := range p.ports {
		repr := fmt.Sprintf("%v:%v/%s/%s", port.PublishedPort, port.TargetPort, port.Protocol, port.PublishMode)
		ports = append(ports, repr)
	}
	return strings.Join(ports, ", ")
}

// Value returns the ports
func (p *PortOpt) Value() []swarm.PortConfig {
	return p.ports
}
