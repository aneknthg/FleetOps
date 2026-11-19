# Terraform infrastructure

## Local kind environments

The reusable `kind-cluster` module creates one control plane and two workers by default. `local-mgmt`, `local-dev`, and `local-prod` use separate local state. `make up` applies each environment in order; `make down` destroys them in reverse order.

```sh
make up
kubectl --context kind-dev get nodes
make down
```

The pinned local node image is `kindest/node:v1.36.6`. A kind image change recreates the cluster and is not an in-place Kubernetes upgrade.

## Optional AWS EKS exercise

The EKS code is an implementation template and has not been applied. Before using it, create a $10 monthly AWS budget alert in the account, verify `aws sts get-caller-identity`, and review AWS pricing for the selected region. EKS control plane, EC2, NAT Gateway, public IPv4, data transfer, and related resources may incur charges. Do not apply until the budget alert is active and the complete plan has been reviewed.

The EKS module pins `terraform-aws-modules/eks/aws` 21.29.0 and `terraform-aws-modules/vpc/aws` 6.7.3. It uses two availability zones, private worker subnets, one NAT gateway, two `t3.small` on-demand workers, IRSA/OIDC, and a public API endpoint restricted to caller-supplied CIDRs.

### Bootstrap state and account guardrails

Choose a globally unique S3 bucket name and use an AWS identity with the required permissions. The bootstrap state is local and the bucket is versioned, encrypted, blocked from public access, and protected from accidental Terraform destroy.

```sh
cd terraform/bootstrap-state
terraform init
terraform plan -var='bucket_name=<unique-name>' -var='region=us-east-1'
terraform apply -var='bucket_name=<unique-name>' -var='region=us-east-1'
```

Create the $10 budget alert in AWS Budgets before continuing. Then initialize the EKS environment with the bucket details. Terraform's S3 backend lockfile is enabled; do not use the bucket for unrelated state.

```sh
cd ../envs/aws-eks
terraform init -reconfigure \
  -backend-config='bucket=<unique-name>' \
  -backend-config='key=fleetops/aws-eks/terraform.tfstate' \
  -backend-config='region=us-east-1' \
  -backend-config='use_lockfile=true'
cp terraform.tfvars.example terraform.tfvars
# Edit endpoint_public_access_cidrs to your current public IP/32.
terraform plan -out=fleetops.tfplan
terraform show fleetops.tfplan
# Apply only after reviewing the saved plan:
terraform apply fleetops.tfplan
aws eks update-kubeconfig --region us-east-1 --name fleetops-lab
```

To practice a minor upgrade, change `kubernetes_version` in a reviewed pull request, plan, apply, and inspect node versions. Destroy the EKS environment the same day and verify no related resources remain in AWS. Keep `terraform.tfvars`, plans, credentials, and state out of Git.

### Local validation

```sh
terraform fmt -check -recursive terraform
terraform -chdir=terraform/envs/local-dev init -backend=false
terraform -chdir=terraform/envs/local-dev validate
```

The GitHub Actions workflow validates the local environments, state bootstrap, and EKS module configuration without applying resources. `.github/workflows/terraform-plan.yml` provides a manual, plan-only EKS workflow after you configure an AWS OIDC role and state bucket; it never applies changes.
