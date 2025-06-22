package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DeviceType string

const (
	TV        DeviceType = "tv"
	Recording DeviceType = "recording"
	Tuner     DeviceType = "tuner"
	Playback  DeviceType = "playback"
	Audio     DeviceType = "audio"
)

type KeyboardActionType string

const (
	ActionPress KeyboardActionType = "press"
	ActionDown  KeyboardActionType = "down"
	ActionUp    KeyboardActionType = "up"
)

type clickAction struct{}

type Config struct {
	Adapter  string       `yaml:"adapter"`
	Name     string       `yaml:"name"`
	Type     DeviceType   `yaml:"type"`
	Mappings []MappingDef `yaml:"mappings"`
}

type MappingAction struct {
	Keyboard *KeyboardAction `yaml:"keyboard"`
	Mouse    *MouseAction    `yaml:"mouse"`
}

type MappingDef struct {
	CecCode int             `yaml:"cecCode"`
	Actions []MappingAction `yaml:"actions"`
}
type KeyboardAction struct {
	Type KeyboardActionType `yaml:"type"`
	Code uint16             `yaml:"code"`
}

type MouseAction struct {
	LeftClick   *clickAction `yaml:"leftClick"`
	RightClick  *clickAction `yaml:"rightClick"`
	MiddleClick *clickAction `yaml:"middleClick"`
	SideClick   *clickAction `yaml:"sideClick"`
	MoveX       *int32       `yaml:"moveX"`
	MoveY       *int32       `yaml:"moveY"`
}

func LoadConfig(filename string) (*cecConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validated(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	mapping := make(map[int][]MappingAction, len(config.Mappings))
	for _, mappingDef := range config.Mappings {
		mapping[mappingDef.CecCode] = mappingDef.Actions
	}

	return &cecConfig{
		Adapter: config.Adapter,
		Name:    config.Name,
		Type:    string(config.Type),
		Mapping: mapping,
	}, nil
}

func validated(config *Config) error {
	switch config.Type {
	case TV, Recording, Tuner, Playback, Audio:

	default:
		return fmt.Errorf("invalid CEC device type '%s', must be one of tv, recording, tuner, playback, audio", config.Type)
	}

	seenCodes := make(map[int]struct{}, len(config.Mappings))

	for index, mapping := range config.Mappings {
		if mapping.CecCode < 0 {
			return fmt.Errorf("invalid CEC code %d at index %d, must be a non-negative integer", mapping.CecCode, index)
		}

		if _, exists := seenCodes[mapping.CecCode]; exists {
			return fmt.Errorf("duplicate CEC code %d found at index %d", mapping.CecCode, index)
		}
		seenCodes[mapping.CecCode] = struct{}{}

		if len(mapping.Actions) == 0 {
			return fmt.Errorf("no actions defined for CEC key code %d at index %d", mapping.CecCode, index)
		}

		for jindex, action := range mapping.Actions {
			if action.Mouse != nil {
				if action.Keyboard != nil {
					return fmt.Errorf("both keyboard and mouse actions defined for CEC key code %d at index %d, action index %d", mapping.CecCode, index, jindex)
				}

				registeredActions := make([]string, 0, 6)

				for name, isDefined := range map[string]bool{
					"leftClick":   action.Mouse.LeftClick != nil,
					"rightClick":  action.Mouse.RightClick != nil,
					"middleClick": action.Mouse.MiddleClick != nil,
					"sideClick":   action.Mouse.SideClick != nil,
					"moveX":       action.Mouse.MoveX != nil,
					"moveY":       action.Mouse.MoveY != nil,
				} {
					if isDefined {
						registeredActions = append(registeredActions, name)
					}
				}

				if len(registeredActions) == 0 {
					return fmt.Errorf("no mouse actions defined for CEC key code %d at index %d, action index %d", mapping.CecCode, index, jindex)
				}
				if len(registeredActions) > 1 {
					return fmt.Errorf("multiple mouse actions defined for CEC key code %d at index %d, action index %d: %v", mapping.CecCode, index, jindex, registeredActions)
				}
			} else if action.Keyboard != nil {
				if action.Mouse != nil {
					return fmt.Errorf("both keyboard and mouse actions defined for CEC key code %d at index %d, action index %d", mapping.CecCode, index, jindex)
				}

				switch action.Keyboard.Type {
				case ActionPress, ActionDown, ActionUp:
				default:
					return fmt.Errorf("invalid keyboard action type '%s' for CEC key code %d at index %d, action index %d", action.Keyboard.Type, mapping.CecCode, index, jindex)
				}
			} else {
				return fmt.Errorf("no keyboard or mouse actions defined for CEC key code %d at index %d, action index %d", mapping.CecCode, index, jindex)
			}
		}
	}

	return nil
}
