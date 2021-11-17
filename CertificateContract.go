package tmchain

import (
	"bytes"
	"encoding/json"
	"strconv"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

type SmartContract struct {
}

func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "init" {
		return s.Init(stub)
	} else if function == "saveCertificate" {
		return s.saveCertificate(stub, args)
	} else if function == "getAllCertificates" {
		return s.getAllCertificates(stub, args)
	} else if function == "queryCertificate" {
		return s.queryCertificate(stub, args)
	} else if function == "deleteCertificate" {
		return s.deleteCertificate(stub, args)
	} else if function == "getCertificateHistory" {
		return s.getCertificateHistory(stub, args)
	} else if function == "verifyCertificate" {
		return s.verifyCertificate(stub, args)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) saveCertificate(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	certificateJSON := Certificate{}

	err := json.Unmarshal([]byte(args[0]), &certificateJSON)

	if err != nil {
		return shim.Error("Invalid assetJSON")
	}

	certificateAsBytes, _ := json.Marshal(certificateJSON)
	stub.PutState(certificateJSON.SerialNumber, certificateAsBytes)

	return shim.Success(certificateAsBytes)
}

func (s *SmartContract) verifyCertificate(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	serialNumber := args[0]
	certificateHash := args[1]
	var buffer bytes.Buffer
	
	certificateAsBytes, err := stub.GetState(serialNumber)
	if err != nil || len(certificateAsBytes) == 0 {
		buffer.WriteString("{")
		buffer.WriteString("\"matched\":false","\"serialNumber\":false")
		buffer.WriteString("}")
		return shim.Success(buffer.Bytes())
	}

	certificateJSON := Certificate{}

	_ = json.Unmarshal(assetAsBytes, &certificateJSON)

	buffer.WriteString("{")

	if certificateJSON.CertificateHash == certificateHash {
		buffer.WriteString("\"matched\":true","\"serialNumber\":true")
	} else {
		buffer.WriteString("\"matched\":false","\"serialNumber\":true")
	}

	buffer.WriteString("}")

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryCertificate(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	serialNumber := args[0]

	certificateAsBytes, err := stub.GetState(serialNumber)
	if err != nil || len(certificateAsBytes) == 0 {
		return shim.Error("The SerialNumber " + serialNumber + " does not exist")
	}

	return shim.Success(certificateAsBytes)
}

func (s *SmartContract) deleteCertificate(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	serialNumber := args[0]

	certificateAsBytes, err := stub.GetState(serialNumber)
	if err != nil || len(assetAsBytes) == 0 {
		return shim.Error("The asset " + serialNumber + " does not exist")
	}

	stub.DelState(serialNumber)

	return shim.Success(nil)
}

func (s *SmartContract) getCertificateHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	serialNumber := args[0]

	resultsIterator, err := stub.GetHistoryForKey(serialNumber)
	if err != nil {
		return shim.Error("The serialNumber " + serialNumber + " does not exist")
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
		val := string(queryResponse.Value)
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(val)
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) getAllCertificates(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	startKey := args[0]
	endKey := args[1]
	pageSize, _ := strconv.Atoi(args[2])

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() && pageSize > 0 {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		val := string(queryResponse.Value)
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(val)
		bArrayMemberAlreadyWritten = true
		pageSize--
	}
	buffer.WriteString("]")

	var response bytes.Buffer

	response.WriteString("{")
	response.WriteString("\"page\": {")
	response.WriteString("\"pageSize\": " + args[2] + ",")
	response.WriteString("\"currentStartKey\": \"" + args[0] + "\",")
	response.WriteString("\"currentEndKey\": \"" + args[1] + "\",")
	response.WriteString("\"nextStartKey\": \"")
	if resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		certificateJSON := Certificate{}

		_ = json.Unmarshal(queryResponse.Value, &certificateJSON)

		response.WriteString(strconv.Itoa(certificateJSON.SerialNumber))
	} else {
		response.WriteString("")
	}
	response.WriteString("\"")
	response.WriteString("},")
	response.WriteString("\"content\": " + buffer.String())
	response.WriteString("}")

	return shim.Success(response.Bytes())
}
