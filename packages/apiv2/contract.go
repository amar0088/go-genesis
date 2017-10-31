// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package apiv2

import (
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type contractResult struct {
	Hash string `json:"hash"`
}

func contract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var (
		hash, publicKey []byte
		toSerialize     interface{}
	)
	contract, parerr, err := validateSmartContract(r, data, nil)
	if err != nil {
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusBadRequest, parerr)
		}
		return errorAPI(w, err, http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)

	key := &model.Key{}
	key.SetTablePrefix(data.state)
	err = key.Get(data.wallet)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if len(key.PublicKey) == 0 {
		if _, ok := data.params[`pubkey`]; ok && len(data.params[`pubkey`].([]byte)) > 0 {
			publicKey = data.params[`pubkey`].([]byte)
			lenpub := len(publicKey)
			if lenpub > 64 {
				publicKey = publicKey[lenpub-64:]
			}
		}
		if len(publicKey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			return errorAPI(w, `E_EMPTYPUBLIC`, http.StatusBadRequest)
		}
	} else {
		logger.Warning("public key for wallet not found")
		publicKey = []byte("null")
	}
	signature := data.params[`signature`].([]byte)
	if len(signature) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signature is empty")
		return errorAPI(w, `E_EMPTYSIGN`, http.StatusBadRequest)
	}
	idata := make([]byte, 0)
	if info.Tx != nil {
	fields:
		for _, fitem := range *info.Tx {
			val := strings.TrimSpace(r.FormValue(fitem.Name))
			if strings.Contains(fitem.Tags, `address`) {
				val = converter.Int64ToStr(converter.StringToAddress(val))
			}
			switch fitem.Type.String() {
			case `[]interface {}`:
				var list []string
				for key, values := range r.Form {
					if key == fitem.Name+`[]` {
						for _, value := range values {
							list = append(list, value)
						}
					}
				}
				idata = append(idata, converter.EncodeLength(int64(len(list)))...)
				for _, ilist := range list {
					blist := []byte(ilist)
					idata = append(append(idata, converter.EncodeLength(int64(len(blist)))...), blist...)
				}
			case `uint64`:
				valInt, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": val}).Error("paring value to int")
					valInt = 0
				}
				converter.BinMarshal(&idata, uint64(valInt))
			case `int64`:
				valInt, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": val}).Error("paring value to int")
					valInt = 0
				}
				converter.EncodeLenInt64(&idata, valInt)
			case `float64`:
				valFloat, err := strconv.ParseFloat(val, 64)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": val}).Error("paring value to float")
					valFloat = 0
				}
				converter.BinMarshal(&idata, valFloat)
			case `string`, script.Decimal:
				idata = append(append(idata, converter.EncodeLength(int64(len(val)))...), []byte(val)...)
			case `[]uint8`:
				var bytes []byte
				bytes, err = hex.DecodeString(val)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": val}).Error("decoding value from hex")
					break fields
				}
				idata = append(append(idata, converter.EncodeLength(int64(len(bytes)))...), bytes...)
			}
		}
	}
	timeInt, err := strconv.ParseInt(data.params["time"].(string), 10, 64)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": data.params["time"].(string)}).Error("decoding value from hex")
		timeInt = 0
	}
	toSerialize = tx.SmartContract{
		Header: tx.Header{Type: int(info.ID), Time: timeInt,
			UserID: data.wallet, StateID: data.state, PublicKey: publicKey,
			BinSignatures: converter.EncodeLengthPlusData(signature)},
		TokenEcosystem: data.params[`token_ecosystem`].(int64),
		MaxSum:         data.params[`max_sum`].(string),
		PayOver:        data.params[`payover`].(string),
		Data:           idata,
	}
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if hash, err = model.SendTx(int64(info.ID), data.wallet,
		append([]byte{128}, serializedData...)); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("error sending tx")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	data.result = &contractResult{Hash: hex.EncodeToString(hash)} // !!! string(converter.BinToHex(hash))}
	return nil
}
