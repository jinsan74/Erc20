package balanceof

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

var compositKeyIdx string = "balanceOf"

// 요청한 주소의 현재 토큰 수 조회
// params - ownerAddress, queryAddress
// Returns 토큰수
func BalanceOf(stub shim.ChaincodeStubInterface, params []string) sc.Response {

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
