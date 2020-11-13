package transfer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"sejongtelecom.net/erc20/erc20_function/query/balanceof"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// 토큰 전송시 발생하는 이벤트 구조체
type TransferEvent struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    uint64 `json:"amount"`
}

var compositKeyIdx string = "balanceOf"

// 토큰전송 - High Throughput 적용
// params - ownerAddress, toAddress, Amount
func Transfer(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	callerAddress, recipientAddress, transferAmount := params[0], params[1], params[2]
	transferAmount = strings.Trim(transferAmount, " ")
	fmt.Println("TRANSFER:", callerAddress+":"+recipientAddress+":"+transferAmount)
	// Token Value가 양수 인지 조사
	transferAmountInt, err := strconv.ParseUint(transferAmount, 10, 64)
	//fmt.Println("TRANSFER INT :", transferAmountInt)
	if err != nil {
		return shim.Error("transfer amount must be integer, err" + err.Error() + ":" + string(transferAmountInt))
	}

	if transferAmountInt <= 0 {
		return shim.Error("transfer amount must be positive")
	}

	//--보내는 금액보다 보유한 금액이 많은지 체크 하는 부분------
	//해당 로직이 있으면 High Throughput 기능 동작 안함
	newParam := []string{callerAddress}
	callerAmount := balanceof.BalanceOf(stub, newParam)
	callerAmountInt, err := strconv.ParseUint(string(callerAmount.GetPayload()), 10, 64)
	if err != nil {
		return shim.Error("ParseIntErr, err" + err.Error())
	}
	// check callerResult Amount is positive
	if callerAmountInt < transferAmountInt {
		return shim.Error("caller's balance is not sufficient")
	}
	//------------------------------------------------

	// Save the Reipient DATA
	txid := stub.GetTxID()
	recipientCompositeKey, err := stub.CreateCompositeKey(compositKeyIdx, []string{recipientAddress, "+", transferAmount, txid})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a Receiver composite key for %s: %s", recipientAddress, err.Error()))
	}
	recipientCompositePutErr := stub.PutState(recipientCompositeKey, []byte{0x00})
	if recipientCompositePutErr != nil {
		return shim.Error(fmt.Sprintf("[RECEIVER]Could not put operation for %s in the ledger: %s", recipientAddress, recipientCompositePutErr.Error()))
	}

	// Save the Caller DATA
	callerCompositeKey, err := stub.CreateCompositeKey(compositKeyIdx, []string{callerAddress, "-", transferAmount, txid})
	if err != nil {

		deltaRowDelErr := stub.DelState(recipientCompositeKey)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}

		return shim.Error(fmt.Sprintf("Could not create a caller composite key for %s: %s", callerAddress, err.Error()))
	}
	callerCompositePutErr := stub.PutState(callerCompositeKey, []byte{0x00})
	if callerCompositePutErr != nil {

		deltaRowDelErr := stub.DelState(recipientCompositeKey)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}

		return shim.Error(fmt.Sprintf("[Caller]Could not put operation for %s in the ledger: %s", callerAddress, callerCompositePutErr.Error()))
	}

	// Emit Transfer Event
	transferEvent := TransferEvent{Sender: callerAddress, Recipient: recipientAddress, Amount: transferAmountInt}
	transferEventBytes, err := json.Marshal(transferEvent)

	err = stub.SetEvent("transferEvent", transferEventBytes)
	if err != nil {
		return shim.Error("failed to SetEvent of TransgerEvent, err: " + err.Error())
	}

	fmt.Println(callerAddress + " send " + transferAmount + " to " + recipientAddress)

	return shim.Success([]byte("transfer success"))
}
