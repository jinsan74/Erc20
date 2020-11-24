package transferfrom

import (
	"strconv"

	"github.com/jinsan74/Erc20/wallet"

	"github.com/jinsan74/Erc20/erc20_function/invoke/approve"
	"github.com/jinsan74/Erc20/erc20_function/invoke/transfer"
	"github.com/jinsan74/Erc20/erc20_function/query/allowance"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// From 계좌에서 To 계좌로 Admount 토큰을 전송 , 단 Approve 함수를 통해 권한을 위임 받은 Spender만 할 수 있음
// parmas - spenderAddress(실행 ADDRESS), ownerAddress(원래토큰주인) , recipientAddress(받을사람), Amount
func TransferFrom(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// Vaild Wallet을 호출하여 Parameter 추출 및 유효성 검사.
	params = wallet.CallVaildWallet(stub)
	if params == nil {
		return sc.Response{Status: 501, Message: "Vaild Wallet Error!", Payload: nil}
	}

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

	allowanceResponse := allowance.Allowance(stub, []string{ownerAddress, spenderAddress})
	if allowanceResponse.GetStatus() >= 400 {
		return shim.Error("failed to get allowance, error: " + allowanceResponse.GetMessage())
	}

	allowanceInt, err := strconv.Atoi(string(allowanceResponse.GetPayload()))
	if err != nil {
		return shim.Error("allowance must be positive")
	}

	transferResponse := transfer.Transfer(stub, []string{ownerAddress, recipientAddress, transferAmount})
	if transferResponse.GetStatus() >= 400 {
		return shim.Error("failed to transfer, error: " + transferResponse.GetMessage())
	}

	approveAmountInt := allowanceInt - transferAmountInt
	approveAmount := strconv.Itoa(approveAmountInt)

	approveResponse := approve.Approve(stub, []string{ownerAddress, spenderAddress, approveAmount})
	if approveResponse.GetStatus() >= 400 {
		return shim.Error("failed to approve, error: " + approveResponse.GetMessage())
	}

	return shim.Success([]byte("transferFrom success"))
}
