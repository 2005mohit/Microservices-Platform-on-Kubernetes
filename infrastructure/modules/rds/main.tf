# Security group: allow Postgres only from inside the VPC
resource "aws_security_group" "rds" {
  name_prefix = "${var.project}-rds-"
  description = "Allow Postgres access from within the VPC"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-rds-sg"
  }
}

resource "aws_db_subnet_group" "this" {
  name       = "${var.project}-rds-subnets"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.project}-rds-subnet-group"
  }
}

resource "aws_db_instance" "this" {
  identifier            = "${var.project}-${var.environment}-db"
  engine                = "postgres"
  engine_version        = "16.14"
  instance_class        = var.db_instance_class
  allocated_storage     = 20
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  backup_retention_period = 1
  skip_final_snapshot     = true

  tags = {
    Name = "${var.project}-${var.environment}-db"
  }
}
