# VM with additional volume example

This example provisions:

- keypair
- security group + SSH rule
- VM instance
- standalone data volume
- volume attachment to the VM

## Required variables

- `huddle_api_key`
- `flavor_id`
- `image_id`
- `ssh_public_key`

Optional:

- `region` (default: `eu2`)
- `volume_size` (default: `50`)
- `volume_type` (default: empty string)

## Volume deletion behaviour on destroy

By default, `huddle_cloud_volume` sets `delete_on_destroy = false`.

This means **`terraform destroy` removes the volume from Terraform state but does NOT delete the volume from the cloud**. Your data is preserved. You must delete the volume manually (e.g. via the Huddle dashboard or API) if you no longer need it.

Set `delete_on_destroy = true` to have Terraform permanently delete the volume:

```hcl
resource "huddle_cloud_volume" "data" {
  name              = "tf-vm-data-volume"
  size              = 50
  region            = var.region
  delete_on_destroy = true
}
```

> **Warning:** `delete_on_destroy = true` permanently destroys the volume and all data on it when `terraform destroy` (or a resource replacement) is run. This cannot be undone.

## Apply

```bash
terraform init
terraform apply
```
