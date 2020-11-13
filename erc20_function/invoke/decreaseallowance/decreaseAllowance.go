package decreaseallowance

import (
	"strconv"

	"sejongtelecom.net/erc20/wallet"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"sejongtelecom.net/erc20/erc20_function/invoke/approve"
	"sejongtelecom.net/erc20/erc20_function/query/allowance"
)

// Allowance 값을 감소
// params - ownerAddress, spenderAddress, Amount
func DecreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// Vaild Wallet을 호출하여 Parameter 추출 및 유효성 검사.
	params = wallet.CallVaildWallet(stub)
	if params == nil {
		return sc.Response{Status: 501, Message: "Vaild Wallet Error!", Payload: nil}
	}

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

	allowanceResponse := allowance.Allowance(stub, []string{ownerAddress, spenderAddress})
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

	approveResponse := approve.Approve(stub, []string{ownerAddress, spenderAddress, resultAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to approve allowance, error: " + approveResponse.GetMessage())
	}

	return shim.Success([]byte("decreaseAllowance success"))
}
