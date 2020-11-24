package approve

import (
	"encoding/json"
	"strconv"

	"github.com/jinsan74/Erc20/wallet"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Approval 이벤트 및 데이터 구조체
type Approval struct {
	Owner     string `json:"owner"`
	Spender   string `json:"spender"`
	Allowance int    `json:"allowance"`
}

// Ownere가 Spender 에서 Amount 만큼 토큰을 인출 할 권리 부여
// params - ownerAddress, spenderAddress, Amount
func Approve(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// Vaild Wallet을 호출하여 Parameter 추출 및 유효성 검사.
	params = wallet.CallVaildWallet(stub)
	if params == nil {
		return sc.Response{Status: 501, Message: "Vaild Wallet Error!", Payload: nil}
	}

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
