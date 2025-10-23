package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LimaDiskResource{}
var _ resource.ResourceWithImportState = &LimaDiskResource{}

func NewLimaDiskResource() resource.Resource {
	return &LimaDiskResource{}
}

// LimaDiskResource defines the resource implementation.
type LimaDiskResource struct{}

// LimaDiskResourceModel describes the resource data model.
type LimaDiskResourceModel struct {
	Name   types.String  `tfsdk:"name"`
	Size   types.Float64 `tfsdk:"size"`
	Format types.String  `tfsdk:"format"`
	Id     types.String  `tfsdk:"id"`
}

func (r *LimaDiskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_disk"
}

func (r *LimaDiskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lima disk resource. Creates and manages a disk using limactl disk create.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the disk.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.Float64Attribute{
				MarkdownDescription: "Size of the disk in GiB. Can be increased (but not decreased) after creation.",
				Required:            true,
			},
			"format": schema.StringAttribute{
				MarkdownDescription: "Disk format. Defaults to 'qcow2'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("qcow2"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Disk identifier (same as name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *LimaDiskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// No client needed for limactl - it's a local command-line tool
	if req.ProviderData == nil {
		return
	}
}

func (r *LimaDiskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LimaDiskResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build limactl disk create command
	args := []string{"disk", "create", data.Name.ValueString()}

	// Add required size flag (convert GiB to string with 'G' suffix)
	args = append(args, fmt.Sprintf("--size=%gG", data.Size.ValueFloat64()))

	// Add format flag
	if !data.Format.IsNull() {
		args = append(args, "--format="+data.Format.ValueString())
	}

	// Add --tty=false to disable interactive mode
	args = append(args, "--tty=false")

	tflog.Debug(ctx, "Creating Lima disk", map[string]any{
		"command": "limactl " + strings.Join(args, " "),
	})

	// Execute limactl disk create command
	cmd := exec.CommandContext(ctx, "limactl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Lima disk",
			fmt.Sprintf("Command: limactl %s\nError: %s\nOutput: %s", strings.Join(args, " "), err, string(output)),
		)
		return
	}

	tflog.Trace(ctx, "Created Lima disk", map[string]any{
		"name": data.Name.ValueString(),
	})

	// Set the ID to the disk name
	data.Id = data.Name

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LimaDiskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LimaDiskResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if disk exists using limactl disk list --json
	cmd := exec.CommandContext(ctx, "limactl", "disk", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list Lima disks",
			fmt.Sprintf("Error: %s\nOutput: %s", err, string(output)),
		)
		return
	}

	// Parse JSON output - limactl disk list --json returns a single JSON object per line
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	found := false

	for _, line := range lines {
		if line == "" {
			continue
		}

		var disk struct {
			Name string `json:"name"`
		}

		if err := json.Unmarshal([]byte(line), &disk); err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse disk list JSON",
				fmt.Sprintf("Error: %s\nLine: %s", err, line),
			)
			return
		}

		if disk.Name == data.Name.ValueString() {
			found = true
			break
		}
	}

	if !found {
		// Disk no longer exists, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LimaDiskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LimaDiskResourceModel
	var state LimaDiskResourceModel

	// Read Terraform plan and state data into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if size has changed
	if !plan.Size.Equal(state.Size) {
		// Validate that size is increasing, not decreasing
		if plan.Size.ValueFloat64() < state.Size.ValueFloat64() {
			resp.Diagnostics.AddError(
				"Cannot decrease disk size",
				fmt.Sprintf("Disk size can only be increased. Current size: %gG, requested size: %gG",
					state.Size.ValueFloat64(), plan.Size.ValueFloat64()),
			)
			return
		}

		tflog.Debug(ctx, "Resizing Lima disk", map[string]any{
			"name":     plan.Name.ValueString(),
			"old_size": state.Size.ValueFloat64(),
			"new_size": plan.Size.ValueFloat64(),
		})

		// Build limactl disk resize command
		args := []string{"disk", "resize", plan.Name.ValueString()}
		args = append(args, fmt.Sprintf("--size=%gG", plan.Size.ValueFloat64()))
		args = append(args, "--tty=false")

		// Execute limactl disk resize command
		cmd := exec.CommandContext(ctx, "limactl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resize Lima disk",
				fmt.Sprintf("Command: limactl %s\nError: %s\nOutput: %s", strings.Join(args, " "), err, string(output)),
			)
			return
		}

		tflog.Trace(ctx, "Resized Lima disk", map[string]any{
			"name": plan.Name.ValueString(),
			"size": plan.Size.ValueFloat64(),
		})
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LimaDiskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LimaDiskResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Lima disk", map[string]any{
		"name": data.Name.ValueString(),
	})

	// Delete the Lima disk
	deleteCmd := exec.CommandContext(ctx, "limactl", "disk", "delete", data.Name.ValueString())
	deleteOutput, deleteErr := deleteCmd.CombinedOutput()
	if deleteErr != nil {
		resp.Diagnostics.AddError(
			"Failed to delete Lima disk",
			fmt.Sprintf("Command: limactl disk delete %s\nError: %s\nOutput: %s", data.Name.ValueString(), deleteErr, string(deleteOutput)),
		)
		return
	}

	tflog.Trace(ctx, "Deleted Lima disk", map[string]any{
		"name": data.Name.ValueString(),
	})
}

func (r *LimaDiskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the disk name
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)

	// Also set the ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
