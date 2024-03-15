# AWS Managed EOA
AWS Managed EOA is an Ethereum EOA(Externally Owned Account) using [Asymmetric Keys of AWS Key Management Service](https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html).

## Using commad line

```sh
$ export AWS_REGION=YOUR_REGION
$ export AWS_PROFILE=YOUR_PROFILE
$ nsuite-kmscli list
# list keys

$ nsuite-kmscli new
# create new key and set alias as address

$ nsuite-kmscli add-tags 01234567-abcd-1234-9876-02468ace79bf tag1:value1 tag2:value2
# tags added to key

$ nsuite-kmscli show-address 01234567-abcd-1234-9876-02468ace79bf
# show address of key (normally address is the same as alias)
```
