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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &LimaInstanceResource{}
var _ resource.ResourceWithImportState = &LimaInstanceResource{}

func NewLimaInstanceResource() resource.Resource {
	return &LimaInstanceResource{}
}

type LimaInstanceResource struct{}

type LimaInstanceResourceModel struct {
	Name          types.String  `tfsdk:"name"`
	Template      types.String  `tfsdk:"template"`
	Arch          types.String  `tfsdk:"arch"`
	Containerd    types.String  `tfsdk:"containerd"`
	Cpus          types.Int64   `tfsdk:"cpus"`
	Disk          types.Float64 `tfsdk:"disk"`
	Memory        types.Float64 `tfsdk:"memory"`
	DNS           types.List    `tfsdk:"dns"`
	Mount         types.List    `tfsdk:"mount"`
	MountInotify  types.Bool    `tfsdk:"mount_inotify"`
	MountNone     types.Bool    `tfsdk:"mount_none"`
	MountType     types.String  `tfsdk:"mount_type"`
	MountWritable types.Bool    `tfsdk:"mount_writable"`
	Network       types.List    `tfsdk:"network"`
	Plain         types.Bool    `tfsdk:"plain"`
	Rosetta       types.Bool    `tfsdk:"rosetta"`
	Video         types.Bool    `tfsdk:"video"`
	VmType        types.String  `tfsdk:"vm_type"`
	Disks         types.List    `tfsdk:"disks"`
	Id            types.String  `tfsdk:"id"`
}

type DisksModel struct {
	Name       types.String `tfsdk:"name"`
	MountPoint types.String `tfsdk:"mount_point"`
}

func (r *LimaInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *LimaInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lima instance resource. Creates and manages a lightweight VM using limactl.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Lima instance. If not specified, defaults to 'default'.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Template to use for the instance. Can be a template name (e.g., 'docker'), local file path, or URL. If not specified, uses the default Ubuntu template.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"arch": schema.StringAttribute{
				MarkdownDescription: "Machine architecture (x86_64, aarch64, riscv64, armv7l, s390x, ppc64le).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"containerd": schema.StringAttribute{
				MarkdownDescription: "Containerd mode (user, system, user+system, none).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cpus": schema.Int64Attribute{
				MarkdownDescription: "Number of CPUs to allocate to the instance.",
				Optional:            true,
			},
			"disk": schema.Float64Attribute{
				MarkdownDescription: "Disk size in GiB.",
				Optional:            true,
			},
			"memory": schema.Float64Attribute{
				MarkdownDescription: "Memory in GiB.",
				Optional:            true,
			},
			"dns": schema.ListAttribute{
				MarkdownDescription: "Custom DNS servers (disables host resolver).",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"mount": schema.ListAttribute{
				MarkdownDescription: "Directories to mount. Suffix ':w' for writable. Do not specify directories that overlap with existing mounts.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"mount_inotify": schema.BoolAttribute{
				MarkdownDescription: "Enable inotify for mounts.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"mount_none": schema.BoolAttribute{
				MarkdownDescription: "Remove all mounts.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"mount_type": schema.StringAttribute{
				MarkdownDescription: "Mount type (reverse-sshfs, 9p, virtiofs).",
				Optional:            true,
			},
			"mount_writable": schema.BoolAttribute{
				MarkdownDescription: "Make all mounts writable.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"network": schema.ListAttribute{
				MarkdownDescription: "Additional networks, e.g., 'vzNAT' or 'lima:shared' to assign vmnet IP.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"plain": schema.BoolAttribute{
				MarkdownDescription: "Plain mode. Disables mounts, port forwarding, containerd, etc.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"rosetta": schema.BoolAttribute{
				MarkdownDescription: "Enable Rosetta (for vz instances on macOS).",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"video": schema.BoolAttribute{
				MarkdownDescription: "Enable video output (has negative performance impact for QEMU).",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"vm_type": schema.StringAttribute{
				MarkdownDescription: "Virtual machine type (qemu, vz).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Instance identifier (same as name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"disks": schema.ListNestedBlock{
				MarkdownDescription: "Additional disks to attach to the instance. Each disk must have a name and mount point.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the additional disk to attach.",
							Required:            true,
						},
						"mount_point": schema.StringAttribute{
							MarkdownDescription: "Mount point for the additional disk (e.g., '/mnt/data').",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *LimaInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// No client needed for limactl - it's a local command-line tool
	if req.ProviderData == nil {
		return
	}
}

func (r *LimaInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LimaInstanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := []string{"create"}

	if !data.Name.IsNull() {
		args = append(args, "--name="+data.Name.ValueString())
	}

	if !data.Arch.IsNull() {
		args = append(args, "--arch="+data.Arch.ValueString())
	}

	if !data.Containerd.IsNull() {
		args = append(args, "--containerd="+data.Containerd.ValueString())
	}

	if !data.Cpus.IsNull() {
		args = append(args, fmt.Sprintf("--cpus=%d", data.Cpus.ValueInt64()))
	}

	if !data.Disk.IsNull() {
		args = append(args, fmt.Sprintf("--disk=%g", data.Disk.ValueFloat64()))
	}

	if !data.Memory.IsNull() {
		args = append(args, fmt.Sprintf("--memory=%g", data.Memory.ValueFloat64()))
	}

	if !data.DNS.IsNull() && len(data.DNS.Elements()) > 0 {
		var dnsServers []string
		resp.Diagnostics.Append(data.DNS.ElementsAs(ctx, &dnsServers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, dns := range dnsServers {
			args = append(args, "--dns="+dns)
		}
	}

	if !data.Mount.IsNull() && len(data.Mount.Elements()) > 0 {
		var mounts []string
		resp.Diagnostics.Append(data.Mount.ElementsAs(ctx, &mounts, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, mount := range mounts {
			args = append(args, "--mount="+mount)
		}
	}

	if data.MountInotify.ValueBool() {
		args = append(args, "--mount-inotify")
	}

	if data.MountNone.ValueBool() {
		args = append(args, "--mount-none")
	}

	if !data.MountType.IsNull() {
		args = append(args, "--mount-type="+data.MountType.ValueString())
	}

	if data.MountWritable.ValueBool() {
		args = append(args, "--mount-writable")
	}

	if !data.Network.IsNull() && len(data.Network.Elements()) > 0 {
		var networks []string
		resp.Diagnostics.Append(data.Network.ElementsAs(ctx, &networks, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, network := range networks {
			args = append(args, "--network="+network)
		}
	}

	if data.Plain.ValueBool() {
		args = append(args, "--plain")
	}

	if data.Rosetta.ValueBool() {
		args = append(args, "--rosetta")
	}

	if data.Video.ValueBool() {
		args = append(args, "--video")
	}

	if !data.VmType.IsNull() {
		args = append(args, "--vm-type="+data.VmType.ValueString())
	}

	if !data.Disks.IsNull() && len(data.Disks.Elements()) > 0 {
		var disks []DisksModel
		resp.Diagnostics.Append(data.Disks.ElementsAs(ctx, &disks, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		type diskJSON struct {
			Name       string `json:"name"`
			MountPoint string `json:"mountPoint"`
		}

		var diskArray []diskJSON
		for _, disk := range disks {
			diskArray = append(diskArray, diskJSON{
				Name:       disk.Name.ValueString(),
				MountPoint: disk.MountPoint.ValueString(),
			})
		}

		diskJSONBytes, err := json.Marshal(diskArray)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to marshal additional disks",
				fmt.Sprintf("Error: %s", err),
			)
			return
		}

		args = append(args, fmt.Sprintf("--set=.additionalDisks=%s", string(diskJSONBytes)))
	}

	// Add --tty=false for non-interactive use (automation)
	args = append(args, "--tty=false")

	if !data.Template.IsNull() {
		template := data.Template.ValueString()
		if !strings.HasPrefix(template, "http") && !strings.HasSuffix(template, ".yaml") {
			template = "template://" + template
		}
		args = append(args, template)
	}

	tflog.Debug(ctx, "Creating Lima instance", map[string]any{
		"command": "limactl " + strings.Join(args, " "),
	})

	cmd := exec.CommandContext(ctx, "limactl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Lima instance",
			fmt.Sprintf("Command: limactl %s\nError: %s\nOutput: %s", strings.Join(args, " "), err, string(output)),
		)
		return
	}

	tflog.Trace(ctx, "Created Lima instance", map[string]any{
		"name": data.Name.ValueString(),
	})

	tflog.Debug(ctx, "Starting Lima instance", map[string]any{
		"name": data.Name.ValueString(),
	})

	startCmd := exec.CommandContext(ctx, "limactl", "start", data.Name.ValueString())
	startOutput, startErr := startCmd.CombinedOutput()
	if startErr != nil {
		// Clean up the created instance if start fails
		tflog.Warn(ctx, "Start failed, cleaning up created instance", map[string]any{
			"name": data.Name.ValueString(),
		})
		deleteCmd := exec.CommandContext(ctx, "limactl", "delete", data.Name.ValueString())
		deleteOutput, deleteErr := deleteCmd.CombinedOutput()
		if deleteErr != nil {
			tflog.Error(ctx, "Failed to clean up instance after start failure", map[string]any{
				"name":   data.Name.ValueString(),
				"error":  deleteErr.Error(),
				"output": string(deleteOutput),
			})
		}

		resp.Diagnostics.AddError(
			"Failed to start Lima instance",
			fmt.Sprintf("Command: limactl start %s\nError: %s\nOutput: %s", data.Name.ValueString(), startErr, string(startOutput)),
		)
		return
	}

	tflog.Trace(ctx, "Started Lima instance", map[string]any{
		"name": data.Name.ValueString(),
	})

	data.Id = data.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LimaInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LimaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if instance exists using limactl list --json
	cmd := exec.CommandContext(ctx, "limactl", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list Lima instances",
			fmt.Sprintf("Error: %s\nOutput: %s", err, string(output)),
		)
		return
	}

	// Parse JSON output - limactl list --json returns a single JSON object per line
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	found := false

	for _, line := range lines {
		if line == "" {
			continue
		}

		var instance struct {
			Name string `json:"name"`
		}

		if err := json.Unmarshal([]byte(line), &instance); err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse instance list JSON",
				fmt.Sprintf("Error: %s\nLine: %s", err, line),
			)
			return
		}

		if instance.Name == data.Name.ValueString() {
			found = true
			break
		}
	}

	if !found {
		// Instance no longer exists, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LimaInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LimaInstanceResourceModel
	var state LimaInstanceResourceModel

	// Read Terraform plan and state data into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build limactl edit command
	args := []string{"edit", plan.Name.ValueString()}

	// Add flags for changed attributes that are supported by limactl edit
	if !plan.Cpus.IsNull() && !plan.Cpus.Equal(state.Cpus) {
		args = append(args, fmt.Sprintf("--cpus=%d", plan.Cpus.ValueInt64()))
	}

	if !plan.Disk.IsNull() && !plan.Disk.Equal(state.Disk) {
		args = append(args, fmt.Sprintf("--disk=%g", plan.Disk.ValueFloat64()))
	}

	if !plan.Memory.IsNull() && !plan.Memory.Equal(state.Memory) {
		args = append(args, fmt.Sprintf("--memory=%g", plan.Memory.ValueFloat64()))
	}

	// DNS flags
	if !plan.DNS.Equal(state.DNS) {
		if !plan.DNS.IsNull() && len(plan.DNS.Elements()) > 0 {
			var dnsServers []string
			resp.Diagnostics.Append(plan.DNS.ElementsAs(ctx, &dnsServers, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			for _, dns := range dnsServers {
				args = append(args, "--dns="+dns)
			}
		}
	}

	// Mount flags
	if !plan.Mount.Equal(state.Mount) {
		if !plan.Mount.IsNull() && len(plan.Mount.Elements()) > 0 {
			var mounts []string
			resp.Diagnostics.Append(plan.Mount.ElementsAs(ctx, &mounts, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			for _, mount := range mounts {
				args = append(args, "--mount="+mount)
			}
		}
	}

	if !plan.MountInotify.Equal(state.MountInotify) && plan.MountInotify.ValueBool() {
		args = append(args, "--mount-inotify")
	}

	if !plan.MountType.IsNull() && !plan.MountType.Equal(state.MountType) {
		args = append(args, "--mount-type="+plan.MountType.ValueString())
	}

	if !plan.MountWritable.Equal(state.MountWritable) && plan.MountWritable.ValueBool() {
		args = append(args, "--mount-writable")
	}

	// Network flags
	if !plan.Network.Equal(state.Network) {
		if !plan.Network.IsNull() && len(plan.Network.Elements()) > 0 {
			var networks []string
			resp.Diagnostics.Append(plan.Network.ElementsAs(ctx, &networks, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			for _, network := range networks {
				args = append(args, "--network="+network)
			}
		}
	}

	if !plan.Rosetta.Equal(state.Rosetta) && plan.Rosetta.ValueBool() {
		args = append(args, "--rosetta")
	}

	if !plan.Video.Equal(state.Video) && plan.Video.ValueBool() {
		args = append(args, "--video")
	}

	// Only proceed with edit if there are actual changes
	if len(args) > 2 { // More than just "edit" and instance name
		tflog.Debug(ctx, "Editing Lima instance", map[string]any{
			"command": "limactl " + strings.Join(args, " "),
		})

		// First stop the instance
		tflog.Debug(ctx, "Stopping Lima instance for edit", map[string]any{
			"name": plan.Name.ValueString(),
		})

		stopCmd := exec.CommandContext(ctx, "limactl", "stop", plan.Name.ValueString())
		stopOutput, stopErr := stopCmd.CombinedOutput()
		if stopErr != nil {
			resp.Diagnostics.AddError(
				"Failed to stop Lima instance for edit",
				fmt.Sprintf("Command: limactl stop %s\nError: %s\nOutput: %s", plan.Name.ValueString(), stopErr, string(stopOutput)),
			)
			return
		}

		// Execute limactl edit command
		cmd := exec.CommandContext(ctx, "limactl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to edit Lima instance",
				fmt.Sprintf("Command: limactl %s\nError: %s\nOutput: %s", strings.Join(args, " "), err, string(output)),
			)
			return
		}

		tflog.Trace(ctx, "Edited Lima instance", map[string]any{
			"name": plan.Name.ValueString(),
		})

		// Start the instance again
		tflog.Debug(ctx, "Starting Lima instance after edit", map[string]any{
			"name": plan.Name.ValueString(),
		})

		startCmd := exec.CommandContext(ctx, "limactl", "start", plan.Name.ValueString())
		startOutput, startErr := startCmd.CombinedOutput()
		if startErr != nil {
			resp.Diagnostics.AddError(
				"Failed to start Lima instance after edit",
				fmt.Sprintf("Command: limactl start %s\nError: %s\nOutput: %s", plan.Name.ValueString(), startErr, string(startOutput)),
			)
			return
		}

		tflog.Trace(ctx, "Started Lima instance after edit", map[string]any{
			"name": plan.Name.ValueString(),
		})
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LimaInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LimaInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Lima instance", map[string]any{
		"name": data.Name.ValueString(),
	})

	// Stop and delete the Lima instance
	// First stop it
	stopCmd := exec.CommandContext(ctx, "limactl", "stop", data.Name.ValueString())
	stopOutput, stopErr := stopCmd.CombinedOutput()
	if stopErr != nil {
		tflog.Warn(ctx, "Failed to stop Lima instance (may already be stopped)", map[string]any{
			"name":   data.Name.ValueString(),
			"error":  stopErr.Error(),
			"output": string(stopOutput),
		})
	}

	// Then delete it
	deleteCmd := exec.CommandContext(ctx, "limactl", "delete", data.Name.ValueString())
	deleteOutput, deleteErr := deleteCmd.CombinedOutput()
	if deleteErr != nil {
		resp.Diagnostics.AddError(
			"Failed to delete Lima instance",
			fmt.Sprintf("Command: limactl delete %s\nError: %s\nOutput: %s", data.Name.ValueString(), deleteErr, string(deleteOutput)),
		)
		return
	}

	tflog.Trace(ctx, "Deleted Lima instance", map[string]any{
		"name": data.Name.ValueString(),
	})
}

func (r *LimaInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the instance name
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
	// Also set the ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
