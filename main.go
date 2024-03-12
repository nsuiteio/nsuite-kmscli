package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/doublejumptokyo/nsuite-kmscli/awseoa"
	"github.com/doublejumptokyo/nsuite-kmscli/kmsutil"
)

var commands = []struct {
	name        string
	description string
	examples    []string
}{
	{"list", "Show list of keys",
		[]string{"-tags=false"}},
	{"new", "Create key",
		[]string{}},
	{"show-address", "Show address of key",
		[]string{"[keyID]"}},
	{"add-tags", "add tag to exist key",
		[]string{"[keyID] [name:value] [name:value]..."}},
}

func buildUsageText() string {
	var buffer bytes.Buffer

	if _, err := fmt.Fprintf(&buffer, "Usage of nsuite-kmscli:\n"); err != nil {
		return ""
	}

	for _, command := range commands {
		if _, err := fmt.Fprintf(&buffer,
			"%12s  %s\n",
			command.name,
			command.description,
		); err != nil {
			return ""
		}
		for _, example := range command.examples {
			if _, err := fmt.Fprintf(&buffer,
				"              %s\n",
				example,
			); err != nil {
				return ""
			}
		}
		if _, err := fmt.Fprintln(&buffer); err != nil {
			return ""
		}
	}

	return buffer.String()
}

func usage() {
	fmt.Println(buildUsageText())
}

var (
	flagTags = true
)

func List(svc *kms.Client) (err error) {

	in := &kms.ListAliasesInput{}
	out, err := svc.ListAliases(context.TODO(), in)
	if err != nil {
		return
	}

	for _, a := range out.Aliases {
		alias := "None"
		if a.AliasName != nil {
			alias = *a.AliasName
		}
		alias = strings.TrimPrefix(alias, "alias/")
		if strings.HasPrefix(alias, "aws/") {
			continue
		}
		keyID := "None"
		if a.TargetKeyId != nil {
			keyID = *a.TargetKeyId
		}

		tags := ""
		if flagTags {
			in := &kms.ListResourceTagsInput{KeyId: a.TargetKeyId}
			out, err := svc.ListResourceTags(context.TODO(), in)
			if err != nil {
				return err
			}

			for _, t := range out.Tags {
				tags += *t.TagKey + ":" + *t.TagValue + "\t"
			}
		}

		fmt.Println(alias, keyID, tags)
	}
	return
}

func AddTag(svc *kms.Client, keyID, tagKey, tagValue string) (err error) {
	in := &kms.TagResourceInput{
		KeyId: aws.String(keyID),
		Tags: []kmstypes.Tag{
			{
				TagKey:   aws.String(tagKey),
				TagValue: aws.String(tagValue),
			},
		},
	}
	_, err = svc.TagResource(context.TODO(), in)
	return
}

func New(svc *kms.Client) (err error) {
	signer, err := awseoa.CreateSigner(svc, big.NewInt(4))
	if err != nil {
		return
	}
	fmt.Println(signer.Address().String(), signer.ID)
	return
}

func ShowAddress(svc *kms.Client, id string) (err error) {
	signer, err := awseoa.NewSigner(svc, id, big.NewInt(4))
	if err != nil {
		return
	}
	fmt.Println(signer.Address().String())
	return
}

func main() {
	var err error
	listFlag := flag.NewFlagSet("list", flag.ExitOnError)
	_ = flag.NewFlagSet("new", flag.ExitOnError)
	_ = flag.NewFlagSet("add-tags", flag.ExitOnError)
	_ = flag.NewFlagSet("show-address", flag.ExitOnError)

	listFlag.BoolVar(&flagTags, "tags", flagTags, "Show tags")

	if len(os.Args) == 1 {
		usage()
		return
	}

	svc, err := kmsutil.NewKMSClient()
	if err != nil {
		panic(err)
	}
	if err := listFlag.Parse(os.Args[2:]); err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "list":
		err = List(svc)
	case "new":
		err = New(svc)
	case "add-tags":
		keyID := os.Args[2]

		for i := 3; i < len(os.Args); i++ {
			parts := strings.Split(os.Args[i], ":")
			err = AddTag(svc, keyID, parts[0], parts[1])
		}
	case "show-address":
		keyID := os.Args[2]
		err = ShowAddress(svc, keyID)
	default:
		usage()
	}

	if err != nil {
		panic(err)
	}
}
