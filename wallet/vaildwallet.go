/*
 * SejongTelecom
 */

package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/jinsan74/Erc20/model"
	"golang.org/x/crypto/ripemd160"
)

/*
 * 트랜잭션 체크 함수 : 지갑주소와 실제 트랜잭션 파라미터를 리턴한다.
 */
func vaildWallet(walletMeta WalletMeta) ([]string, error) {

	var err error
	//--필수 파라미터 체크------------
	if walletMeta.Publickey == "" || walletMeta.Sigmsg == "" || walletMeta.Txtime == "" {
		return nil, model.NewCustomError(model.MandatoryPrameterErrorType, "Wallet Parameter", "incorrect number of transaction parameter")
	}

	publicKeyStr := walletMeta.Publickey
	transData := walletMeta.Transdata
	sigData := walletMeta.Sigmsg
	nowTimeStamp := walletMeta.Nowtime
	txTime := walletMeta.Txtime
	transJdata := ""

	if len(walletMeta.Transjdata) > 0 {
		transJdata = walletMeta.Transjdata
	}

	//nowTimeStamp, _ := strconv.ParseInt(nowTime, 10, 64)
	txTimeStamp, _ := strconv.ParseInt(txTime, 10, 64)

	betweenSec := nowTimeStamp - txTimeStamp
	fmt.Println("BETWEEN SEC:", betweenSec)

	if betweenSec > 10 || betweenSec < -10 {
		//return nil, model.NewCustomError(model.TxTimeStampErrorType, "", "Must Be Tx Time inner +-10 Sec")
	}

	//--Public Key 생성----------
	publicKeySlice := strings.Split(publicKeyStr, ":")
	xHexStr := publicKeySlice[0]
	yHexStr := publicKeySlice[1]
	ePubKey := hexToPublicKey(xHexStr, yHexStr)

	//--ORG SIG MSG--------------------
	orgSigMsg := publicKeyStr + txTime

	//--ORG SIG MSG Hash 처리---------------------------
	orgSigMsgByte := []byte(orgSigMsg)
	orgSigMsgHash := sha256.Sum256(orgSigMsgByte)
	sigMsg := fmt.Sprintf("%x", orgSigMsgHash)
	//--SIGNATURE 비교-----------------
	sigHex, _ := hex.DecodeString(sigData)
	sigok := verifyMySig(ePubKey, sigMsg, sigHex)
	fmt.Println("SIG OK:", sigok)

	if !sigok {
		return nil, model.NewCustomError(model.SignatureErrorType, "", "Signature is Fail")
	}

	//--Double Hash 처리---------------------------
	publicKeyByte1 := []byte(publicKeyStr)
	publicKeyHash1 := sha256.Sum256(publicKeyByte1)

	publicKeyStr2 := fmt.Sprintf("%x", publicKeyHash1)
	publicKeyByte2 := []byte(publicKeyStr2)
	publicKeyHash2 := sha256.Sum256(publicKeyByte2)

	//--Public Key SAH256 처리----
	shaPubKey := fmt.Sprintf("%x", publicKeyHash2)

	//--RIPEMD160 로직 추가---
	h := ripemd160.New()
	h.Write([]byte(shaPubKey))
	shaPubKey = fmt.Sprintf("%x", h.Sum(nil))
	//----------------------

	//--Public Key SAH256 => BASE58Check 처리 ----
	shaPubkeyStr := new(big.Int)
	shaPubkeyStr.SetString(shaPubKey, 32)

	shaPubkeyDigit := fmt.Sprint(shaPubkeyStr)

	walletAddr, err := convertToBase58(shaPubkeyDigit, 10)
	if err != nil {
		return nil, model.NewCustomError(model.SignatureErrorType, "", "Signature is Fail")
	}

	fmt.Println("Wallet Address: ", walletAddr)

	retParams := make([]string, 1)

	if len(transJdata) > 0 {
		retParams = make([]string, 2)
		retParams[0] = walletAddr
		retParams[1] = transJdata
		return retParams, nil
	}

	if len(transData) > 0 {
		returnMsg := walletAddr
		returnMsg = walletAddr + "," + transData
		retParams = strings.Split(returnMsg, ",")
	} else {
		retParams[0] = walletAddr
	}

	return retParams, nil
}

func hexToPublicKey(xHex string, yHex string) *ecdsa.PublicKey {
	xBytes, _ := hex.DecodeString(xHex)
	x := new(big.Int)
	x.SetBytes(xBytes)

	yBytes, _ := hex.DecodeString(yHex)
	y := new(big.Int)
	y.SetBytes(yBytes)

	pub := new(ecdsa.PublicKey)
	pub.X = x
	pub.Y = y

	pub.Curve = elliptic.P256()

	return pub
}

const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var big0 = new(big.Int)
var big58 = big.NewInt(58)

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func convertToBase58(hash string, base int) (string, error) {
	var x, ok = new(big.Int).SetString(hash, base)
	if !ok {
		return "", fmt.Errorf("'%v' is not a valid integer in base '%d'", hash, base)
	}
	var sb strings.Builder
	var rem = new(big.Int)
	for x.Cmp(big0) == 1 {
		x.QuoRem(x, big58, rem)
		r := rem.Int64()
		sb.WriteByte(alphabet[r])
	}
	return reverse(sb.String()), nil
}

type ecdsaSignature struct {
	R, S *big.Int
}

func verifyMySig(pub *ecdsa.PublicKey, msg string, sig []byte) bool {
	digest := sha1.Sum([]byte(msg))

	var esig ecdsaSignature
	asn1.Unmarshal(sig, &esig)

	return ecdsa.Verify(pub, digest[:], esig.R, esig.S)
}
