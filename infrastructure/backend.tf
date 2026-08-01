terraform {
  backend "s3" {
    bucket         = "devplatform-tfstate-102156246344"
    key            = "devplatform/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }
}
