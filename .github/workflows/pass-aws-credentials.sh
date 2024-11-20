#!/bin/bash

# Variables
EC2_USER=$1
EC2_HOST=$2
EC2_SSH_KEY=$3
AWS_ACCESS_KEY_ID=$4
AWS_SECRET_ACCESS_KEY=$5
AWS_SESSION_TOKEN=$6

# SSH into EC2 instance and configure AWS credentials
ssh -i "${EC2_SSH_KEY}" -o StrictHostKeyChecking=no "${EC2_USER}@${EC2_HOST}" << EOF
echo "Setting up AWS credentials on EC2 instance..."

# Write AWS credentials to a script file
cat << CREDENTIALS > ~/aws_env.sh
export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
export AWS_SESSION_TOKEN=${AWS_SESSION_TOKEN}
CREDENTIALS

# Load the credentials
chmod +x ~/aws_env.sh
source ~/aws_env.sh

# Verify AWS credentials
aws sts get-caller-identity || { echo "AWS credentials validation failed"; exit 1; }

echo "AWS credentials successfully configured."
EOF
