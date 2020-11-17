/*
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

//SmartContract is the data structure which represents this contract and on which  various contract lifecycle functions are attached
type SmartContract struct {
}

type Document struct {
	ObjectType string `json:"Type"`
	Document   string `json:"document"`
	Hash       string `json:"hash"`
}

type Admin struct {
	ObjectType string   `json:"Type"`
	AdminID    string   `json:"adminID"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Message    []string `json:"message"`
}

func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("Init Firing!")
	return shim.Success(nil)
}

func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Chaincode Invoke Is Running " + function)
	if function == "addDocument" {
		return t.addDocument(stub, args)
	}
	if function == "queryDocuments" {
		return t.queryDocuments(stub)
	}
	if function == "queryDocumentByHash" {
		return t.queryDocumentByHash(stub, args)
	}
	if function == "addAdmin" {
		return t.addAdmin(stub, args)
	}
	if function == "queryAdmin" {
		return t.queryAdmin(stub, args)
	}
	if function == "updateAdmin" {
		return t.updateAdmin(stub, args)
	}

	fmt.Println("Invoke did not find specified function " + function)
	return shim.Error("Invoke did not find specified function " + function)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func (t *SmartContract) addDocument(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect Number of Aruments. Expecting 1")
	}

	fmt.Println("Adding new Documnet")

	// ==== Input sanitation ====
	if len(args[0]) <= 0 {
		return shim.Error("1st Argument Must be a Non-Empty String")
	}
	document := args[0]

	rand.Seed(time.Now().UnixNano())
	hash := randSeq(30)

	// ======Check if Document Already exists

	DocumentAsBytes, err := stub.GetState(hash)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if DocumentAsBytes != nil {
		return shim.Error("The Inserted document already Exists")
	}

	// ===== Create Document Object and Marshal to JSON

	objectType := "Document"
	Document := &Document{objectType, document, hash}
	DocumentJSONasBytes, err := json.Marshal(Document)

	if err != nil {
		return shim.Error(err.Error())
	}

	// ======= Save Document to State

	err = stub.PutState(document, DocumentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ======= Return Success

	fmt.Println("Successfully Saved Document")
	return shim.Success(nil)
}

func (t *SmartContract) queryDocumentByHash(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	hash := args[0]

	queryString := fmt.Sprintf("{\"selector\":{\"Type\":\"Document\",\"hash\":\"%s\"}}", hash)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func (t *SmartContract) queryDocuments(stub shim.ChaincodeStubInterface) peer.Response {

	queryString := fmt.Sprintf("{\"selector\":{\"Type\":\"Document\"}}")

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func (t *SmartContract) addAdmin(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect Number of Aruments. Expecting 3")
	}

	fmt.Println("Adding new Admin")

	// ==== Input sanitation ====
	if len(args[0]) <= 0 {
		return shim.Error("1st Argument Must be a Non-Empty String")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd Argument Must be a Non-Empty String")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd Argument Must be a Non-Empty String")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4rd Argument Must be a Non-Empty String")
	}

	adminID := args[0]
	username := args[1]
	password := args[2]
	message := args[3]
	// ======Check if admin Already exists

	adminAsBytes, err := stub.GetState(adminID)
	if err != nil {
		return shim.Error("Transaction Failed with Error: " + err.Error())
	} else if adminAsBytes != nil {
		return shim.Error("The Inserted expert ID already Exists: " + adminID)
	}

	// ===== Create admin Object and Marshal to JSON

	objectType := "admin"
	admin := &Admin{objectType, adminID, username, password, append(Admin{}.Message, message)}
	adminJSONasBytes, err := json.Marshal(admin)

	if err != nil {
		return shim.Error(err.Error())
	}

	// ======= Save admin to State

	err = stub.PutState(adminID, adminJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ======= Return Success

	fmt.Println("Successfully Saved admin")
	return shim.Success(nil)
}

func (t *SmartContract) updateAdmin(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	adminID := args[0]
	newMessage := args[1]
	fmt.Println("- start  ", adminID, newMessage)

	responseAsBytes, err := stub.GetState(adminID)
	if err != nil {
		return shim.Error("Failed to get status:" + err.Error())
	} else if responseAsBytes == nil {
		return shim.Error("response does not exist")
	}

	responseToUpdate := Admin{}
	err = json.Unmarshal(responseAsBytes, &responseToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	responseToUpdate.Message = append(responseToUpdate.Message, newMessage) //change the status

	responseJSONasBytes, _ := json.Marshal(responseToUpdate)
	err = stub.PutState(adminID, responseJSONasBytes) //rewrite
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end  (success)")
	return shim.Success(nil)
}

func (t *SmartContract) queryAdmin(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	username := args[0]
	password := args[1]

	queryString := fmt.Sprintf("{\"selector\":{\"Type\":\"admin\",\"username\":\"%s\",\"password\":\"%s\"}}", username, password)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
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

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

//Main Function starts up the Chaincode
func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Smart Contract could not be run. Error Occured: %s", err)
	} else {
		fmt.Println("Smart Contract successfully Initiated")
	}
}
