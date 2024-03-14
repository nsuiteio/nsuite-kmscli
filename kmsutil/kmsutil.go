package kmsutil

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/doublejumptokyo/nsuite-kmscli/awseoa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func NewKMSClient(ctx context.Context) (*kms.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return kms.NewFromConfig(cfg), nil
}

func TransactOptsFromAddress(ctx context.Context, svc *kms.Client, addr common.Address, chainID *big.Int) (*bind.TransactOpts, error) {
	keyID, err := KeyIDFromAddress(ctx, svc, addr)
	if err != nil {
		return nil, err
	}

	return awseoa.NewKMSTransactor(ctx, svc, keyID, chainID)
}

func KeyIDFromAddress(ctx context.Context, svc *kms.Client, addr common.Address) (string, error) {
	in := &kms.ListAliasesInput{}
	out, err := svc.ListAliases(ctx, in)
	if err != nil {
		return "", err
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

		ad := common.HexToAddress(alias)
		if ad.String() != addr.String() {
			continue
		}

		return *a.TargetKeyId, nil
	}

	return "", errors.New("Not found addr: " + addr.String())
}
