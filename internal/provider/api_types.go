package provider

import (
	"encoding/json"
	"strconv"
)

type regionRef struct {
	Name string `json:"name"`
}

type instanceResponseEnvelope struct {
	Instance instancePayload `json:"instance"`
}

type instancePayload struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	FlavorName string           `json:"flavor"`
	VCPUs      float64          `json:"vcpus"`
	RAM        float64          `json:"ram"`
	Status     string           `json:"status"`
	CreatedAt  string           `json:"created_at"`
	Region     regionRef        `json:"region"`
	Networks   instanceNetworks `json:"networks"`
	Image      struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"image"`
}

type instanceNetworks struct {
	V4 []instanceNetwork `json:"v4"`
	V6 []instanceNetwork `json:"v6"`
}

type instanceNetwork struct {
	IPAddress string `json:"ip_address"`
	Type      string `json:"type"`
}

type createInstanceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	VCPUs     float64   `json:"vcpus"`
	RAM       float64   `json:"ram"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
	Region    regionRef `json:"region"`
}

type networkEnvelope struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Network networkPayload `json:"network"`
}

type networkListEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		Networks []networkPayload `json:"networks"`
		Count    int              `json:"count"`
	} `json:"data"`
}

type networkPayload struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Subnets      []string `json:"subnets"`
	AdminStateUp bool     `json:"admin_state_up"`
}

type securityGroupListEnvelope struct {
	SecurityGroups []securityGroupPayload `json:"security_groups"`
}

type securityGroupDetailEnvelope struct {
	SecurityGroup securityGroupPayload `json:"security_group"`
}

type createSecurityGroupEnvelope struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
	Region      regionRef `json:"region"`
}

type securityGroupPayload struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Rules       []securityRuleItem `json:"rules"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	Region      regionRef          `json:"region"`
}

type securityRuleItem struct {
	ID              string `json:"id"`
	Direction       string `json:"direction"`
	EtherType       string `json:"ether_type"`
	Ethertype       string `json:"ethertype"`
	Protocol        string `json:"protocol"`
	PortRangeMin    *int64 `json:"port_range_min"`
	PortRangeMax    *int64 `json:"port_range_max"`
	RemoteIPPrefix  string `json:"remote_ip_prefix"`
	RemoteGroupID   string `json:"remote_group_id"`
	SecurityGroupID string `json:"security_group_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (s securityRuleItem) EffectiveEtherType() string {
	if s.EtherType != "" {
		return s.EtherType
	}
	return s.Ethertype
}

func (s *securityRuleItem) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	_ = unmarshalString(raw["id"], &s.ID)
	_ = unmarshalString(raw["direction"], &s.Direction)
	_ = unmarshalString(raw["ether_type"], &s.EtherType)
	_ = unmarshalString(raw["ethertype"], &s.Ethertype)
	_ = unmarshalString(raw["protocol"], &s.Protocol)
	s.PortRangeMin = parseOptionalInt64(raw["port_range_min"])
	s.PortRangeMax = parseOptionalInt64(raw["port_range_max"])
	_ = unmarshalString(raw["remote_ip_prefix"], &s.RemoteIPPrefix)
	_ = unmarshalString(raw["remote_group_id"], &s.RemoteGroupID)
	_ = unmarshalString(raw["security_group_id"], &s.SecurityGroupID)
	_ = unmarshalString(raw["created_at"], &s.CreatedAt)
	_ = unmarshalString(raw["updated_at"], &s.UpdatedAt)

	return nil
}

func unmarshalString(data json.RawMessage, out *string) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	return json.Unmarshal(data, out)
}

func parseOptionalInt64(data json.RawMessage) *int64 {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		return &n
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			return nil
		}
		v, convErr := strconv.ParseInt(s, 10, 64)
		if convErr == nil {
			return &v
		}
	}

	return nil
}

type createSecurityGroupRuleEnvelope struct {
	ID              string `json:"id"`
	Direction       string `json:"direction"`
	EtherType       string `json:"ether_type"`
	Protocol        string `json:"protocol"`
	PortRangeMin    *int64 `json:"port_range_min"`
	PortRangeMax    *int64 `json:"port_range_max"`
	RemoteIPPrefix  string `json:"remote_ip_prefix"`
	RemoteGroupID   string `json:"remote_group_id"`
	SecurityGroupID string `json:"security_group_id"`
	CreatedAt       string `json:"created_at"`
}

type keyPairListEnvelope struct {
	KeyPairs []keyPairPayload `json:"keypairs"`
}

type keyPairDetailEnvelope struct {
	KeyPair keyPairPayload `json:"keypair"`
}

type createKeyPairEnvelope struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
	CreatedAt   string `json:"created_at"`
}

type keyPairPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type volumeRegionRef struct {
	Name string `json:"name"`
}

type volumeAttachmentPayload struct {
	ServerID string `json:"server_id"`
	Device   string `json:"device"`
}

type volumePayload struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Size        int64                     `json:"size"`
	Status      string                    `json:"status"`
	VolumeType  string                    `json:"volume_type"`
	Bootable    bool                      `json:"bootable"`
	Attachments []volumeAttachmentPayload `json:"attachments"`
	CreatedAt   string                    `json:"created_at"`
	UpdatedAt   string                    `json:"updated_at"`
	Region      volumeRegionRef           `json:"region"`
}

type volumeListEnvelope struct {
	Volumes []volumePayload `json:"volumes"`
}

type volumeDetailEnvelope struct {
	Volume volumePayload `json:"volume"`
}

type createVolumeEnvelope struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Size   int64           `json:"size"`
	Status string          `json:"status"`
	Region volumeRegionRef `json:"region"`
}

type flavorsEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Flavors []flavorPayload `json:"flavors"`
		Count   int             `json:"count"`
	} `json:"data"`
}

type flavorPayload struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	VCPUs        int64   `json:"vcpus"`
	RAM          int64   `json:"ram"`
	Disk         int64   `json:"disk"`
	PricePerHour float64 `json:"price_per_hour"`
}

type imagesEnvelope struct {
	ImageGroups []imageGroupPayload `json:"image_groups"`
}

type imageGroupPayload struct {
	Distro   string              `json:"distro"`
	Versions []imageVersionEntry `json:"versions"`
}

type imageVersionEntry struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type regionsEnvelope map[string]bool
