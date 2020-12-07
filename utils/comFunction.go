package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// 토큰 Transfer
func DoTransfer(stub shim.ChaincodeStubInterface, transParam string, tokenName string) sc.Response {

	_, orgParam := stub.GetFunctionAndParameters()

	var jsonMap map[string]string
	json.Unmarshal([]byte(orgParam[0]), &jsonMap)
	jsonMap["transdata"] = transParam

	newString, _ := json.Marshal(jsonMap)
	fmt.Println(string(newString))

	chainCodeFunc := "transfer"
	invokeArgs := toChaincodeArgs(chainCodeFunc, string(newString))
	channel := stub.GetChannelID()
	response := stub.InvokeChaincode(tokenName, invokeArgs, channel)

	return response

}

// 토큰 balanceOf
func DoBalanceOf(stub shim.ChaincodeStubInterface, toaddress string, tokenName string) sc.Response {

	// 지갑형 트랜잭션 VAILD WALLET CHECK 및 지갑주소/파라미터 파싱
	chainCodeFunc := "balanceOf"
	invokeArgs := toChaincodeArgs(chainCodeFunc, toaddress)
	channel := stub.GetChannelID()
	response := stub.InvokeChaincode(tokenName, invokeArgs, channel)

	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to balanceOf chaincode. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)
		return sc.Response{Status: 501, Message: "balanceOf Fail!", Payload: nil}
	}

	return response
}

// 외부 체인코드 호출시 파라미터 만드는 함수
func ToChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// 현재 시간 반환
func GetNowDt(stub shim.ChaincodeStubInterface) int64 {
	nowTimestamp, _ := stub.GetTxTimestamp()
	nowdt := nowTimestamp.GetSeconds()

	return nowdt
}

// 데이터 저장
func SaveMetaData(stub shim.ChaincodeStubInterface, dataKey string, metaDataBytes []byte]) sc.Response {

	// 저장
	err = stub.PutState(dataKey, metaDataBytes)
	if err != nil {
		return shim.Error("failed to PutState, error: " + err.Error())
	}

	return shim.Success(nil)
}
