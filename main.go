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

func buildCommands() []*flag.FlagSet {
	var results []*flag.FlagSet

	for _, command := range commands {
		flagSet := flag.NewFlagSet(command.name, flag.ExitOnError)
		flagSet.Usage = func() {
			if _, err := fmt.Fprintf(flagSet.Output(),
				"%s  %s\n",
				command.name,
				command.description,
			); err != nil {
				panic(err)
			}
			indent := strings.Repeat(" ", len(command.name))
			for _, example := range command.examples {
				if _, err := fmt.Fprintf(flagSet.Output(),
					"%s  %s\n",
					indent,
					example,
				); err != nil {
					panic(err)
				}
			}
		}
		results = append(results, flagSet)
	}
	return results
}

func usage(name string, commandList []*flag.FlagSet) {
	var buffer bytes.Buffer

	if _, err := fmt.Fprintf(&buffer, "Usage of %s:\n", name); err != nil {
		panic(err)
	}

	var maxLen int
	for _, command := range commandList {
		nameLen := len(command.Name())
		if maxLen < nameLen {
			maxLen = nameLen
		}
	}

	for _, command := range commandList {
		var tmp bytes.Buffer
		command.SetOutput(&tmp)
		command.Usage()
		command.SetOutput(os.Stdout)
		indent := strings.Repeat(" ", maxLen-len(command.Name()))
		for _, line := range strings.Split(tmp.String(), "\n") {
			if _, err := fmt.Fprintf(&buffer,
				"%s%s\n",
				indent,
				line,
			); err != nil {
				panic(err)
			}
		}
	}

	fmt.Println(buffer.String())
}

func List(ctx context.Context, svc *kms.Client, flagTags bool) (err error) {

	var aliases []kmstypes.AliasListEntry
	in := &kms.ListAliasesInput{}
	for {
		out, err := svc.ListAliases(ctx, in)
		if err != nil {
			return err
		}
		aliases = append(aliases, out.Aliases...)
		if out.NextMarker == nil {
			break
		}
		in.Marker = out.NextMarker
	}

	for _, a := range aliases {
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
			out, err := svc.ListResourceTags(ctx, in)
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

func AddTag(ctx context.Context, svc *kms.Client, keyID, tagKey, tagValue string) (err error) {
	in := &kms.TagResourceInput{
		KeyId: aws.String(keyID),
		Tags: []kmstypes.Tag{
			{
				TagKey:   aws.String(tagKey),
				TagValue: aws.String(tagValue),
			},
		},
	}
	_, err = svc.TagResource(ctx, in)
	return
}

func New(ctx context.Context, svc *kms.Client) (err error) {
	signer, err := awseoa.CreateSigner(ctx, svc, big.NewInt(4))
	if err != nil {
		return
	}
	fmt.Println(signer.Address(ctx).String(), signer.ID)
	return
}

func ShowAddress(ctx context.Context, svc *kms.Client, id string) (err error) {
	signer, err := awseoa.NewSigner(ctx, svc, id, big.NewInt(4))
	if err != nil {
		return
	}
	fmt.Println(signer.Address(ctx).String())
	return
}

func main() {
	var err error
	ctx := context.Background()
	commandList := buildCommands()

	myName := ""
	if len(os.Args) > 0 {
		myName = os.Args[0]
	}
	if len(os.Args) < 2 {
		usage(myName, commandList)
		return
	}
	commandName := os.Args[1]
	var command *flag.FlagSet
	for _, cmd := range commandList {
		if cmd.Name() == commandName {
			command = cmd
			break
		}
	}
	if command == nil {
		usage(myName, commandList)
		return
	}

	svc, err := kmsutil.NewKMSClient(ctx)
	if err != nil {
		panic(err)
	}

	switch command.Name() {
	case "list":
		flagTags := true
		command.BoolVar(&flagTags, "tags", flagTags, "Show tags")
		if err := command.Parse(os.Args[2:]); err != nil {
			command.Usage()
			panic(err)
		}
		err = List(ctx, svc, flagTags)
	case "new":
		err = New(ctx, svc)
	case "add-tags":
		keyID := os.Args[2]

		for i := 3; i < len(os.Args); i++ {
			parts := strings.Split(os.Args[i], ":")
			err = AddTag(ctx, svc, keyID, parts[0], parts[1])
		}
	case "show-address":
		keyID := os.Args[2]
		err = ShowAddress(ctx, svc, keyID)
	default:
		usage(myName, commandList)
	}

	if err != nil {
		panic(err)
	}
}
