package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &securityGroupRuleResource{}
var _ resource.ResourceWithConfigure = &securityGroupRuleResource{}
var _ resource.ResourceWithImportState = &securityGroupRuleResource{}

type securityGroupRuleResource struct {
	client *apiClient
}

type securityGroupRuleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	Direction       types.String `tfsdk:"direction"`
	EtherType       types.String `tfsdk:"ether_type"`
	Protocol        types.String `tfsdk:"protocol"`
	PortRangeMin    types.Int64  `tfsdk:"port_range_min"`
	PortRangeMax    types.Int64  `tfsdk:"port_range_max"`
	RemoteIPPrefix  types.String `tfsdk:"remote_ip_prefix"`
	RemoteGroupID   types.String `tfsdk:"remote_group_id"`
	Region          types.String `tfsdk:"region"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func NewSecurityGroupRuleResource() resource.Resource {
	return &securityGroupRuleResource{}
}

func (r *securityGroupRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_rule"
}

func (r *securityGroupRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"security_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("ingress", "egress"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ether_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("IPv4", "IPv6"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port_range_min": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"port_range_max": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"remote_ip_prefix": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remote_group_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *securityGroupRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*apiClient)
}

func (r *securityGroupRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, plan.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	// Some security groups are created with default rules. If an identical rule
	// already exists, adopt it into state and skip create.
	if matched := r.findMatchingRule(ctx, plan.SecurityGroupID.ValueString(), region, plan); matched != nil {
		plan.ID = types.StringValue(matched.ID)
		plan.Region = types.StringValue(region)
		plan.CreatedAt = types.StringValue(matched.CreatedAt)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	body := map[string]any{
		"direction":  plan.Direction.ValueString(),
		"ether_type": plan.EtherType.ValueString(),
	}
	if v := stringOrEmpty(plan.Protocol); v != "" {
		body["protocol"] = v
	}
	if v := int64OrZero(plan.PortRangeMin); v != nil {
		body["port_range_min"] = *v
	}
	if v := int64OrZero(plan.PortRangeMax); v != nil {
		body["port_range_max"] = *v
	}
	if v := stringOrEmpty(plan.RemoteIPPrefix); v != "" {
		body["remote_ip_prefix"] = v
	}
	if v := stringOrEmpty(plan.RemoteGroupID); v != "" {
		body["remote_group_id"] = v
	}

	path := "/security-groups/" + url.PathEscape(plan.SecurityGroupID.ValueString()) + "/rules"
	var out createSecurityGroupRuleEnvelope
	if err := r.client.post(ctx, path, queryWithRegion(region), body, generateIdempotencyKey(), &out); err != nil {
		if isConflict(err) {
			// Some backends return 409 when an identical rule already exists.
			// Retry-read briefly to handle eventual consistency, then adopt.
			if matched := r.findMatchingRuleWithRetry(ctx, plan.SecurityGroupID.ValueString(), region, plan, 5, time.Second); matched != nil {
				plan.ID = types.StringValue(matched.ID)
				plan.Region = types.StringValue(region)
				plan.CreatedAt = types.StringValue(matched.CreatedAt)
				resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
				return
			}
		}
		resp.Diagnostics.AddError("Create security group rule failed", describeAPIError(err))
		return
	}

	plan.ID = types.StringValue(out.ID)
	plan.Region = types.StringValue(region)
	plan.CreatedAt = types.StringValue(out.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupRuleResource) findMatchingRuleWithRetry(ctx context.Context, securityGroupID, region string, plan securityGroupRuleResourceModel, attempts int, delay time.Duration) *securityRuleItem {
	if attempts <= 0 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		if matched := r.findMatchingRule(ctx, securityGroupID, region, plan); matched != nil {
			return matched
		}
		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(delay):
			}
		}
	}
	return nil
}

func (r *securityGroupRuleResource) findMatchingRule(ctx context.Context, securityGroupID, region string, plan securityGroupRuleResourceModel) *securityRuleItem {
	path := "/security-groups/" + url.PathEscape(securityGroupID)
	var out securityGroupDetailEnvelope
	if err := r.client.get(ctx, path, queryWithRegion(region), &out); err != nil {
		return nil
	}

	for i := range out.SecurityGroup.Rules {
		rule := &out.SecurityGroup.Rules[i]
		if sameRule(rule, plan) {
			return rule
		}
	}
	return nil
}

func sameRule(rule *securityRuleItem, plan securityGroupRuleResourceModel) bool {
	if !strings.EqualFold(rule.Direction, plan.Direction.ValueString()) {
		return false
	}
	if !strings.EqualFold(rule.EffectiveEtherType(), plan.EtherType.ValueString()) {
		return false
	}
	if !strings.EqualFold(rule.Protocol, stringOrEmpty(plan.Protocol)) {
		return false
	}
	if !sameInt64PtrValue(rule.PortRangeMin, int64OrZero(plan.PortRangeMin)) {
		return false
	}
	if !sameInt64PtrValue(rule.PortRangeMax, int64OrZero(plan.PortRangeMax)) {
		return false
	}
	if rule.RemoteIPPrefix != stringOrEmpty(plan.RemoteIPPrefix) {
		return false
	}
	if rule.RemoteGroupID != stringOrEmpty(plan.RemoteGroupID) {
		return false
	}
	return true
}

func sameInt64PtrValue(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (r *securityGroupRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	path := "/security-groups/" + url.PathEscape(state.SecurityGroupID.ValueString())
	var out securityGroupDetailEnvelope
	err = r.client.get(ctx, path, queryWithRegion(region), &out)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read security group rule failed", describeAPIError(err))
		return
	}

	var matched *securityRuleItem
	for i := range out.SecurityGroup.Rules {
		if out.SecurityGroup.Rules[i].ID == state.ID.ValueString() {
			matched = &out.SecurityGroup.Rules[i]
			break
		}
	}
	if matched == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Direction = types.StringValue(matched.Direction)
	// Use EffectiveEtherType() so both "ether_type" and "ethertype" JSON fields
	// are handled — the API returns one or the other depending on the backend.
	state.EtherType = types.StringValue(matched.EffectiveEtherType())
	// protocol, remote_ip_prefix, remote_group_id are Optional (not Computed).
	// Store null rather than "" so state stays in sync with an unset config field.
	if matched.Protocol == "" {
		state.Protocol = types.StringNull()
	} else {
		state.Protocol = types.StringValue(matched.Protocol)
	}
	state.PortRangeMin = ptrInt64ToTerraform(matched.PortRangeMin)
	state.PortRangeMax = ptrInt64ToTerraform(matched.PortRangeMax)
	if matched.RemoteIPPrefix == "" {
		state.RemoteIPPrefix = types.StringNull()
	} else {
		state.RemoteIPPrefix = types.StringValue(matched.RemoteIPPrefix)
	}
	if matched.RemoteGroupID == "" {
		state.RemoteGroupID = types.StringNull()
	} else {
		state.RemoteGroupID = types.StringValue(matched.RemoteGroupID)
	}
	state.CreatedAt = types.StringValue(matched.CreatedAt)
	state.Region = types.StringValue(region)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityGroupRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported update", "Security group rule updates are not supported by API. Recreate the resource.")
}

func (r *securityGroupRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := effectiveRegion(r.client, state.Region)
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())
		return
	}

	path := "/security-groups/" + url.PathEscape(state.SecurityGroupID.ValueString()) + "/rules/" + url.PathEscape(state.ID.ValueString())
	err = r.client.delete(ctx, path, queryWithRegion(region), generateIdempotencyKey(), nil)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Delete security group rule failed", describeAPIError(err))
	}
}

func (r *securityGroupRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sgID, ruleID := parseRuleImportID(req.ID)
	if sgID == "" || ruleID == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected <security_group_id>/<rule_id>, got %q", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("security_group_id"), sgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ruleID)...)
	if r.client.defaultRegion != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), r.client.defaultRegion)...)
	}
}
