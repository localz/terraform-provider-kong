# TODO
Terraform currently doesn't support nested TypeMaps ([see here](https://github.com/hashicorp/hil/pull/42), and [here](https://github.com/hashicorp/terraform/pull/11704)).
This makes it difficult to add support for Kong plugins like [Request Transformer]() which have complex bodies, as it's not possibly to have Terraform config like:

```
config {
  add {
    headers = ["x-new-headers:some,value"]
  }
}
```

The above results in `expected type 'string', got unconvertible type '[]map[string]interface {}'`

It's possible to work around this by using the `TypeSet` type, and storing a list of key value pairs (`TypeList` of `TypeSet`) and converting them to maps, but I think this is overly complex and it would be better to wait for Hashicorp to add support for complex structures.

# Terraform provider for Kong

Uses [Terraform](http://www.terraform.io) to configure APIs in [Kong](http://www.getkong.org). It fully supports creating APIs and consumers, but plugins and credentials are not complete (most plugins will work though).

## Example usage

```Terraform
provider "kong" {
   address = "http://localhost:8001"
}

provider "kong" {
    address = "http://192.168.99.100:8001"
}

resource "kong_api" "api" {
    name               = "test"
    upstream_url       = "http://api.local"
    uris               = ["/api"]
    strip_uri          = true
}

resource "kong_consumer" "consumer" {
    username  = "user"
    custom_id = "123456"
}

resource "kong_api_plugin" "basic_auth" {
    api = "${kong_api.api.id}"
    name = "basic-auth"
}

resource "kong_api_plugin" "jwt" {
    api = "${kong_api.api.id}"
    name = "jwt"
}

resource "kong_api_plugin" "rate_limiting" {
    api  = "${kong_api.api.id}"
    name = "rate-limiting"

    config {
        minute = "100"
    }
}

resource "kong_consumer_basic_auth_credential" "basic_auth_credential" {
    consumer = "${kong_consumer.consumer.id}"
    username = "user123"
    password = "password"
}

resource "kong_consumer_jwt_credential" "jwt_credential" {
    consumer = "${kong_consumer.consumer.id}"
    secret   = "secret"
}
```
