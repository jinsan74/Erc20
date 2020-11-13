package allowance

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// owner 가 spender 에게 인출을 허락한 토큰의 개수는 몇개인지 조회
// params - ownerAddress, spenderAddress
// Returns 토큰 수
func Allowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

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
