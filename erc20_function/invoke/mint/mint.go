package mint

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"sejongtelecom.net/erc20/wallet"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// TOKEN 이름을 지정해 준다 - 토큰이름은 아무 의미가 없음
const tokenName = "TOKEN"

//토큰 정보 저장 구조체
type ERC20Metadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Owner       string `json:"owner"`
	TotalSupply uint64 `json:"totalSupply"`
}

// 관리자 Address 저장 구조체
type AdminMetadata struct {
	Adminaddress string `json:"adminaddress"`
}

// Owner의 토큰을 증가시키고 전체 발행량을 증가(토큰 추가 발행) - 관리자만 실행가능
// params - ownerAddress, Amount
func Mint(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// Vaild Wallet을 호출하여 Parameter 추출 및 유효성 검사.
	params = wallet.CallVaildWallet(stub)
	if params == nil {
		return sc.Response{Status: 501, Message: "Vaild Wallet Error!", Payload: nil}
	}

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
