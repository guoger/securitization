package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"golang.org/x/crypto/sha3"
)

const TRADER = "TRADER"

type Trader struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Balance int    `json:"balance"`
}

func (t *Trader) Store(stub shim.ChaincodeStubInterface) error {
	key, err := stub.CreateCompositeKey(TRADER, []string{t.ID})
	if err != nil {
		return err
	}

	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	if err = stub.PutState(key, data); err != nil {
		return err
	}

	return nil
}

func Enroll(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if err := enroll(stub, params); err != nil {
		fmt.Printf("Failed to enroll trader: %s\n", err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Enroll creates a new user in trade network
func enroll(stub shim.ChaincodeStubInterface, params []string) error {
	if len(params) != 1 {
		return fmt.Errorf("expect 1 parameter (Name), but got %d: %+v", len(params), params)
	}
	name := params[0]
	id, err := getID(stub)
	if err != nil {
		return err
	}

	key, err := stub.CreateCompositeKey(TRADER, []string{id})
	if err != nil {
		return err
	}

	data, err := stub.GetState(key)
	if err != nil {
		return err
	} else if len(data) != 0 {
		return fmt.Errorf("already enrolled")
	}

	trader := &Trader{
		Name:    name,
		ID:      id,
		Balance: 10000,
	}

	if err = trader.Store(stub); err != nil {
		return err
	}

	return nil
}

func getID(stub shim.ChaincodeStubInterface) (string, error) {
	idBytes, err := stub.GetCreator()
	if err != nil {
		return "", err
	}

	id := sha3.Sum256(idBytes)
	return hex.EncodeToString(id[:]), nil
}

func GetTrader(stub shim.ChaincodeStubInterface, id string) (*Trader, error) {
	key, err := stub.CreateCompositeKey(TRADER, []string{id})
	if err != nil {
		return nil, err
	}

	data, err := stub.GetState(key)
	if err != nil {
		return nil, err
	} else if len(data) == 0 {
		return nil, nil
	}

	var trader Trader
	if err = json.Unmarshal(data, &trader); err != nil {
		return nil, err
	}

	return &trader, nil
}
