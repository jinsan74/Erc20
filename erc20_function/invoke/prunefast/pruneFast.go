package prunefast

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"sejongtelecom.net/erc20/wallet"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// 관리자 Address 저장 구조체
type AdminMetadata struct {
	Adminaddress string `json:"adminaddress"`
}

var compositKeyIdx string = "balanceOf"

// High Throughput 가비지 데이터를 정리 해줌
// params - ownerAddress, targetAddress
func PruneFast(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Vaild Wallet을 호출하여 Parameter 추출 및 유효성 검사.
	args = wallet.CallVaildWallet(stub)
	if args == nil {
		return sc.Response{Status: 501, Message: "Vaild Wallet Error!", Payload: nil}
	}

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 2")
	}

	targetAddress := args[0]

	// AMDIN 인지 확인
	/*
		isAdmin := checkAdmin(stub, ownerAddress)
		fmt.Println("IS ADMIN:", isAdmin)
		if !isAdmin {
			return shim.Error("This Function Only Excute Admin!")
		}
	*/

	deltaResultsIterator, deltaErr := stub.GetStateByPartialCompositeKey(compositKeyIdx, []string{targetAddress})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", targetAddress, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No variable by the name %s exists", targetAddress))
	}

	var finalVal float64
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {

		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		_, keyParts, splitKeyErr := stub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}

		operation := keyParts[1]
		valueStr := keyParts[2]

		value, convErr := strconv.ParseFloat(valueStr, 64)
		if convErr != nil {
			return shim.Error(convErr.Error())
		}

		deltaRowDelErr := stub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}

		switch operation {
		case "+":
			finalVal += value
		case "-":
			finalVal -= value
		default:
			return shim.Error(fmt.Sprintf("Unrecognized operation %s", operation))
		}
	}

	// 마지막 값으로 렛저에 데이터 저장
	updateResp := update(stub, []string{targetAddress, strconv.FormatFloat(finalVal, 'f', -1, 64), "+"})
	if updateResp.Status == 200 {
		return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", args[0], finalVal, i)))
	}

	return shim.Error(fmt.Sprintf("Failed to prune variable: all rows deleted but could not update value to %f, variable no longer exists in ledger", finalVal))
}

//렛저 데이터 업데이트
func update(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, expecting 3")
	}

	address := args[0]
	op := args[2]
	_, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Provided value was not a number")
	}

	if op != "+" && op != "-" {
		return shim.Error(fmt.Sprintf("Operator %s is unrecognized", op))
	}

	txid := stub.GetTxID()

	compositeKey, compositeErr := stub.CreateCompositeKey(compositKeyIdx, []string{address, op, args[1], txid})
	if compositeErr != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s: %s", address, compositeErr.Error()))
	}

	compositePutErr := stub.PutState(compositeKey, []byte{0x00})
	if compositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", address, compositePutErr.Error()))
	}

	return shim.Success([]byte(fmt.Sprintf("Successfully added %s%s to %s", op, args[1], address)))
}

//Admin 지갑인지 확인
func checkAdmin(stub shim.ChaincodeStubInterface, chkAddress string) bool {

	//-----AMDIN 인지 확인----------------------------------
	adminMeta := AdminMetadata{}
	adminMetaBytes, err := stub.GetState("ADMINADDRESS")
	if err != nil {
		fmt.Println("ERR1")
		return false
	}
	err = json.Unmarshal(adminMetaBytes, &adminMeta)
	if err != nil {
		fmt.Println("ERR2")
		return false
	}

	adminAddressBytes, err := json.Marshal(adminMeta.Adminaddress)
	if err != nil {
		fmt.Println("ERR3")
		return false
	}

	realAddress := string(adminAddressBytes)
	realAddress = strings.Replace(realAddress, "\"", "", -1)

	fmt.Println("REALADDR:" + realAddress)
	fmt.Println("CHKADDR:" + chkAddress)

	if realAddress == chkAddress {
		return true
	} else {
		return false
	}
	//----------------------------------------------------
}
