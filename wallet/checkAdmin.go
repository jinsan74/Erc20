package wallet

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// 관리자 Address 저장 구조체
type AdminMetadata struct {
	Adminaddress string `json:"adminaddress"`
}

//Admin 지갑인지 확인
func CheckAdmin(stub shim.ChaincodeStubInterface, chkAddress string) bool {

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
