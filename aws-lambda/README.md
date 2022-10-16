# aws-lambda

Basic AWS lambda example, deployed using terraform.

Initial deploy:

```bash
make all
terraform init
terraform apply
```

Afterwads you can build, zip and upload to AWS directly with:

```bash
make deploy
```
