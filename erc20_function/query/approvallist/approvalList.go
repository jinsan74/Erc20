package approvallist

import (
	"encoding/json"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Approval 이벤트 및 데이터 구조체
type Approval struct {
	Owner     string `json:"owner"`
	Spender   string `json:"spender"`
	Allowance int    `json:"allowance"`
}

// 권한을 위힘한 Approval 리스트 조회
// params - ownerAddress
// Returns Approval List
func ApprovalList(stub shim.ChaincodeStubInterface, params []string) sc.Response {

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
