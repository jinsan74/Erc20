package utils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"

	"github.com/jinsan74/Erc20/wallet"
)

// DoTransfer is 토큰 Transfer
func DoTransfer(stub shim.ChaincodeStubInterface, transParam string, tokenName string) sc.Response {

	_, orgParam := stub.GetFunctionAndParameters()

	walletMeta := wallet.WalletMeta{}
	json.Unmarshal([]byte(orgParam[0]), &walletMeta)
	walletMeta.Transdata = transParam

	realTrans, _ := json.Marshal(walletMeta)

	chainCodeFunc := "transfer"
	invokeArgs := ToChaincodeArgs(chainCodeFunc, string(realTrans))
	channel := stub.GetChannelID()
	response := stub.InvokeChaincode(tokenName, invokeArgs, channel)

	return response
}

// DoBalanceOf is 토큰 balanceOf
func DoBalanceOf(stub shim.ChaincodeStubInterface, toaddress string, tokenName string) sc.Response {

	// 지갑형 트랜잭션 VAILD WALLET CHECK 및 지갑주소/파라미터 파싱
	chainCodeFunc := "balanceOf"
	invokeArgs := ToChaincodeArgs(chainCodeFunc, toaddress)
	channel := stub.GetChannelID()
	response := stub.InvokeChaincode(tokenName, invokeArgs, channel)

	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to balanceOf chaincode. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)
		return sc.Response{Status: 501, Message: "balanceOf Fail!", Payload: nil}
	}

	return response
}

// DoTransferMulti is 토큰 TransferMulti
func DoTransferMulti(stub shim.ChaincodeStubInterface, stTransferMetaArr []wallet.TransferMeta, tokenName string) sc.Response {

	_, orgParam := stub.GetFunctionAndParameters()

	walletMeta := wallet.WalletMeta{}
	json.Unmarshal([]byte(orgParam[0]), &walletMeta)

	stTransferStr, _ := json.Marshal(stTransferMetaArr)
	walletMeta.Transjdata = string(stTransferStr)
	realTrans, _ := json.Marshal(walletMeta)

	chainCodeFunc := "transferMulti"
	invokeArgs := ToChaincodeArgs(chainCodeFunc, string(realTrans))
	channel := stub.GetChannelID()
	response := stub.InvokeChaincode(tokenName, invokeArgs, channel)

	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to transfer chaincode. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)
		return sc.Response{Status: 501, Message: "transfer Fail!", Payload: nil}
	}

	return response
}

// ToChaincodeArgs is 외부 체인코드 호출시 파라미터 만드는 함수
func ToChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// GetNowDt is 현재 시간 반환
func GetNowDt(stub shim.ChaincodeStubInterface) int64 {
	nowTimestamp, _ := stub.GetTxTimestamp()
	nowdt := nowTimestamp.GetSeconds()

	return nowdt
}

// JsonFromQueryResponse 은 iterator 를 json 으로 변환
// query result iterator 와 응답 메타데이터를 넘기면, 리턴할 json 으로 변환해 줌
func JsonFromQueryResponse(resultsIterator shim.StateQueryIteratorInterface, responseMetadata *sc.QueryResponseMetadata) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	//buffer.WriteString("[")
	buffer.WriteString("{\"bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"")
	buffer.WriteString(",")
	buffer.WriteString("\"recordcnt\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"reqlist\":")
	buffer.WriteString("[")
	var i int64 = 0
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// 두번째 array 부터는 , 붙이기
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// json object 를 string 으로 변환해서
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
		i++
	}
	buffer.WriteString("]")
	buffer.WriteString("}")
	return &buffer, nil
}

// SaveMetaData is 데이터 저장
func SaveMetaData(stub shim.ChaincodeStubInterface, dataKey string, metaDataBytes []byte) sc.Response {

	// 저장
	err := stub.PutState(dataKey, metaDataBytes)
	if err != nil {
		return shim.Error("failed to PutState, error: " + err.Error())
	}
	return shim.Success(nil)
}
