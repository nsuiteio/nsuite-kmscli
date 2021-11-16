# AWS Managed EOA
AWS Managed EOA is an Ethereum EOA(Externally Owned Account) using [Asymmetric Keys of AWS Key Management Service](https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html).

## Using commad line

```sh
$ export AWS_REGION=YOUR_REGION
$ export AWS_PROFILE=YOUR_PROFILE
$ awseoa list
# list keys

$ awseoa new
# create new key and set alias as address
```
