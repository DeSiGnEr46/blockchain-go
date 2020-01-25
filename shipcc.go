/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SmartContract struct {
}

type Shipment struct {
	Producto   string `json:"producto"`
	Modelo  string `json:"modelo"`
	Tipo string `json:"tipo"`
	Dimensiones  string `json:"dimensiones"`
	FechaFabric string `json:"fechafab"`
	Materiales string `json:"materiales"`
	Descripcion string `json:"descripcion"`
	Cantidad string `json:"cantidad"`
	Precio_ud string `json:"precio_ud"`
	Precio_total string `json:"precio_tot"`
	EntOrg string `json:"origen"`
	EntDst string `json:"dst"`
	Orderer string `json:orderer`
	FechaEnv string `json:"fechaenv"`
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	//Only return success
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately

	if fn == "set" {
		return s.set(stub, args)
	} else if fn == "getAll" {
		return s.getAll(stub, args)
	} else if fn == "get" {
		return s.get(stub, args)
	} else if fn == "getHist" {
		return s.getHistory(stub, args)
	}

	// Return the result as success payload
	return shim.Error("Invalid Smart Contract function name.")
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func (s *SmartContract) set(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 15 {
		return shim.Error("Incorrect arguments. Expecting 13 arguments")
	}

	var shipment = Shipment{Producto: args[1], Modelo: args[2], Tipo: args[3], Dimensiones: args[4], FechaFabric: args[5],
	Materiales: args[6], Descripcion: args[7], Cantidad: args[8], Precio_ud: args[9], Precio_total: args[10], EntOrg: args[11], EntDst: args[12], 
	Orderer: args[13], FechaEnv: args[14]}
	
	shipAsBytes, _ := json.Marshal(shipment)

	err := stub.PutState(args[0], shipAsBytes)
	if err != nil {
		return shim.Error("Failed to set asset")
	}
	fmt.Printf("Transición insertada con éxito")

	err2 := stub.PutState("lastKey",[]byte(args[0]))
	if err2 != nil {
		return shim.Error("Failed to set asset")
	}
	fmt.Printf("Last key actualizada con éxito")
	return shim.Success(nil)
}

// Get returns the value of the specified asset key
func (s *SmartContract) get(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key")
	}

	shipAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get asset")
	}

	return shim.Success(shipAsBytes);
}

// Get returns the value of the specified asset key
func (s *SmartContract) getHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key")
	}

	histIterator, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return shim.Error("Failed to get asset")
	}

	defer histIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for histIterator.HasNext() {
		response, err := histIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForMarble returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) getAll(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Failed to get asset: %s with error")
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllShipments:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
