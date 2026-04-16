package provider

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func queryWithRegion(region string) url.Values {
	q := url.Values{}
	q.Set("region", region)
	return q
}

func stringSliceToTerraform(values []string) []types.String {
	out := make([]types.String, 0, len(values))
	for _, v := range values {
		out = append(out, types.StringValue(v))
	}
	return out
}

func ptrInt64ToTerraform(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

func listStringToSlice(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	out := make([]string, 0, len(list.Elements()))
	for _, el := range list.Elements() {
		v, ok := el.(types.String)
		if !ok {
			continue
		}
		if v.IsNull() || v.IsUnknown() {
			continue
		}
		out = append(out, v.ValueString())
	}
	return out
}

func boolOrDefault(v types.Bool, fallback bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return fallback
	}
	return v.ValueBool()
}

func int64OrZero(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	val := v.ValueInt64()
	return &val
}

func stringOrEmpty(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func parseRuleImportID(v string) (string, string) {
	parts := strings.SplitN(v, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func waitForInstanceStatus(ctx context.Context, client *apiClient, instanceID string, region string, timeout time.Duration, terminalStatus string) error {
	return waitForCondition(ctx, timeout, 5*time.Second, func(c context.Context) (bool, error) {
		var payload instanceResponseEnvelope
		err := client.get(c, "/instances/"+instanceID, queryWithRegion(region), &payload)
		if err != nil {
			return false, err
		}
		return strings.EqualFold(payload.Instance.Status, terminalStatus), nil
	})
}

func waitForVolumeDeleted(ctx context.Context, client *apiClient, volumeID string, region string, timeout time.Duration) error {
	return waitForCondition(ctx, timeout, 2*time.Second, func(c context.Context) (bool, error) {
		var payload volumeDetailEnvelope
		err := client.get(c, "/volumes/"+volumeID, queryWithRegion(region), &payload)
		if err != nil {
			if isNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}

func waitForVolumeStatus(ctx context.Context, client *apiClient, volumeID string, region string, timeout time.Duration, terminalStatuses ...string) error {
	target := make([]string, 0, len(terminalStatuses))
	for _, s := range terminalStatuses {
		if strings.TrimSpace(s) == "" {
			continue
		}
		target = append(target, strings.ToLower(strings.TrimSpace(s)))
	}
	if len(target) == 0 {
		return nil
	}

	return waitForCondition(ctx, timeout, 2*time.Second, func(c context.Context) (bool, error) {
		var payload volumeDetailEnvelope
		err := client.get(c, "/volumes/"+volumeID, queryWithRegion(region), &payload)
		if err != nil {
			return false, err
		}
		current := strings.ToLower(strings.TrimSpace(payload.Volume.Status))
		for _, desired := range target {
			if current == desired {
				return true, nil
			}
		}
		return false, nil
	})
}
