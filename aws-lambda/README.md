# aws-lambda

Basic AWS lambda example, deployed using terraform.

- Run `./scripts/build.sh` to build the binary.
- Run `./scripts/zip.sh` to build and create a zip.

## Deploy using terraform

```bash
./scripts/zip.sh
terraform init
terraform apply
```
