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
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

func ParseInt(value, ty string) (v *big.Int) {
	switch ty {
	case "Integer":
		v, _ = new(big.Int).SetString(value, 10)
	default:
		v, _ = new(big.Int).SetString(HexStringReverse(value), 16)
	}
	return
}

func HexReverse(arr []byte) []byte {
	l := len(arr)
	x := make([]byte, 0)
	for i := l - 1; i >= 0; i-- {
		x = append(x, arr[i])
	}
	return x
}

func HexStringReverse(value string) string {
	aa, _ := hex.DecodeString(value)
	bb := HexReverse(aa)
	return hex.EncodeToString(bb)
}

type InvalidUnlockEvent struct {
	*DstTx
	Error error
}

func (o *InvalidUnlockEvent) Format() (title string, keys []string, values []interface{}, buttons []map[string]string) {
	keys = []string{"Amount", "Asset", "To", "DstChain", "PolyHash", "DstHash", "Error"}
	values = []interface{}{o.Amount.String(), o.DstAsset, o.To, o.DstChainId, o.PolyTx, o.DstTx.DstTx, o.Error}
	title = fmt.Sprintf("Suspicious unlock on chain %d", o.DstChainId)
	return
}

type SetManagerProxyEvent struct {
	TxHash   string
	Contract string
	ChainId  uint64
	Manager  string
	Operator string
}

func (o *SetManagerProxyEvent) Format() (title string, keys []string, values []interface{}, buttons []map[string]string) {
	title = fmt.Sprintf("Suspicious set manager proxy event on chain %v", o.ChainId)
	keys = []string{"Hash", "Contract", "ChainId", "New Manager"}
	values = []interface{}{o.TxHash, o.Contract, o.ChainId, o.Manager}
	return
}

type BindProxyEvent struct {
	Contract  string
	TxHash    string
	ChainId   uint64
	ToChainId uint64
	ToProxy   string
	Operator  string
}

func (o *BindProxyEvent) Format() (title string, keys []string, values []interface{}, buttons []map[string]string) {
	title = fmt.Sprintf("Suspicious bind proxy event on chain %v", o.ChainId)
	keys = []string{"Hash", "Contract", "ChainId", "ToChainId", "ToProxy"}
	values = []interface{}{o.TxHash, o.Contract, o.ChainId, o.ToChainId, o.ToProxy}
	return
}

type TxEvent struct {
	TxHash  string
	ChainId string
	From    string
	To      string
	Path    string
	Value   string
	Message string
}

func (o *TxEvent) Format() (title string, keys []string, values []interface{}, buttons []map[string]string) {
	title = fmt.Sprintf("Tracking address event on chain %v", o.ChainId)
	keys = []string{"Hash", "From", "ChainId", "To", "Value", "Message"}
	values = []interface{}{o.TxHash, o.From, o.ChainId, o.To, o.Value, o.Message}
	return
}

type ChainHeightStuckEvent struct {
	Chain         string
	Duration      time.Duration
	CurrentHeight uint64
	Nodes         []string
}

func (o *ChainHeightStuckEvent) Format() (title string, keys []string, values []interface{}, buttons []map[string]string) {
	title = fmt.Sprintf("Chain node height stopped for %s", o.Chain)
	keys = []string{"CurrentHeight", "StuckFor", "Nodes"}
	values = []interface{}{o.CurrentHeight, o.Duration, o.Nodes}
	return
}

type BindAssetEvent struct {
	TxHash        string
	Contract      string
	ChainId       uint64
	FromAsset     string
	ToChainId     uint64
	Asset         string
	InitialAmount *big.Int
	Operator      string
}

func (o *BindAssetEvent) Format() (title string, keys []string, values []interface{}, buttons []map[string]string) {
	title = fmt.Sprintf("Suspicious bind asset event on chain %v", o.ChainId)
	keys = []string{"Hash", "Contract", "ChainId", "FromAsset", "ToChainId", "ToAsset", "InitialAmount"}
	values = []interface{}{o.TxHash, o.Contract, o.ChainId, o.FromAsset, o.ToChainId, o.Asset, o.InitialAmount}
	return
}
