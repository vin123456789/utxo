/*
 * Hyperledger Fabric Chaincode exercise
 */

package main

/* Imports
 * 6 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// SmartContract defines the Smart Contract structure
type SmartContract struct {
}

// UTXO
type UTXO struct {
	Txid    string `json:"txid"`
	Index   string `json:"index"`
	Amount  string `json:"amount"`
	Address string `json:"address"`
	inOrOut string `json:"inOrOut"`
}

// Tranction
type Transaction struct {
	id      string `json:"id"`
	inputs  []UTXO `json:"input"`
	outputs []UTXO `json:"output"`
}

// Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
// Best practice is to have any Ledger initialization in separate function -- see initLedger()
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke method is called as a result of an application request to run the Smart Contract
// The calling application program has also specified the particular smart contract function to be called, with arguments
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "init" {
		return s.initState(APIstub)
	} else if function == "queryUTXO" {
		return s.queryUTXO(APIstub, args)
	} else if function == "queryUTXOByAddr" {
		return s.queryUTXOByAddr(APIstub, args)
	} else if function == "queryTransaction" {
		return s.queryTransaction(APIstub, args)
	} else if function == "getAllUTXO" {
		return s.getAllUTXO(APIstub)
	} else if function == "getAllTransaction" {
		return s.getAllTransaction(APIstub)
	} else if function == "transferUTXO" {
		return s.transferUTXO(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

//initialize coins
func (s *SmartContract) initState(APIstub shim.ChaincodeStubInterface) sc.Response {

	txid := APIstub.GetTxID()

	//math.MaxUint32 means it's a coinbase
	inputs := []UTXO{
		makeUTXO(txid, strconv.Itoa(math.MaxUint32), "50", "Coinbase", "in"),
	}

	outputs := []UTXO{
		makeUTXO(txid, "1", "50", "User A", "out"),
	}

	transaction := makeTransaction(txid, inputs, outputs)

	//store utxo
	err := storeUTXO(APIstub, txid, outputs[0])

	if err != nil {
		return shim.Error(err.Error())
	}

	//store transaction (It's not necessary to store transaction here, transaction info can be found in block.)
	err = storeTransaction(APIstub, txid, transaction)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func makeUTXO(txid string, index string, amount string, address string, inOrOut string) UTXO {

	utxo := UTXO{txid, index, amount, address, inOrOut}

	return utxo
}

func storeUTXO(APIstub shim.ChaincodeStubInterface, txid string, utxo UTXO) error {

	utxoKey := txid + ":" + utxo.Index
	utxoAsBYtes, _ := json.Marshal(utxo)

	//UTXO key is transaction id:index
	return APIstub.PutState(utxoKey, utxoAsBYtes)
}

func makeTransaction(id string, inputs []UTXO, outputs []UTXO) Transaction {

	transaction := Transaction{id, inputs, outputs}

	return transaction
}

func storeTransaction(APIstub shim.ChaincodeStubInterface, txid string, transaction Transaction) error {

	transactionAsBYtes, _ := json.Marshal(transaction)

	return APIstub.PutState(txid, transactionAsBYtes)
}

// Get all utxo
func (s *SmartContract) getAllUTXO(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := ""
	endKey := ""

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)

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
			return shim.Error(err.Error())
		}

		//Values contain comma are the UTXO objects.
		if strings.Contains(queryResponse.Key, ":") {
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
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllUtxo:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// Get all transaction
func (s *SmartContract) getAllTransaction(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := ""
	endKey := ""

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)

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
		//skip utxo
		if strings.Contains(queryResponse.Key, ":") {
			continue
		}
		if err != nil {
			return shim.Error(err.Error())
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

	fmt.Printf("- queryAllTransaction:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// Query transaction by transaction id
// it's not needed since it can be found in block, here just for easy tracking
func (s *SmartContract) queryTransaction(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	transactionAsBytes, err := APIstub.GetState(args[0])

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(transactionAsBytes)
}

// Query utxo by key (txid:index)
func (s *SmartContract) queryUTXO(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	utxoAsBytes, err := APIstub.GetState(args[0])

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(utxoAsBytes)
}

// Query utxo by address
func (s *SmartContract) queryUTXOByAddr(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	address := args[0]

	//Get all UTXO
	startKey := ""
	endKey := ""
	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)

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
			return shim.Error(err.Error())
		}

		if strings.Contains(queryResponse.Key, ":") {
			var utxo UTXO
			json.Unmarshal(queryResponse.Value, &utxo)

			//check utxo owner ship, only return the one with address equal to with args[0]
			if checkUTXOOwnerShip(utxo, address) == false {
				continue
			}

			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			//transaction value, Record is a JSON object, so we write as-is
			buffer.WriteString("\"")
			buffer.WriteString(", \"Record\":")
			buffer.WriteString(string(queryResponse.Value))

			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- queryUTXOByAddr:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// Check if one utxo belong to one particular address
// At this point, there is no encoding and decoding in address, so just check the address and spent
func checkUTXOOwnerShip(utxo UTXO, address string) bool {
	return utxo.Address == address && utxo.inOrOut == "out"
}

// Transfer utxo from one address to another
func (s *SmartContract) transferUTXO(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	addrFrom := args[0]
	addrTo := args[1]
	amount, _ := strconv.ParseFloat(args[2], 64)

	//get all utxo
	startKey := ""
	endKey := ""
	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	inputs := []UTXO{}
	utfoKeysToRemove := []string{}
	currValue := 0.0
	txid := APIstub.GetTxID()

	for resultsIterator.HasNext() && currValue < amount {

		//loop the utxo one by one
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		//prepare utxo to spend
		var utxo UTXO
		json.Unmarshal(queryResponse.Value, &utxo)

		//skip the utxo belong to others
		if checkUTXOOwnerShip(utxo, addrFrom) == false {
			continue
		}

		utxo.inOrOut = "in"

		inputs = append(inputs, utxo)

		utfoKeysToRemove = append(utfoKeysToRemove, queryResponse.Key)

		//accumulate currValue
		newValue, _ := strconv.ParseFloat(utxo.Amount, 64)
		currValue += newValue
	}

	//no enough utxo to spend
	if currValue < amount {
		return shim.Error("No enough amount to spend")
	}

	for _, v := range utfoKeysToRemove {
		//delete utxo which have been spent
		APIstub.DelState(v)
	}

	//create new outputs
	outputs := []UTXO{}
	utxo1 := makeUTXO(txid, "1", args[2], addrTo, "out")
	outputs = append(outputs, utxo1)

	//give the extract amount back to addrFrom
	var utxo2 UTXO
	if currValue > amount {
		utxo2 = makeUTXO(txid, "2", strconv.FormatFloat(float64(currValue-amount), 'f', 2, 64), addrFrom, "out")
		outputs = append(outputs, utxo2)
	}

	//store new utxo
	storeUTXO(APIstub, txid, utxo1)
	storeUTXO(APIstub, txid, utxo2)

	//store new transaction
	transaction := makeTransaction(txid, inputs, outputs)
	storeTransaction(APIstub, txid, transaction)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}