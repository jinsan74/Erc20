package transferothertoken

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// 다른 Chain 코드의 토큰을 이동
// params - ownerAddress, chaincode name, toAddress, Amount
func TransferOtherToken(stub shim.ChaincodeStubInterface, params []string) sc.Response {

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
