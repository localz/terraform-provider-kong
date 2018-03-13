#!/usr/bin/env bash
TF_LOG='DEBUG'
echo ${1}

rm -f terraform/${1}/terraform-provider-kong

go build -ldflags -w -o terraform/${1}/terraform-provider-kong
chmod +x terraform/${1}/terraform-provider-kong

cd terraform/${1}

terraform init
terraform ${2}

rm -f terraform/${1}/terraform-provider-kong

