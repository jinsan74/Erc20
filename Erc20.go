/*
 * SejongTelecom 코어기술개발팀
 * @author JinSan
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"sejongtelecom.net/erc20/erc20_function/invoke/approve"
	"sejongtelecom.net/erc20/erc20_function/invoke/burn"
	"sejongtelecom.net/erc20/erc20_function/invoke/decreaseallowance"
	"sejongtelecom.net/erc20/erc20_function/invoke/increaseallowance"
	"sejongtelecom.net/erc20/erc20_function/invoke/mint"
	"sejongtelecom.net/erc20/erc20_function/invoke/prunefast"
	"sejongtelecom.net/erc20/erc20_function/invoke/transfer"
	"sejongtelecom.net/erc20/erc20_function/invoke/transferfrom"
	"sejongtelecom.net/erc20/erc20_function/invoke/transferothertoken"

	"sejongtelecom.net/erc20/erc20_function/query/allowance"
	"sejongtelecom.net/erc20/erc20_function/query/approvallist"
	"sejongtelecom.net/erc20/erc20_function/query/balanceof"
	"sejongtelecom.net/erc20/erc20_function/query/rowcount"
	"sejongtelecom.net/erc20/erc20_function/query/totalsupply"

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

	switch fcn {
	case "totalSupply":
		return totalsupply.TotalSupply(stub, orgParam)
	case "balanceOf":
		return balanceof.BalanceOf(stub, orgParam)
	case "transfer":
		return transfer.Transfer(stub, orgParam)
	case "allowance":
		return allowance.Allowance(stub, orgParam)
	case "approve":
		return approve.Approve(stub, orgParam)
	case "approvalList":
		return approvallist.ApprovalList(stub, orgParam)
	case "transferFrom":
		return transferfrom.TransferFrom(stub, orgParam)
	case "transferOtherToken":
		return transferothertoken.TransferOtherToken(stub, orgParam)
	case "increaseAllowance":
		return increaseallowance.IncreaseAllowance(stub, orgParam)
	case "decreaseAllowance":
		return decreaseallowance.DecreaseAllowance(stub, orgParam)
	case "mint":
		return mint.Mint(stub, orgParam)
	case "burn":
		return burn.Burn(stub, orgParam)
	case "rowCount":
		return rowcount.RowCount(stub, orgParam)
	case "pruneFast":
		return prunefast.PruneFast(stub, orgParam)
	default:
		return sc.Response{Status: 404, Message: "404 Not Found", Payload: nil}
	}
}
