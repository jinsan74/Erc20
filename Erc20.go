/*
 * SejongTelecom 코어기술개발팀
 * @author JinSan
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// TOKEN 이름을 지정해 준다 - 토큰이름은 아무 의미가 없음
const tokenName = "TOKEN"

type ERC20Chaincode struct {
}

// 초기데이터를 한번만 적용하기 위한 데이터 구조체
type InitMetadata struct {
	IsInit bool `json:"isinit"`
}

// 관리자 Address 저장 구조체
type AdminMetadata struct {
	Adminaddress string `json:"adminaddress"`
}

//토큰 정보 저장 구조체
type ERC20Metadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Owner       string `json:"owner"`
	TotalSupply uint64 `json:"totalSupply"`
}

// 토큰 전송시 발생하는 이벤트 구조체
type TransferEvent struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    uint64 `json:"amount"`
}

// Approval 이벤트 및 데이터 구조체
type Approval struct {
	Owner     string `json:"owner"`
	Spender   string `json:"spender"`
	Allowance int    `json:"allowance"`
}

var compositKeyIdx string = "balanceOf"

// ERC20 토큰 초기화
// params - owner(address), symbol, amount
func (cc *ERC20Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	_, params := stub.GetFunctionAndParameters()

	// 이미 초기화가 되었다면 Success Return
	initDataBytes, err := stub.GetState("INIT")
	if err != nil {
		return shim.Error("failed to GetState, err: " + err.Error())
	}

	if initDataBytes != nil {
		fmt.Println("Already Init Success")
		return shim.Success(nil)
	}

	// 초기화를 한번만 하기 위해 데이터 영역에 INIT == Y 로 기록함----
	err = stub.PutState("INIT", []byte("Y"))
	if err != nil {
		return shim.Error("failed to PutState of INIT Data, err: " + err.Error())
	}

	// 파라미터 체크 후 기본 토큰 데이터 세팅
	if len(params) != 3 {
		return shim.Error("incorrect number of transaction parameter")
	}

	owner, symbol, amount := params[0], params[1], params[2]

	amountUint, err := strconv.ParseUint(string(amount), 10, 64)
	if err != nil {
		return shim.Error("amount must be a number or amount cannot be negative")
	}

	if len(symbol) == 0 || len(owner) == 0 {
		return shim.Error("Symbol or owner cannot be emtpy")
	}

	erc20 := &ERC20Metadata{Name: tokenName, Symbol: symbol, Owner: owner, TotalSupply: amountUint}
	erc20Bytes, err := json.Marshal(erc20)
	if err != nil {
		return shim.Error("failed to Marshal erc20, error: " + err.Error())
	}

	// 토큰 데이터 저장
	err = stub.PutState(tokenName, erc20Bytes)
	if err != nil {
		return shim.Error("failed to PutState, error: " + err.Error())
	}

	// Owner Address 에 생성한 모든 코인 저장(High Throughput 적용)
	txid := stub.GetTxID()
	ownerCompositeKey, err := stub.CreateCompositeKey(compositKeyIdx, []string{owner, "+", amount, txid})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a caller composite key for %s: %s", owner, err.Error()))
	}
	ownerCompositePutErr := stub.PutState(ownerCompositeKey, []byte{0x00})
	if ownerCompositePutErr != nil {
		return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", owner, ownerCompositePutErr.Error()))
	}

	// Owner Address 를 관리자 Address로 등록
	adminMeta := &AdminMetadata{Adminaddress: owner}
	adminMetaBytes, err := json.Marshal(adminMeta)
	if err != nil {
		return shim.Error("failed to Marshal AdminSave, error: " + err.Error())
	}
	err = stub.PutState("ADMINADDRESS", adminMetaBytes)
	if err != nil {
		return shim.Error("failed to PutState, error: " + err.Error())
	}

	return shim.Success(nil)
}

// Invoke ChainCode
func (cc *ERC20Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, orgParam := stub.GetFunctionAndParameters()

	if len(orgParam) != 1 {
		return shim.Error("incorrect number of parameter")
	}

	getParams := callVaildWallet(stub)

	if getParams == nil {
		return sc.Response{Status: 501, Message: "Vaild Wallet Error!", Payload: nil}
	}

	switch fcn {
	case "totalSupply":
		return cc.totalSupply(stub, getParams)
	case "balanceOf":
		return cc.balanceOf(stub, getParams)
	case "transfer":
		return cc.transfer(stub, getParams)
	case "allowance":
		return cc.allowance(stub, getParams)
	case "approve":
		return cc.approve(stub, getParams)
	case "approvalList":
		return cc.approvalList(stub, getParams)
	case "transferFrom":
		return cc.transferFrom(stub, getParams)
	case "transferOtherToken":
		return cc.transferOtherToken(stub, getParams)
	case "increaseAllowance":
		return cc.increaseAllowance(stub, getParams)
	case "decreaseAllowance":
		return cc.decreaseAllowance(stub, getParams)
	case "mint":
		return cc.mint(stub, getParams)
	case "burn":
		return cc.burn(stub, getParams)
	case "rowCount":
		return cc.rowCount(stub, getParams)
	case "pruneFast":
		return cc.pruneFast(stub, getParams)
	default:
		return sc.Response{Status: 404, Message: "404 Not Found", Payload: nil}
	}
}

// vaildWallet 호출 함수
func callVaildWallet(stub shim.ChaincodeStubInterface) []string {

	fcn, orgParam := stub.GetFunctionAndParameters()

	isInvoke := "Y"

	switch fcn {
	case "totalSupply":
		isInvoke = "N"
	case "balanceOf":
		isInvoke = "N"
	case "transfer":
		isInvoke = "Y"
	case "allowance":
		isInvoke = "N"
	case "approve":
		isInvoke = "Y"
	case "approvalList":
		isInvoke = "N"
	case "transferFrom":
		isInvoke = "Y"
	case "transferOtherToken":
		isInvoke = "Y"
	case "increaseAllowance":
		isInvoke = "Y"
	case "decreaseAllowance":
		isInvoke = "Y"
	case "mint":
		isInvoke = "Y"
	case "burn":
		isInvoke = "Y"
	case "rowCount":
		isInvoke = "N"
	case "pruneFast":
		isInvoke = "Y"
	default:
		return nil
	}

	retParams := orgParam

	if strings.Compare(isInvoke, "Y") == 0 {

		orgString := orgParam[0]

		//--트랜잭션 String 변환 : realfunc 추가----
		var jsonMap map[string]string
		json.Unmarshal([]byte(orgString), &jsonMap)
		jsonMap["realfunc"] = fcn
		newString, _ := json.Marshal(jsonMap)
		fmt.Println(string(newString))
		//--------------------------------------

		// 지갑형 트랜잭션 VAILD WALLET CHECK 및 지갑주소/파라미터 파싱
		chainCodeFunc := "vaildWallet"
		invokeArgs := toChaincodeArgs(chainCodeFunc, string(newString))
		channel := stub.GetChannelID()
		response := stub.InvokeChaincode("vaildWallet", invokeArgs, channel)

		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to vaildWallet chaincode. Got error: %s", string(response.Payload))
			fmt.Printf(errStr)
			return nil
			//return sc.Response{Status: 501, Message: "vaild Wallet Fail!", Payload: nil}
		}
		//-----------------------------------------------------------------------

		retParams = strings.Split(string(response.Payload), ",")
	}

	return retParams
}

// 외부 체인코드 호출시 파라미터 만드는 함수
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// 전체 토큰 발행량 조회
// params - ownerAddress
// Returns 토큰 발행량
func (cc *ERC20Chaincode) totalSupply(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 1 {
		return shim.Error("incorrect number of parameter")
	}

	// Get ERC20 Metadata
	erc20 := ERC20Metadata{}
	erc20Bytes, err := stub.GetState(tokenName)
	if err != nil {
		return shim.Error("failed to GetState, error: " + err.Error())
	}
	err = json.Unmarshal(erc20Bytes, &erc20)
	if err != nil {
		return shim.Error("failed to Unmarshal, error: " + err.Error())
	}

	// Convert TotalSupply to Bytes
	totalSupplyBytes, err := json.Marshal(erc20.TotalSupply)
	if err != nil {
		return shim.Error("failed to Marshal totalSupply, error: " + err.Error())
	}
	fmt.Println(tokenName + "'s totalSupply is " + string(totalSupplyBytes))

	return shim.Success(totalSupplyBytes)
}

// 요청한 주소의 현재 토큰 수 조회
// params - ownerAddress, queryAddress
// Returns 토큰수
func (cc *ERC20Chaincode) balanceOf(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 1 {
		return shim.Error("incorrect number of parameters")
	}

	queryAddress := params[0]

	balanceResultsIterator, err := stub.GetStateByPartialCompositeKey(compositKeyIdx, []string{queryAddress})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve balance value for %s: %s", queryAddress, err.Error()))
	}

	defer balanceResultsIterator.Close()

	// 데이터가 없으면 balance = 0
	if !balanceResultsIterator.HasNext() {
		fmt.Println("NO DATA")
		return shim.Success([]byte("0"))
	}

	var finalVal int
	var i int
	for i = 0; balanceResultsIterator.HasNext(); i++ {
		responseRange, nextErr := balanceResultsIterator.Next()
		if nextErr != nil {
			fmt.Println("NO DATA1")
			return shim.Error(nextErr.Error())
		}

		_, keyParts, splitKeyErr := stub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			fmt.Println("NO DATA2")
			return shim.Error(splitKeyErr.Error())
		}

		operation := keyParts[1]
		valueStr := keyParts[2]

		//fmt.Println("DATA:" + operation + ":" + valueStr)

		value, convErr := strconv.Atoi(valueStr)
		if convErr != nil {
			return shim.Error(convErr.Error())
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
	return shim.Success([]byte(strconv.Itoa(finalVal)))
}

// 토큰전송 - High Throughput 적용
// params - ownerAddress, toAddress, Amount
func (cc *ERC20Chaincode) transfer(stub shim.ChaincodeStubInterface, params []string) sc.Response {

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
	callerAmount := cc.balanceOf(stub, newParam)
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

func (cc *ERC20Chaincode) rowCount(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if len(params) != 1 {
		return shim.Error("incorrect number of parameters")
	}

	queryAddress := params[0]

	balanceResultsIterator, err := stub.GetStateByPartialCompositeKey(compositKeyIdx, []string{queryAddress})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve balance value for %s: %s", queryAddress, err.Error()))
	}

	defer balanceResultsIterator.Close()

	// 데이터가 없으면 balance = 0
	if !balanceResultsIterator.HasNext() {
		fmt.Println("NO DATA")
		return shim.Success([]byte("0"))
	}

	var finalVal int
	var i int
	for i = 0; balanceResultsIterator.HasNext(); i++ {
		responseRange, nextErr := balanceResultsIterator.Next()
		if nextErr != nil {
			fmt.Println("NO DATA1")
			return shim.Error(nextErr.Error())
		}

		_, keyParts, splitKeyErr := stub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			fmt.Println("NO DATA2")
			return shim.Error(splitKeyErr.Error())
		}

		operation := keyParts[1]
		valueStr := keyParts[2]

		fmt.Println("DATA:" + operation + ":" + valueStr)

		finalVal++
	}
	return shim.Success([]byte(strconv.Itoa(finalVal)))
}

// owner 가 spender 에게 인출을 허락한 토큰의 개수는 몇개인지 조회
// params - ownerAddress, spenderAddress
// Returns 토큰 수
func (cc *ERC20Chaincode) allowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 2 {
		return shim.Error("incorrect number of parameters")
	}

	ownerAddress, spenderAddress := params[0], params[1]

	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	if err != nil {
		return shim.Error("failed to CreateCompositeKey for approval")
	}

	amountBytes, err := stub.GetState(approvalKey)
	if err != nil {
		return shim.Error("failed to GetState for amount")
	}
	if amountBytes == nil {
		amountBytes = []byte("0")
	}

	return shim.Success(amountBytes)

}

// Ownere가 Spender 에서 Amount 만큼 토큰을 인출 할 권리 부여
// params - ownerAddress, spenderAddress, Amount
func (cc *ERC20Chaincode) approve(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 3 {
		return shim.Error("incorrect number of parameters")
	}

	ownerAddress, spenderAddress, allowanceAmount := params[0], params[1], params[2]

	// Admount가 숫자인지 양수인지 조사
	allowanceAmountInt, err := strconv.Atoi(allowanceAmount)
	if err != nil {
		return shim.Error("allowance amount must be integer")
	}
	if allowanceAmountInt < 0 {
		return shim.Error("allowance amount must be positve")
	}

	// Create composite key For Allowance - approval/{owner}/{spender}
	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	if err != nil {
		return shim.Error("failed to CreateCompositeKey for approval")
	}

	err = stub.PutState(approvalKey, []byte(allowanceAmount))
	if err != nil {
		return shim.Error("failed to PutState for approval")
	}

	// Emit Approval Event
	approvalEvent := Approval{Owner: ownerAddress, Spender: spenderAddress, Allowance: allowanceAmountInt}
	approvalBytes, err := json.Marshal(approvalEvent)
	if err != nil {
		return shim.Error("failed to SetEvent for ApprovalEvent")
	}
	err = stub.SetEvent("approvalEvent", approvalBytes)
	if err != nil {
		return shim.Error("failed to SetEvent for ApprovalEvent")
	}

	return shim.Success([]byte("approve success"))
}

// 권한을 위힘한 Approval 리스트 조회
// params - ownerAddress
// Returns Approval List
func (cc *ERC20Chaincode) approvalList(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 1 {
		return shim.Error("incorrect number of params")
	}

	ownerAddress := params[0]

	approvalIterator, err := stub.GetStateByPartialCompositeKey("approval", []string{ownerAddress})
	if err != nil {
		return shim.Error("failed to GetStateByPartialCompositeKey for approval iterationm error: " + err.Error())
	}

	approvalSlice := []Approval{}

	defer approvalIterator.Close()
	if approvalIterator.HasNext() {
		for approvalIterator.HasNext() {
			approvalKV, _ := approvalIterator.Next()

			_, addresses, err := stub.SplitCompositeKey(approvalKV.GetKey())
			if err != nil {
				return shim.Error("failed to SplitCompositeKey, error: " + err.Error())
			}
			spenderAddress := addresses[1]

			amountBytes := approvalKV.GetValue()
			amountInt, err := strconv.Atoi(string(amountBytes))
			if err != nil {
				return shim.Error("failed to get amount, error: " + err.Error())
			}

			approval := Approval{Owner: ownerAddress, Spender: spenderAddress, Allowance: amountInt}
			approvalSlice = append(approvalSlice, approval)
		}
	}

	response, err := json.Marshal(approvalSlice)
	if err != nil {
		return shim.Error("failed to Marshal approvalSlice, error: " + err.Error())
	}

	return shim.Success(response)
}

// From 계좌에서 To 계좌로 Admount 토큰을 전송 , 단 Approve 함수를 통해 권한을 위임 받은 Spender만 할 수 있음
// parmas - spenderAddress(실행 ADDRESS), ownerAddress(원래토큰주인) , recipientAddress(받을사람), Amount
func (cc *ERC20Chaincode) transferFrom(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 4 {
		return shim.Error("incorrect number of params")
	}

	spenderAddress, ownerAddress, recipientAddress, transferAmount := params[0], params[1], params[2], params[3]

	transferAmountInt, err := strconv.Atoi(transferAmount)
	if err != nil {
		return shim.Error("amount must be integer")
	}
	if transferAmountInt <= 0 {
		return shim.Error("amount must be positve")
	}

	allowanceResponse := cc.allowance(stub, []string{ownerAddress, spenderAddress})
	if allowanceResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error: " + allowanceResponse.GetMessage())
	}

	allowanceInt, err := strconv.Atoi(string(allowanceResponse.GetPayload()))
	if err != nil {
		return shim.Error("allowance must be positive")
	}

	transferResponse := cc.transfer(stub, []string{ownerAddress, recipientAddress, transferAmount})
	if transferResponse.GetStatus() >= 400 {
		return shim.Error("failed to transfer, error: " + transferResponse.GetMessage())
	}

	approveAmountInt := allowanceInt - transferAmountInt
	approveAmount := strconv.Itoa(approveAmountInt)

	approveResponse := cc.approve(stub, []string{ownerAddress, spenderAddress, approveAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to approve, error: " + approveResponse.GetMessage())
	}

	return shim.Success([]byte("transferFrom success"))
}

// 다른 Chain 코드의 토큰을 이동
// params - ownerAddress, chaincode name, toAddress, Amount
func (cc *ERC20Chaincode) transferOtherToken(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 4 {
		return shim.Error("incorrect number of params")
	}

	ownerAddress, chaincodeName, recipientAddress, transferAmount := params[0], params[1], params[2], params[3]

	args := [][]byte{[]byte("transfer"), []byte(ownerAddress), []byte(recipientAddress), []byte(transferAmount)}

	// 현재 Channel 조회
	channel := stub.GetChannelID()

	// 타 체인코드 토큰 이동
	transferResponse := stub.InvokeChaincode(chaincodeName, args, channel)
	if transferResponse.GetStatus() >= 400 {
		return shim.Error(fmt.Sprintf("failed to transfer %s, error: %s", chaincodeName, transferResponse.GetMessage()))
	}

	return shim.Success([]byte("transfer other token success"))
}

// Allowance 값을 증가
// params -  ownerAddress, spenderAddress, Amount
func (cc *ERC20Chaincode) increaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 3 {
		return shim.Error("incorrect number of params")
	}

	ownerAddress, spenderAddress, increaseAmount := params[0], params[1], params[2]

	increaseAmountInt, err := strconv.Atoi(increaseAmount)
	if err != nil {
		return shim.Error("amount must be integer")
	}
	if increaseAmountInt <= 0 {
		return shim.Error("amount must be positve")
	}

	allowanceResponse := cc.allowance(stub, []string{ownerAddress, spenderAddress})
	if allowanceResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error: " + allowanceResponse.GetMessage())
	}

	allowanceInt, err := strconv.Atoi(string(allowanceResponse.GetPayload()))
	if err != nil {
		return shim.Error("allowance must be positive")
	}

	resultAmountInt := allowanceInt + increaseAmountInt
	resultAmount := strconv.Itoa(resultAmountInt)

	approveResponse := cc.approve(stub, []string{ownerAddress, spenderAddress, resultAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to approve allowance, error: " + approveResponse.GetMessage())
	}

	return shim.Success([]byte("increaseAllowance success"))
}

// Allowance 값을 감소
// params - ownerAddress, spenderAddress, Amount
func (cc *ERC20Chaincode) decreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 3 {
		return shim.Error("incorrect number of params")
	}

	ownerAddress, spenderAddress, decreaseAmount := params[0], params[1], params[2]

	decreaseAmountInt, err := strconv.Atoi(decreaseAmount)
	if err != nil {
		return shim.Error("decrease amount must be integer")
	}
	if decreaseAmountInt <= 0 {
		return shim.Error("decrease amount must be positive")
	}

	allowanceResponse := cc.allowance(stub, []string{ownerAddress, spenderAddress})
	if allowanceResponse.Status >= 400 {
		return shim.Error("failed to get allowance, error: " + allowanceResponse.GetMessage())
	}

	allowanceInt, err := strconv.Atoi(string(allowanceResponse.GetPayload()))
	if err != nil {
		return shim.Error("allowance must be positive")
	}

	resultAmountInt := allowanceInt - decreaseAmountInt
	if resultAmountInt < 0 {
		resultAmountInt = 0
	}
	resultAmount := strconv.Itoa(resultAmountInt)

	approveResponse := cc.approve(stub, []string{ownerAddress, spenderAddress, resultAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to approve allowance, error: " + approveResponse.GetMessage())
	}

	return shim.Success([]byte("decreaseAllowance success"))
}

// Owner의 토큰을 증가시키고 전체 발행량을 증가(토큰 추가 발행) - 관리자만 실행가능
// params - ownerAddress, Amount
func (cc *ERC20Chaincode) mint(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 2 {
		return shim.Error("incorrect number of params")
	}

	ownerAddress, mintAmount := params[0], params[1]

	// AMDIN 인지 확인
	isAdmin := checkAdmin(stub, ownerAddress)
	fmt.Println("IS ADMIN:", isAdmin)

	if !isAdmin {
		return shim.Error("This Function Only Excute Admin!")
	}

	mintAmountInt, err := strconv.Atoi(mintAmount)
	if err != nil {
		return shim.Error("mint amount must be integer")
	}
	if mintAmountInt <= 0 {
		return shim.Error("mint amount must be positive")
	}

	erc20 := ERC20Metadata{}
	erc20Bytes, err := stub.GetState(tokenName)
	if err != nil {
		return shim.Error("failed to GetState, error: " + err.Error())
	}
	err = json.Unmarshal(erc20Bytes, &erc20)
	if err != nil {
		return shim.Error("failed to Unmarshal, error: " + err.Error())
	}

	resultTotalSupply := erc20.TotalSupply + uint64(*&mintAmountInt)

	erc20 = ERC20Metadata{Name: erc20.Name, Symbol: erc20.Symbol, Owner: erc20.Owner, TotalSupply: resultTotalSupply}
	erc20Bytes, err = json.Marshal(erc20)
	if err != nil {
		return shim.Error("failed to Marshal erc20, error: " + err.Error())
	}

	err = stub.PutState(tokenName, erc20Bytes)
	if err != nil {
		return shim.Error("failed to PutState, error: " + err.Error())
	}

	recipientAmount, err := stub.GetState(ownerAddress)
	if err != nil {
		return shim.Error("failed to GetState, error: " + err.Error())
	}
	if recipientAmount == nil {
		recipientAmount = []byte("0")
	}
	recipientAmountInt, err := strconv.Atoi(string(recipientAmount))
	if err != nil {
		return shim.Error("caller amount must be integer")
	}

	// Owner Balance 증가
	resultBalance := recipientAmountInt + mintAmountInt

	err = stub.PutState(ownerAddress, []byte(strconv.Itoa(resultBalance)))
	if err != nil {
		return shim.Error("failed to PutState of caller, error: " + err.Error())
	}

	return shim.Success([]byte("mint success"))
}

// 지정한 계정의 토큰을 소각함 - 관리자만 실행 가능(관리자가 남의 코인을 삭제 가능 한 것이 맞는지?)
// params - ownerAddress, recipientAddress, Amount
func (cc *ERC20Chaincode) burn(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 3 {
		return shim.Error("incorrect number of params")
	}

	ownerAddress, recipientAddress, burnAmount := params[0], params[1], params[2]

	// AMDIN 인지 확인
	isAdmin := checkAdmin(stub, ownerAddress)
	fmt.Println("IS ADMIN:", isAdmin)

	if !isAdmin {
		return shim.Error("This Function Only Excute Admin!")
	}

	burnAmountInt, err := strconv.Atoi(burnAmount)
	if err != nil {
		return shim.Error("burn amount must be integer")
	}
	if burnAmountInt <= 0 {
		return shim.Error("burn amount must be positive")
	}

	erc20 := ERC20Metadata{}
	erc20Bytes, err := stub.GetState(tokenName)
	if err != nil {
		return shim.Error("failed to GetState, error: " + err.Error())
	}
	err = json.Unmarshal(erc20Bytes, &erc20)
	if err != nil {
		return shim.Error("failed to Unmarshal, error: " + err.Error())
	}

	resultTotalSupply := erc20.TotalSupply - uint64(*&burnAmountInt)
	if resultTotalSupply < 0 {
		return shim.Error("TotalSupply must be positive")
	}

	erc20 = ERC20Metadata{Name: erc20.Name, Symbol: erc20.Symbol, Owner: erc20.Owner, TotalSupply: resultTotalSupply}
	erc20Bytes, err = json.Marshal(erc20)
	if err != nil {
		return shim.Error("failed to Marshal erc20, error: " + err.Error())
	}

	err = stub.PutState(tokenName, erc20Bytes)
	if err != nil {
		return shim.Error("failed to PutState, error: " + err.Error())
	}

	recipientAmount, err := stub.GetState(recipientAddress)
	if err != nil {
		return shim.Error("failed to GetState, error: " + err.Error())
	}
	if recipientAmount == nil {
		recipientAmount = []byte("0")
	}
	recipientAmountInt, err := strconv.Atoi(string(recipientAmount))
	if err != nil {
		return shim.Error("caller amount must be integer")
	}

	resultBalance := recipientAmountInt - burnAmountInt
	if resultBalance < 0 {
		return shim.Error("Balance must be positive")
	}

	err = stub.PutState(recipientAddress, []byte(strconv.Itoa(resultBalance)))
	if err != nil {
		return shim.Error("failed to PutState of caller, error: " + err.Error())
	}

	return shim.Success([]byte("burn success"))
}

// High Throughput 가비지 데이터를 정리 해줌
// params - ownerAddress, targetAddress
func (cc *ERC20Chaincode) pruneFast(stub shim.ChaincodeStubInterface, args []string) sc.Response {

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
	updateResp := cc.update(stub, []string{targetAddress, strconv.FormatFloat(finalVal, 'f', -1, 64), "+"})
	if updateResp.Status == 200 {
		return shim.Success([]byte(fmt.Sprintf("Successfully pruned variable %s, final value is %f, %d rows pruned", args[0], finalVal, i)))
	}

	return shim.Error(fmt.Sprintf("Failed to prune variable: all rows deleted but could not update value to %f, variable no longer exists in ledger", finalVal))
}

//렛저 데이터 업데이트
func (cc *ERC20Chaincode) update(stub shim.ChaincodeStubInterface, args []string) sc.Response {

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
