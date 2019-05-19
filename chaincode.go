/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Chaincode is the definition of the chaincode structure.
type Chaincode struct {
}

// Init is called when the chaincode is instantiated by the blockchain network.
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Init()", fcn, params)
	return shim.Success(nil)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Invoke()", fcn, params)

	switch fcn {
	case "enroll":
		return Enroll(stub, params)
	case "list":
		return List(stub)
	case "create":
		return Create(stub, params)
	case "sell":
		return Sell(stub, params)
	case "buy":
		return Buy(stub, params)
	default:
		return shim.Error(fmt.Sprintf("Unexpected method: %s", fcn))
	}
}

func List(stub shim.ChaincodeStubInterface) sc.Response {
	msg, err := list(stub)
	if err != nil {
		fmt.Printf("Failed to list trader and assets: %s\n", err)
		return shim.Error(err.Error())
	}

	return shim.Success(msg)
}

// List lists the info of current user and all assets
func list(stub shim.ChaincodeStubInterface) ([]byte, error) {
	id, err := getID(stub)
	if err != nil {
		return nil, err
	}

	trader, err := GetTrader(stub, id)
	if err != nil {
		return nil, err
	} else if trader == nil {
		return nil, fmt.Errorf("trader not exist")
	}

	iter, err := stub.GetStateByPartialCompositeKey(ASSET, []string{})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	res := struct {
		Trader *Trader          `json:"trader"`
		Assets map[string]Asset `json:"assets"`
	}{
		Trader: trader,
		Assets: make(map[string]Asset),
	}

	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if err = json.Unmarshal(kv.Value, &asset); err != nil {
			return nil, err
		}

		_, keys, err := stub.SplitCompositeKey(kv.Key)
		if err != nil {
			return nil, err
		}

		res.Assets[keys[0]] = asset
	}

	data, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return data, nil
}
