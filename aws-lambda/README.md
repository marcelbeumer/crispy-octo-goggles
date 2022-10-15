# aws-lambda

Basic AWS lambda example, deployed using terraform.

- Run `./scripts/build.sh` to build the binary.
- Run `./scripts/zip.sh` to create a zip from the build.
- Run `./scripts/upload.sh` to update the lambda by uploading the zip (requires terraform deploy).
- Run `./scripts/deploy.sh` does build, zip and then upload.

## Deploy using terraform

```bash
./scripts/zip.sh
terraform init
terraform apply
```
