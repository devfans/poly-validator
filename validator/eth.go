/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package validator

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	eccm "github.com/polynetwork/bridge-common/abi/eccm_abi"
	lockproxy "github.com/polynetwork/bridge-common/abi/lock_proxy_abi"
	"github.com/polynetwork/bridge-common/chains/eth"
)

type EthValidator struct {
	sdk   *eth.SDK
	conf  *ChainConfig
	proxy []*lockproxy.LockProxy
	ccm   *eccm.EthCrossChainManager
}

func (v *EthValidator) LatestHeight() (uint64, error) {
	return v.sdk.Node().GetLatestHeight()
}

func (v *EthValidator) Setup(cfg *ChainConfig) (err error) {
	v.conf = cfg
	v.sdk, err = eth.NewSDK(cfg.ChainId, cfg.Nodes, time.Minute, 1)
	if err != nil {
		return
	}

	for _, address := range v.conf.ProxyContracts {
		contract, err := lockproxy.NewLockProxy(common.HexToAddress(address), v.sdk.Node().Client)
		if err != nil {
			return err
		}
		v.proxy = append(v.proxy, contract)
	}
	v.ccm, err = eccm.NewEthCrossChainManager(common.HexToAddress(v.conf.CCMContract), v.sdk.Node().Client)
	return
}

func (v *EthValidator) Scan(height uint64) (txs []*DstTx, err error) {
	h := height
	opt := &bind.FilterOpts{
		Start:   h,
		End:     &h,
		Context: context.Background(),
	}
	ccmUnlocks, err := v.ccm.FilterVerifyHeaderAndExecuteTxEvent(opt)
	if err != nil {
		return nil, err
	}

	unlocks := map[string]DstTx{}
	txs = []*DstTx{}
	for ccmUnlocks.Next() {
		evt := ccmUnlocks.Event
		hash := evt.Raw.TxHash.String()[2:]
		unlocks[hash] = DstTx{
			SrcChainId: evt.FromChainID,
			SrcTx:      HexStringReverse(hex.EncodeToString(evt.FromChainTxHash)),
			PolyTx:     HexStringReverse(hex.EncodeToString(evt.CrossChainTxHash)),
			DstHeight:  evt.Raw.BlockNumber,
		}
	}

	for _, p := range v.proxy {
		unlockEvents, err := p.FilterUnlockEvent(opt)
		if err != nil {
			return nil, err
		}
		for unlockEvents.Next() {
			evt := unlockEvents.Event
			tx := &DstTx{
				Amount:     evt.Amount,
				DstTx:      evt.Raw.TxHash.String()[2:],
				DstAsset:   strings.ToLower(evt.ToAssetHash.String()[2:]),
				To:         strings.ToLower(evt.ToAddress.String()[2:]),
				DstChainId: v.conf.ChainId,
			}
			ccmTx, ok := unlocks[tx.DstTx]
			if ok {
				tx.SrcChainId = ccmTx.SrcChainId
				tx.SrcTx = ccmTx.SrcTx
				tx.PolyTx = ccmTx.PolyTx
				tx.DstHeight = ccmTx.DstHeight
			}
			txs = append(txs, tx)
		}
	}

	return
}

func (v *EthValidator) Validate(tx *DstTx) (err error) {
	data, err := v.sdk.Node().TransactionReceipt(context.Background(), common.HexToHash(tx.SrcTx))
	if err != nil {
		return err
	}
	height := uint64(data.BlockNumber.Int64())
	opt := &bind.FilterOpts{
		Start:   height,
		End:     &height,
		Context: context.Background(),
	}

	for _, p := range v.proxy {
		locks, err := p.FilterLockEvent(opt)
		if err != nil {
			return err
		}
		for locks.Next() {
			evt := locks.Event
			amount := evt.Amount
			address := string(evt.ToAddress)
			chainId := evt.ToChainId
			asset := string(evt.ToAssetHash)

			logs.Info("Comparing %v %v %v %v %v", *tx, amount, address, chainId, asset)
			if amount.Cmp(tx.Amount) == 0 && address == tx.To && chainId == tx.DstChainId && asset == tx.DstAsset {
				logs.Info("Successfully validated tx %s to %s asset %v amount %s", tx.SrcTx, address, asset, amount.String())
				return nil
			}
		}
	}
	err = fmt.Errorf("Failed to validate tx %s to %s asset %v amount %s", tx.SrcTx, tx.To, tx.DstAsset, tx.Amount.String())
	return
}
