package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

const ASSET = "ASSET"

type Asset struct {
	Name    string `json:"name"`
	Price   int    `json:"price"`
	Owner   string `json:"owner"`
	ForSale bool   `json:"forsale"`
}

// ID calculates asset ID
func (a *Asset) ID() string {
	id := sha3.Sum256([]byte(a.Name))
	return hex.EncodeToString(id[0:8])
}

func (a *Asset) Store(stub shim.ChaincodeStubInterface) error {
	key, err := stub.CreateCompositeKey(ASSET, []string{a.ID()})
	if err != nil {
		return err
	}

	data, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return stub.PutState(key, data)
}

func Create(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if err := create(stub, params); err != nil {
		fmt.Printf("Failed to create asset: %s\n", err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func create(stub shim.ChaincodeStubInterface, params []string) error {
	if len(params) != 1 {
		return fmt.Errorf("expect 1 parameter, got %d", len(params))
	}

	name := params[0]
	id, err := getID(stub)
	if err != nil {
		return err
	}

	trader, err := GetTrader(stub, id)
	if err != nil {
		return err
	} else if trader == nil {
		return fmt.Errorf("trader not exist")
	}

	asset := &Asset{
		Name:    name,
		Price:   0,
		Owner:   id,
		ForSale: false,
	}

	err = asset.Store(stub)
	if err != nil {
		return err
	}

	return nil
}

func Sell(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if err := sell(stub, params); err != nil {
		fmt.Printf("Failed to put asset on sale: %s\n", err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func sell(stub shim.ChaincodeStubInterface, params []string) error {
	if len(params) != 2 {
		return fmt.Errorf("expect 2 params, got %d", len(params))
	}

	price, err := strconv.Atoi(params[1])
	if err != nil {
		return err
	}

	id, err := getID(stub)
	if err != nil {
		return err
	}

	trader, err := GetTrader(stub, id)
	if err != nil {
		return err
	}

	if trader == nil {
		return fmt.Errorf("trader not exist")
	}

	key, err := stub.CreateCompositeKey(ASSET, params[0:1])
	if err != nil {
		return err
	}

	data, err := stub.GetState(key)
	if err != nil {
		return err
	} else if len(data) == 0 {
		return fmt.Errorf("asset does not exist")
	}

	var asset Asset
	if err = json.Unmarshal(data, &asset); err != nil {
		return err
	}

	if asset.Owner != id {
		return fmt.Errorf("asset does not belong to you")
	}

	asset.Price = price
	asset.ForSale = true
	err = asset.Store(stub)
	if err != nil {
		return err
	}

	return nil
}

func Buy(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if err := buy(stub, params); err != nil {
		fmt.Printf("Failed to purchase asset: %s\n", err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func buy(stub shim.ChaincodeStubInterface, params []string) error {
	if len(params) != 1 {
		return fmt.Errorf("expect 1 param, got %d", len(params))
	}

	id, err := getID(stub)
	if err != nil {
		return err
	}

	buyer, err := GetTrader(stub, id)
	if err != nil {
		return err
	}

	if buyer == nil {
		return fmt.Errorf("buyer not exist")
	}

	asset, err := GetAsset(stub, params[0])
	if err != nil {
		return err
	}

	if asset == nil {
		return fmt.Errorf("asset not exist")
	}

	if asset.Owner == id {
		return fmt.Errorf("cannot buy your own asset")
	}

	if !asset.ForSale {
		return fmt.Errorf("asset not for sale")
	}

	if buyer.Balance < asset.Price {
		return fmt.Errorf("not enough balance")
	}

	seller, err := GetTrader(stub, asset.Owner)
	if err != nil {
		return err
	}

	if seller == nil {
		return fmt.Errorf("seller not exist")
	}

	buyer.Balance -= asset.Price
	seller.Balance += asset.Price
	asset.ForSale = false

	if err = buyer.Store(stub); err != nil {
		return err
	}

	if err = seller.Store(stub); err != nil {
		return err
	}

	if err = asset.Store(stub); err != nil {
		return err
	}

	return nil
}

func GetAsset(stub shim.ChaincodeStubInterface, id string) (*Asset, error) {
	key, err := stub.CreateCompositeKey(ASSET, []string{id})
	if err != nil {
		return nil, err
	}

	data, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, err
	}

	var asset Asset
	if err = json.Unmarshal(data, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}
