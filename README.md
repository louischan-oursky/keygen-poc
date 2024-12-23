- [Keygen.sh PoC](#keygensh-poc)
  * [Prerequisite](#prerequisite)
  * [Getting started](#getting-started)
  * [Import the Postman collection](#import-the-postman-collection)
  * [Obtain an Admin token](#obtain-an-admin-token)
  * [Create a product, and its policy](#create-a-product-and-its-policy)
  * [Create a license](#create-a-license)
  * [Play around with the license with Postman](#play-around-with-the-license-with-postman)
    + [Try validating the license](#try-validating-the-license)
    + [Activate a license](#activate-a-license)
    + [Deactivate a license](#deactivate-a-license)
    + [Re-activate a license on another machine](#re-activate-a-license-on-another-machine)
  * [What value should we use for fingerprint?](#what-value-should-we-use-for-fingerprint)
  * [Play around with the license in a Go program](#play-around-with-the-license-in-a-go-program)
  * [Play around with the license in a containerized environment](#play-around-with-the-license-in-a-containerized-environment)

# Keygen.sh PoC

## Prerequisite

- Docker
- [mkcert](https://github.com/FiloSottile/mkcert)
  - You can also use `openssl` to generate self-signed TLS certificate, if you are familiar with it.
  - mkcert is used here because it automates the task of generate a CA, and then generate TLS certificate signed by that CA, and put the CA certificate in a platform-specific location such that programs, like Chrome, Firefox, Safari, curl, can respect.
- [Postman](https://www.postman.com/)

## Getting started

Run the following

```sh
make setup
```

It runs `make .env` and `make tls` under the hood.

You only need to run `make setup` once.

And then you run

```sh
make start
```

It runs keygen server.

## Import the Postman collection

Keygen.sh does not come with a web interface, nor it comes with a OpenAPI doc page.
So I have prepared a Postman collection to allow us to invoke the API we are going to use.

The Postman collection is at [./keygen.postman_collection.json](./keygen.postman_collection.json).

## Obtain an Admin token

A admin username and a admin password is set up in .env. But their usage is only for creating tokens.

To call the API, you **MUST** first generate an Admin token.
In the Postman collection, call `Create an admin token`.
You should get a response like this.

```json
{
    "data": {
        "id": "b4e18a3c-ef0e-4c16-b13a-e52e7210f02e",
        "type": "tokens",
        "attributes": {
            "kind": "admin-token",
            "token": "admin-7904475e2c629284a812cd197b85ba5600138fe1ca3eb0ac519bafb6cf358d63v3",
            "expiry": null,
            "created": "2024-12-16T06:53:44.919Z",
            "updated": "2024-12-16T06:53:44.930Z"
        }
    }
}
```

You extract `data.attributes.token` (the one that starts with `admin-`), and set it to the Postman collection variable `admin_token`. **THIS STEP IS IMPORTANT.**

## Create a product, and its policy

The next step is to create a product.
A product is a term in Keygen.sh to represent something you sell to customer.
A policy defines the licensing model of product.
A product can have more than 1 policy.

- In the Postman collection, call `Create a product`.
- In the response, take note of the product ID. We need it in the next step.
- In the Postman collection, open `Create a policy`.
- Replace the product ID in the request body.
- Invoke the updated request of `Create a policy`.

## Create a license

The next step is to create a license.

- In the Postman collection, open `Create a license`.
- Replace the policy ID in the request body.
- Invoke the updated request of `Create a license`.
- In the response, take note of the license key (`data.attributes.key`). We need it later.

## Play around with the license with Postman

In this section, we are going to explore the Keygen.sh with Postman.

### Try validating the license

- In the Postman collection, open `Validate a license`.
- Replace the license key in the request body.
- Invoke the updated request of `Validate a license`.
- You should get a response like

```json
{
    "data": {
        "id": "426f192a-3111-4454-a35f-5af85ab3bd71",
        "type": "licenses",
        "attributes": {
            "key": "A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3",
        }
    },
    "meta": {
        "ts": "2024-12-16T07:07:24.323Z",
        "valid": false,
        "detail": "fingerprint scope is required",
        "code": "FINGERPRINT_SCOPE_REQUIRED"
    }
}
```

You see `meta.valid` is `false` and `meta.detail` says `fingerprint scope is required`.
This is because the license is not activated yet.

### Activate a license

In Keygen.sh, activating a license requires two steps:

1. Create a machine with a fingerprint.
2. Validate the license with the machine information.

Do the following:

- In the Postman collection, open `Create a machine`.
- Update the header `Authorization` with the license key.
- Replace the license ID in the request body.
- Invoke the updated request of `Create a machine`.

Try invoke the request again, you will get the following response

```json
{
    "errors": [
        {
            "title": "Unprocessable resource",
            "detail": "has already been taken",
            "code": "FINGERPRINT_TAKEN",
            "source": {
                "pointer": "/data/attributes/fingerprint"
            },
            "links": {
                "about": "https://keygen.sh/docs/api/machines/#machines-object-attrs-fingerprint"
            }
        },
        {
            "title": "Unprocessable resource",
            "detail": "machine count has exceeded maximum allowed for license (1)",
            "code": "MACHINE_LIMIT_EXCEEDED",
            "source": {
                "pointer": "/data"
            },
            "links": {
                "about": "https://keygen.sh/docs/api/machines/#machines-object"
            }
        }
    ],
    "meta": {
        "id": "e4d8bd3e-a776-4526-813f-421258c6d09c"
    }
}
```

It says `FINGERPRINT_TAKEN`. It is because we repeated the request with the same fingerprint.

Try invoke the request again. This time, we change the fingerprint. You will get the following response

```json
{
    "errors": [
        {
            "title": "Unprocessable resource",
            "detail": "machine count has exceeded maximum allowed for license (1)",
            "code": "MACHINE_LIMIT_EXCEEDED",
            "source": {
                "pointer": "/data"
            },
            "links": {
                "about": "https://keygen.sh/docs/api/machines/#machines-object"
            }
        }
    ],
    "meta": {
        "id": "26ed50be-aa3e-4b87-9496-13b346d90fae"
    }
}
```

It says `MACHINE_LIMIT_EXCEEDED`. It is because the policy is configured to allow at most 1 machine.

Now that we have created a machine, it is time to activate the license.

- In the Postman collection, open `Activate a license`.
- Update the fingerprint of the request body to the same you used in creating the machine.
- Invoke the updated request of `Activate a license`. You should see a response like

```json
{
    "data": {
        "id": "426f192a-3111-4454-a35f-5af85ab3bd71",
        "type": "licenses",
        "attributes": {
            "key": "A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3",
        }
    },
    "meta": {
        "ts": "2024-12-16T07:22:14.585Z",
        "valid": true,
        "detail": "is valid",
        "code": "VALID",
        "scope": {
            "fingerprint": "some-fingerprint"
        }
    }
}
```

This time, we finally see `meta.valid` is `true`.

Try invoke the request again, you should see the same result. This request is idempotent.

Try invoke the request again, with fingerprint changed to something else. You should see the following response

```json
{
    "data": {
        "id": "426f192a-3111-4454-a35f-5af85ab3bd71",
        "type": "licenses",
        "attributes": {
            "key": "A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3",
        }
    },
    "meta": {
        "ts": "2024-12-16T07:24:02.826Z",
        "valid": false,
        "detail": "fingerprint is not activated (does not match any associated machines)",
        "code": "FINGERPRINT_SCOPE_MISMATCH",
        "scope": {
            "fingerprint": "some-fingerprint123"
        }
    }
}
```

It is because the fingerprint does not associate with any machine.

### Deactivate a license

In Keygen.sh, deactivating a license reverses the things we have done in activating a license.

1. Delete the machine with a previously used fingerprint.

Do the following:

- In the Postman collection, open `List machines of a license`
- Update the header `Authorization` with the license key.
- Replace the license ID in the URL.
- Invoke the updated request of `List machines of a license`.
- Take note of the machine ID.
- In the Postman collection, open `Delete a machine`.
- Update the header `Authorization` with the license key.
- Replace the machine ID in the URL.
- Invoke the updated request of `Delete a machine`.

Note that the license is still expiring. It means the validity period is still counting.

### Re-activate a license on another machine

It is simple. Just repeat [Activate a license](#activate-a-license).

## What value should we use for fingerprint?

The [official documentation](https://keygen.sh/docs/activating-machines/#how-to-identify-a-device) gives some advice on what value to be used as fingerprint. In short, use [github.com/denisbrodbeck/machineid](https://github.com/denisbrodbeck/machineid) in Golang. However, checking the issue list of that package reveals that in a containerized environment, machine-id is not available. So the best we can do in a containerized environment is to generate a cryptographically secure random machine ID, and store it a a writable directory.

The above method works fine in case of non-containerized environment. But if the machine is a VM with shared hardware, then two VMs may share the same machine ID. In this case, follow the [official documentation](https://keygen.sh/docs/choosing-a-licensing-model/floating-licenses/#licensing-virtual-machines), and implement [https://keygen.sh/docs/api/machines/#machines-actions-ping](https://keygen.sh/docs/api/machines/#machines-actions-ping).

## Play around with the license in a Go program

This section showcases the use of Keygen Go SDK.

This Go program is a `cat` program, except that it requires a license key to run :)

First, you need to build it with the following command.

```sh
make cat KEYGEN_ACCOUNT_ID=account KEYGEN_PRODUCT_ID=product
```

- `KEYGEN_ACCOUNT_ID`: You can find it in `.env`.
- `KEYGEN_PRODUCT_ID`: You need to use the product ID you created previously.

Then you can invoke our `cat` with the following command.

```sh
./cat ./hello.txt
panic: an error occurred: id=4907adad-7b42-43cb-9352-c443ec240f3f status=401 size=237 body={"meta":{"id":"4907adad-7b42-43cb-9352-c443ec240f3f"},"errors":[{"title":"Unauthorized","detail":"You must be authenticated to complete the request","code":"TOKEN_MISSING","links":{"about":"https://keygen.sh/docs/api/authentication/"}}]}

goroutine 1 [running]:
main.main()
        /Users/louischan/keygen-poc/main.go:124 +0x3dc
```

It is because you have not supplied `--license`. In practice, the license key is written in a persistent location. It is not required to specify it in command line.

Let's try again.

```sh
./cat --license A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3 ./hello.txt
hello
```

It works.

Let's simulate the case that the license was used on another machine.
This can be achieved with the flag `--fingerprint`.

```sh
./cat --license A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3 --fingerprint somefingerprint ./hello.txt
machine limit exceeded. If you want to deactivate the previous machine, and activate this machine instead, add --force-activate
```

It correctly reported that the license has been used already. In practice, the end-user just want to activate this machine anyway. To facilitate this, add the flag `--force-activate`.

```sh
./cat --license A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3 --fingerprint somefingerprint --force-activate ./hello.txt
hello
```

It has deactivated the previous machine under the hood. To verify this, you run

```
./cat --license A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3 ./hello.txt
machine limit exceeded. If you want to deactivate the previous machine, and activate this machine instead, add --force-activate
```

## Play around with the license in a containerized environment

This section showcases Keygen also works in a containerized environment.

First, you need to build the image with the following command.

```sh
make cat-image KEYGEN_ACCOUNT_ID=account KEYGEN_PRODUCT_ID=product
```

- `KEYGEN_ACCOUNT_ID`: You can find it in `.env`.
- `KEYGEN_PRODUCT_ID`: You need to use the product ID you created previously.

Then you can run the image with the following command:

```sh
docker run --rm --network host -v "$HOME/Library/Application Support/mkcert/rootCA.pem:/etc/ssl/certs/mkcert.pem" cat:latest keygen-cat --license A0F2D8-359DCA-FFC1F1-FAFD8E-15CCDC-V3 /usr/src/app/hello.txt
```

- `--rm`: Remove the container after exit.
- `--network host`: Use host network because the endpoint is hard-coded as `localhost` in the program.
- `-v ...:/etc/ssl/certs/mkcert.pem`: Mount the CA certificate of mkcert into the container. It is because in our setup, the TLS certificate is signed by that CA.

The rest is just the same as running a Go binary in non-containerized environment, so I do not repeat here.
