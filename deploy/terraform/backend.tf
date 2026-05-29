terraform {
  backend "s3" {
    bucket         = "roundup-platform-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "roundup-platform-terraform-locks"
  }
}
