package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type InfoShare struct {
}

type info struct {
	ObjectType string `json:"docType"`
	InfoID     string `json:"InfoID"`
	InfoType   string `json:"InfoType"`
	Content    string `json:"Content"`
	UploadTime string `json:"UploadTime"`
	Uploader   string `json:"Uploader"`
	Department string `json:"Department"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(InfoShare))
	if err != nil {
		fmt.Printf("Error starting InfoShare: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *InfoShare) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Entry point for Invocations
// ========================================
func (t *InfoShare) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initInfo" { //create a new info
		return t.initInfo(stub, args)
	} else if function == "readInfo" { //read an info
		return t.readInfo(stub, args)
	} else if function == "queryInfoByDepartment" { //query an info by department
		return t.queryInfoByDepartment(stub, args)
	} else if function == "queryInfoByUploader" { //query an info by uploader
		return t.queryInfoByUploader(stub, args)
	} else if function == "queryInfoByInfoType" { //query an info by infotype
		return t.queryInfoByInfoType(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initInfo - create a new info, store into chaincode state
// ============================================================
func (t *InfoShare) initInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//     0        1          2        3       4          5
	// "420106","weather"", "sunny", "10:10", "bob",  "airforce"
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init info")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	InfoID := args[0]
	InfoType := strings.ToLower(args[1])
	Content := args[2]
	UploadTime := args[3]
	Uploader := strings.ToLower(args[4])
	Department := strings.ToLower(args[5])

	// ==== Check if info already exists ====
	infoAsBytes, err := stub.GetState(InfoID)
	if err != nil {
		return shim.Error("Failed to get info: " + err.Error())
	} else if infoAsBytes != nil {
		return shim.Error("This info already exists: " + InfoID)
	}

	// ==== Create info object and marshal to JSON ====
	objectType := "info"
	info := &info{objectType, InfoID, InfoType, Content, UploadTime, Uploader, Department}
	infoJSONasBytes, err := json.Marshal(info)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save info to state ===
	err = stub.PutState(InfoID, infoJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== info saved and indexed. Return success ====
	fmt.Println("- end init info")
	return shim.Success(nil)
}

// ===============================================
// readInfo - read an info from chaincode state
// ===============================================
func (t *InfoShare) readInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ID, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting ID of the info to query")
	}

	ID = args[0]
	valAsbytes, err := stub.GetState(ID) //get the info from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + ID + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Info does not exist: " + ID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *InfoShare) queryInfoByDepartment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//     0
	// "airforce"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	department := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"info\",\"department\":\"%s\"}}", department)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *InfoShare) queryInfoByUploader(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	uploader := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"info\",\"uploader\":\"%s\"}}", uploader)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *InfoShare) queryInfoByInfoType(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//     0
	// "weather"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	infotype := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"info\",\"infotype\":\"%s\"}}", infotype)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
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

	return &buffer, nil
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
