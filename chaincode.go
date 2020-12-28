/*
 * SejongTelecom 코어기술개발팀
 * @author JinSan
 */

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"github.com/jinsan74/Erc20/wallet"
)

// Chaincode is the definition of the chaincode structure.
type Chaincode struct {
}

// Init is called when the chaincode is instantiated by the blockchain network.
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {

	return shim.Success(nil)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

	fcn, params := stub.GetFunctionAndParameters()

	switch fcn {
	case "walletTest":
		return cc.WalletTest(stub, params)
	default:
		return sc.Response{Status: 404, Message: "404 Not Found", Payload: nil}
	}
	//return shim.Success(nil)
}

// WalletTest is 지갑형 트랜잭션 테스트
// params -
func (cc *Chaincode) WalletTest(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	params := wallet.CallVaildWallet(stub)
	fmt.Println("PARAM LEN:", len(params))
	fmt.Println("WAddress:", params[0])

	return shim.Success(nil)
}
