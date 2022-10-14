locals {
  function_name    = "go-playground-aws-lambda"
  api_gateway_name = "go-playground-aws-lambda"
}

terraform {
  required_version = ">= 1.2.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.34.0"
    }
  }
}

module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "4.1.0"

  function_name = local.function_name
  description   = "Basic lambda from go-playground."
  handler       = "main"
  runtime       = "go1.x"

  create_package         = false
  local_existing_package = "dist/main.zip"

  attach_policies    = true
  number_of_policies = 2
  policies           = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    "arn:aws:iam::aws:policy/CloudWatchLambdaInsightsExecutionRolePolicy"
  ]
}

module "lambda_alias" {
  source  = "terraform-aws-modules/lambda/aws//modules/alias"
  version = "4.1.0"


  name             = "prod"
  function_name    = module.lambda_function.lambda_function_name
  function_version = module.lambda_function.lambda_function_version
  refresh_alias    = false

  allowed_triggers = {
    APIGatewayInternal = {
      principal = "apigateway.amazonaws.com"
      # Setting the source arn somehow does not work, though when adding it manually does work.
      # Terraform plan does not show differences compared to manually adding. Perhaps order/dep issue?
      # source_arn = "${module.api_gateway.apigatewayv2_api_execution_arn}/*/*/${local.function_name}"
    }
  }
}

module "api_gateway" {
  source  = "terraform-aws-modules/apigateway-v2/aws"
  version = "2.2.0"

  create_api_domain_name = false

  name        = local.api_gateway_name
  description = "Basic lambda from go-playground"

  integrations = {
    "$default" = {
      lambda_arn = module.lambda_alias.lambda_alias_arn
    }
  }
}

output "lambda_endpoint" {
  description = "Endpoint of the lambda."
  value       = "${module.api_gateway.default_apigatewayv2_stage_invoke_url}${local.function_name}"
}
