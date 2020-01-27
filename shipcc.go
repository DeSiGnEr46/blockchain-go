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

// Init se llama cuando se instancia o actualiza un chaincode
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	//Only return success
	return shim.Success(nil)
}

// Invoke se llama por cada interacción con el chaincode.
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	// Se extraen los argumentos enviados al chaincode
	fn, args := stub.GetFunctionAndParameters()

	// Enruta a la función elegida
	if fn == "set" {
		return s.set(stub, args)
	} else if fn == "getAll" {
		return s.getAll(stub, args)
	} else if fn == "get" {
		return s.get(stub, args)
	} else if fn == "getHist" {
		return s.getHistory(stub, args)
	} else if fn == "edit" {
		return s.edit(stub, args)
	}

	// Return the result as success payload
	return shim.Error("Invalid Smart Contract function name.")
}

// Almacena un nuevo valor en el 'ledger'. Además, actualiza el valor de la 'lastKey'.
func (s *SmartContract) set(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 15 {
		return shim.Error("Incorrect arguments. Expecting 13 arguments")
	}

	result , err := stub.GetState(args[0])
	if len(result) == 0 { //Si la longitud es 0 es porque no existe
		//Se inserta
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
			return shim.Error("Failed to set lastKey")
		}
		fmt.Printf("Last key actualizada con éxito")
		return shim.Success(nil)
	} else {
		fmt.Printf("Asset already exists")
		return shim.Error(err.Error())
	}
}

// Edita un valor del 'ledger' con una clave dada
func (s *SmartContract) edit(stub shim.ChaincodeStubInterface, args []string) sc.Response {
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

	return shim.Success(nil)
}

// Obtiene los valores asociados a una clave
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

// Obtiene el historial de transacciones para una clave dada
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

//Obtiene todos los valores para un rango de claves
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

// Función que inicia el chaincode en el contenedor
func main() {
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
